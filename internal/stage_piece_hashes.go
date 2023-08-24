package internal

import (
	"fmt"
	"math/rand"
	"path"
	"strings"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

func testPieceHashes(stageHarness *tester_utils.StageHarness) error {
	initRandom()

	logger := stageHarness.Logger
	executable := stageHarness.Executable

	torrentFilename := "test.torrent"
	tempDir, err := createTempDir(executable)
	if err != nil {
		logger.Errorf("Couldn't create temp directory")
		return err
	}

	pieceHashes, err := createPieceHashes()
	if err != nil {
		logger.Errorf("error creating piece hashes")
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
	expectedInfoHash, err := torrent.writeToFile(destinationPath)
	if err != nil {
		return err
	}

	logger.Infof("Running ./your_bittorrent.sh info %s", torrentFilename)
	result, err := executable.Run("info", torrentFilename)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	output := []string{
		fmt.Sprintf("Tracker URL: %s", trackerURL),
		fmt.Sprintf("Length: %d", fileLengthBytes),
		fmt.Sprintf("Info Hash: %x", expectedInfoHash),
		fmt.Sprintf("Piece Length: %d", pieceLengthBytes),
		"Piece Hashes:"}
	output = append(output, pieceHashes...)

	expected := strings.Join(output, "\n") + "\n"

	if err = assertStdout(result, expected); err != nil {
		return err
	}

	return nil
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
