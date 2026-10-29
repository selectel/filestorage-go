package share

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	v1 "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/selectel/filestorage-go/pkg/v1/internal/testhttp"
	"github.com/stretchr/testify/require"
)

const (
	testShareID   = "4f3c2b1a-0e9d-8c7b-6a59-483726150abc"
	testNetworkID = "net-1"
	testSubnetID  = "subnet-1"
	testOwnerKey  = "owner"

	apiShareJSON = `{"share": {
		"id": "4f3c2b1a-0e9d-8c7b-6a59-483726150abc",
		"status": "creating",
		"name": "data",
		"description": "a share",
		"share_proto": "NFS",
		"size": 50,
		"share_type": "type-uuid",
		"share_type_name": "basic.ru-9a",
		"availability_zone": "ru-9a",
		"share_network_id": "sn-1",
		"metadata": {"owner": "team", "delete_empty_share_network": "enabled"},
		"created_at": "2026-10-29T10:15:30.000000",
		"links": []
	}}`
)

func sharesPath() string { return testhttp.BasePath + "/shares" }

func TestShare_Create(t *testing.T) {
	t.Run("SendsFlatBodyWithEmptyAccessList", func(t *testing.T) {
		httpClient := &testhttp.Client{Answers: []testhttp.Answer{{Status: http.StatusOK, Body: apiShareJSON}}}

		created, response, err := Create(t.Context(), testhttp.NewSDKClient(httpClient), CreateOpts{
			ShareProto:  "NFS",
			Size:        50,
			ShareType:   "basic.ru-9a",
			Name:        "data",
			Description: "a share",
			Metadata:    map[string]string{testOwnerKey: "team"},
			NetworkID:   testNetworkID,
			SubnetID:    testSubnetID,
			IP:          "10.0.0.5",
		})
		require.NoError(t, err)

		request := httpClient.LastRequest(t)
		require.Equal(t, http.MethodPost, request.Method)
		require.Equal(t, sharesPath(), request.URL.Path)
		require.Equal(t, v1.Microversion, request.Header.Get("X-OpenStack-Manila-API-Version"))
		require.JSONEq(t, `{
			"share_proto": "NFS",
			"size": 50,
			"share_type": "basic.ru-9a",
			"name": "data",
			"description": "a share",
			"metadata": {"owner": "team"},
			"neutron_net_id": "net-1",
			"neutron_subnet_id": "subnet-1",
			"ip": "10.0.0.5",
			"access_list": []
		}`, httpClient.LastBody(t))

		require.Equal(t, http.StatusOK, response.StatusCode)
		require.Equal(t, testShareID, created.ID)
		require.Equal(t, "type-uuid", created.ShareType)
		require.Equal(t, "basic.ru-9a", created.ShareTypeName)
		require.Equal(t, "2026-10-29T10:15:30.000000", created.CreatedAt)
		require.Equal(t, "sn-1", created.ShareNetworkID)
		require.Equal(t, "ru-9a", created.AvailabilityZone)
		require.Equal(t, map[string]string{testOwnerKey: "team", "delete_empty_share_network": "enabled"}, created.Metadata)
	})

	t.Run("OmitsOptionalFields", func(t *testing.T) {
		httpClient := &testhttp.Client{Answers: []testhttp.Answer{{Status: http.StatusOK, Body: apiShareJSON}}}

		_, _, err := Create(t.Context(), testhttp.NewSDKClient(httpClient), CreateOpts{
			ShareProto: "CIFS", Size: 1, NetworkID: testNetworkID, SubnetID: testSubnetID,
		})
		require.NoError(t, err)
		require.JSONEq(t, `{
			"share_proto": "CIFS",
			"size": 1,
			"neutron_net_id": "net-1",
			"neutron_subnet_id": "subnet-1",
			"access_list": []
		}`, httpClient.LastBody(t))
	})

	t.Run("StructuredRefusal", func(t *testing.T) {
		httpClient := &testhttp.Client{Answers: []testhttp.Answer{{
			Status: http.StatusBadRequest,
			Body:   `{"badRequest": {"code": 400, "message": "IP address: 10.0.0.5 is already in use."}}`,
		}}}

		created, _, err := Create(t.Context(), testhttp.NewSDKClient(httpClient), CreateOpts{ShareProto: "NFS", Size: 1})
		require.Nil(t, created)

		var sdkErr *v1.Error
		require.ErrorAs(t, err, &sdkErr)
		require.Equal(t, v1.KindInvalidRequest, sdkErr.Kind)
		require.True(t, sdkErr.StructuredFault)
		require.Contains(t, sdkErr.Message, "already in use")
	})
}

