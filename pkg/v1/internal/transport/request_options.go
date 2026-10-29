package transport

import "net/url"

type requestOptions struct {
	query            url.Values
	responseEnvelope string
}

type RequestOption func(*requestOptions)

func newRequestOptions(options []RequestOption) *requestOptions {
	result := &requestOptions{}

	for _, option := range options {
		option(result)
	}

	return result
}

func WithQuery(query url.Values) RequestOption {
	return func(options *requestOptions) {
		options.query = query
	}
}

func WithResponseEnvelope(name string) RequestOption {
	return func(options *requestOptions) {
		options.responseEnvelope = name
	}
}
