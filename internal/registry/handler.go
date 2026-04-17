package registry

import (
	"encoding/json"
	"net/http"
)

func decodeJSON[T any](req *http.Request, dst *T) error {
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields() // Strict decoding to prevent unexpected fields
	return dec.Decode(dst)
}

func (r *Registry) HandleRegister(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	// Only accepts POST /register
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode request body into RegisterRequest struct
	var registerReq RegisterRequest
	if err := decodeJSON(req, &registerReq); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Process registration
	if err := r.Register(registerReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (r *Registry) HandleDeregister(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	// Only accepts POST /deregister
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode request body into DeregisterRequest struct
	var deregisterReq DeregisterRequest
	if err := decodeJSON(req, &deregisterReq); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Process deregistration
	if err := r.Deregister(deregisterReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
