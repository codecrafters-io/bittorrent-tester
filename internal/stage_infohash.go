package internal

import (
	"fmt"
	"os"
	"path"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

func testInfoHash(stageHarness *tester_utils.StageHarness) error {
	initRandom()

	logger := stageHarness.Logger
	executable := stageHarness.Executable
	torrent := randomTorrent()

	tempDir, err := os.MkdirTemp("", "torrents")
	if err != nil {
		return err
	}

	if err := copyTorrent(tempDir, torrent.filename); err != nil {
		logger.Errorf("Couldn't copy torrent file")
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

	expected := fmt.Sprintf("Info Hash: %s", torrent.infohash)

	if err = assertStdoutContains(result, expected); err != nil {
		return err
	}

	return nil
}
