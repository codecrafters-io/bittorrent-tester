package metafile

import (
	"fmt"
	"io"
	"os"

	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/bencode"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/internal"
)

const PieceHashLength int = 20

type Metafile struct {
	Filename     string
	Tracker      string
	Name         string
	Length       int
	PieceLength  int
	PiecesString string
	PiecesHash   []string
	HashedInfo   string
}

func NewMetafile(filename string) (*Metafile, error) {
	meta := &Metafile{
		Filename: filename,
	}

	handle, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("Unable to open file: %s\nErr: %s", filename, err.Error())
	}
	content, err := io.ReadAll(handle)
	if err != nil {
		return nil, fmt.Errorf("Unable to read file: %s\nErr: %s", filename, err.Error())
	}
	decodedContent, _, err := bencode.DecodeBencode(string(content))
	if err != nil {
		return nil, fmt.Errorf("Unable to parse file %s", filename)
	}

	parsedContent, ok := decodedContent.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Torrent file %s is invalid", filename)
	}

	parsedInfo, ok := parsedContent["info"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Torrent file %s info is invalid", filename)

	}

	meta.Tracker, ok = parsedContent["announce"].(string)
	meta.Name, ok = parsedInfo["name"].(string)
	meta.Length, ok = parsedInfo["length"].(int)
	meta.PieceLength, ok = parsedInfo["piece length"].(int)
	meta.PiecesString, ok = parsedInfo["pieces"].(string)
	meta.HashedInfo = internal.HashInfo(parsedInfo)

	if !ok {
		return nil, fmt.Errorf("Torrent file %s info is invalid", filename)
	}

	meta.parsePieces()

	return meta, nil
}

func (m *Metafile) parsePieces() {
	pieces := []byte(m.PiecesString)
	for i := 0; i < len(pieces); i += PieceHashLength {
		m.PiecesHash = append(
			m.PiecesHash,
			fmt.Sprintf("%x\n", pieces[i:i+PieceHashLength]),
		)
	}
}
