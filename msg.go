package main

import (
	"bytes"
	"encoding/binary"
	"etcord/common"
	"fmt"
)

type MsgType int

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
	//Serialize() ([]byte, error)
	Deserialize(common.Buffer) error
	//String() string
}

type Error struct {
	Code    int16  `json:"id"`
	Message string `json:"message"`
}

type LoginRequest struct {
	Name string `json:"name"`
}

type GetClientsRequest struct {
	Type      uint8    `json:"type"`
	ClientID  uint16   `json:"clientId,omitempty"`
	Count     uint16   `json:"count,omitempty"`
	ClientIDs []uint16 `json:"clientIds,omitempty"`
}

type GetClientsResponse struct {
	Count   uint16   `json:"count"`
	Clients []Client `json:"clients"`
}

type GetChannelsRequest struct {
}

type GetChannelsResponse struct {
	Count    uint16    `json:"count"`
	Channels []Channel `json:"channels"`
}

type GetChatHistoryRequest struct {
	ChannelID uint16 `json:"channelId"`
	Count uint16 `json:"count"`
	OffsetID uint16 `json:"offsetId"`
}

type GetChatHistoryResponse struct {
	ChannelID uint16 `json:"channelId"`
	Count uint16 `json:"count"`
	Messages []Message `json:"messages"`
}

type ChatMessageRequest struct {
	ChannelID uint16 `json:"channelId"`
	Content string `json:"content"`
}

type ChatMessageResponse struct {
	ChannelID uint16 `json:"channelId"`
	Message Message `json:"message"`
}

type VoiceChannelJoinRequest struct {
	ChannelID uint16 `json:"channelId"`
}

type VoiceChannelJoinResponse struct {
	ChannelID uint16 `json:"channelId"`
	ClientID uint16 `json:"clientId"`
}

const (
	GetClientsAll = 0
	GetClientsOne = 1
	GetClientsMany = 2
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

// Deserializes deserializes a raw packet to an Etcord protocol message
func Deserialize(tmp *bytes.Buffer) (Msg, error) {
	lb := tmp.Next(2)
	if len(lb) != 2 {
		return nil, fmt.Errorf("invalid packet length")
	}
	msgLen := binary.BigEndian.Uint16(lb)

	buf := common.Buffer{bytes.NewBuffer(tmp.Next(int(msgLen)))}

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

func (m *Error) Deserialize(buf common.Buffer) error {
	var err error
	if m.Code, err = buf.ReadInt16(); err != nil {
		return err
	}
	m.Message = buf.String()
	return nil
}

func (m *LoginRequest) Deserialize(buf common.Buffer) error {
	m.Name = buf.String()
	if len(m.Name) == 0 {
		return fmt.Errorf("empty name")
	}
	return nil
}

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

func (m *GetChannelsRequest) Deserialize(common.Buffer) error { return nil }

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

func (m *VoiceChannelJoinRequest) Deserialize(buf common.Buffer) error {
	var err error
	if m.ChannelID, err = buf.ReadUint16(); err != nil {
		return err
	}
	return nil
}

