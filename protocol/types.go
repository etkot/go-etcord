package protocol

type MsgType uint8

const (
	ErrorType MsgType = iota
	LoginType
	ClientConnectedType
	ClientDisconnectedType
	GetClientsType
	GetChannelsType
	GetChatHistoryType

	ChatMessageType MsgType = iota + 10

	VoiceChannelJoinType MsgType = iota + 20
	VoiceChannelLeaveType
)

func (mt MsgType) String() string {
	switch mt {
	case ErrorType:
		return "Error"
	case LoginType:
		return "Login"
	case ClientConnectedType:
		return "ClientConnectedType"
	case ClientDisconnectedType:
		return "ClientDisconnectedType"
	case GetClientsType:
		return "GetClients"
	case GetChannelsType:
		return "GetChannels"
	case GetChatHistoryType:
		return "GetChatHistoryType"
	case ChatMessageType:
		return "ChatMessage"
	case VoiceChannelJoinType:
		return "VoiceChannelJoin"
	case VoiceChannelLeaveType:
		return "VoiceChannelLeave"
	}
	return ""
}

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

type ChatMessage struct {
	MessageID  uint16 `json:"messageId"`
	SenderID   uint16 `json:"senderId"`
	SenderName string `json:"senderName"`
	Content    string `json:"content"`
}
