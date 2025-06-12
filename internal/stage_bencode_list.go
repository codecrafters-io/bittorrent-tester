package internal

import (
	"fmt"
	"math/rand"
	"path"
	"strings"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testBencodeList(stageHarness *test_case_harness.TestCaseHarness) error {
	logger := stageHarness.Logger
	executable := stageHarness.Executable

	// Test empty list
	emptyListEncoded := "le"
	emptyListDecoded := "[]"
	logger.Infof("Running ./%s decode %s", path.Base(executable.Path), emptyListEncoded)
	logger.Infof("Expected output: %s", emptyListDecoded)
	result, err := executable.Run("decode", emptyListEncoded)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}
	if err = assertStdout(result, emptyListDecoded+"\n"); err != nil {
		return err
	}

	// Test list with random word and random number
	randomWord := randomWord()
	randomWordEncoded := fmt.Sprintf("%d:%s", len(randomWord), randomWord)
	randomNumber := rand.Intn(1000)
	randomNumberEncoded := fmt.Sprintf("i%de", randomNumber)
	listEncoded := fmt.Sprintf("l%s%se", randomWordEncoded, randomNumberEncoded)

	list := []string{
		fmt.Sprintf("[\"%s\",%d]\n", randomWord, randomNumber),
		fmt.Sprintf("[\"%s\", %d]\n", randomWord, randomNumber),
	}

	logger.Infof("Running ./%s decode %s", path.Base(executable.Path), listEncoded)
	logger.Infof("Expected output: %s", strings.TrimSpace(list[0]))
	result, err = executable.Run("decode", listEncoded)
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
	nestedListEncoded := fmt.Sprintf("ll%s%see", randomNumberEncoded, randomWordEncoded)

	nestedList := []string{
		fmt.Sprintf("[[%d,\"%s\"]]\n", randomNumber, randomWord),
		fmt.Sprintf("[[%d, \"%s\"]]\n", randomNumber, randomWord),
	}

	logger.Infof("Running ./%s decode %s", path.Base(executable.Path), nestedListEncoded)
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

	// Test for a nested list: [[4], 5]
	nestedListEncoded = "lli4eei5ee"
	nestedList = []string{
		"[[4],5]\n",
		"[[4], 5]\n",
	}
	logger.Infof("Running ./%s decode %s", path.Base(executable.Path), nestedListEncoded)
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
