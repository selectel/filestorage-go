package sharetype

import (
	"net/http"
	"testing"

	"github.com/selectel/filestorage-go/pkg/v1/internal/testhttp"
	"github.com/stretchr/testify/require"
)

func TestShareType_List(t *testing.T) {
	httpClient := &testhttp.Client{Answers: []testhttp.Answer{{Status: http.StatusOK, Body: `{
		"share_types": [{
			"id": "type-1",
			"name": "basic.ru-9a",
			"description": "Basic storage",
			"extra_specs": {"driver_handles_share_servers": "True"},
			"is_default": false
		}],
		"volume_types": []
	}`}}}

	types, _, err := List(t.Context(), testhttp.NewSDKClient(httpClient))
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, httpClient.LastRequest(t).Method)
	require.Equal(t, testhttp.BasePath+"/types", httpClient.LastRequest(t).URL.Path)
	require.Equal(t, []View{{ID: "type-1", Name: "basic.ru-9a", Description: "Basic storage"}}, types)
}
