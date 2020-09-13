package rcon

import (
	"bytes"
	"encoding/binary"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	var buf bytes.Buffer

	inputPacket := Packet{
		Header: PacketHeader{
			ID:   1,
			Type: PacketTypeData,
		},
		Body: "test",
	}

	err := inputPacket.EncodeBinary(&buf)
	if err != nil {
		t.Fatal(err)
	}

	var outputPacket Packet
	err = outputPacket.DecodeBinary(&buf)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(inputPacket, outputPacket) {
		t.Errorf("Expected:\n%+v\nGot:\n%+v", inputPacket, outputPacket)
	}
}

func TestInvalidPackets(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})

	var pkt Packet
	err := pkt.DecodeBinary(buf)
	if err == nil {
		t.Error("Expected to fail to decode empty message.")
	} else if !strings.Contains(err.Error(), "EOF") {
		t.Error("Unexpected error: ", err)
	}

	size := int32(0)
	binary.Write(buf, binary.LittleEndian, size)
	err = pkt.DecodeBinary(buf)
	if err == nil {
		t.Error("Expected to fail to decode zero message.")
	} else if !strings.Contains(err.Error(), "minimum") {
		t.Error("Unexpected error: ", err)
	}

	size = int32(math.MaxInt32)
	binary.Write(buf, binary.LittleEndian, size)
	err = pkt.DecodeBinary(buf)
	if err == nil {
		t.Error("Expected to fail to decode exceptionally large message.")
	} else if !strings.Contains(err.Error(), "maximum") {
		t.Error("Unexpected error: ", err)
	}
}
