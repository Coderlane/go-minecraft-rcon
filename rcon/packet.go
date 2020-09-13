package rcon

import (
	"encoding/binary"
	"fmt"
	"io"
)

type packetType int32

const (
	// PacketTypeAuth is sent by the client to the server during initial auth.
	PacketTypeAuth packetType = 3
	// AuthResponsePacket is sent by the server to the client during initial auth.
	AuthResponsePacket packetType = 2
	// PacketTypeData is sent by the client to the server for standard messages.
	PacketTypeData packetType = 2
	// PacketTypeDataResponse is sent by the server to the client for replies to
	// standard messages.
	PacketTypeDataResponse packetType = 0
)

const (
	packetSizeBase int32 = 10

	// PacketMaxSize is the maximum size of the body of a message.
	PacketMaxSize int32 = 4096
	// PacketIDInvalid is the packet ID used for invalid responses.
	PacketIDInvalid int32 = -1
)

// PacketHeader contains the constant-size portion of the packet.
type PacketHeader struct {
	ID   int32
	Type packetType
}

// Packet represents a raw RCON packet
type Packet struct {
	Header PacketHeader
	Body   string
}

func (pkt Packet) size() int32 {
	return packetSizeBase + int32(len(pkt.Body))
}

// EncodeBinary encodes a packet into its wire format.
func (pkt Packet) EncodeBinary(writer io.Writer) error {
	if err := binary.Write(writer, binary.LittleEndian, pkt.size()); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.LittleEndian, pkt.Header); err != nil {
		return err
	}
	body := append([]byte(pkt.Body), make([]byte, 2)...)
	if err := binary.Write(writer, binary.LittleEndian, body); err != nil {
		return err
	}
	return nil
}

// DecodeBinary decodes a packet from its wire format.
func (pkt *Packet) DecodeBinary(reader io.Reader) error {
	var size int32
	if err := binary.Read(reader, binary.LittleEndian, &size); err != nil {
		return err
	}
	if size < packetSizeBase {
		return fmt.Errorf("packet size: %d smaller than minimum: %d",
			size, packetSizeBase)
	}
	if size > packetSizeBase+PacketMaxSize {
		return fmt.Errorf("packet size: %d greater than maximum: %d",
			size, packetSizeBase)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pkt.Header); err != nil {
		return err
	}
	body := make([]byte, size-packetSizeBase)
	if err := binary.Read(reader, binary.LittleEndian, &body); err != nil {
		return err
	}
	var pad int16
	if err := binary.Read(reader, binary.LittleEndian, &pad); err != nil {
		return err
	}
	pkt.Body = string(body)
	return nil
}
