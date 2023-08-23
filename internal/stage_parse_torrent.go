package internal

import (
	"fmt"
	"os"
	"strings"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

func testParseTorrent(stageHarness *tester_utils.StageHarness) error {
	logger := stageHarness.Logger
	executable := stageHarness.Executable
	torrent := randomTorrent()

	tempDir, err := createTempDir(executable)
	if err != nil {
		logger.Errorf("Couldn't create temp directory")
		return err
	}

	if err := copyTorrent(tempDir, torrent.filename); err != nil {
		logger.Errorf("TESTER DIR: %s", os.Getenv("TESTER_DIR"))
		logger.Errorf("Couldn't copy torrent file")
		return err
	}

	logger.Debugf("Running ./your_bittorrent.sh info %s", torrent.filename)
	result, err := executable.Run("info", torrent.filename)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	expected := strings.Join([]string{
		fmt.Sprintf("Tracker URL: %s", torrent.tracker),
		fmt.Sprintf("Length: %s", torrent.length)}, "\n") + "\n"

	if err = assertStdoutContains(result, expected); err != nil {
		return err
	}

	return nil
}
