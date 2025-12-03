package solr

import "transform.com/m/core"

type SolrErrorFunction func(core.Training, error) error
type AnalyticsErrorFunction func(error) error

type Solr struct {
	ID   int    `json:"id,omitempty"`
	Host string `json:"host"`
	Port string `json:"port"`
}

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

type SolrJobManager map[int]*SolrJob

func (j SolrJobManager) GetJob(id int) *SolrJob {
	return j[id]
}

func (j SolrJobManager) RegisterJob(id int, job *SolrJob) {
	j[id] = job
}

func (j SolrJobManager) GetNextId() int {
	return len(j) + 1
}
