package protocol

import (
	"fmt"
	"math"

	"etcord/common"
	"etcord/types"
)

const (
	PacketHeaderLen = 3 // length + id fields
)

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

type Serializer interface {
	Serialize() []byte
	Deserialize(common.Buffer) error
}

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

// Serialize serializes an Etcord protocol message to a raw packet
func Serialize(m Serializer) ([]byte, error) {
	mb := m.Serialize()
	if len(mb) > math.MaxUint16 {
		return nil, fmt.Errorf("packet content length overflows uint16")
	}

	buf := common.NewBuffer(make([]byte, 0, PacketHeaderLen+len(mb)))
	buf.WriteUint16(1 + uint16(len(mb)))
	buf.Write([]byte{uint8(GetMsgType(m))})
	buf.Write(mb)

	return buf.Bytes(), nil
}

// Deserializes deserializes a raw packet to an Etcord protocol message
func Deserialize(tmp common.Buffer) (Serializer, error) {
	msgLen, err := tmp.ReadUint16()
	if err != nil {
		return nil, err
	}

	buf := common.NewBuffer(tmp.Next(int(msgLen)))

	tb, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	t := MsgType(tb)

	var m Serializer
	switch t {
	case ErrorType:
		m = &Error{}
	case LoginType:
		m = &LoginRequest{}
	case GetClientsType:
		m = &GetClientsRequest{}
	case GetChannelsType:
		m = &GetChannelsRequest{}
	case GetChatHistoryType:
		m = &GetChatHistoryRequest{}
	case ChatMessageType:
		m = &ChatMessageRequest{}
	case VoiceChannelJoinType:
		m = &VoiceChannelJoinRequest{}
	default:
		return nil, fmt.Errorf("unknown type")
	}

	if err = m.Deserialize(buf); err != nil {
		return nil, err
	}

	return m, nil
}

type Error struct {
	Code    uint16 `json:"id"`
	Message string `json:"message"`
}

func (m *Error) Serialize() []byte {
	l := 2 + len(m.Message)
	buf := common.NewBuffer(make([]byte, 0, l))
	buf.WriteUint16(m.Code)
	buf.WriteNullTerminatedString(m.Message)
	return buf.Bytes()
}

func (m *Error) Deserialize(buf common.Buffer) error {
	var err error
	if m.Code, err = buf.ReadUint16(); err != nil {
		return err
	}
	m.Message = buf.ReadNullTerminatedString()
	return nil
}

type LoginRequest struct {
	Name string `json:"name"`
}

func (m *LoginRequest) Serialize() []byte {
	b := []byte(m.Name)
	return b
}

func (m *LoginRequest) Deserialize(buf common.Buffer) error {
	if m.Name = buf.ReadNullTerminatedString(); len(m.Name) == 0 {
		return fmt.Errorf("empty name")
	}
	return nil
}

const (
	GetClientsAll  = 0
	GetClientsOne  = 1
	GetClientsMany = 2
)

type GetClientsRequest struct {
	Type      uint8    `json:"type"`
	ClientID  uint16   `json:"clientId,omitempty"`
	Count     uint16   `json:"count,omitempty"`
	ClientIDs []uint16 `json:"clientIds,omitempty"`
}

func (m *GetClientsRequest) Serialize() []byte { return nil } // TODO

func (m *GetClientsRequest) Deserialize(buf common.Buffer) error {
	var err error

	if m.Type, err = buf.ReadByte(); err != nil {
		return err
	}

	switch m.Type {
	case GetClientsAll:
		break
	case GetClientsOne:
		if m.ClientID, err = buf.ReadUint16(); err != nil {
			return err
		}
	case GetClientsMany:
		if m.Count, err = buf.ReadUint16(); err != nil {
			return err
		}

		var id uint16
		m.ClientIDs = make([]uint16, 0, m.Count)
		for i := 0; i < int(m.Count); i++ {
			if id, err = buf.ReadUint16(); err != nil {
				return err
			}
			m.ClientIDs = append(m.ClientIDs, id)
		}
	}

	return nil
}

type GetClientsResponse struct {
	Count   uint16         `json:"count"`
	Clients []types.Client `json:"clients"`
}

func (m *GetClientsResponse) Serialize() []byte { return nil } // TODO

func (m *GetClientsResponse) Deserialize(common.Buffer) error { return nil }

type GetChannelsRequest struct {
}

func (m *GetChannelsRequest) Serialize() []byte { return nil } // TODO

func (m *GetChannelsRequest) Deserialize(common.Buffer) error { return nil }

type GetChannelsResponse struct {
	Count    uint16          `json:"count"`
	Channels []types.Channel `json:"channels"`
}

func (m *GetChannelsResponse) Serialize() []byte { return nil } // TODO

func (m *GetChannelsResponse) Deserialize(common.Buffer) error { return nil }

