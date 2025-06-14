package parser

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/codecrafters-io/grep-starter-go/bencode"
	"github.com/codecrafters-io/grep-starter-go/torrent"
)

/*
type bencodeFileInfo struct {
	Length   int64    `bencode:"length"` // BEP3
	Path     []string `bencode:"path"`   // BEP3
	PathUtf8 []string `bencode:"path.utf-8,omitempty"`
}
*/

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length,omitempty"`
	Name        string `bencode:"name"`
	NameUtf8    string `bencode:"name.utf-8,omitempty"`
	Private     int    `bencode:"private,omitempty"`
	Source      string `bencode:"source,omitempty"`
	//Files       []bencodeFileInfo `bencode:"files,omitempty"` // BEP3, mutually exclusive with Length
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

type BencodeTrackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func (torrent *bencodeTorrent) writeToFile(outputPath string) ([20]byte, error) {
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

// Open parses a torrent file
func Open(path string) (torrent.TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return torrent.TorrentFile{}, err
	}
	defer file.Close()

	bto := bencodeTorrent{}
	err = bencode.Unmarshal(file, &bto)
	if err != nil {
		return torrent.TorrentFile{}, err
	}
	return bto.toTorrentFile()
}

func (i *bencodeInfo) hash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *i)
	if err != nil {
		return [20]byte{}, err
	}
	h := sha1.Sum(buf.Bytes())
	return h, nil
}

func (i *bencodeInfo) splitPieceHashes() ([][20]byte, error) {
	hashLen := 20 // Length of SHA-1 hash
	buf := []byte(i.Pieces)
	if len(buf)%hashLen != 0 {
		err := fmt.Errorf("Received malformed pieces of length %d", len(buf))
		return nil, err
	}
	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}
	return hashes, nil
}

func (bto *bencodeTorrent) toTorrentFile() (torrent.TorrentFile, error) {
	infoHash, err := bto.Info.hash()
	if err != nil {
		return torrent.TorrentFile{}, err
	}
	pieceHashes, err := bto.Info.splitPieceHashes()
	if err != nil {
		return torrent.TorrentFile{}, err
	}
	t := torrent.TorrentFile{
		Announce:    bto.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bto.Info.PieceLength,
		Length:      bto.Info.Length,
		Name:        bto.Info.Name,
	}
	return t, nil
}

func DecodeInfoHash(infoHashStr string) ([20]byte, error) {
	var infoHash [20]byte
	decodedBytes, err := hex.DecodeString(infoHashStr)
	if err != nil {
		return infoHash, err
	}
	copy(infoHash[:], decodedBytes)
	return infoHash, nil
}

func FromByteArray(data []byte, announceUrl string) (torrent.TorrentFile, error) {
	bi := bencodeInfo{}
	err := bencode.Unmarshal(bytes.NewReader(data), &bi)
	if err != nil {
		return torrent.TorrentFile{}, err
	}
	bto := bencodeTorrent{
		Announce: announceUrl,
		Info:     bi,
	}
	return bto.toTorrentFile()
}
