package internal

import (
	"path"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

func testDownloadFile(stageHarness *tester_utils.StageHarness) error {
	initRandom()

	logger := stageHarness.Logger
	executable := stageHarness.Executable

	t := randomTorrent()

	// TODO: Remove, don't change working directory
	tempDir, err := createTempDir(executable)
	if err != nil {
		logger.Errorf("Couldn't create temp directory")
		return err
	}

	if err := copyTorrent(tempDir, t.filename); err != nil {
		logger.Errorf("Couldn't copy torrent file")
		return err
	}

	logger.Infof("Running ./your_bittorrent.sh download -o %s %s", t.outputFilename, t.filename)
	result, err := executable.Run("download", "-o", t.outputFilename, t.filename)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	downloadedFilePath := path.Join(tempDir, t.outputFilename)
	if err = assertFileSize(downloadedFilePath, t.length); err != nil {
		return err
	}

	if err = assertFileSHA1(downloadedFilePath, t.expectedSha1); err != nil {
		return err
	}

	return nil
}
