package share

// CreateOpts describe a share in a private network. The share is created with no access rules.
type CreateOpts struct {
	// ShareProto is NFS or CIFS.
	ShareProto string

	// Size is the requested size in GiB.
	Size int

	// ShareType is the name of the share type.
	ShareType   string
	Name        string
	Description string
	Metadata    map[string]string

	// NetworkID and SubnetID identify the private network and subnet of the share.
	NetworkID string
	SubnetID  string

	// IP is an optional IPv4 address in the subnet. The API chooses an address when it is empty.
	IP string
}

// Nil fields are not changed.
type UpdateOpts struct {
	Name        *string
	Description *string
}

func (opts UpdateOpts) isEmpty() bool {
	return opts.Name == nil && opts.Description == nil
}

// ListOpts are server-side filters. Empty fields are not sent.
type ListOpts struct {
	Name               string
	Description        string
	Status             string
	ExportLocationPath string

	// Metadata matches shares that have every given key and value.
	Metadata map[string]string

	// Limit is the page size. ListDetail still reads all pages.
	Limit int
}

// createBody is the flat body of the fork: the share fields are not wrapped in a "share" envelope.
type createBody struct {
	ShareProto  string            `json:"share_proto"`
	Size        int               `json:"size"`
	ShareType   string            `json:"share_type,omitempty"`
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	NetworkID   string            `json:"neutron_net_id"`
	SubnetID    string            `json:"neutron_subnet_id"`
	IP          string            `json:"ip,omitempty"`

	// AccessList is always empty: at 2.75 it creates a share that nobody can access.
	AccessList []accessRule `json:"access_list"`
}

type accessRule struct{}

type updateRequest struct {
	Share updateBody `json:"share"`
}

type updateBody struct {
	Name        *string `json:"display_name,omitempty"`
	Description *string `json:"display_description,omitempty"`
}

type extendRequest struct {
	Extend extendBody `json:"extend"`
}

type extendBody struct {
	NewSize int `json:"new_size"`
}

type metadataEnvelope struct {
	Metadata map[string]string `json:"metadata"`
}
