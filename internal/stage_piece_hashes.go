package internal

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strings"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

func testPieceHashes(stageHarness *tester_utils.StageHarness) error {
	initRandom()

	logger := stageHarness.Logger
	executable := stageHarness.Executable

	torrentFilename := "test.torrent"
	tempDir, err := os.MkdirTemp("", "torrents")
	if err != nil {
		logger.Errorln("Couldn't create temp directory")
		return err
	}

	pieceHashes, err := createPieceHashes()
	if err != nil {
		logger.Errorln("internal error creating piece hashes")
		return err
	}
	trackerURL := "http://bttracker.debian.org:6969/announce"
	pieceLengthBytes := 256 * 1024
	fileLengthBytes := pieceLengthBytes * len(pieceHashes)
	torrent := TorrentFile{
		Announce: trackerURL,
		Info: TorrentFileInfo{
			Name:        "faketorrent.iso",
			Length:      fileLengthBytes,
			Pieces:      toPiecesStr(pieceHashes),
			PieceLength: pieceLengthBytes,
		},
	}

	destinationPath := path.Join(tempDir, torrentFilename)
	_, err = torrent.writeToFile(destinationPath)
	if err != nil {
		return err
	}

	logger.Infof("Running ./your_bittorrent.sh info %s", destinationPath)
	result, err := executable.Run("info", destinationPath)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	expectedPieceLengthOutput := fmt.Sprintf("Piece Length: %d", pieceLengthBytes)
	if err = assertStdoutContains(result, expectedPieceLengthOutput); err != nil {
		return err
	}

	for _, pieceHash := range pieceHashes {
		if err = assertStdoutContains(result, pieceHash); err != nil {
			if strings.Contains(string(result.Stdout), hashWithoutLeadingZeros(pieceHash)) {
				logger.Errorln("Your piece hash value is shorter than 40 characters, it's missing a leading zero.")
			}
			return err
		}
	}

	return nil
}

func hashWithoutLeadingZeros(hexString string) string {
	bytes, decodeErr := hex.DecodeString(hexString)
	if decodeErr != nil {
		return ""
	}
	var withoutLeadingZeros string
	for _, b := range bytes {
		withoutLeadingZeros += fmt.Sprintf("%x", b)
	}
	return withoutLeadingZeros
}

func createPieceHashes() ([]string, error) {
	size := rand.Intn(7) + 2
	pieceHashes := make([]string, size)
	for i := 0; i < size; i++ {
		hash, err := randomHash()
		if err != nil {
			return []string{}, err
		}
		pieceHashes[i] = fmt.Sprintf("%x", hash)
	}
	return pieceHashes, nil
}
