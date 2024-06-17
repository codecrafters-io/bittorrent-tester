package internal

import (
	"os"
	"path"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testDownloadFile(stageHarness *test_case_harness.TestCaseHarness) error {
	initRandom()

	logger := stageHarness.Logger
	executable := stageHarness.Executable

	t := randomTorrent()

	tempDir, err := os.MkdirTemp("", "torrents")
	logger.Infof("Temp dir: %s", tempDir)
	if err != nil {
		logger.Errorln("Couldn't create temp directory")
		return err
	}

	if err := copyTorrent(tempDir, t.filename); err != nil {
		logger.Errorln("Couldn't copy torrent file")
		return err
	}

	torrentFilePath := path.Join(tempDir, t.filename)
	downloadedFilePath := path.Join(tempDir, t.outputFilename)

	logger.Infof("Running ./%s download -o %s %s", path.Base(executable.Path), downloadedFilePath, torrentFilePath)
	result, err := executable.Run("download", "-o", path.Base(executable.Path), downloadedFilePath, torrentFilePath)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	if err = assertFileSize(downloadedFilePath, t.length); err != nil {
		return err
	}

	if err = assertFileSHA1(downloadedFilePath, t.expectedSha1); err != nil {
		return err
	}

	return nil
}
