package solr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"

	"golang.org/x/sync/errgroup"
	"transform.com/m/core"
)

type SolrJobStatus int

const (
	INIT SolrJobStatus = iota
	RUNNING
	COMPLETED
	FAILED
	PENDING
)

func (s SolrJobStatus) String() string {
	return [...]string{"INIT", "RUNNING", "COMPLETED", "FAILED", "PENDING"}[s]
}

func (s SolrJobStatus) GetStatus() string {
	return s.String()
}

type SolrJob struct {
	s              *Solr
	j              *SolrCore
	Status         SolrJobStatus
	prvCursorMark  string
	currentSolrURL string
	ErrorMessage   string
	wg             *sync.WaitGroup
}

func GetSolrJob(j *SolrCore, s *Solr) *SolrJob {

	si := &SolrJob{
		s:              s,
		j:              j,
		Status:         PENDING,
		prvCursorMark:  "",
		currentSolrURL: "",
		wg:             &sync.WaitGroup{},
	}
	return si
}

func (s SolrJob) GetStatus() SolrJobStatus {
	return s.Status
}

func (s *SolrJob) SetStatus(newStatus SolrJobStatus) {
	s.Status = newStatus
}

func (s SolrJob) getSolrUrl(cursorMark string) string {
	solrBaseURL := fmt.Sprintf("http://%s:%s", s.s.Host, s.s.Port)
	collection := s.j.Collection
	query := "*:*"
	rows := s.j.Rows
	sort := s.j.Sort
	s.prvCursorMark = cursorMark

	if cursorMark == "*" {
		solrURL := fmt.Sprintf("%s/solr/%s/select?q=%s&wt=json&rows=%d&sort=%s&cursorMark=%s",
			solrBaseURL, collection, url.QueryEscape(query), rows, url.QueryEscape(sort), url.QueryEscape(cursorMark))
		return solrURL
	}

	solrURL := fmt.Sprintf("%s/solr/%s/select?q=%s&wt=json&rows=%d&sort=%s&nextCursorMark=%s",
		solrBaseURL, collection, url.QueryEscape(query), rows, url.QueryEscape(sort), url.QueryEscape(cursorMark))

	return solrURL
}

func (s SolrJob) getTargetURL(endpoint string) string {
	targetURL := fmt.Sprintf("http://%s:%s", s.j.Host, s.j.Port, endpoint)
	return targetURL
}

func (s *SolrJob) Init() error {

	// Initialize connection to analytics engine if needed

	s.SetStatus(INIT)

	target := s.getTargetURL("/stream/status")
	isRunning := core.TrainingStatus{}

	data, err := core.GetRequest(target)

	if err != nil {
		return err
	}

	err = json.Unmarshal(data, isRunning)
	if err != nil {
		return err
	}

	if isRunning.Running == false {
		target = s.getTargetURL("/stream/start")
		data, err := core.PostRequest(target, nil)
		if err != nil {
			return err
		}
		fmt.Println("Analytics engine started:", data)
	}
	s.Status = PENDING
	return nil
}

func (s SolrJob) Operation() error {

	var trainData chan core.Training = make(chan core.Training)
	s.Status = RUNNING
	ctx := context.Background()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		defer close(trainData)
		return s.PullSolrData("*", func(data core.Training, err error) error {
			if err != nil {
				return err
			}
			select {
			case trainData <- data:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	})

	c := make(chan *core.Result)
	for i := 0; i < 10; i++ {
		g.Go(func() error {
			for data := range trainData {
				res, err := s.trainEngineOperation(data)
				if err != nil {
					return err
				}

				select {
				case c <- res:
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
			}
			return nil
		})
	}

	go func() {
		g.Wait()
	}()

	for r := range c {
		fmt.Println(r)
	}

	if err := g.Wait(); err != nil {
		s.Status = FAILED
		return err
	}

	s.Status = COMPLETED
	return nil
}

func (s SolrJob) PullSolrData(cursor string, fn SolrErrorFunction) error {
	mark, traindata, err := s.solrPuller(cursor)

	if err != nil {
		err = fn(traindata, err)
	} else {
		err = s.PullSolrData(mark, fn)
	}
	return err
}

func (s SolrJob) solrPuller(cursorMark string) (string, core.Training, error) {

	var mark string
	var sampleData core.Training

	solrResponse := SolrResponse{}

	data, err := core.GetRequest(s.getSolrUrl(cursorMark))

	if err != nil {
		return mark, sampleData, err
	}

	err = json.Unmarshal(data, solrResponse)

	if err != nil {
		return mark, sampleData, err
	}

	if len(solrResponse.Response.Docs) == 0 {
		return mark, sampleData, fmt.Errorf("No more documents to fetch")
	}

	var docs []core.TrainExamples = make([]core.TrainExamples, len(solrResponse.Response.Docs))

	// Process solrResponse as needed
	for i := 0; i < len(solrResponse.Response.Docs); i++ {
		doc := solrResponse.Response.Docs[i]
		docs[i].Text = doc.Content
		docs[i].Label = s.j.Label
	}

	sampleData.Examples = docs

	newCursorMark := solrResponse.ResponseHeader.Params["nextCursorMark"]

	return newCursorMark, sampleData, nil
}

func (s SolrJob) trainEngineOperation(data core.Training) (*core.Result, error) {

	res := &core.Result{}

	target := s.getTargetURL("/stream/train")

	payload, err := json.Marshal(data)

	if err != nil {
		return res, err
	}

	res, err = core.PostRequest(target, payload)

	if err != nil {
		return res, err
	}

	return res, err
}
