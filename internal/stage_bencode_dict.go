package internal

import (
	tester_utils "github.com/codecrafters-io/tester-utils"
)

// TODO: Randomize
func testBencodeDict(stageHarness *tester_utils.StageHarness) error {
	logger := stageHarness.Logger
	executable := stageHarness.Executable

	logger.Debugf("Running ./your_bittorrent.sh decode d3:foo3:bar5:helloi52ee")
	result, err := executable.Run("decode", "d3:foo3:bar5:helloi52ee")
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	list := []string{
		"{\"foo\":\"bar\",\"hello\":52}\n",
		"{\"foo\": \"bar\", \"hello\": 52}\n",
	}
	if err = assertStdoutList(result, list); err != nil {
		return err
	}

	return nil
}
