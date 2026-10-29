package share

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	v1 "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/selectel/filestorage-go/pkg/v1/internal/pagination"
	"github.com/selectel/filestorage-go/pkg/v1/internal/transport"
)

const (
	basePath   = "shares"
	detailPath = "shares/detail"
)

// Create sends the flat create body of the fork with an empty access list, so the new share has
// no access. A structured Manila fault means that nothing was created; any other failure leaves
// the outcome unknown.
func Create(ctx context.Context, client *v1.Client, opts CreateOpts) (*View, *v1.Response, error) {
	request := createBody{
		ShareProto:  opts.ShareProto,
		Size:        opts.Size,
		ShareType:   opts.ShareType,
		Name:        opts.Name,
		Description: opts.Description,
		Metadata:    opts.Metadata,
		NetworkID:   opts.NetworkID,
		SubnetID:    opts.SubnetID,
		IP:          opts.IP,
		AccessList:  []accessRule{},
	}

	envelope := &viewEnvelope{}

	response, err := transport.DoRequest(
		ctx, client, http.MethodPost, basePath, http.StatusOK, request, envelope,
		transport.WithResponseEnvelope("share"),
	)
	if err != nil {
		return nil, response, err
	}

	return &envelope.Share, response, nil
}

// Manila returns not found for shares hidden from the token.
func Get(ctx context.Context, client *v1.Client, shareID string) (*View, *v1.Response, error) {
	path, err := sharePath(shareID)
	if err != nil {
		return nil, nil, err
	}

	envelope := &viewEnvelope{}

	response, err := transport.DoRequest(
		ctx, client, http.MethodGet, path, http.StatusOK, nil, envelope,
		transport.WithResponseEnvelope("share"),
	)
	if err != nil {
		return nil, response, err
	}

	return &envelope.Share, response, nil
}

// ListDetail returns all pages of the detailed share list and never returns a partial result.
func ListDetail(ctx context.Context, client *v1.Client, opts ListOpts) ([]View, error) {
	query, err := opts.query()
	if err != nil {
		return nil, err
	}

	return pagination.ReadAll[View](ctx, client, detailPath, "shares", query, opts.Limit, func() *pageEnvelope {
		return &pageEnvelope{}
	})
}

func Update(ctx context.Context, client *v1.Client, shareID string, opts UpdateOpts) (*View, *v1.Response, error) {
	path, err := sharePath(shareID)
	if err != nil {
		return nil, nil, err
	}

	if opts.isEmpty() {
		return nil, nil, &v1.Error{
			Kind:    v1.KindInvalidRequest,
			Message: "an update needs the name, the description or both",
		}
	}

	request := updateRequest{Share: updateBody(opts)}
	envelope := &viewEnvelope{}

	response, err := transport.DoRequest(
		ctx, client, http.MethodPut, path, http.StatusOK, request, envelope,
		transport.WithResponseEnvelope("share"),
	)
	if err != nil {
		return nil, response, err
	}

	return &envelope.Share, response, nil
}

// Extend only grows a share; shrinking is an administrative operation.
func Extend(ctx context.Context, client *v1.Client, shareID string, newSize int) (*v1.Response, error) {
	path, err := shareActionPath(shareID)
	if err != nil {
		return nil, err
	}

	if newSize <= 0 {
		return nil, &v1.Error{
			Kind:    v1.KindInvalidRequest,
			Message: "the new size of the share must be a positive number of gibibytes",
		}
	}

	request := extendRequest{Extend: extendBody{NewSize: newSize}}

	return transport.DoRequest(ctx, client, http.MethodPost, path, http.StatusAccepted, request, nil)
}

// Delete also removes the access rules; the fork removes its service share network afterwards.
func Delete(ctx context.Context, client *v1.Client, shareID string) (*v1.Response, error) {
	path, err := sharePath(shareID)
	if err != nil {
		return nil, err
	}

	return transport.DoRequest(ctx, client, http.MethodDelete, path, http.StatusAccepted, nil, nil)
}

// SetMetadata merges the given keys into the share metadata and returns the resulting metadata.
// Keys that are not given are kept, including the service keys of the fork.
func SetMetadata(
	ctx context.Context, client *v1.Client, shareID string, metadata map[string]string,
) (map[string]string, *v1.Response, error) {
	path, err := sharePath(shareID)
	if err != nil {
		return nil, nil, err
	}

	if len(metadata) == 0 {
		return nil, nil, &v1.Error{Kind: v1.KindInvalidRequest, Message: "at least one metadata key is required"}
	}

	envelope := &metadataEnvelope{}

	response, err := transport.DoRequest(
		ctx, client, http.MethodPost, path+"/metadata", http.StatusOK, metadataEnvelope{Metadata: metadata}, envelope,
		transport.WithResponseEnvelope("metadata"),
	)
	if err != nil {
		return nil, response, err
	}

	return envelope.Metadata, response, nil
}

func DeleteMetadatum(ctx context.Context, client *v1.Client, shareID, key string) (*v1.Response, error) {
	path, err := sharePath(shareID)
	if err != nil {
		return nil, err
	}

	if key == "" {
		return nil, &v1.Error{Kind: v1.KindInvalidRequest, Message: "the metadata key is required"}
	}

	return transport.DoRequest(
		ctx, client, http.MethodDelete, path+"/metadata/"+url.PathEscape(key), http.StatusOK, nil, nil,
	)
}

// ListExportLocations returns the paths visible to the token. Administrative paths are hidden.
func ListExportLocations(
	ctx context.Context, client *v1.Client, shareID string,
) ([]ExportLocation, *v1.Response, error) {
	path, err := sharePath(shareID)
	if err != nil {
		return nil, nil, err
	}

	envelope := &exportLocationsEnvelope{}

	response, err := transport.DoRequest(
		ctx, client, http.MethodGet, path+"/export_locations", http.StatusOK, nil, envelope,
	)
	if err != nil {
		return nil, response, err
	}

	return envelope.ExportLocations, response, nil
}

func (opts ListOpts) query() (url.Values, error) {
	query := url.Values{}

	for key, value := range map[string]string{
		"name":                 opts.Name,
		"description":          opts.Description,
		"status":               opts.Status,
		"export_location_path": opts.ExportLocationPath,
	} {
		if value != "" {
			query.Set(key, value)
		}
	}

	if len(opts.Metadata) != 0 {
		// Manila parses the filter as a Python literal; a JSON object of strings is one.
		encoded, err := json.Marshal(opts.Metadata)
		if err != nil {
			return nil, &v1.Error{Kind: v1.KindInvalidRequest, Message: "unable to encode the metadata filter", Err: err}
		}

		query.Set("metadata", string(encoded))
	}

	return query, nil
}

func sharePath(shareID string) (string, error) {
	if shareID == "" {
		return "", &v1.Error{Kind: v1.KindInvalidRequest, Message: "the share ID is required"}
	}

	return basePath + "/" + url.PathEscape(shareID), nil
}

func shareActionPath(shareID string) (string, error) {
	path, err := sharePath(shareID)
	if err != nil {
		return "", err
	}

	return path + "/action", nil
}
