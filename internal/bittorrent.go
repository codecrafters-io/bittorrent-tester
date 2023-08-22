package internal

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/jackpal/bencode-go"
)

type Handshake struct {
	ProtocolStr string
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

func readHandshake(r io.Reader) (*Handshake, error) {
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

	if protocolNameLength == 0 {
		return nil, err
	}

	handshakeBuffer := make([]byte, protocolNameLength+8+20+20)
	_, err = io.ReadFull(r, handshakeBuffer)
	if err != nil {
		return nil, err
	}

	var infoHash, peerID [20]byte
	copy(infoHash[:], handshakeBuffer[protocolNameLength+8:protocolNameLength+8+20])
	copy(peerID[:], handshakeBuffer[protocolNameLength+8+20:])

	handshake := Handshake{
		ProtocolStr: string(handshakeBuffer[0:protocolNameLength]),
		InfoHash:    infoHash,
		PeerID:      peerID,
	}

	return &handshake, nil
}

func sendHandshake(conn net.Conn, infoHash [20]byte, peerID [20]byte) {
	handshake := []byte{19}                                         // Protocol name length
	handshake = append(handshake, []byte("BitTorrent protocol")...) // Protocol name
	handshake = append(handshake, make([]byte, 8)...)               // Reserved bytes
	handshake = append(handshake, infoHash[:]...)                   // Info hash
	handshake = append(handshake, peerID[:]...)                     // Peer ID

	_, err := conn.Write(handshake)
	if err != nil {
		log.Fatal("Error sending handshake:", err)
	}
}
