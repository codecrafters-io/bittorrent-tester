package internal

import (
	"fmt"
	"path"
	"strings"

	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testBencodeInt(stageHarness *test_case_harness.TestCaseHarness) error {
	logger := stageHarness.Logger
	executable := stageHarness.Executable

	randomNumber := fmt.Sprintf("%d", random.RandomInt(0, 2147483647))
	randomNumberEncoded := fmt.Sprintf("i%se", randomNumber)
	logger.Infof("Running ./%s decode %s", path.Base(executable.Path), randomNumberEncoded)
	result, err := executable.Run("decode", randomNumberEncoded)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	expected := randomNumber + "\n"
	if err = assertStdout(result, expected); err != nil {
		actual := string(result.Stdout)
		if strings.Contains(actual, "\""+randomNumber+"\"") {
			logger.Errorln("You need to print the number without quotes")
		}
		return err
	}

	largeNumber := "4294967300"
	largeNumberEncoded := fmt.Sprintf("i%se", largeNumber)
	logger.Infof("Running ./%s decode %s", path.Base(executable.Path), largeNumberEncoded)
	result, err = executable.Run("decode", largeNumberEncoded)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	expected = largeNumber + "\n"
	if err = assertStdout(result, expected); err != nil {
		return err
	}

	logger.Infof("Running ./%s decode i-52e", path.Base(executable.Path))
	result, err = executable.Run("decode", "i-52e")
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	expected = "-52" + "\n"
	if err = assertStdout(result, expected); err != nil {
		return err
	}

	return nil
}
