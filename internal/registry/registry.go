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
// This implementation assumes the given InstanceID's are unique per instance, otherwise ID conflicts will cause the newest entry to overwrite the existing one.
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
	instance := &Instance{ID: req.InstanceID, Address: req.Address, Status: StatusHealthy}
	instance.LastSeen.Store(time.Now().Unix())
	r.services[req.ServiceName][req.InstanceID] = instance

	return nil
}

// Important:: This method is idempotent. If the instance is not found, it will simply return nil without error.
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

func (r *Registry) ReceiveHeartbeat(req HeartbeatRequest) error {
	// Validate input
	if req.ServiceName == "" {
		return ErrMissingServiceName
	}
	if req.InstanceID == "" {
		return ErrMissingInstanceID
	}

	// Only read lock required since Instance.LastSeen is atomic
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check instanceID exists
	instances, ok := r.services[req.ServiceName]
	if !ok {
		return nil
	}
	instance, exists := instances[req.InstanceID]
	if !exists {
		return nil
	}

	// Update LastSeen timestamp atomically
	instance.LastSeen.Store(time.Now().Unix())
	return nil
}

func (r *Registry) GetInstances(req GetInstancesRequest) ([]InstanceSnapshot, error) {

	// Validate input
	if req.ServiceName == "" {
		return nil, ErrMissingServiceName
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check service exists
	instances, ok := r.services[req.ServiceName]
	if !ok {
		return nil, ErrServiceNotFound
	}

	// Copy instance data to snapshot list to avoid exposing internal state and registry-specific data
	snapshots := make([]InstanceSnapshot, 0, len(instances))
	for _, instance := range instances {
		snapshots = append(snapshots, InstanceSnapshot{
			ID:      instance.ID,
			Address: instance.Address,
			Status:  instance.Status, // Included for potential future use, not currently used in registry or loadbalancer logic
		})
	}
	return snapshots, nil

}

// Instance cleanup logic that scans for instances that haven't sent a heartbeat within the expected interval (ttl) and removes them from the registry.
// Uses two-pass approach to reduce time spent holding a write lock: eliminates write lock entirely if no instances are stale and only iterates through marked instances for cleanup rather than entire registry.
func (r *Registry) CleanupStaleInstances(ttl time.Duration) {

	markedInstances := make([]Pair[string, *Instance], 0) // List of instances to be removed after recheck
	now := time.Now().Unix()

	// Search through instance list for instances that haven't received a heartbeat/reregister in >ttl and add to markedInstances
	r.mu.RLock()
	for serviceName, instances := range r.services {
		for _, instance := range instances {
			if now-instance.LastSeen.Load() > int64(ttl.Seconds()) {
				markedInstances = append(markedInstances, Pair[string, *Instance]{First: serviceName, Second: instance}) // Uses Pair struct to store serviceName and instance reference for later cleanup
			}
		}
	}
	r.mu.RUnlock()

	// If no instances are marked for cleanup, return early to avoid unnecessary write lock
	if len(markedInstances) == 0 {
		return
	}

	// Remove marked instances, cleanup if needed
	recheckNow := time.Now().Unix()
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, markedInstance := range markedInstances {
		// Recheck for updated heartbeats between initial check and newly acquired write lock
		if recheckNow-markedInstance.Second.LastSeen.Load() > int64(ttl.Seconds()) {
			delete(r.services[markedInstance.First], markedInstance.Second.ID)
			// Clean up service map if empty
			if len(r.services[markedInstance.First]) == 0 {
				delete(r.services, markedInstance.First)
			}
		}
	}

}
