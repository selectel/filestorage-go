package availabilityzone

type View struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type listEnvelope struct {
	AvailabilityZones []View `json:"availability_zones"`
}
