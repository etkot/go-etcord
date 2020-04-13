package types

type ChannelType uint8

const (
	NoneChannelType ChannelType = iota // TODO this would be a category
	TextChannelType
	VoiceChannelType
	MultiChannelType // text + voice
)

type Channel struct {
	ID       uint16      `json:"channelId"`
	ParentID uint16      `json:"parentId"`
	Name     string      `json:"name"`
	Type     ChannelType `json:"type"`
}

type Client struct {
	UserID uint16 `json:"userId"`
	Name   string `json:"name"`
}

type Message struct {
	MessageID  uint16 `json:"messageId"`
	SenderID   uint16 `json:"senderId"`
	SenderName string `json:"senderName"`
	Content    string `json:"content"`
}
