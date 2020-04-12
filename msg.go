package main

import (
	"fmt"
	"math"

	"etcord/common"
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

type Msg interface {
	Serialize() []byte
	Deserialize(common.Buffer) error
	//String() string
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
func Serialize(m Msg) ([]byte, error) {
	mb := m.Serialize()
	if len(mb) > math.MaxUint16 {
		return nil, fmt.Errorf("packet content length overflows uint16")
	}

	buf := common.NewBuffer(make([]byte, 0, PacketHeaderLen+len(mb)))
	buf.WriteUint16(1 + uint16(len(mb)))
	buf.Write([]byte{getPacketID(m)})
	buf.Write(mb)

	return buf.Bytes(), nil
}

// Deserializes deserializes a raw packet to an Etcord protocol message
func Deserialize(tmp common.Buffer) (Msg, error) {
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

	// TODO add response deserialization?
	var m Msg
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
	buf.Write([]byte(m.Message))
	return buf.Bytes()
}

func (m *Error) Deserialize(buf common.Buffer) error {
	var err error
	if m.Code, err = buf.ReadUint16(); err != nil {
		return err
	}
	m.Message = buf.String()
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
	m.Name = buf.String()
	if len(m.Name) == 0 {
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
	Count   uint16   `json:"count"`
	Clients []Client `json:"clients"`
}

func (m *GetClientsResponse) Serialize() []byte { return nil } // TODO

func (m *GetClientsResponse) Deserialize(common.Buffer) error { return nil }

type GetChannelsRequest struct {
}

func (m *GetChannelsRequest) Serialize() []byte { return nil } // TODO

func (m *GetChannelsRequest) Deserialize(common.Buffer) error { return nil }

type GetChannelsResponse struct {
	Count    uint16    `json:"count"`
	Channels []Channel `json:"channels"`
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
	ChannelID uint16    `json:"channelId"`
	Count     uint16    `json:"count"`
	Messages  []Message `json:"messages"`
}

func (m *GetChatHistoryResponse) Serialize() []byte { return nil } // TODO

func (m *GetChatHistoryResponse) Deserialize(common.Buffer) error { return nil } // TODO

type ChatMessageRequest struct {
	ChannelID uint16 `json:"channelId"`
	Content   string `json:"content"`
}

func (m *ChatMessageRequest) Serialize() []byte { return nil } // TODO

func (m *ChatMessageRequest) Deserialize(buf common.Buffer) error {
	var err error
	if m.ChannelID, err = buf.ReadUint16(); err != nil {
		return err
	}
	m.Content = buf.String()
	if len(m.Content) == 0 {
		return fmt.Errorf("empty message")
	}
	return nil
}

type ChatMessageResponse struct {
	ChannelID uint16  `json:"channelId"`
	Message   Message `json:"message"`
}

func (m *ChatMessageResponse) Serialize() []byte { return nil } // TODO

func (m *ChatMessageResponse) Deserialize(common.Buffer) error { return nil }

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

func getPacketID(m Msg) uint8 {
	switch m.(type) {
	case *Error:
		return uint8(ErrorType)
	case *LoginRequest:
		return uint8(LoginType)
	case *GetClientsRequest:
		return uint8(GetClientsType)
	case *GetClientsResponse:
		return uint8(GetClientsType)
	case *GetChannelsRequest:
		return uint8(GetChannelsType)
	case *GetChannelsResponse:
		return uint8(GetChannelsType)
	case *GetChatHistoryRequest:
		return uint8(GetChatHistoryType)
	case *GetChatHistoryResponse:
		return uint8(GetChatHistoryType)
	case *ChatMessageRequest:
		return uint8(ChatMessageType)
	case *ChatMessageResponse:
		return uint8(ChatMessageType)
	case *VoiceChannelJoinRequest:
		return uint8(VoiceChannelJoinType)
	case *VoiceChannelJoinResponse:
		return uint8(VoiceChannelJoinType)
	}
	return 0
}
