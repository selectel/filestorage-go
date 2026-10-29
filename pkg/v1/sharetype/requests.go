package sharetype

import (
	"context"
	"net/http"

	v1 "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/selectel/filestorage-go/pkg/v1/internal/transport"
)

func List(ctx context.Context, client *v1.Client) ([]View, *v1.Response, error) {
	envelope := &listEnvelope{}

	response, err := transport.DoRequest(
		ctx, client, http.MethodGet, "types", http.StatusOK, nil, envelope,
		transport.WithResponseEnvelope("share_types"),
	)
	if err != nil {
		return nil, response, err
	}

	return envelope.ShareTypes, response, nil
}
