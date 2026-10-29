package shareaccess

import (
	"net/http"
	"testing"

	v1 "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/selectel/filestorage-go/pkg/v1/internal/testhttp"
	"github.com/stretchr/testify/require"
)

const (
	testShareID     = "4f3c2b1a-0e9d-8c7b-6a59-483726150abc"
	testAccessID    = "a1b2c3d4-0000-1111-2222-333344445555"
	testCreateKey   = "selectel_tf_create_token"
	testCreateValue = "value-1"

	apiAccessJSON = `{"access": {
		"id": "a1b2c3d4-0000-1111-2222-333344445555",
		"share_id": "4f3c2b1a-0e9d-8c7b-6a59-483726150abc",
		"access_type": "ip",
		"access_to": "10.0.0.0/24",
		"access_level": "rw",
		"state": "queued_to_apply",
		"access_key": null,
		"created_at": "2026-10-29T10:00:00.000000",
		"updated_at": null,
		"metadata": {"selectel_tf_create_token": "value-1"}
	}}`
)

func TestShareAccess_Allow(t *testing.T) {
	httpClient := &testhttp.Client{Answers: []testhttp.Answer{{Status: http.StatusOK, Body: apiAccessJSON}}}

	rule, response, err := Allow(t.Context(), testhttp.NewSDKClient(httpClient), testShareID, AllowOpts{
		AccessTo:    "10.0.0.0/24",
		AccessLevel: "rw",
		Metadata:    map[string]string{testCreateKey: testCreateValue},
	})
	require.NoError(t, err)

	request := httpClient.LastRequest(t)
	require.Equal(t, http.MethodPost, request.Method)
	require.Equal(t, testhttp.BasePath+"/shares/"+testShareID+"/action", request.URL.Path)
	require.Equal(t, v1.Microversion, request.Header.Get("X-OpenStack-Manila-API-Version"))
	require.JSONEq(t, `{"allow_access": {
		"access_type": "ip",
		"access_to": "10.0.0.0/24",
		"access_level": "rw",
		"metadata": {"selectel_tf_create_token": "value-1"}
	}}`, httpClient.LastBody(t))

	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, &View{
		ID:          testAccessID,
		ShareID:     testShareID,
		AccessType:  "ip",
		AccessTo:    "10.0.0.0/24",
		AccessLevel: "rw",
		State:       "queued_to_apply",
		Metadata:    map[string]string{testCreateKey: testCreateValue},
	}, rule)
}

func TestShareAccess_AllowRejectsEmptyInput(t *testing.T) {
	httpClient := &testhttp.Client{}
	client := testhttp.NewSDKClient(httpClient)

	_, _, err := Allow(t.Context(), client, "", AllowOpts{AccessTo: "10.0.0.1"})
	require.True(t, v1.IsKind(err, v1.KindInvalidRequest))

	_, _, err = Allow(t.Context(), client, testShareID, AllowOpts{})
	require.True(t, v1.IsKind(err, v1.KindInvalidRequest))
	require.Empty(t, httpClient.Requests)
}

func TestShareAccess_Deny(t *testing.T) {
	httpClient := &testhttp.Client{Answers: []testhttp.Answer{{Status: http.StatusAccepted}}}

	response, err := Deny(t.Context(), testhttp.NewSDKClient(httpClient), testShareID, testAccessID)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, response.StatusCode)

	request := httpClient.LastRequest(t)
	require.Equal(t, http.MethodPost, request.Method)
	require.Equal(t, testhttp.BasePath+"/shares/"+testShareID+"/action", request.URL.Path)
	require.JSONEq(t, `{"deny_access": {"access_id": "`+testAccessID+`"}}`, httpClient.LastBody(t))
}

func TestShareAccess_Get(t *testing.T) {
	httpClient := &testhttp.Client{Answers: []testhttp.Answer{{Status: http.StatusOK, Body: apiAccessJSON}}}

	rule, _, err := Get(t.Context(), testhttp.NewSDKClient(httpClient), testAccessID)
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, httpClient.LastRequest(t).Method)
	require.Equal(t, testhttp.BasePath+"/share-access-rules/"+testAccessID, httpClient.LastRequest(t).URL.Path)
	require.Equal(t, testCreateValue, rule.Metadata[testCreateKey])
}

func TestShareAccess_List(t *testing.T) {
	t.Run("FiltersByShareAndMetadata", func(t *testing.T) {
		httpClient := &testhttp.Client{Answers: []testhttp.Answer{{Status: http.StatusOK, Body: `{"access_list": [
			{"id": "r-1", "access_to": "10.0.0.1", "access_level": "ro", "state": "active",
			 "metadata": {"selectel_tf_create_token": "value-1"}}
		]}`}}}

		rules, _, err := List(t.Context(), testhttp.NewSDKClient(httpClient), testShareID, ListOpts{
			Metadata: map[string]string{testCreateKey: testCreateValue},
		})
		require.NoError(t, err)

		request := httpClient.LastRequest(t)
		require.Equal(t, testhttp.BasePath+"/share-access-rules", request.URL.Path)
		require.Equal(t, testShareID, request.URL.Query().Get("share_id"))
		require.JSONEq(t, `{"selectel_tf_create_token": "value-1"}`, request.URL.Query().Get("metadata"))
		require.Equal(t, []View{{
			ID: "r-1", AccessTo: "10.0.0.1", AccessLevel: "ro", State: "active",
			Metadata: map[string]string{testCreateKey: testCreateValue},
		}}, rules)
	})

	t.Run("RequiresShareID", func(t *testing.T) {
		httpClient := &testhttp.Client{}

		_, _, err := List(t.Context(), testhttp.NewSDKClient(httpClient), "", ListOpts{})
		require.True(t, v1.IsKind(err, v1.KindInvalidRequest))
		require.Empty(t, httpClient.Requests)
	})

	t.Run("OmitsEmptyMetadataFilter", func(t *testing.T) {
		httpClient := &testhttp.Client{Answers: []testhttp.Answer{{Status: http.StatusOK, Body: `{"access_list": []}`}}}

		rules, _, err := List(t.Context(), testhttp.NewSDKClient(httpClient), testShareID, ListOpts{})
		require.NoError(t, err)
		require.Empty(t, rules)
		require.False(t, httpClient.LastRequest(t).URL.Query().Has("metadata"))
	})
}
