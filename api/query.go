package api

import (
	"net/url"

	query "github.com/google/go-querystring/query"
)

// AddQueryOptions encodes struct fields tagged with `url:"name,omitempty"`
// as query parameters and appends them to the given path.
func AddQueryOptions(path string, opts any) (string, error) {
	if opts == nil {
		return path, nil
	}

	params, err := query.Values(opts)
	if err != nil {
		return path, err
	}

	if len(params) == 0 {
		return path, nil
	}

	u, err := url.Parse(path)
	if err != nil {
		return path, err
	}
	q := u.Query()
	for k, vs := range params {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}
