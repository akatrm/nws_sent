// Package core contains lightweight types and helpers used by the
// middleware for coordinating analytics work (training payloads,
// status objects, etc.). These types are intentionally small and
// expressed as plain structs so they can be JSON (un)marshalled when
// communicating with external services.

package core

// TrainingStatus represents the current training state of the
// analytics engine. It is serialized as JSON when querying
// the engine's `/stream/status` endpoint.
type TrainingStatus struct {
	Running bool `json:"running"`
}

// TrainExamples is a single example used by streaming training. The
// field names are JSON-tagged to match the expected payload format
// sent from the middleware to the analytics engine.
type TrainExamples struct {
	Label string `json:"label"`
	Text  string `json:"text"`
}

// Training is the top-level wrapper for a batch of training examples
// that will be sent to the analytics engine for incremental update.
type Training struct {
	Examples []TrainExamples `json:"examples"`
}
