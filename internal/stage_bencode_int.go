package internal

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testBencodeInt(stageHarness *test_case_harness.TestCaseHarness) error {
	initRandom()

	logger := stageHarness.Logger
	executable := stageHarness.Executable

	randomNumber := fmt.Sprintf("%d", rand.Intn(2147483647))
	randomNumberEncoded := fmt.Sprintf("i%se", randomNumber)
	logger.Infof("Running ./your_bittorrent.sh decode %s", randomNumberEncoded)
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
	logger.Infof("Running ./your_bittorrent.sh decode %s", largeNumberEncoded)
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

	logger.Infof("Running ./your_bittorrent.sh decode i-52e")
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
