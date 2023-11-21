package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type server struct {
	domain  string
	apiRoot string
	srv     *http.Server
	router  *http.ServeMux
	logger  *log.Logger
	auth    *authApi
	gateway *gateway
}

func main() {
	s := server{
		domain:  "my.example.com",
		apiRoot: "/api/therm/",
	}

	// Logger
	s.logger = log.New(os.Stdout, "therm: ", log.LstdFlags)

	// Config
	srvListen, ok := os.LookupEnv("THERMADDR")
	if !ok {
        s.logger.Fatal("Error: missing THERMADDR env config eg. 127.0.xxx.xxx:yyy")
	}
	authApiUrl, ok := os.LookupEnv("THERMAUTHAPIURL")
	if !ok {
		s.logger.Fatal("Error: missing THERMAUTHAPIURL env config")
	}
	authApiKey, ok := os.LookupEnv("THERMAUTHAPIKEY")
	if !ok {
		s.logger.Fatal("Error: missing THERMAUTHAPIKEY env config (base64-encoded)")
	}
	authCookieName, ok := os.LookupEnv("THERMAUTHCOOKIE")
	if !ok {
		s.logger.Fatal("Error: missing THERMAUTHCOOKIE env config")
	}
	thermApiUrl, ok := os.LookupEnv("THERMTHERMAPIURL")
	if !ok {
		s.logger.Fatal("Error: missing THERMTHERMAPIURL env config")
	}
	thermApiKey, ok := os.LookupEnv("THERMTHERMAPIKEY")
	if !ok {
		s.logger.Fatal("Error: missing THERMTHERMAPIKEY env config")
	}
	thermApiCert, ok := os.LookupEnv("THERMTHERMAPICERT")
	if !ok {
		s.logger.Fatal("Error: missing THERMTHERMAPICERT env config")
	}

	// Auth API
	s.auth = &authApi{
		app:        "therm",
		url:        authApiUrl,
		key:        authApiKey,
		CookieName: authCookieName,
	}

	// Gateway
	apiCertPem, err := ioutil.ReadFile(thermApiCert)
	if err != nil {
		s.logger.Fatal("Error reading Therm API Certificate file:", err.Error())
	}
	s.gateway, err = newGetway(thermApiUrl, thermApiKey, apiCertPem)
	if err != nil {
		s.logger.Fatal("Error setting-up the gateway:", err.Error())
	}
	defer s.gateway.Close()

	// Router
	s.router = http.NewServeMux()
	s.routes()

	// Server
	s.srv = &http.Server{
		Addr:         srvListen,
		Handler:      s.logMid(s.router),
		ErrorLog:     s.logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	// Handling graceful shutdown & starting server
	srvClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		<-sigint
		s.logger.Println("Received interrupt, shutting down server...")
		if err := s.srv.Shutdown(context.Background()); err != nil {
			s.logger.Fatal("Error shutting down server:", err)
		}
		close(srvClosed)
	}()
	s.logger.Println("Starting server:", s.srv)
	if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
		s.logger.Fatal("Server error:", err)
	}

	<-srvClosed
	s.logger.Println("Server closed. Bye!")
}
