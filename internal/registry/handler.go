package registry

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// Possible future refactoring for registry mocking and testing
// -------------------------------------------------------------
// type RegistryService interface {
// 	Register(RegisterRequest) error
// 	Deregister(DeregisterRequest) error
// 	// GetInstances(serviceName string) ([]registry.Instance, error) // Optional: For future extension to support service discovery
// }

// type Handler struct {
// 	registry RegistryService
// }
// -------------------------------------------------------------

func decodeJSON[T any](req *http.Request, dst *T) error {
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields() // Strict decoding to prevent unexpected fields
	return dec.Decode(dst)
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	// Guard against empty data
	if data == nil {
		w.WriteHeader(statusCode)
		return
	}

	// Encode the response to a buffer first to handle encoding errors gracefully
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	buf.WriteTo(w)
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

	// Response
	writeJSON(w, http.StatusOK, nil) // Expand later to return a message or status
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

	// Response
	writeJSON(w, http.StatusOK, nil) // Expand later to return a message or status
}

func (r *Registry) HandleHeartbeat(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	// Only accepts POST /heartbeat
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode request body into HeartbeatRequest struct
	var heartbeatReq HeartbeatRequest
	if err := decodeJSON(req, &heartbeatReq); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Process heartbeat
	if err := r.ReceiveHeartbeat(heartbeatReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Response
	writeJSON(w, http.StatusOK, nil) // Expand later to return updated instance status, and potentially a time till next expected heartbeat
}

func (r *Registry) HandleGetInstances(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	// Only accepts POST /get-instances
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode request body into GetInstanceRequest struct
	var getInstanceReq GetInstanceRequest
	if err := decodeJSON(req, &getInstanceReq); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Process get instances
	instances, err := r.GetInstances(getInstanceReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Response
	writeJSON(w, http.StatusOK, instances)
}
