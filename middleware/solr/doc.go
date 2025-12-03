package solr

type ResponseHeader struct {
	Status int               `json:"status"`
	QTime  int               `json:"QTime"`
	Params map[string]string `json:"params"`
}

type SolrResponse struct {
	ResponseHeader ResponseHeader `json:"responseHeader"`
	Response       Response       `json:"response"`
}

type SolrDocument struct {
	Tstamp  string   `json:"tstamp"`
	Anchor  []string `json:"anchor"`
	Digest  string   `json:"digest"`
	Boost   float64  `json:"boost"`
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	URL     string   `json:"url"`
	Content string   `json:"content"`
	Version int      `json:"_version_"`
}

// Replace 'SolrDocument' fields as needed for your schema
type Response struct {
	NumFound      int            `json:"numFound"`
	Start         int            `json:"start"`
	NumFoundExact bool           `json:"numFoundExact"`
	Docs          []SolrDocument `json:"docs"`
}

type TrainerData struct {
	Text  string `json:"text"`
	Label int    `json:"label"`
}

type TrainerInput struct {
	Examples []TrainerData `json:"examples"`
}