type GetChatHistoryRequest struct {
	ChannelID uint16 `json:"channelId"`
	Count     uint16 `json:"count"`
	OffsetID  uint16 `json:"offsetId"`
}

func (m *GetChatHistoryRequest) Serialize() []byte { return nil } // TODO

func (m *GetChatHistoryRequest) Deserialize(buf common.Buffer) error {
	var err error
	if m.ChannelID, err = buf.ReadUint16(); err != nil {
		return err
	}
	if m.Count, err = buf.ReadUint16(); err != nil {
		return err
	}
	if m.OffsetID, err = buf.ReadUint16(); err != nil {
		return err
	}
	return nil
}

type GetChatHistoryResponse struct {
	ChannelID uint16              `json:"channelId"`
	Count     uint16              `json:"count"`
	Messages  []types.ChatMessage `json:"messages"`
}

func (m *GetChatHistoryResponse) Serialize() []byte { return nil } // TODO

func (m *GetChatHistoryResponse) Deserialize(common.Buffer) error { return nil } // TODO

type ChatMessageRequest struct {
	ChannelID uint16 `json:"channelId"`
	Content   string `json:"content"`
}

func (m *ChatMessageRequest) Serialize() []byte {
	l := 2 + len(m.Content)
	buf := common.NewBuffer(make([]byte, 0, l))
	buf.WriteUint16(m.ChannelID)
	buf.WriteNullTerminatedString(m.Content)
	return buf.Bytes()
}

func (m *ChatMessageRequest) Deserialize(buf common.Buffer) error {
	var err error
	if m.ChannelID, err = buf.ReadUint16(); err != nil {
		return err
	}
	if m.Content = buf.ReadNullTerminatedString(); len(m.Content) == 0 {
		return fmt.Errorf("empty message")
	}
	return nil
}

type ChatMessageResponse struct {
	ChannelID uint16            `json:"channelId"`
	Message   types.ChatMessage `json:"message"`
}

func (m *ChatMessageResponse) Serialize() []byte {
	l := 6 + len(m.Message.SenderName) + 1 + len(m.Message.Content)
	buf := common.NewBuffer(make([]byte, 0, l))
	buf.WriteUint16(m.ChannelID)
	buf.WriteUint16(m.Message.MessageID)
	buf.WriteUint16(m.Message.SenderID)
	buf.WriteNullTerminatedString(m.Message.SenderName)
	buf.WriteNullTerminatedString(m.Message.Content)
	return buf.Bytes()
}

func (m *ChatMessageResponse) Deserialize(buf common.Buffer) error {
	var err error

	if m.ChannelID, err = buf.ReadUint16(); err != nil {
		return err
	}
	if m.Message.MessageID, err = buf.ReadUint16(); err != nil {
		return err
	}
	if m.Message.SenderID, err = buf.ReadUint16(); err != nil {
		return err
	}
	if m.Message.SenderName = buf.ReadNullTerminatedString(); len(m.Message.SenderName) == 0 {
		return fmt.Errorf("empty sender name")
	}
	if m.Message.Content = buf.ReadNullTerminatedString(); len(m.Message.Content) == 0 {
		return fmt.Errorf("empty content")
	}
	return nil
}

type VoiceChannelJoinRequest struct {
	ChannelID uint16 `json:"channelId"`
}

func (m *VoiceChannelJoinRequest) Serialize() []byte { return nil } // TODO

func (m *VoiceChannelJoinRequest) Deserialize(buf common.Buffer) error {
	var err error
	if m.ChannelID, err = buf.ReadUint16(); err != nil {
		return err
	}
	return nil
}

type VoiceChannelJoinResponse struct {
	ChannelID uint16 `json:"channelId"`
	ClientID  uint16 `json:"clientId"`
}

func (m *VoiceChannelJoinResponse) Serialize() []byte { return nil } // TODO

func (m *VoiceChannelJoinResponse) Deserialize(common.Buffer) error { return nil } // TODO

func GetMsgType(m Serializer) MsgType {
	switch m.(type) {
	case *Error:
		return ErrorType
	case *LoginRequest:
		return LoginType
	case *GetClientsRequest:
		return GetClientsType
	case *GetClientsResponse:
		return GetClientsType
	case *GetChannelsRequest:
		return GetChannelsType
	case *GetChannelsResponse:
		return GetChannelsType
	case *GetChatHistoryRequest:
		return GetChatHistoryType
	case *GetChatHistoryResponse:
		return GetChatHistoryType
	case *ChatMessageRequest:
		return ChatMessageType
	case *ChatMessageResponse:
		return ChatMessageType
	case *VoiceChannelJoinRequest:
		return VoiceChannelJoinType
	case *VoiceChannelJoinResponse:
		return VoiceChannelJoinType
	}
	return 0
}
