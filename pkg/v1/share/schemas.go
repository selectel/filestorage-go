package share

import "github.com/selectel/filestorage-go/pkg/v1/internal/pagination"

type View struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ShareProto  string `json:"share_proto"`

	// Size is the share size in GiB.
	Size int `json:"size"`

	// ShareType is the share type ID; ShareTypeName is its name.
	ShareType        string `json:"share_type"`
	ShareTypeName    string `json:"share_type_name"`
	AvailabilityZone string `json:"availability_zone"`

	// ShareNetworkID identifies the service share network created by the fork.
	ShareNetworkID string `json:"share_network_id"`

	// Metadata includes service keys added by the fork.
	Metadata map[string]string `json:"metadata"`

	// CreatedAt is the creation time in UTC without a zone, such as 2026-10-29T10:15:30.000000.
	CreatedAt string `json:"created_at"`
}

// ExportLocation is a path that clients mount, such as 10.0.0.5:/shares/share-<id> for NFS.
type ExportLocation struct {
	ID        string `json:"id"`
	Path      string `json:"path"`
	Preferred bool   `json:"preferred"`
}

type viewEnvelope struct {
	Share View `json:"share"`
}

type exportLocationsEnvelope struct {
	ExportLocations []ExportLocation `json:"export_locations"`
}

type pageEnvelope struct {
	Shares []View            `json:"shares"`
	Links  []pagination.Link `json:"shares_links"`
}

func (p *pageEnvelope) Items() []View { return p.Shares }

func (p *pageEnvelope) NextHref() string { return pagination.NextHref(p.Links) }
