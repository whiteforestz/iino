package tglistener

const (
	getUpdatesUpdateTypeMessage = "message"
)

type getUpdatesIn struct {
	Offset         int64    `json:"offset,omitempty"`
	Limit          int64    `json:"limit,omitempty"`
	Timeout        int64    `json:"timeout,omitempty"`
	AllowedUpdates []string `json:"allowed_updates,omitempty"`
}

type getUpdatesOut struct {
	Result *[]dtoUpdate
}

const (
	sendMessageParseModeMarkdownV2 = "MarkdownV2"
)

type sendMessageIn struct {
	ChatID              int64  `json:"chat_id"`
	Text                string `json:"text"`
	ParseMode           string `json:"parse_mode,omitempty"`
	DisableNotification bool   `json:"disable_notification,omitempty"`
}

type sendMessageOut dtoMessage

type dtoUpdate struct {
	UpdateID int64       `json:"update_id"`
	Message  *dtoMessage `json:"message,omitempty"`
}

type dtoMessage struct {
	MessageID int64    `json:"message_id"`
	From      *dtoUser `json:"from,omitempty"`
	Chat      dtoChat  `json:"chat,omitempty"`
	Text      string   `json:"text"`
}

type dtoUser struct {
	ID int64 `json:"id"`
}

const (
	chatTypePrivate = "private"
)

type dtoChat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}
