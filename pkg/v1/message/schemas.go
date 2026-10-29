package message

// View is a user message about a failed asynchronous operation.
type View struct {
	ID           string `json:"id"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	MessageLevel string `json:"message_level"`
	UserMessage  string `json:"user_message"`
	CreatedAt    string `json:"created_at"`
}

type listEnvelope struct {
	Messages []View `json:"messages"`
}
