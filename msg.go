package main

import (
	"bytes"
	"fmt"
)

type MsgType int

const (
	Error MsgType = iota
	GetClients
	GetChannels
	GetTextHistory

	ChatMessage MsgType = iota + 10

	VoiceChannelJoin MsgType = iota + 20
	VoiceChannelLeave
)

func (mt MsgType) String() string {
	switch mt {
	case Error:
		return "Error"
	case GetClients:
		return "GetClients"
	case GetChannels:
		return "GetChannels"
	case GetTextHistory:
		return "GetTextHistory"
	case ChatMessage:
		return "ChatMessage"
	case VoiceChannelJoin:
		return "VoiceChannelJoin"
	case VoiceChannelLeave:
		return "VoiceChannelLeave"
	}
	return ""
}

type Msg struct {
	Type MsgType
	Content []byte
}

func NewMsg() *Msg {
	return &Msg{}
}

func (m *Msg) String() string {
	return fmt.Sprintf("[%s][%s]", m.Type, string(m.Content))
}

func (m *Msg) Serialize() ([]byte, error) {
	return nil, nil
}

func (m *Msg) Deserialize(buf *bytes.Buffer) error {
	tb, err := buf.ReadByte()
	if err != nil {
		return err
	}
	m.Type = MsgType(tb)

	// TODO

	return nil
}