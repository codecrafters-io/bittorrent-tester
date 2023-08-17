package internal

import (
	"io/ioutil"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

func testBencodeString(stageHarness *tester_utils.StageHarness) error {
	logger := stageHarness.Logger
	executable := stageHarness.Executable

	tempDir, err := ioutil.TempDir("", "worktree")
	if err != nil {
		return err
	}

	executable.WorkingDir = tempDir

	logger.Debugf("Running ./your_bittorrent.sh decode 5:hello")
	result, err := executable.Run("decode", "5:hello")
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	expected := "\"hello\"" + "\n"
	if err = assertStdout(result, expected); err != nil {
		return err
	}

	return nil
}
