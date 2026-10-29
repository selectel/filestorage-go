package sharenetwork

import (
	"net/http"
	"testing"

	v1 "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/selectel/filestorage-go/pkg/v1/internal/testhttp"
	"github.com/stretchr/testify/require"
)

func TestShareNetwork_Get(t *testing.T) {
	httpClient := &testhttp.Client{Answers: []testhttp.Answer{{Status: http.StatusOK, Body: `{"share_network": {
		"id": "sn-1",
		"name": "share-network-1",
		"status": "active",
		"share_network_subnets": [{
			"id": "sns-1",
			"availability_zone": null,
			"neutron_net_id": "net-1",
			"neutron_subnet_id": "subnet-1",
			"cidr": "10.0.0.0/24",
			"mtu": 1500
		}]
	}}`}}}

	network, _, err := Get(t.Context(), testhttp.NewSDKClient(httpClient), "sn-1")
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, httpClient.LastRequest(t).Method)
	require.Equal(t, testhttp.BasePath+"/share-networks/sn-1", httpClient.LastRequest(t).URL.Path)
	require.Equal(t, &View{
		ID:   "sn-1",
		Name: "share-network-1",
		Subnets: []Subnet{{
			ID: "sns-1", NetworkID: "net-1", SubnetID: "subnet-1", CIDR: "10.0.0.0/24",
		}},
	}, network)
}

func TestShareNetwork_GetRequiresID(t *testing.T) {
	httpClient := &testhttp.Client{}

	_, _, err := Get(t.Context(), testhttp.NewSDKClient(httpClient), "")
	require.True(t, v1.IsKind(err, v1.KindInvalidRequest))
	require.Empty(t, httpClient.Requests)
}
