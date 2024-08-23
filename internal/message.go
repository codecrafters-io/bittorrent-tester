package internal

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"

	logger "github.com/codecrafters-io/tester-utils/logger"
	"github.com/jackpal/bencode-go"
)

type messageID uint8
type Message struct {
    ID      messageID
    Payload []byte
}

const (
    HandshakeExtendedID uint8 = 0
    RequestMetadataExtensionMsgType uint8 = 0
    DataMetadataExtensionMsgType    uint8 = 1

    MsgBitfield messageID = 5
    MsgExtended messageID = 20
)

func sendBitfieldMessage(conn net.Conn, payload []byte, logger *logger.Logger) (err error) {
    defer logOnExit(logger, &err)

    logger.Debugln("Sending bitfield message")
    req := Message { ID: MsgBitfield, Payload: payload }
    serialized := req.Serialize()
    _, err = conn.Write(serialized)
    return err
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

func sendExtensionHandshake(conn net.Conn, metadataID uint8, metadataSize int, logger *logger.Logger) (err error) {
    defer logOnExit(logger, &err)

    logger.Debugln("Sending extension handshake")
    req := createExtensionHandshake(metadataID, metadataSize, logger)
    serialized := req.Serialize()
    _, err = conn.Write(serialized)
    return err
}

func createExtensionHandshake(metadataID uint8, metadataSize int, logger *logger.Logger) *Message {
    dict := make(map[string]interface{})
    inner := make(map[string]int64)
    inner["ut_metadata"] = int64(metadataID)
    dict["m"] = inner
    dict["metadata_size"] = metadataSize
    var buf bytes.Buffer
    err := bencode.Marshal(&buf, dict)
    if err != nil {
        logger.Errorf("Error encoding: %v", err)
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

// Read parses a message from a stream. Returns `nil` on keep-alive message
func readMessage(r io.Reader) (*Message, error) {
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