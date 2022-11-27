package schemes

// Create a struct that mimics the webhook response body
// https://core.telegram.org/bots/api#update
type WebhookBody struct {
	Message struct {
		MessageID int    `json:"message_id"`
		Text      string `json:"text"`
		Chat      struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}