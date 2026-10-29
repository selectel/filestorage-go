package sharenetwork

type View struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Subnets []Subnet `json:"share_network_subnets"`
}

type Subnet struct {
	ID               string `json:"id"`
	AvailabilityZone string `json:"availability_zone"`
	NetworkID        string `json:"neutron_net_id"`
	SubnetID         string `json:"neutron_subnet_id"`
	CIDR             string `json:"cidr"`
}

type viewEnvelope struct {
	ShareNetwork View `json:"share_network"`
}
