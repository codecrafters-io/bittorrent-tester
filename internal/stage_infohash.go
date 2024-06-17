package internal

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"strings"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testInfoHash(stageHarness *test_case_harness.TestCaseHarness) error {
	initRandom()

	logger := stageHarness.Logger
	executable := stageHarness.Executable

	tempDir, err := os.MkdirTemp("", "torrents")
	if err != nil {
		return err
	}

	shuffled := make([]TestTorrentInfo, len(testTorrents))
	copy(shuffled, testTorrents)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	for _, torrent := range shuffled {
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

		expected := fmt.Sprintf("Info Hash: %s", torrent.infohash)

		if err = assertStdoutContains(result, expected); err != nil {
			output := string(result.Stdout)
			for _, incorrectHash := range torrent.incorrectSha1 {
				if strings.Contains(output, incorrectHash) {
					logger.Errorln("WARNING: In your bencoded info dictionary, ensure that keys appear in sorted order.")
					break
				}
			}
			return err
		}
	}

	return nil
}
