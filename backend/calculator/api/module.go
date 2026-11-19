// Package api is the public boundary of the calculator bounded context.
package api

import (
	"context"

	"github.com/victorinolavida/calculator/internal/application/ports"
	"github.com/victorinolavida/pkg/lifecycle"
)

// Module is the in-process calculator module.
type Module struct {
	security ports.SecurityProvider
}

var _ lifecycle.Service = (*Module)(nil)

// New wires the calculator module on top of the securities provider.
func New(security ports.SecurityProvider) *Module {
	return &Module{security: security}
}

// Start runs the calculator module until the context is cancelled.
func (m *Module) Start(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Close releases the calculator module.
func (m *Module) Close(context.Context) error {
	return nil
}
