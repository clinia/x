package persistencex

type ListRequestBuilder struct {
	page    *int
	perPage *int
}

func NewListRequestBuilder() *ListRequestBuilder {
	return &ListRequestBuilder{}
}

func (b *ListRequestBuilder) WithPage(page *int) *ListRequestBuilder {
	b.page = page
	return b
}

func (b *ListRequestBuilder) WithPerPage(perPage *int) *ListRequestBuilder {
	b.perPage = perPage
	return b
}

func (b *ListRequestBuilder) Build() ListRequest {
	rq := ListRequest{
		Page:    0,
		PerPage: 20,
	}

	if b.page != nil {
		rq.Page = *b.page
	}

	// PerPage must be between 0 and 100
	// If it's not, we'll default to 20
	if b.perPage != nil && *b.perPage >= 0 && *b.perPage <= 100 {
		rq.PerPage = *b.perPage
	}

	return rq
}
