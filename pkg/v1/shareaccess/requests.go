package shareaccess

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	v1 "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/selectel/filestorage-go/pkg/v1/internal/transport"
)

const basePath = "share-access-rules"

// Allow adds an IP rule to the share. A structured Manila fault means that no rule was created; any
// other failure leaves the outcome unknown, and List with the metadata of the request finds the rule.
func Allow(ctx context.Context, client *v1.Client, shareID string, opts AllowOpts) (*View, *v1.Response, error) {
	path, err := shareActionPath(shareID)
	if err != nil {
		return nil, nil, err
	}

	if opts.AccessTo == "" {
		return nil, nil, &v1.Error{Kind: v1.KindInvalidRequest, Message: "the access target is required"}
	}

	request := allowRequest{AllowAccess: allowBody{
		AccessType:  accessTypeIP,
		AccessTo:    opts.AccessTo,
		AccessLevel: opts.AccessLevel,
		Metadata:    opts.Metadata,
	}}
	envelope := &viewEnvelope{}

	response, err := transport.DoRequest(
		ctx, client, http.MethodPost, path, http.StatusOK, request, envelope,
		transport.WithResponseEnvelope("access"),
	)
	if err != nil {
		return nil, response, err
	}

	return &envelope.Access, response, nil
}

// Deny removes the rule from the share. The fork answers not found when the rule belongs to another share.
func Deny(ctx context.Context, client *v1.Client, shareID, accessID string) (*v1.Response, error) {
	path, err := shareActionPath(shareID)
	if err != nil {
		return nil, err
	}

	if accessID == "" {
		return nil, &v1.Error{Kind: v1.KindInvalidRequest, Message: "the access rule ID is required"}
	}

	request := denyRequest{DenyAccess: denyBody{AccessID: accessID}}

	return transport.DoRequest(ctx, client, http.MethodPost, path, http.StatusAccepted, request, nil)
}

func Get(ctx context.Context, client *v1.Client, accessID string) (*View, *v1.Response, error) {
	if accessID == "" {
		return nil, nil, &v1.Error{Kind: v1.KindInvalidRequest, Message: "the access rule ID is required"}
	}

	envelope := &viewEnvelope{}

	response, err := transport.DoRequest(
		ctx, client, http.MethodGet, basePath+"/"+url.PathEscape(accessID), http.StatusOK, nil, envelope,
		transport.WithResponseEnvelope("access"),
	)
	if err != nil {
		return nil, response, err
	}

	return &envelope.Access, response, nil
}

// List returns the rules of the share. Manila answers a bad request, not not found, for a missing share.
func List(ctx context.Context, client *v1.Client, shareID string, opts ListOpts) ([]View, *v1.Response, error) {
	if shareID == "" {
		return nil, nil, &v1.Error{Kind: v1.KindInvalidRequest, Message: "the share ID is required"}
	}

	query := url.Values{"share_id": {shareID}}

	if len(opts.Metadata) != 0 {
		// Manila parses the filter as a Python literal; a JSON object of strings is one.
		encoded, err := json.Marshal(opts.Metadata)
		if err != nil {
			return nil, nil, &v1.Error{
				Kind: v1.KindInvalidRequest, Message: "unable to encode the metadata filter", Err: err,
			}
		}

		query.Set("metadata", string(encoded))
	}

	envelope := &listEnvelope{}

	response, err := transport.DoRequest(
		ctx, client, http.MethodGet, basePath, http.StatusOK, nil, envelope,
		transport.WithQuery(query), transport.WithResponseEnvelope("access_list"),
	)
	if err != nil {
		return nil, response, err
	}

	return envelope.AccessList, response, nil
}

func shareActionPath(shareID string) (string, error) {
	if shareID == "" {
		return "", &v1.Error{Kind: v1.KindInvalidRequest, Message: "the share ID is required"}
	}

	return "shares/" + url.PathEscape(shareID) + "/action", nil
}
