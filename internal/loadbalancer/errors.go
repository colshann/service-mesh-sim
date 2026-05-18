package loadbalancer

import (
	"errors"
)

var (
	ErrMissingServiceName   = errors.New("missing service name in path")
	ErrNoInstancesAvailable = errors.New("no healthy instances available for the requested service")
	ErrRegistryError        = errors.New("error communicating with registry")
)
