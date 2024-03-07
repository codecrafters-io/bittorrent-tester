package internal

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"strings"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

type DownloadPieceTest struct {
	torrentFilename string
	pieceIndex      int
	pieceHash       string
	pieceLength     int64
}

var downloadPieceTests = [][]DownloadPieceTest{
	{
		{
			torrentFilename: "congratulations.gif.torrent",
			pieceIndex:      2,
			pieceHash:       "76869e6c9c1f101f94f39de153e468be6a638f4f",
			pieceLength:     262144,
		},
		{
			torrentFilename: "congratulations.gif.torrent",
			pieceIndex:      3,
			pieceHash:       "bded68d02de011a2b687f75b5833f46cce8e3e9c",
			pieceLength:     34460,
		},
	},
	{

		{
			torrentFilename: "itsworking.gif.torrent",
			pieceIndex:      9,
			pieceHash:       "7affc94f0985b985eb888a36ec92652821a21be4",
			pieceLength:     190404,
		},
		{
			torrentFilename: "itsworking.gif.torrent",
			pieceIndex:      1,
			pieceHash:       "838f703cf7f6f08d1c497ed390df78f90d5f7566",
			pieceLength:     262144,
		},
	},
	{

		{
			torrentFilename: "codercat.gif.torrent",
			pieceIndex:      0,
			pieceHash:       "3c34309faebf01e49c0f63c90b7edcc2259b6ad0",
			pieceLength:     262144,
		},
		{
			torrentFilename: "codercat.gif.torrent",
			pieceIndex:      11,
			pieceHash:       "3d8db9e34db63b4ba1be27930911aa37b3f997dd",
			pieceLength:     110536,
		},
	},
}

func testDownloadPiece(stageHarness *test_case_harness.TestCaseHarness) error {
	initRandom()

	randomIndex := rand.Intn(len(downloadPieceTests))
	tests := downloadPieceTests[randomIndex]

	logger := stageHarness.Logger
	executable := stageHarness.Executable

	tempDir, err := os.MkdirTemp("", "torrents")
	if err != nil {
		logger.Errorln("Couldn't create temp directory")
		return err
	}

	if err := copyTorrent(tempDir, tests[0].torrentFilename); err != nil {
		logger.Errorf("Couldn't copy torrent file: %s", err)
		return err
	}

	for _, t := range tests {
		torrentFilePath := path.Join(tempDir, t.torrentFilename)
		expectedFilename := fmt.Sprintf("piece-%d", t.pieceIndex)
		downloadedFilePath := path.Join(tempDir, expectedFilename)

		logger.Infof("Running ./your_bittorrent.sh download_piece -o %s %s %d", downloadedFilePath, torrentFilePath, t.pieceIndex)
		result, err := executable.Run("download_piece", "-o", downloadedFilePath, torrentFilePath, fmt.Sprintf("%d", t.pieceIndex))
		resultStr := string(result.Stdout)

		if strings.Contains(resultStr, "Connection reset by peer") || strings.Contains(resultStr, "EOF") {
			logger.Infoln("Connection reset by peer or EOF type of errors can be transient issues (try again), but they can also happen when peers receive unexpected messages (sending a REQUEST message with an incorrect offset/length, block size larger than 16 kb etc.)")
			if t.pieceLength != 262144 {
				logger.Infof("Last piece of the file can be less than piece length specified in torrent metadata. For instance, the length of this piece is %d", t.pieceLength)
			}
		}

		if err != nil {
			return err
		}

		if err = assertExitCode(result, 0); err != nil {
			return err
		}

		if err = assertFileSize(downloadedFilePath, t.pieceLength); err != nil {
			return err
		}

		if err = assertFileSHA1(downloadedFilePath, t.pieceHash); err != nil {
			return err
		}
	}

	return nil
}
