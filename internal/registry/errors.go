package registry

import "errors"

var (
	ErrInvalidRequest     = errors.New("invalid request")
	ErrMissingServiceName = errors.New("missing service name")
	ErrMissingInstanceID  = errors.New("missing instance ID")
	ErrMissingURL         = errors.New("missing URL")
	ErrServiceNotFound    = errors.New("service not found")
)
