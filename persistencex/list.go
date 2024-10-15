package persistencex

import (
	"math"
)

type ListRequest struct {
	// The page to retrieve.
	// Default: 0
	Page int `json:"page,omitempty"`

	// The number of items per page.
	// Default: 20
	PerPage int `json:"perPage,omitempty"`
}

type ListResponse[T any] struct {
	Data []T              `json:"data"`
	Meta ListResponseMeta `json:"meta"`
}

type ListResponseMeta struct {
	Page     int   `json:"page"`
	PerPage  int   `json:"perPage"`
	Total    int64 `json:"total"`
	NumPages int   `json:"numPages"`
}

func NewListResponseMeta(page, perPage int, total int64) ListResponseMeta {
	numPages := 0
	if perPage != 0 {
		numPages = int(math.Ceil(float64(total) / float64(perPage)))
	}

	return ListResponseMeta{
		Page:     page,
		PerPage:  perPage,
		Total:    total,
		NumPages: numPages,
	}
}
