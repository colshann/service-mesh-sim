package registry

import (
	"time"
)

type Instance struct {
	ID       string
	Address  string
	Status   Status
	LastSeen time.Time
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
