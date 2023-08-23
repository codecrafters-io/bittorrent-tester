package internal

import (
	"path"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

// TODO: Use private tracker and own torrents
func testDownloadPiece(stageHarness *tester_utils.StageHarness) error {
	torrentFile := "test.torrent"
	pieceIndex := "0"
	expectedFilename := "test-iso-piece-0"
	expectedSha1 := "ddf33172599fda84f0a209a3034f79f0b8aa5e22"
	expectedFileSize := int64(262144)

	logger := stageHarness.Logger
	executable := stageHarness.Executable

	tempDir, err := createTempDir(executable)
	if err != nil {
		logger.Errorf("Couldn't create temp directory")
		return err
	}

	if err := copyTorrent(tempDir, torrentFile); err != nil {
		logger.Errorf("Couldn't copy torrent file", err)
		return err
	}

	logger.Debugf("Running ./your_bittorrent.sh download_piece -o %s %s %s", expectedFilename, torrentFile, pieceIndex)
	result, err := executable.Run("download_piece", "-o", expectedFilename, torrentFile, pieceIndex)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	downloadedFilePath := path.Join(tempDir, expectedFilename)
	if err = assertFileSize(downloadedFilePath, expectedFileSize); err != nil {
		return err
	}

	if err = assertFileSHA1(downloadedFilePath, expectedSha1); err != nil {
		return err
	}

	return nil
}
