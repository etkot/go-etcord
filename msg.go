package main

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"etcord/types"
)

type MsgType int

const (
	ErrorType MsgType = iota
	LoginType
	GetClientsType
	GetChannelsType
	GetTextHistoryType

	ChatMessageType MsgType = iota + 10

	VoiceChannelJoinType MsgType = iota + 20
	VoiceChannelLeaveType
)

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
	case GetClientsType:
		return "GetClients"
	case GetChannelsType:
		return "GetChannels"
	case GetTextHistoryType:
		return "GetTextHistory"
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
// TODO errors
func Deserialize(tmp *bytes.Buffer) (types.Msg, error) {
	lb := tmp.Next(2)
	if len(lb) != 2 {
		return nil, fmt.Errorf("invalid packet length")
	}
	msgLen := binary.BigEndian.Uint16(lb)

	buf := bytes.NewBuffer(tmp.Next(int(msgLen)))

	tb, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	t := MsgType(tb)

	// TODO move to separate functions
	switch t {
	case ErrorType:
		m := &types.Error{}

		b := buf.Next(2)
		if len(b) != 2 {
			return nil, fmt.Errorf("invalid field length")
		}
		m.Code = int16(binary.BigEndian.Uint16(b))

		m.Message = buf.String()

		return m, nil

	case GetClientsType:
		m := &types.GetClientsRequest{}

		b, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		m.Type = b

		switch m.Type {
		case GetClientsAll:
			break

		case GetClientsOne:
			b := buf.Next(2)
			if len(b) != 2 {
				return nil, fmt.Errorf("invalid field length")
			}
			m.ClientID = binary.BigEndian.Uint16(b)

		case GetClientsMany:
			b := buf.Next(2)
			if len(b) != 2 {
				return nil, fmt.Errorf("invalid field length")
			}
			m.Count = binary.BigEndian.Uint16(b)
			for i := 0; i < int(m.Count); i++ {
				b := buf.Next(2)
				if len(b) != 2 {
					return nil, fmt.Errorf("invalid field length")
				}
				id := binary.BigEndian.Uint16(b)
				m.ClientIDs = append(m.ClientIDs, id)
			}
		}

		return m, nil
	}
	// TODO rest of the messages

	return nil, fmt.Errorf("unknown type")
}