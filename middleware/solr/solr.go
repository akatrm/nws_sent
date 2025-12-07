// Package solr contains higher-level abstractions for interacting with
// Solr collections and converting Solr documents into training payloads
// consumed by the analytics engine.
package solr

import "transform.com/m/core"

// SolrErrorFunction is a callback used by Solr iterators to report
// training batches and optionally an error. The function should return
// a non-nil error to stop the iteration.
type SolrErrorFunction func(core.Training, error) error

// AnalyticsErrorFunction describes a callback that translates an
// analytics/processing level error into a higher-level value. It is
// provided for symmetry with the Solr error callback shape.
type AnalyticsErrorFunction func(error) error

// Solr represents a Solr instance configuration (host/port and id).
type Solr struct {
	ID   int    `json:"id,omitempty"`
	Host string `json:"host"`
	Port string `json:"port"`
}

// SolrCore binds a collection and query parameters to a logical job
// configuration. It contains pagination/sorting settings and the label
// that should be applied to documents when turned into training
// examples.
type SolrCore struct {
	ID         int    `json:"id,omitempty"`
	SolrID     int    `json:"solr_id"`
	Collection string `json:"collection"`
	Sort       string `json:"sort"`
	CursorMark string `json:"cursorMark"`
	Rows       int    `json:"rows"`
	Endpoint   string `json:"endpoint"`
	Label      string `json:"label"`
	Host       string `json:"host"`
	Port       string `json:"port"`
}

// SolrJobManager is a lightweight map-based registry for active jobs.
// It provides convenience helpers to register and retrieve jobs by id.
type SolrJobManager map[int]*SolrJob

// GetJob returns the Solr job registered with the given id or nil if
// none exists.
func (j SolrJobManager) GetJob(id int) *SolrJob {
	return j[id]
}

// RegisterJob stores a job instance under the provided id.
func (j SolrJobManager) RegisterJob(id int, job *SolrJob) {
	j[id] = job
}

// GetNextId computes a new id for a job based on the current map
// size. This is a convenience function for simple in-memory
// registries.
func (j SolrJobManager) GetNextId() int {
	return len(j) + 1
}
