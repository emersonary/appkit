package messenger

import (
	"encoding/json"
	"time"
)

type User struct {
	ID          string
	TenantID    string
	UserType    string
	AccountID   *string
	DisplayName string
	AvatarURL   *string
	ExternalRef *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Device struct {
	ID              string
	MessengerUserID string
	DeviceKey       string
	Platform        string
	PushToken       *string
	LastSeen        time.Time
}

type Channel struct {
	ID              string
	TenantID        string
	Kind            string
	Title           *string
	AvatarURL       *string
	CreatedByUserID string
	UnreadCount     int32
	LastPreview     *string
	UpdatedAt       time.Time
	Members         []Member
}

type Member struct {
	ID              string
	ChannelID       string
	MessengerUserID string
	Role            string
	LastReadMessage *string
	UnreadCount     int32
	User            User
}

type Message struct {
	ID           string
	ChannelID    string
	Kind         string
	SenderUserID *string
	Sender       *User
	Body         *string
	EventType    *string
	EventPayload json.RawMessage
	ClientRef    *string
	ReplyToID    *string
	CreatedAt    time.Time
}

type ReadReceipt struct {
	ID              string
	MessageID       string
	ChannelID       string
	MessengerUserID string
	DeviceID        *string
	ReadAt          time.Time
}

type Handoff struct {
	ID                  string
	ChannelID           string
	FromMessengerUserID string
	ToMessengerUserID   string
	Reason              *string
	ContextSummary      *string
	CreatedAt           time.Time
}
