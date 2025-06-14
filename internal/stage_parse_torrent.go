package internal

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testParseTorrent(stageHarness *test_case_harness.TestCaseHarness) error {
	logger := stageHarness.Logger
	executable := stageHarness.Executable
	torrent := randomTorrent()

	tempDir, err := os.MkdirTemp("", "torrents")
	if err != nil {
		return err
	}

	if err := copyTorrent(tempDir, torrent.filename); err != nil {
		logger.Errorln("Couldn't copy torrent file")
		return err
	}

	torrentPath := path.Join(tempDir, torrent.filename)

	logger.Infof("Running ./%s info %s", path.Base(executable.Path), torrentPath)
	result, err := executable.Run("info", torrentPath)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	expectedTrackerURLValue := fmt.Sprintf("Tracker URL: %s", torrent.tracker)
	expectedLengthValue := fmt.Sprintf("Length: %d", torrent.length)

	logger.Debugf("Checking for tracker URL (%v)", expectedTrackerURLValue)

	if err = assertStdoutContains(result, expectedTrackerURLValue); err != nil {
		actual := string(result.Stdout)
		if strings.Contains(actual, fmt.Sprintf("Tracker URL:%s", torrent.tracker)) {
			logger.Errorln("There needs to be a space character after Tracker URL:")
		}
		return err
	}

	logger.Successf("Tracker URL is correct")

	logger.Debugf("Checking for length (%v)", expectedLengthValue)

	if err = assertStdoutContains(result, expectedLengthValue); err != nil {
		actual := string(result.Stdout)
		if strings.Contains(actual, fmt.Sprintf("Length:%d", torrent.length)) {
			logger.Errorln("There needs to be a space character after Length:")
		}
		return err
	}

	logger.Successf("Length is correct")

	return nil
}
