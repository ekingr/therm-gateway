package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Periodic update of therm-ctrl state
const updatePeriod = 1 * time.Second

type thermState struct {
	Rel1   bool `json:"rel1"`
	Sens01 bool `json:"sens01"`
	Sens11 bool `json:"sens11"`
	Rel2   bool `json:"rel2"`
	Sens02 bool `json:"sens02"`
	Sens12 bool `json:"sens12"`
	Rel3   bool `json:"rel3"`
	Sens03 bool `json:"sens03"`
	Sens13 bool `json:"sens13"`
}

type gateway struct {
	state            thermState
	updateOk         bool
	updateStatus     string
	updateStatusCode int
	updateTime       int64
	running          bool
	mu               sync.RWMutex
	apiUrl           string
	apiKey           string
	httpClient       *http.Client
}

type ThermApiError struct {
	Status     string
	StatusCode int
}

func (e *ThermApiError) Error() string {
	return "Error from Therm API: " + e.Status
}

func newGetway(apiUrl, apiKey string, apiCert []byte) (g *gateway, err error) {
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(apiCert)
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{RootCAs: certPool}
	g = &gateway{
		httpClient:       &http.Client{Transport: transport},
		apiUrl:           apiUrl,
		apiKey:           apiKey,
		running:          true,
		state:            thermState{},
		updateStatusCode: 0,
		updateStatus:     "000: Not started yet",
	}

	// Periodic update of stored state
	// on top on interrupts, for good measure
	go func() {
		for g.running {
			g.updateState()
			time.Sleep(updatePeriod)
		}
	}()
	return
}

func (g *gateway) Close() {
	g.running = false
}

func (g *gateway) updateState() error {
	r, err := g.httpClient.Get(g.apiUrl + "/status.json?key=" + g.apiKey)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	g.mu.Lock()
	defer g.mu.Unlock()
	g.updateStatus = r.Status
	g.updateStatusCode = r.StatusCode
	if r.StatusCode < 200 || r.StatusCode > 299 {
		g.updateOk = false
		return &ThermApiError{r.Status, r.StatusCode}
	}
	g.updateOk = true
	g.updateTime = time.Now().Unix()
	return json.NewDecoder(r.Body).Decode(&g.state)
}

type apiState struct {
	Therm            thermState `json:"therm"`
	UpdateOk         bool       `json:"updateOk"`
	UpdateStatus     string     `json:"updateStatus"`
	UpdateStatusCode int        `json:"updateStatusCode"`
	UpdateTime       int64      `json:"updateTime"`
}

func (g *gateway) GetState() apiState {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return apiState{g.state, g.updateOk, g.updateStatus, g.updateStatusCode, g.updateTime}
}

type postReq struct {
	State thermState `json:"state"`
	Key   string     `json:"key"`
}

func (g *gateway) SetState(s thermState) error {
	body := postReq{s, g.apiKey}
	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(body)
	if err != nil {
		return &ThermApiError{"999: Unable to encode thermState to JSON", 999}
	}
	r, err := g.httpClient.Post(g.apiUrl+"/set", "application/json", &b)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode < 200 || r.StatusCode > 299 {
		return &ThermApiError{r.Status, r.StatusCode}
	}
	return g.updateState()
}
