package internal

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	logger "github.com/codecrafters-io/tester-utils/logger"
	"github.com/jackpal/bencode-go"
)

type Handshake struct {
	ProtocolStr string
	Reserved    [8]byte
	InfoHash    [20]byte
	PeerID      [20]byte
}

type TorrentFile struct {
	Announce string          `bencode:"announce"`
	Info     TorrentFileInfo `bencode:"info"`
}

type TorrentFileInfo struct {
	Name        string `bencode:"name"`
	Length      int    `bencode:"length"`
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	NameUtf8    string `bencode:"name.utf-8,omitempty"`
	Private     int    `bencode:"private,omitempty"`
	Source      string `bencode:"source,omitempty"`
}

const ProtocolName = "BitTorrent protocol"

func (i *TorrentFileInfo) hash() ([20]byte, error) {
	var buffer bytes.Buffer
	if err := bencode.Marshal(&buffer, *i); err != nil {
		return [20]byte{}, err
	}
	hash := sha1.Sum(buffer.Bytes())
	return hash, nil
}

func (torrent *TorrentFile) writeToFile(outputPath string) ([20]byte, error) {
	torrentFile, err := os.Create(outputPath)
	if err != nil {
		return [20]byte{}, err
	}
	defer torrentFile.Close()

	err = bencode.Marshal(torrentFile, *torrent)
	if err != nil {
		return [20]byte{}, err
	}

	infoHash, err := torrent.Info.hash()
	if err != nil {
		return [20]byte{}, err
	}
	return infoHash, nil
}

func toPiecesStr(pieces []string) string {
	piecesByteArray := make([]byte, 20*len(pieces))
	for i, piece := range pieces {
		hex.Decode(piecesByteArray[i*20:(i+1)*20], []byte(piece))
	}
	return string(piecesByteArray)
}

func createPiecesStrFromFile(filePath string, pieceLengthBytes int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buffer := make([]byte, pieceLengthBytes)
	hasher := sha1.New()
	pieceIndex := 0
	var pieceHashes []string
	for {
		bytesRead, err := file.Read(buffer)
		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}
		_, err = hasher.Write(buffer[:bytesRead])
		if err != nil {
			return "", err
		}
		if bytesRead == pieceLengthBytes {
			hashSum := hasher.Sum(nil)
			pieceHashes = append(pieceHashes, string(hashSum))
			hasher.Reset()
			pieceIndex++
		}
	}
	// Handle the last piece if it's less than piece size
	if hasher.Size() > 0 {
		hashSum := hasher.Sum(nil)
		pieceHashes = append(pieceHashes, string(hashSum))
	}
	return strings.Join(pieceHashes, ""), nil
}

func readHandshake(r io.Reader, logger *logger.Logger) (*Handshake, error) {
	// Handshake message contents:
	// 1 byte protocol string length
	// x byte protocol string
	// 8 byte reserved
	// 20 byte info hash
	// 20 byte peer id
	lengthBuffer := make([]byte, 1)
	_, err := io.ReadFull(r, lengthBuffer)
	if err != nil {
		return nil, err
	}
	protocolNameLength := int(lengthBuffer[0])
	expectedProtocolNameLength := len(ProtocolName)

	if protocolNameLength != expectedProtocolNameLength {
		if protocolNameLength == 49 {
			logger.Errorln("Are you sending 19 as a string? 19 should be encoded as an 8 bit unsigned number (1 byte), not as ASCII-encoded characters")
		}
		return nil, fmt.Errorf("first byte of handshake needs to be protocol string length, expected value %d but received: %d", expectedProtocolNameLength, protocolNameLength)
	}

	handshakeBuffer := make([]byte, protocolNameLength+8+20+20)
	_, err = io.ReadFull(r, handshakeBuffer)
	if err != nil {
		if err == io.ErrUnexpectedEOF {
			logger.Errorln("Your handshake message might be shorter than expected. Expected to read 68 bytes (1 byte for protocol length, 19 bytes for protocol string, 8 bytes for reserved, 20 for info hash, 20 for peer id) but no more input was available.")
		}
		return nil, err
	}

	protocolStr := string(handshakeBuffer[0:protocolNameLength])
	if protocolStr != ProtocolName {
		return nil, fmt.Errorf("unknown protocol name, expected: %s, actual: %s", ProtocolName, protocolStr)
	}

	expectedReservedBytes := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	actualReservedBytes := handshakeBuffer[protocolNameLength : protocolNameLength+8]
	if !bytes.Equal(expectedReservedBytes, actualReservedBytes) {
		logger.Infof("Did you send reserved bytes? expected bytes: %v but received: %v\n", expectedReservedBytes, actualReservedBytes)
	}

	var reservedBytes [8]byte
	copy(reservedBytes[:], handshakeBuffer[protocolNameLength : protocolNameLength+8])
	var infoHash, peerID [20]byte
	copy(infoHash[:], handshakeBuffer[protocolNameLength+8:protocolNameLength+8+20])
	copy(peerID[:], handshakeBuffer[protocolNameLength+8+20:])

	handshake := Handshake{
		ProtocolStr: string(handshakeBuffer[0:protocolNameLength]),
		Reserved:    reservedBytes,
		InfoHash:    infoHash,
		PeerID:      peerID,
	}

	return &handshake, nil
}

func sendHandshake(conn net.Conn, reserved [8]byte, infoHash [20]byte, peerID [20]byte) error {
	handshake := []byte{19}                                // Protocol name length
	handshake = append(handshake, []byte(ProtocolName)...) // Protocol name
	handshake = append(handshake, reserved[:]...)          // Reserved bytes
	handshake = append(handshake, infoHash[:]...)          // Info hash
	handshake = append(handshake, peerID[:]...)            // Peer ID

	_, err := conn.Write(handshake)
	if err != nil {
		return fmt.Errorf("error sending handshake: %v", err)
	}
	return nil
}
