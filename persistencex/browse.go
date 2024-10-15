package persistencex

import "context"

// Browse is a helper function that will fetch all the data from a list request.
func Browse[T any](req func(context.Context, ListRequest) (*ListResponse[T], error), ctx context.Context, opts ListRequest) ([]T, error) {
	data := []T{}

	res, err := req(ctx, opts)
	if err != nil {
		return nil, err
	}

	data = append(data, res.Data...)

	// If we have all the data, return it
	if res.Meta.Total == int64(len(res.Data)) {
		return data, nil
	}

	// Otherwise, fetch the rest of the data
	for len(res.Data) == 0 || res.Meta.Total != int64(len(data)) {
		opts.Page++

		res, err = req(ctx, opts)
		if err != nil {
			return nil, err
		}

		data = append(data, res.Data...)
	}

	return data, nil
}
