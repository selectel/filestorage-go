package shareaccess

type View struct {
	ID          string `json:"id"`
	ShareID     string `json:"share_id"`
	AccessType  string `json:"access_type"`
	AccessTo    string `json:"access_to"`
	AccessLevel string `json:"access_level"`

	// State is queued_to_apply, applying, active, error, queued_to_deny or denying.
	State    string            `json:"state"`
	Metadata map[string]string `json:"metadata"`
}

type viewEnvelope struct {
	Access View `json:"access"`
}

type listEnvelope struct {
	AccessList []View `json:"access_list"`
}
