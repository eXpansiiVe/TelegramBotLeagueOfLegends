package schemes

type ApiTelegramMessage struct {
	Ok     bool `json:"ok"`
	Result struct {
		MessageID int `json:"message_id"`
		From      struct {
			ID        int64  `json:"id"`
			IsBot     bool   `json:"is_bot"`
			FirstName string `json:"first_name"`
			Username  string `json:"username"`
		} `json:"from"`
		Chat struct {
			ID    int64  `json:"id"`
			Title string `json:"title"`
			Type  string `json:"type"`
		} `json:"chat"`
		Date int    `json:"date"`
		Text string `json:"text"`
	} `json:"result"`
}