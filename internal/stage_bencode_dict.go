package internal

import (
	"fmt"
	"path"
	"strings"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testBencodeDict(stageHarness *test_case_harness.TestCaseHarness) error {
	logger := stageHarness.Logger
	executable := stageHarness.Executable

	emptyDictEncoded := "de"
	emptyDictExpectedJson := "{}"
	logger.Infof("Running ./%s decode %s", path.Base(executable.Path), emptyDictEncoded)
	logger.Infof("Expected output: %s", emptyDictExpectedJson)
	result, err := executable.Run("decode", emptyDictEncoded)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	if err = assertStdout(result, emptyDictExpectedJson+"\n"); err != nil {
		return err
	}

	randomWord := randomWord()
	randomWordEncoded := fmt.Sprintf("%d:%s", len(randomWord), randomWord)
	// Keys must be strings and appear in sorted order
	randomDictEncoded := fmt.Sprintf("d3:foo%s5:helloi52ee", randomWordEncoded)
	randomDictExpectedJsonList := []string{
		fmt.Sprintf("{\"foo\":\"%s\",\"hello\":52}\n", randomWord),
		fmt.Sprintf("{\"foo\": \"%s\", \"hello\": 52}\n", randomWord),
	}

	logger.Infof("Running ./%s decode %s", path.Base(executable.Path), randomDictEncoded)
	logger.Infof("Expected output: %s", strings.TrimSpace(randomDictExpectedJsonList[0]))
	result, err = executable.Run("decode", randomDictEncoded)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	if err = assertStdoutList(result, randomDictExpectedJsonList); err != nil {
		return err
	}

	dictWithinDictEncoded := "d10:inner_dictd4:key16:value14:key2i42e8:list_keyl5:item15:item2i3eeee"
	dictWithinDictExpectedJsonList := []string{
		"{\"inner_dict\":{\"key1\":\"value1\",\"key2\":42,\"list_key\":[\"item1\",\"item2\",3]}}\n",
		"{\"inner_dict\": {\"key1\": \"value1\", \"key2\": 42, \"list_key\": [\"item1\", \"item2\", 3]}}\n",
	}
	logger.Infof("Running ./%s decode %s", path.Base(executable.Path), dictWithinDictEncoded)
	logger.Infof("Expected output: %s", strings.TrimSpace(dictWithinDictExpectedJsonList[0]))
	result, err = executable.Run("decode", dictWithinDictEncoded)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	if err = assertStdoutList(result, dictWithinDictExpectedJsonList); err != nil {
		return err
	}

	return nil
}
