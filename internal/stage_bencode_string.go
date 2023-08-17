package internal

import (
	"fmt"
	"io/ioutil"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

func testBencodeString(stageHarness *tester_utils.StageHarness) error {
	initRandom()

	logger := stageHarness.Logger
	executable := stageHarness.Executable

	tempDir, err := ioutil.TempDir("", "worktree")
	if err != nil {
		return err
	}

	executable.WorkingDir = tempDir

	randomWord := randomWord()
	randomWordEncoded := fmt.Sprintf("%d:%s", len(randomWord), randomWord)

	logger.Infof("Running ./your_bittorrent.sh decode %s", randomWordEncoded)
	result, err := executable.Run("decode", randomWordEncoded)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	expected := fmt.Sprintf("\"%s\"\n", randomWord)
	if err = assertStdout(result, expected); err != nil {
		return err
	}

	return nil
}
