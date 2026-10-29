package message

import (
	"net/http"
	"testing"

	"github.com/selectel/filestorage-go/pkg/v1/internal/testhttp"
	"github.com/stretchr/testify/require"
)

func TestMessage_List(t *testing.T) {
	httpClient := &testhttp.Client{Answers: []testhttp.Answer{{Status: http.StatusOK, Body: `{
		"messages": [{
			"id": "message-1",
			"project_id": "project-1",
			"action_id": "002",
			"detail_id": "004",
			"message_level": "ERROR",
			"created_at": "2026-10-28T22:05:34.182953",
			"resource_type": "SHARE",
			"resource_id": "share-1",
			"user_message": "create: Could not find an existing share server."
		}]
	}`}}}

	messages, _, err := List(t.Context(), testhttp.NewSDKClient(httpClient), ListOpts{ResourceID: "share-1"})
	require.NoError(t, err)

	request := httpClient.LastRequest(t)
	require.Equal(t, http.MethodGet, request.Method)
	require.Equal(t, testhttp.BasePath+"/messages", request.URL.Path)
	require.Equal(t, "resource_id=share-1&sort_dir=desc&sort_key=created_at", request.URL.RawQuery)
	require.Equal(t, []View{{
		ID:           "message-1",
		ResourceType: "SHARE",
		ResourceID:   "share-1",
		MessageLevel: "ERROR",
		UserMessage:  "create: Could not find an existing share server.",
		CreatedAt:    "2026-10-28T22:05:34.182953",
	}}, messages)
}
