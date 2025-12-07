// Package core holds small shared types and helpers used across the
// middleware components. Keep these types minimal and JSON-friendly
// so they can be used when communicating with the analytics engine.

package core

// Application contains basic runtime configuration for the middleware
// process. Fields are unexported here because configuration is
// managed internally; update to exported fields when external
// marshaling is required.
type Application struct {
	host string
	port string
}
