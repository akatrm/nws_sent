// Package solr contains types used to decode Solr HTTP responses and
// helper payloads exchanged between the middleware and the analytics
// engine. These structs mirror the JSON structures returned by Solr's
// JSON API and the small training payloads sent to the analytics
// service.

package solr

// ResponseHeader represents the `responseHeader` block returned by
// Solr. It contains metadata about the query (status, QTime and the
// params used to produce the response).
type ResponseHeader struct {
	Status int               `json:"status"`
	QTime  int               `json:"QTime"`
	Params map[string]string `json:"params"`
}

// SolrDocument models a single Solr document in the response `docs`
// array. Fields should be adjusted to match the schema used by your
// Solr collection; this is a compact example including the text
// content and a few metadata fields.
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

// SolrResponse is a lightweight representation of the JSON object
// returned by the `/select` endpoint when `wt=json` is requested.
type SolrResponse struct {
	ResponseHeader ResponseHeader `json:"responseHeader"`
	Response       Response       `json:"response"`
}

// Response mirrors the `response` block inside a Solr JSON reply and
// contains the list of matching documents and paging information.
type Response struct {
	NumFound      int            `json:"numFound"`
	Start         int            `json:"start"`
	NumFoundExact bool           `json:"numFoundExact"`
	Docs          []SolrDocument `json:"docs"`
}

// TrainerData represents a single training example sent to the
// analytics engine: a short piece of text and an integer label.
type TrainerData struct {
	Text  string `json:"text"`
	Label int    `json:"label"`
}

// TrainerInput is the top-level payload structure posted to the
// analytics engine's `/stream/train` endpoint. It contains the
// `examples` array.
type TrainerInput struct {
	Examples []TrainerData `json:"examples"`
}
