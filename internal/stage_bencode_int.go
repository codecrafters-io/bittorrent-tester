package internal

import (
	tester_utils "github.com/codecrafters-io/tester-utils"
)

// TODO: Randomize
func testBencodeInt(stageHarness *tester_utils.StageHarness) error {
	logger := stageHarness.Logger
	executable := stageHarness.Executable

	logger.Debugf("Running ./your_bittorrent.sh decode i52e")
	result, err := executable.Run("decode", "i52e")
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	expected := "52" + "\n"
	if err = assertStdout(result, expected); err != nil {
		return err
	}

	logger.Debugf("Running ./your_bittorrent.sh decode i-52e")
	result, err = executable.Run("decode", "i-52e")
	if err != nil {
		return err
	}
	expected = "-52" + "\n"
	if err = assertStdout(result, expected); err != nil {
		return err
	}

	return nil
}
