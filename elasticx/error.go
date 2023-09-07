package elasticx

import (
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

func isElasticError(err error) (*types.ElasticsearchError, bool) {
	eserror, ok := err.(*types.ElasticsearchError)
	if !ok {
		return nil, false
	}

	return eserror, true
}

func isElasticAlreadyExistsError(err error) bool {
	eserror, ok := isElasticError(err)
	if !ok {
		return false
	}

	return eserror.Status == 409
}
