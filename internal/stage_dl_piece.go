package internal

import (
	"fmt"
	"math/rand"
	"path"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

type DownloadPieceTest struct {
	torrentFilename string
	pieceIndex      int
	pieceHash       string
	pieceLength     int64
}

var downloadPieceTests = []DownloadPieceTest{
	{
		torrentFilename: "congratulations.gif.torrent",
		pieceIndex:      3,
		pieceHash:       "bded68d02de011a2b687f75b5833f46cce8e3e9c",
		pieceLength:     34460,
	},
	{
		torrentFilename: "itsworking.gif.torrent",
		pieceIndex:      1,
		pieceHash:       "838f703cf7f6f08d1c497ed390df78f90d5f7566",
		pieceLength:     262144,
	},
	{
		torrentFilename: "codercat.gif.torrent",
		pieceIndex:      0,
		pieceHash:       "3c34309faebf01e49c0f63c90b7edcc2259b6ad0",
		pieceLength:     262144,
	},
}

func testDownloadPiece(stageHarness *tester_utils.StageHarness) error {
	initRandom()

	randomIndex := rand.Intn(len(downloadPieceTests))
	t := downloadPieceTests[randomIndex]

	logger := stageHarness.Logger
	executable := stageHarness.Executable

	tempDir, err := createTempDir(executable)
	if err != nil {
		logger.Errorf("Couldn't create temp directory")
		return err
	}

	if err := copyTorrent(tempDir, t.torrentFilename); err != nil {
		logger.Errorf("Couldn't copy torrent file", err)
		return err
	}

	expectedFilename := fmt.Sprintf("piece-%d", t.pieceIndex)
	logger.Debugf("Running ./your_bittorrent.sh download_piece -o %s %s %d", expectedFilename, t.torrentFilename, t.pieceIndex)
	result, err := executable.Run("download_piece", "-o", expectedFilename, t.torrentFilename, fmt.Sprintf("%d", t.pieceIndex))
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	downloadedFilePath := path.Join(tempDir, expectedFilename)
	if err = assertFileSize(downloadedFilePath, t.pieceLength); err != nil {
		return err
	}

	if err = assertFileSHA1(downloadedFilePath, t.pieceHash); err != nil {
		return err
	}

	return nil
}
