package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

//
// get status (GET)
func (s *server) getStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json; charset=UTF-8")
		err := json.NewEncoder(w).Encode(s.gateway.GetState())
		if err != nil {
			s.logger.Println("Error marshalling response status to JSON:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

//
// set status (POST)
func (s *server) postSet() http.HandlerFunc {
	type request struct {
		State thermState `json:"state"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		// Parsing and checking JSON request
		var req request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			s.logger.Println("Error parsing input: ", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// Setting relay state
		if err := s.gateway.SetState(req.State); err != nil {
			s.logger.Println("Error setting state to", req.State, ":", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Writing response
		s.logger.Println("State set to", req.State)
		fmt.Fprint(w, "OK")
	}
}
