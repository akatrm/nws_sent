package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"transform.com/m/core"
	"transform.com/m/solr"
)

type Manager struct {
	solr map[int]solr.Solr
	j    *solr.SolrJobManager
	q    *core.Queue
}

func GetManagerInstance(q *core.Queue, j *solr.SolrJobManager) *Manager {
	return &Manager{
		solr: make(map[int]solr.Solr),
		j:    j,
		q:    q,
	}
}

func (m Manager) StartHTTPServer() {
	http.HandleFunc("/register_solr", m.RegisterSolr)
	http.HandleFunc("/submit_job", m.SubmitJob)
	http.HandleFunc("/get_job_status", m.GetJobStatus)
	http.HandleFunc("/cancel_job", m.CancelJob)

	fmt.Println("Starting HTTP server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}

func (m Manager) RegisterSolr(w http.ResponseWriter, r *http.Request) {

	var response SolrRegisterResponse = SolrRegisterResponse{}

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	bodyLength := r.ContentLength
	var body []byte = make([]byte, bodyLength)

	total, err := r.Body.Read(body)

	if err != nil {
		errorMessage := GetError(err.Error())
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	if total != int(bodyLength) {
		errString := fmt.Sprintf("incomplete read: expected %d bytes, got %d bytes", bodyLength, total)
		errorMessage := GetError(errString)
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	var solrReq solr.Solr
	if err := json.Unmarshal(body, &solrReq); err != nil {
		errorMessage := GetError("Invalid JSON payload")
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	solrReq.ID = len(m.solr) + 1
	m.solr[solrReq.ID] = solrReq

	response.SolrID = solrReq.ID
	response.Message = "Solr configuration registered successfully"

	data, _ := json.Marshal(response)

	// Access your fields, do validation or process as needed
	fmt.Fprintf(w, "Received: %+v\n", solrReq)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)

}

func (m Manager) SubmitJob(w http.ResponseWriter, r *http.Request) {

	var response SolrJobRegisterResponse = SolrJobRegisterResponse{}

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	bodyLength := r.ContentLength
	var body []byte = make([]byte, bodyLength)

	total, err := r.Body.Read(body)

	if err != nil {
		errorMessage := GetError(err.Error())
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	if total != int(bodyLength) {
		errString := fmt.Sprintf("incomplete read: expected %d bytes, got %d bytes", bodyLength, total)
		errorMessage := GetError(errString)
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	var solrJobRequest SolrJobRegisterRequest
	if err := json.Unmarshal(body, &solrJobRequest); err != nil {
		errorMessage := GetError("Invalid JSON payload")
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	if _, exists := m.solr[solrJobRequest.SolrID]; exists != true {
		errorMessage := GetError("No such Solr configuraion present")
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	solrJobRequest.ID = m.j.GetNextId()

	var solrCore solr.SolrCore = solrJobRequest.SolrCore

	solrInstance := m.solr[solrCore.SolrID]

	jobInstance := solr.GetSolrJob(&solrCore, &solrInstance)

	m.j.RegisterJob(solrCore.ID, jobInstance)

	container := core.GetContainer(jobInstance)

	m.q.PushBack(container)

	// Access your fields, do validation or process as needed
	fmt.Fprintf(w, "Received: %+v\n", solrCore)

	response.JobID = solrCore.ID
	response.Message = "Solr job registered successfully"
	data, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)

}

func (m Manager) GetJobStatus(w http.ResponseWriter, r *http.Request) {
	var jobStatusRequest JobStatusRequest = JobStatusRequest{}
	var jobStatusResponse JobStatusResponse = JobStatusResponse{}

	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}

	bodyLength := r.ContentLength
	var body []byte = make([]byte, bodyLength)

	total, err := r.Body.Read(body)

	if err != nil {
		errorMessage := GetError(err.Error())
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	if total != int(bodyLength) {
		errString := fmt.Sprintf("incomplete read: expected %d bytes, got %d bytes", bodyLength, total)
		errorMessage := GetError(errString)
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &jobStatusRequest); err != nil {
		errorMessage := GetError("Invalid JSON payload")
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	jobInstance := m.j.GetJob(jobStatusRequest.JobID)

	if jobInstance == nil {
		errorMessage := GetError("No such job present")
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	jobStatusResponse.JobID = jobStatusRequest.JobID
	jobStatusResponse.JobStatus = jobInstance.GetStatus().GetStatus()

	data, _ := json.Marshal(jobStatusResponse)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (m Manager) CancelJob(w http.ResponseWriter, r *http.Request) {
}

func GetError(message string) string {
	errorResponse := HTTPError{Message: message}
	data, _ := json.Marshal(errorResponse)
	return string(data)
}