func TestShare_Get(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		httpClient := &testhttp.Client{Answers: []testhttp.Answer{{Status: http.StatusOK, Body: apiShareJSON}}}

		view, _, err := Get(t.Context(), testhttp.NewSDKClient(httpClient), testShareID)
		require.NoError(t, err)
		require.Equal(t, http.MethodGet, httpClient.LastRequest(t).Method)
		require.Equal(t, sharesPath()+"/"+testShareID, httpClient.LastRequest(t).URL.Path)
		require.Equal(t, "NFS", view.ShareProto)
		require.Equal(t, 50, view.Size)
	})

	t.Run("NotFound", func(t *testing.T) {
		httpClient := &testhttp.Client{Answers: []testhttp.Answer{{
			Status: http.StatusNotFound,
			Body:   `{"itemNotFound": {"code": 404, "message": "Share could not be found."}}`,
		}}}

		_, _, err := Get(t.Context(), testhttp.NewSDKClient(httpClient), testShareID)
		require.True(t, v1.IsKind(err, v1.KindNotFound))
	})

	t.Run("RequiresID", func(t *testing.T) {
		httpClient := &testhttp.Client{}

		_, _, err := Get(t.Context(), testhttp.NewSDKClient(httpClient), "")
		require.True(t, v1.IsKind(err, v1.KindInvalidRequest))
		require.Empty(t, httpClient.Requests)
	})
}

func TestShare_ListDetail(t *testing.T) {
	t.Run("ReadsAllPages", func(t *testing.T) {
		httpClient := &testhttp.Client{Answers: []testhttp.Answer{
			{Status: http.StatusOK, Body: apiSharePage([]string{"s-1", "s-2"}, "s-2")},
			{Status: http.StatusOK, Body: apiSharePage([]string{"s-3"}, "")},
		}}

		shares, err := ListDetail(t.Context(), testhttp.NewSDKClient(httpClient), ListOpts{Limit: 2})
		require.NoError(t, err)
		require.Len(t, shares, 3)
		require.Equal(t, "s-3", shares[2].ID)
		require.Equal(t, sharesPath()+"/detail", httpClient.Requests[0].URL.Path)
		require.Equal(t, "limit=2", httpClient.Requests[0].URL.RawQuery)
		require.Equal(t, "limit=2&marker=s-2", httpClient.Requests[1].URL.RawQuery)
	})

	t.Run("SendsFilters", func(t *testing.T) {
		httpClient := &testhttp.Client{Answers: []testhttp.Answer{{Status: http.StatusOK, Body: apiSharePage(nil, "")}}}

		_, err := ListDetail(t.Context(), testhttp.NewSDKClient(httpClient), ListOpts{
			Name:               "data",
			Description:        "a share",
			Status:             "available",
			ExportLocationPath: "10.0.0.5:/shares/share-1",
			Metadata:           map[string]string{"selectel_tf_create_token": "token-1"},
		})
		require.NoError(t, err)

		query := httpClient.LastRequest(t).URL.Query()
		require.Equal(t, "data", query.Get("name"))
		require.Equal(t, "a share", query.Get("description"))
		require.Equal(t, "available", query.Get("status"))
		require.Equal(t, "10.0.0.5:/shares/share-1", query.Get("export_location_path"))
		require.JSONEq(t, `{"selectel_tf_create_token": "token-1"}`, query.Get("metadata"))
	})

	t.Run("NoFiltersWhenEmpty", func(t *testing.T) {
		httpClient := &testhttp.Client{Answers: []testhttp.Answer{{Status: http.StatusOK, Body: apiSharePage(nil, "")}}}

		_, err := ListDetail(t.Context(), testhttp.NewSDKClient(httpClient), ListOpts{})
		require.NoError(t, err)
		require.Empty(t, httpClient.LastRequest(t).URL.RawQuery)
	})

	t.Run("NoPartialResult", func(t *testing.T) {
		httpClient := &testhttp.Client{Answers: []testhttp.Answer{
			{Status: http.StatusOK, Body: apiSharePage([]string{"s-1"}, "s-1")},
			{Status: http.StatusInternalServerError, Body: `{"computeFault": {"code": 500, "message": "boom"}}`},
		}}

		shares, err := ListDetail(t.Context(), testhttp.NewSDKClient(httpClient), ListOpts{})
		require.Nil(t, shares)
		require.True(t, v1.IsKind(err, v1.KindIncompleteList))
		require.True(t, v1.IsKind(err, v1.KindServerError))
	})
}

func TestShare_Update(t *testing.T) {
	t.Run("SendsDisplayFields", func(t *testing.T) {
		httpClient := &testhttp.Client{Answers: []testhttp.Answer{{Status: http.StatusOK, Body: apiShareJSON}}}
		name, description := "renamed", ""

		_, _, err := Update(t.Context(), testhttp.NewSDKClient(httpClient), testShareID, UpdateOpts{
			Name: &name, Description: &description,
		})
		require.NoError(t, err)
		require.Equal(t, http.MethodPut, httpClient.LastRequest(t).Method)
		require.Equal(t, sharesPath()+"/"+testShareID, httpClient.LastRequest(t).URL.Path)
		require.JSONEq(t, `{"share": {"display_name": "renamed", "display_description": ""}}`, httpClient.LastBody(t))
	})

	t.Run("RequiresAField", func(t *testing.T) {
		httpClient := &testhttp.Client{}

		_, _, err := Update(t.Context(), testhttp.NewSDKClient(httpClient), testShareID, UpdateOpts{})
		require.True(t, v1.IsKind(err, v1.KindInvalidRequest))
		require.Empty(t, httpClient.Requests)
	})
}

