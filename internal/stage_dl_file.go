package internal

import (
	"path"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

func testDownloadFile(stageHarness *tester_utils.StageHarness) error {
	torrentFile := "alpine-minirootfs-3.18.3-aarch64.tar.gz.torrent"
	expectedFilename := "alpine-minirootfs-3.18.3-aarch64.tar.gz"
	expectedSha1 := "e3fae2149ee08c75ced032561ea8ee5bd9b01d14"
	expectedFileSize := int64(3201634)

	logger := stageHarness.Logger
	executable := stageHarness.Executable

	tempDir, err := createTempDir(executable)
	if err != nil {
		logger.Errorf("Couldn't create temp directory")
		return err
	}

	if err := copyTorrent(tempDir, torrentFile); err != nil {
		logger.Errorf("Couldn't copy torrent file")
		return err
	}

	logger.Debugf("Running ./your_bittorrent.sh download -o %s %s", expectedFilename, torrentFile)
	result, err := executable.Run("download", "-o", expectedFilename, torrentFile)
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
