package availabilityzone

import (
	"net/http"
	"testing"

	"github.com/selectel/filestorage-go/pkg/v1/internal/testhttp"
	"github.com/stretchr/testify/require"
)

func TestAvailabilityZone_List(t *testing.T) {
	httpClient := &testhttp.Client{Answers: []testhttp.Answer{{Status: http.StatusOK, Body: `{"availability_zones": [
		{"id": "az-1", "name": "ru-9a", "created_at": "2026-01-01T00:00:00.000000", "updated_at": null}
	]}`}}}

	zones, _, err := List(t.Context(), testhttp.NewSDKClient(httpClient))
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, httpClient.LastRequest(t).Method)
	require.Equal(t, testhttp.BasePath+"/availability-zones", httpClient.LastRequest(t).URL.Path)
	require.Equal(t, []View{{ID: "az-1", Name: "ru-9a"}}, zones)
}
