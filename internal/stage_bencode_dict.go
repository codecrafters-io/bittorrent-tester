package internal

import (
	"fmt"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

func testBencodeDict(stageHarness *tester_utils.StageHarness) error {
	initRandom()

	logger := stageHarness.Logger
	executable := stageHarness.Executable

	randomWord := randomWord()
	randomWordEncoded := fmt.Sprintf("%d:%s", len(randomWord), randomWord)
	// Keys must be strings and appear in sorted order
	randomDictEncoded := fmt.Sprintf("d3:foo%s5:helloi52ee", randomWordEncoded)

	logger.Infof("Running ./your_bittorrent.sh decode %s", randomDictEncoded)
	result, err := executable.Run("decode", randomDictEncoded)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	list := []string{
		fmt.Sprintf("{\"foo\":\"%s\",\"hello\":52}\n", randomWord),
		fmt.Sprintf("{\"foo\": \"%s\", \"hello\": 52}\n", randomWord),
	}
	if err = assertStdoutList(result, list); err != nil {
		return err
	}

	return nil
}
