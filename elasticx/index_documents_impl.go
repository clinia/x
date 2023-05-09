package elasticx

import (
	"context"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/segmentio/ksuid"
)

func (i *index) DocumentExists(ctx context.Context, id string) (bool, error) {
	// TODO
	return false, nil
}

func (i *index) ReadDocument(ctx context.Context, id string, result interface{}) (*DocumentMeta, error) {
	// TODO
	return &DocumentMeta{}, nil
}

func (i *index) CreateDocument(ctx context.Context, document interface{}, options *DocumentOptions) (*DocumentMeta, error) {
	res, err := esapi.IndexRequest{
		Index:      i.relPath(),
		DocumentID: ksuid.New().String(),
		Refresh:    getRefresh(options),
		Body:       esutil.NewJSONReader(document),
	}.Do(ctx, i.es)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.IsError() {
		return nil, withElasticError(res)
	}

	var meta DocumentMeta
	if err := json.NewDecoder(res.Body).Decode(&meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

func (i *index) ReplaceDocument(ctx context.Context, key string, document interface{}, options *DocumentOptions) (*DocumentMeta, error) {

	res, err := esapi.IndexRequest{
		Index:      i.relPath(),
		DocumentID: key,
		Refresh:    getRefresh(options),
		Body:       esutil.NewJSONReader(document),
	}.Do(ctx, i.es)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.IsError() {
		return nil, withElasticError(res)
	}

	var meta DocumentMeta
	if err := json.NewDecoder(res.Body).Decode(&meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

func (i *index) DeleteDocument(ctx context.Context, key string, options *DocumentOptions) error {
	res, err := esapi.DeleteRequest{
		Index:      i.relPath(),
		DocumentID: key,
		Refresh:    getRefresh(options),
	}.Do(ctx, i.es)
	
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		return withElasticError(res)
	}

	return nil
}

func getRefresh(options *DocumentOptions) string {
	refresh := ""
	if options != nil {
		refresh = options.Refresh
	}
	return refresh
}
