package internal

import (
	"fmt"
	"os"
	"path"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

func testParseTorrent(stageHarness *tester_utils.StageHarness) error {
	initRandom()

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

	logger.Infof("Running ./your_bittorrent.sh info %s", torrentPath)
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
		return err
	}

	logger.Successf("Tracker URL is correct")

	logger.Debugf("Checking for length (%v)", expectedLengthValue)

	if err = assertStdoutContains(result, expectedLengthValue); err != nil {
		return err
	}

	logger.Successf("Length is correct")

	return nil
}
