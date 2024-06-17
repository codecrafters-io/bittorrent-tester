package internal

import (
	"fmt"
	"path"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testBencodeString(stageHarness *test_case_harness.TestCaseHarness) error {
	initRandom()

	logger := stageHarness.Logger
	executable := stageHarness.Executable

	randomWord := randomWord()
	randomWordEncoded := fmt.Sprintf("%d:%s", len(randomWord), randomWord)

	logger.Infof("Running ./%s decode %s", path.Base(executable.Path), randomWordEncoded)
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

	trackerURL := "http://bittorrent-test-tracker.codecrafters.io/announce"
	trackerURLEncoded := fmt.Sprintf("%d:%s", len(trackerURL), trackerURL)
	logger.Infof("Running ./%s decode %s", path.Base(executable.Path), trackerURLEncoded)
	result, err = executable.Run("decode", trackerURLEncoded)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	expected = fmt.Sprintf("\"%s\"\n", trackerURL)
	if err = assertStdout(result, expected); err != nil {
		return err
	}

	return nil
}
