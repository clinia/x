package elasticx

import (
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

const IndexNotFoundException = "index_not_found_exception"

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

	if eserror.Status == 409 {
		return true
	}

	if eserror.Status == 400 && eserror.ErrorCause.Type == "resource_already_exists_exception" {
		return true
	}

	return false
}

func isElasticNotFoundError(err error) bool {
	eserror, ok := isElasticError(err)
	if !ok {
		return false
	}

	return eserror.Status == 404
}
