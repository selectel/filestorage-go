package message

import (
	"context"
	"net/http"
	"net/url"

	v1 "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/selectel/filestorage-go/pkg/v1/internal/transport"
)

// ListOpts selects the messages to list.
type ListOpts struct {
	// ResourceID limits the list to the messages about one resource, such as a share.
	ResourceID string
}

// List returns the project messages, newest first.
func List(ctx context.Context, client *v1.Client, opts ListOpts) ([]View, *v1.Response, error) {
	query := url.Values{"sort_key": {"created_at"}, "sort_dir": {"desc"}}
	if opts.ResourceID != "" {
		query.Set("resource_id", opts.ResourceID)
	}

	envelope := &listEnvelope{}

	response, err := transport.DoRequest(
		ctx, client, http.MethodGet, "messages", http.StatusOK, nil, envelope,
		transport.WithQuery(query), transport.WithResponseEnvelope("messages"),
	)
	if err != nil {
		return nil, response, err
	}

	return envelope.Messages, response, nil
}
