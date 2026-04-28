package registry

import (
	"time"
)

type Instance struct {
	ID       string
	Address  string
	Status   Status
	LastSeen time.Time // Future Optimization: change to atomic.Int64
}

type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
)

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
