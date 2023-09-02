package internal

import (
	"fmt"
	"strings"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

func testBencodeList(stageHarness *tester_utils.StageHarness) error {
	initRandom()

	logger := stageHarness.Logger
	executable := stageHarness.Executable

	randomWord := randomWord()
	randomWordEncoded := fmt.Sprintf("%d:%s", len(randomWord), randomWord)
	listEncoded := fmt.Sprintf("l%si52ee", randomWordEncoded)

	list := []string{
		fmt.Sprintf("[\"%s\",52]\n", randomWord),
		fmt.Sprintf("[\"%s\", 52]\n", randomWord),
	}

	logger.Infof("Running ./your_bittorrent.sh decode %s", listEncoded)
	logger.Infof("Expected output: %s", strings.TrimSpace(list[0]))
	result, err := executable.Run("decode", listEncoded)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	if err = assertStdoutList(result, list); err != nil {
		return err
	}

	// Test for a nested list
	nestedListEncoded := fmt.Sprintf("ll%si52eee", randomWordEncoded)

	nestedList := []string{
		fmt.Sprintf("[[\"%s\",52]]\n", randomWord),
		fmt.Sprintf("[[\"%s\", 52]]\n", randomWord),
	}

	logger.Infof("Running ./your_bittorrent.sh decode %s", nestedListEncoded)
	logger.Infof("Expected output: %s", strings.TrimSpace(nestedList[0]))
	result, err = executable.Run("decode", nestedListEncoded)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	if err = assertStdoutList(result, nestedList); err != nil {
		return err
	}

	return nil
}
