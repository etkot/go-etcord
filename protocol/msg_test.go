package protocol

import (
	"bytes"
	"testing"

	"etcord/common"
)

func TestSerialize(t *testing.T) {
	m := &Error{
		Code:    123,
		Message: "oh woops i dropped my monster condom that i use for my magnum dong",
	}
	b, err := Serialize(m)
	if err != nil {
		t.Fatal(err)
	}

	m_, err := Deserialize(common.NewBuffer(b))
	if err != nil {
		t.Fatal(err)
	}
	b_, err := Serialize(m_)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(b, b_) != 0 {
		t.Fatal("serialized packets do not match")
	}
}
