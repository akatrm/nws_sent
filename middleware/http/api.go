package http

import "transform.com/m/solr"

// Register new solr server
type SolrRegister struct {
	solr.Solr
}

// New solr server registration response
type SolrRegisterResponse struct {
	SolrID  int    `json:"solr_id"`
	Message string `json:"message"`
}

// Register new solr job
type SolrJobRegisterRequest struct {
	solr.SolrCore
}

// New solr job registration response
type SolrJobRegisterResponse struct {
	JobID   int    `json:"job_id"`
	Message string `json:"message"`
}

// Get job status request/response
type JobStatusRequest struct {
	JobID int `json:"job_id"`
}

// Get job status response
type JobStatusResponse struct {
	JobID        int    `json:"job_id"`
	JobStatus    string `json:"job_status"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// Cancel job request
type cancelJobRequest struct {
	JobID int `json:"job_id"`
}

// cancel job response
type CancelJobResponse struct {
	JobID   int    `json:"job_id"`
	Message string `json:"message"`
}

type HTTPError struct {
	Message string `json:"message"`
}
