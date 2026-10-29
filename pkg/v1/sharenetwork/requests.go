package sharenetwork

import (
	"context"
	"net/http"
	"net/url"

	v1 "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/selectel/filestorage-go/pkg/v1/internal/transport"
)

func Get(ctx context.Context, client *v1.Client, shareNetworkID string) (*View, *v1.Response, error) {
	if shareNetworkID == "" {
		return nil, nil, &v1.Error{Kind: v1.KindInvalidRequest, Message: "the share network ID is required"}
	}

	envelope := &viewEnvelope{}

	response, err := transport.DoRequest(
		ctx, client, http.MethodGet, "share-networks/"+url.PathEscape(shareNetworkID), http.StatusOK, nil, envelope,
		transport.WithResponseEnvelope("share_network"),
	)
	if err != nil {
		return nil, response, err
	}

	return &envelope.ShareNetwork, response, nil
}
