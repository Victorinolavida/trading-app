// Package module defines the contract that every feature module of the
// gateway implements.
//
// A module is a self contained group of routes that is mounted by the http
// server under BasePath. Features depend on this small interface instead of
// the concrete http server, so the transport stays an implementation detail.
package module

import "net/http"

// Module is a feature module mounted under a base path.
type Module interface {
	// BasePath is the path where the module is mounted, e.g. "/auth".
	BasePath() string
	// Routes is the handler serving the routes of the module.
	Routes() http.Handler
}
