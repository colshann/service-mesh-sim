package registry

import (
	"sync/atomic"
)

// ----------- Utility structs -----------
type Pair[u, v any] struct {
	First  u
	Second v
}

// ----------- Instance structs for registry data -----------

type Instance struct {
	ID       string
	Address  string
	Status   Status
	LastSeen atomic.Int64 // Unix timestamp in seconds for LastSeen to ensure atomic updates
}

type InstanceSnapshot struct {
	ID       string
	Address  string
	Status   Status
	LastSeen int64 // Unix timestamp in seconds for snapshot
}

// ----------- Status structs -----------

type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
)

// ----------- Request structs for json unmarshalling -----------

type RegisterRequest struct {
	ServiceName string `json:"service_name"`
	InstanceID  string `json:"instance_id"`
	Address     string `json:"address"`
}

type DeregisterRequest struct {
	ServiceName string `json:"service_name"`
	InstanceID  string `json:"instance_id"`
}

type HeartbeatRequest struct {
	ServiceName string `json:"service_name"`
	InstanceID  string `json:"instance_id"`
}

type GetInstanceRequest struct {
	ServiceName string `json:"service_name"`
}
