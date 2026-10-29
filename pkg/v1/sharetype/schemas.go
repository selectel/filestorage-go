package sharetype

type View struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type listEnvelope struct {
	ShareTypes []View `json:"share_types"`
}
