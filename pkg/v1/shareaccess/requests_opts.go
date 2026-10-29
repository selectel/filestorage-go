package shareaccess

const accessTypeIP = "ip"

type AllowOpts struct {
	// AccessTo is an IPv4 address or CIDR, such as 10.0.0.0/24.
	AccessTo string

	// AccessLevel is rw or ro.
	AccessLevel string
	Metadata    map[string]string
}

// ListOpts are server-side filters. Empty fields are not sent.
type ListOpts struct {
	// Metadata selects rules that have all the given keys and values.
	Metadata map[string]string
}

type allowRequest struct {
	AllowAccess allowBody `json:"allow_access"`
}

type allowBody struct {
	AccessType  string            `json:"access_type"`
	AccessTo    string            `json:"access_to"`
	AccessLevel string            `json:"access_level,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type denyRequest struct {
	DenyAccess denyBody `json:"deny_access"`
}

type denyBody struct {
	AccessID string `json:"access_id"`
}
