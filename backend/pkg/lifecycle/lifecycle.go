// Package lifecycle defines the small contract every long-running component
// exposes so the composition root can start and stop it uniformly.
package lifecycle

import "context"

// Service is a component that owns background work and/or resources.
//
// Start blocks until ctx is cancelled or the component fails. Close releases
// the resources owned by the component and is safe to call after Start has
// returned.
type Service interface {
	Start(ctx context.Context) error
	Close(ctx context.Context) error
}
