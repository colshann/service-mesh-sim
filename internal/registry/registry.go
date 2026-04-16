package registry

import (
	"sync"
	"time"
)

type Registry struct {
	services map[string]map[string]*Instance
	mu       sync.RWMutex // Read-Write mutex due to high read-to-write ratio
}

func NewRegistry() *Registry {
	return &Registry{
		services: make(map[string]map[string]*Instance),
	}
}

// Important:: This method uses UPSERT pattern. If registration is duplicated on the same instance ID, it will overwrite the existing entry, acting as an update.
func (r *Registry) Register(req RegisterRequest) error {
	// Validate input
	if req.ServiceName == "" {
		return ErrMissingServiceName
	}
	if req.InstanceID == "" {
		return ErrMissingInstanceID
	}
	if req.Address == "" {
		return ErrMissingAddress
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Initialize service map if it doesn't exist
	if r.services[req.ServiceName] == nil {
		r.services[req.ServiceName] = make(map[string]*Instance)
	}

	// Register the instance
	r.services[req.ServiceName][req.InstanceID] = &Instance{ID: req.InstanceID, Address: req.Address, Status: StatusHealthy, LastSeen: time.Now()}

	return nil
}

// Important:: This method is idempotent. If the instance is not found, it will simply return nil without error.
// Potential Breach: Address is not required for deregistration, which could lead to accidental deregistration if instance ID is reused across different services. Consider adding a check to ensure instance ID is unique across all services or require address for deregistration as well.
func (r *Registry) Deregister(req DeregisterRequest) error {
	// Validate input
	if req.ServiceName == "" {
		return ErrMissingServiceName
	}
	if req.InstanceID == "" {
		return ErrMissingInstanceID
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Idempotency check
	instances, ok := r.services[req.ServiceName]
	if !ok {
		return nil
	}

	if _, ok := instances[req.InstanceID]; !ok {
		return nil
	}

	// Deregister the instance
	delete(r.services[req.ServiceName], req.InstanceID)

	// Clean up service map if empty
	if len(r.services[req.ServiceName]) == 0 {
		delete(r.services, req.ServiceName)
	}

	return nil
}
