package message

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/codecrafters-io/grep-starter-go/bencode"
)

type messageID uint8

const (
	// MsgChoke chokes the receiver
	MsgChoke messageID = 0
	// MsgUnchoke unchokes the receiver
	MsgUnchoke messageID = 1
	// MsgInterested expresses interest in receiving data
	MsgInterested messageID = 2
	// MsgNotInterested expresses disinterest in receiving data
	MsgNotInterested messageID = 3
	// MsgHave alerts the receiver that the sender has downloaded a piece
	MsgHave messageID = 4
	// MsgBitfield encodes which pieces that the sender has downloaded
	MsgBitfield messageID = 5
	// MsgRequest requests a block of data from the receiver
	MsgRequest messageID = 6
	// MsgPiece delivers a block of data to fulfill a request
	MsgPiece messageID = 7
	// MsgCancel cancels a request
	MsgCancel messageID = 8
	// Extension messages (BEP 10)
	MsgExtended messageID = 20
)

// Message stores ID and payload of a message
type Message struct {
	ID      messageID
	Payload []byte
}

type bencodeMetadataExtensionMsg struct {
	Piece     int   `bencode:"piece"`
	TotalSize int   `bencode:"total_size,omitempty"`
	Type      uint8 `bencode:"msg_type"`
}

const (
	HandshakeExtendedID uint8 = 0

	RequestMetadataExtensionMsgType uint8 = 0
	DataMetadataExtensionMsgType    uint8 = 1
	RejectMetadataExtensionMsgType  uint8 = 2
)

// FormatRequest creates a REQUEST message
func FormatRequest(index, begin, length int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &Message{ID: MsgRequest, Payload: payload}
}

// FormatHave creates a HAVE message
func FormatHave(index int) *Message {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(index))
	return &Message{ID: MsgHave, Payload: payload}
}

// ParsePiece parses a PIECE message and copies its payload into a buffer
func ParsePiece(index int, buf []byte, msg *Message) (int, error) {
	if msg.ID != MsgPiece {
		return 0, fmt.Errorf("Expected PIECE (ID %d), got ID %d", MsgPiece, msg.ID)
	}
	if len(msg.Payload) < 8 {
		return 0, fmt.Errorf("Payload too short. %d < 8", len(msg.Payload))
	}
	parsedIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if parsedIndex != index {
		return 0, fmt.Errorf("Expected index %d, got %d", index, parsedIndex)
	}
	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if begin >= len(buf) {
		return 0, fmt.Errorf("Begin offset too high. %d >= %d", begin, len(buf))
	}
	data := msg.Payload[8:]
	if begin+len(data) > len(buf) {
		return 0, fmt.Errorf("Data too long [%d] for offset %d with length %d", len(data), begin, len(buf))
	}
	//fmt.Printf("received piece beginning at %d for piece %d\n", begin, index)
	copy(buf[begin:], data)
	return len(data), nil
}

// ParseHave parses a HAVE message
func ParseHave(msg *Message) (int, error) {
	if msg.ID != MsgHave {
		return 0, fmt.Errorf("Expected HAVE (ID %d), got ID %d", MsgHave, msg.ID)
	}
	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("Expected payload length 4, got length %d", len(msg.Payload))
	}
	index := int(binary.BigEndian.Uint32(msg.Payload))
	return index, nil
}

// Serialize serializes a message into a buffer of the form
// <length prefix><message ID><payload>
// Interprets `nil` as a keep-alive message
func (m *Message) Serialize() []byte {
	if m == nil {
		return make([]byte, 4)
	}
	length := uint32(len(m.Payload) + 1) // +1 for id
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)
	return buf
}

// Read parses a message from a stream. Returns `nil` on keep-alive message
func Read(r io.Reader) (*Message, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuf)

	// keep-alive message
	if length == 0 {
		return nil, nil
	}

	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}

	m := Message{
		ID:      messageID(messageBuf[0]),
		Payload: messageBuf[1:],
	}

	return &m, nil
}

func (m *Message) name() string {
	if m == nil {
		return "KeepAlive"
	}
	switch m.ID {
	case MsgChoke:
		return "Choke"
	case MsgUnchoke:
		return "Unchoke"
	case MsgInterested:
		return "Interested"
	case MsgNotInterested:
		return "NotInterested"
	case MsgHave:
		return "Have"
	case MsgBitfield:
		return "Bitfield"
	case MsgRequest:
		return "Request"
	case MsgPiece:
		return "Piece"
	case MsgCancel:
		return "Cancel"
	default:
		return fmt.Sprintf("Unknown#%d", m.ID)
	}
}

func (m *Message) String() string {
	if m == nil {
		return m.name()
	}
	return fmt.Sprintf("%s [%d]", m.name(), len(m.Payload))
}

func FormatExtensionHandshake() *Message {
	dict := make(map[string]interface{})
	inner := make(map[string]int64)
	inner["ut_metadata"] = 9 // random value
	dict["m"] = inner
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, dict)
	if err != nil {
		fmt.Println("Error encoding", err)
	}
	payload := formatExtendedPayload(buf, HandshakeExtendedID)
	return &Message{ID: MsgExtended, Payload: payload}
}

func formatExtendedPayload(buf bytes.Buffer, extensionId uint8) []byte {
	payload := make([]byte, 1+buf.Len())
	payload[0] = uint8(extensionId)
	copy(payload[1:], buf.Bytes())
	return payload
}

func (m *Message) FindMetadataPayloadIndex() int {
	reader := bytes.NewReader(m.Payload[1:])
	torr, _ := bencode.Decode(reader)
	var buf bytes.Buffer
	inner := torr.(map[string]interface{})
	err := bencode.Marshal(&buf, bencodeMetadataExtensionMsg{
		Piece:     int(inner["piece"].(int64)),
		Type:      uint8(inner["msg_type"].(int64)),
		TotalSize: int(inner["total_size"].(int64)),
	})
	if err != nil {
		fmt.Println("Err", err)
	}
	return buf.Len()
}

func FormatMetadataExtensionMessage(extensionID uint8, msgType uint8, piece int) *Message {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, bencodeMetadataExtensionMsg{
		Piece: piece,
		Type:  msgType,
	})
	if err != nil {
		fmt.Println("Error encoding:", err)
	}
	payload := formatExtendedPayload(buf, extensionID)
	fmt.Println("extended message payload", string(payload))
	return &Message{ID: MsgExtended, Payload: payload}
}
