package internal

import (
	"fmt"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

func testBencodeList(stageHarness *tester_utils.StageHarness) error {
	initRandom()

	logger := stageHarness.Logger
	executable := stageHarness.Executable

	randomWord := randomWord()
	randomWordEncoded := fmt.Sprintf("%d:%s", len(randomWord), randomWord)
	listEncoded := fmt.Sprintf("l%si52ee", randomWordEncoded)

	logger.Debugf("Running ./your_bittorrent.sh decode %s", listEncoded)
	result, err := executable.Run("decode", listEncoded)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	list := []string{
		fmt.Sprintf("[\"%s\",52]\n", randomWord),
		fmt.Sprintf("[\"%s\", 52]\n", randomWord),
	}
	if err = assertStdoutList(result, list); err != nil {
		return err
	}

	return nil
}