func TestShare_Extend(t *testing.T) {
	t.Run("SendsAction", func(t *testing.T) {
		httpClient := &testhttp.Client{Answers: []testhttp.Answer{{Status: http.StatusAccepted}}}

		_, err := Extend(t.Context(), testhttp.NewSDKClient(httpClient), testShareID, 100)
		require.NoError(t, err)
		require.Equal(t, http.MethodPost, httpClient.LastRequest(t).Method)
		require.Equal(t, sharesPath()+"/"+testShareID+"/action", httpClient.LastRequest(t).URL.Path)
		require.JSONEq(t, `{"extend": {"new_size": 100}}`, httpClient.LastBody(t))
	})

	t.Run("RejectsNonPositiveSize", func(t *testing.T) {
		httpClient := &testhttp.Client{}

		_, err := Extend(t.Context(), testhttp.NewSDKClient(httpClient), testShareID, 0)
		require.True(t, v1.IsKind(err, v1.KindInvalidRequest))
		require.Empty(t, httpClient.Requests)
	})
}

func TestShare_Delete(t *testing.T) {
	httpClient := &testhttp.Client{Answers: []testhttp.Answer{{Status: http.StatusAccepted}}}

	_, err := Delete(t.Context(), testhttp.NewSDKClient(httpClient), testShareID)
	require.NoError(t, err)
	require.Equal(t, http.MethodDelete, httpClient.LastRequest(t).Method)
	require.Equal(t, sharesPath()+"/"+testShareID, httpClient.LastRequest(t).URL.Path)
}

func TestShare_Metadata(t *testing.T) {
	t.Run("SetMergesWithPost", func(t *testing.T) {
		httpClient := &testhttp.Client{Answers: []testhttp.Answer{{
			Status: http.StatusOK,
			Body:   `{"metadata": {"owner": "ops", "delete_empty_share_network": "enabled"}}`,
		}}}

		metadata, _, err := SetMetadata(
			t.Context(), testhttp.NewSDKClient(httpClient), testShareID, map[string]string{testOwnerKey: "ops"},
		)
		require.NoError(t, err)
		require.Equal(t, http.MethodPost, httpClient.LastRequest(t).Method)
		require.Equal(t, sharesPath()+"/"+testShareID+"/metadata", httpClient.LastRequest(t).URL.Path)
		require.JSONEq(t, `{"metadata": {"owner": "ops"}}`, httpClient.LastBody(t))
		require.Equal(t, map[string]string{testOwnerKey: "ops", "delete_empty_share_network": "enabled"}, metadata)
	})

	t.Run("SetRequiresKeys", func(t *testing.T) {
		httpClient := &testhttp.Client{}

		_, _, err := SetMetadata(t.Context(), testhttp.NewSDKClient(httpClient), testShareID, nil)
		require.True(t, v1.IsKind(err, v1.KindInvalidRequest))
		require.Empty(t, httpClient.Requests)
	})

	t.Run("DeleteOneKey", func(t *testing.T) {
		httpClient := &testhttp.Client{Answers: []testhttp.Answer{{Status: http.StatusOK}}}

		_, err := DeleteMetadatum(t.Context(), testhttp.NewSDKClient(httpClient), testShareID, "owner/team")
		require.NoError(t, err)
		require.Equal(t, http.MethodDelete, httpClient.LastRequest(t).Method)
		require.Equal(t, sharesPath()+"/"+testShareID+"/metadata/owner%2Fteam", httpClient.LastRequest(t).URL.EscapedPath())
	})
}

func TestShare_ListExportLocations(t *testing.T) {
	httpClient := &testhttp.Client{Answers: []testhttp.Answer{{Status: http.StatusOK, Body: `{"export_locations": [
		{"id": "el-1", "path": "10.0.0.5:/shares/share-1", "preferred": true, "share_instance_id": "si-1"},
		{"id": "el-2", "path": "10.0.0.6:/shares/share-1", "preferred": false}
	]}`}}}

	locations, _, err := ListExportLocations(t.Context(), testhttp.NewSDKClient(httpClient), testShareID)
	require.NoError(t, err)
	require.Equal(t, sharesPath()+"/"+testShareID+"/export_locations", httpClient.LastRequest(t).URL.Path)
	require.Equal(t, []ExportLocation{
		{ID: "el-1", Path: "10.0.0.5:/shares/share-1", Preferred: true},
		{ID: "el-2", Path: "10.0.0.6:/shares/share-1", Preferred: false},
	}, locations)
}

func apiSharePage(ids []string, marker string) string {
	shares := make([]string, 0, len(ids))
	for _, id := range ids {
		shares = append(shares, fmt.Sprintf(`{"id": %q, "metadata": {}}`, id))
	}

	page := fmt.Sprintf(`{"shares": [%s]`, strings.Join(shares, ","))
	if marker != "" {
		page += fmt.Sprintf(`, "shares_links": [{"rel": "next", "href": %q}]`,
			testhttp.Endpoint+"/shares/detail?limit=2&marker="+marker)
	}

	return page + "}"
}
