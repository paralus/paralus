package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	v6Client "github.com/elastic/go-elasticsearch"
)

type elasticSearchQuery struct {
	url          string
	indexPattern string
	logPrefix    string
	esClient     *v6Client.Client
}

type ElasticSearchQuery interface {
	Handle(bytes.Buffer) (map[string]interface{}, error)
}

func NewElasticSearchQuery(url string, indexPattern string, logPrefix string) (ElasticSearchQuery, error) {
	cfg := v6Client.Config{
		Addresses: []string{
			url,
		},
	}
	esClient, err := v6Client.NewClient(cfg)
	if err != nil {
		_log.Errorw("NewElasticSearchQuery: Not able to get New Elastic Search Client")
		return nil, err
	}
	// res, err := esClient.Info()
	// if err != nil {
	// 	_log.Errorw("NewElasticSearchQuery: Error in getting ES client Info")
	// 	return nil, err
	// }
	// _log.Infow(logPrefix+":Connected to elastic search ", "cluster", res, "index", indexPattern)
	esQuery := &elasticSearchQuery{
		url:          url,
		indexPattern: indexPattern,
		logPrefix:    logPrefix,
		esClient:     esClient,
	}
	return esQuery, nil
}

// Handle Fires the search query
func (q *elasticSearchQuery) Handle(msg bytes.Buffer) (map[string]interface{}, error) {
	_log.Debugw("Searching elastic search: ", "index", q.indexPattern, "url", q.url, "q", q)
	res, err := q.esClient.Search(
		q.esClient.Search.WithContext(context.Background()),
		q.esClient.Search.WithIndex(q.indexPattern),
		q.esClient.Search.WithBody(&msg),
		q.esClient.Search.WithTrackTotalHits(true),
	)

	if err != nil {
		_log.Errorw("Error getting response:", "err", err)
		return nil, err
	}

	defer res.Body.Close()

	if res.IsError() {
		_log.Warnw(q.logPrefix+" Error in search request ", "request", msg.String(), "response", res.String())
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			_log.Errorw("Error parsing the response body", "err", err, "request", msg.String(), "response", res.String())
			return nil, err
		} else {
			if e["error"].(map[string]interface{})["type"] == "index_not_found_exception" {
				_log.Warnw(q.logPrefix+"Skipping this query as its a new setup", "request", msg.String(), "response", res.String())
				return nil, nil
			} else {
				_log.Errorw("Error received from ES", "request", msg.String(), "response", res.String())
				return nil, errors.New(res.Status() + " received from ES")
			}
		}
	}
	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		_log.Errorw("Error parsing the response body", "err", err)
		return nil, err
	}
	return r, nil
}
