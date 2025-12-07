// Package http defines request and response payloads used by the
// middleware's HTTP server. These structs map to JSON bodies sent by
// clients when interacting with the middleware API.

package http

import "transform.com/m/solr"

// SolrRegister wraps a Solr configuration used when registering a
// new Solr instance with the middleware.
type SolrRegister struct {
	solr.Solr
}

// SolrRegisterResponse is returned after a successful Solr
// registration.
type SolrRegisterResponse struct {
	SolrID  int    `json:"solr_id"`
	Message string `json:"message"`
}

// SolrJobRegisterRequest represents the JSON payload required to
// register a new Solr job with the middleware queue.
type SolrJobRegisterRequest struct {
	solr.SolrCore
}

// SolrJobRegisterResponse acknowledges that a Solr job was accepted
// and queued.
type SolrJobRegisterResponse struct {
	JobID   int    `json:"job_id"`
	Message string `json:"message"`
}

// JobStatusRequest requests the status of a previously submitted job.
type JobStatusRequest struct {
	JobID int `json:"job_id"`
}

// JobStatusResponse contains the current status of a job and an
// optional error message when a job has failed.
type JobStatusResponse struct {
	JobID        int    `json:"job_id"`
	JobStatus    string `json:"job_status"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// cancelJobRequest is an internal representation of a cancel request.
type cancelJobRequest struct {
	JobID int `json:"job_id"`
}

// CancelJobResponse acknowledges a cancellation request.
type CancelJobResponse struct {
	JobID   int    `json:"job_id"`
	Message string `json:"message"`
}

// HTTPError is used to serialize error messages to clients.
type HTTPError struct {
	Message string `json:"message"`
}
