package internal

import (
	"fmt"
	"os"
	"strings"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

func assertStdoutList(result tester_utils.ExecutableResult, expected []string) error {
	actual := string(result.Stdout)
	if indexOf(expected, actual) == -1 {
		return fmt.Errorf("Expected %q as stdout, got: %q", expected[0], actual)
	}

	return nil
}

func indexOf(slice []string, target string) int {
	for i, v := range slice {
		if v == target {
			return i
		}
	}
	return -1
}

func assertStdout(result tester_utils.ExecutableResult, expected string) error {
	actual := string(result.Stdout)
	if expected != actual {
		return fmt.Errorf("Expected %q as stdout, got: %q", expected, actual)
	}

	return nil
}

func assertStderr(result tester_utils.ExecutableResult, expected string) error {
	actual := string(result.Stderr)
	if expected != actual {
		return fmt.Errorf("Expected %q as stderr, got: %q", expected, actual)
	}

	return nil
}

func assertStdoutContains(result tester_utils.ExecutableResult, expectedSubstring string) error {
	actual := string(result.Stdout)
	if !strings.Contains(actual, expectedSubstring) {
		return fmt.Errorf("Expected stdout to contain %q, got: %q", expectedSubstring, actual)
	}

	return nil
}

func assertStderrContains(result tester_utils.ExecutableResult, expectedSubstring string) error {
	actual := string(result.Stderr)
	if !strings.Contains(actual, expectedSubstring) {
		return fmt.Errorf("Expected stderr to contain %q, got: %q", expectedSubstring, actual)
	}

	return nil
}

func assertExitCode(result tester_utils.ExecutableResult, expected int) error {
	actual := result.ExitCode
	if expected != actual {
		return fmt.Errorf("Expected %d as exit code, got: %d", expected, actual)
	}

	return nil
}

func assertFileSize(downloadedFilePath string, expectedFileSize int64) error {
	fileInfo, err := os.Stat(downloadedFilePath)
	if err != nil {
		return err
	}

	fileSize := fileInfo.Size()
	if fileSize != expectedFileSize {
		return fmt.Errorf("File size does not match expected file size. Expected: %d Actual: %d", expectedFileSize, fileSize)
	}
	return nil
}

func assertFileSHA1(downloadedFilePath string, expectedSha1 string) error {
	sha1, err := calculateSHA1(downloadedFilePath)
	if err != nil {
		return err
	}
	if sha1 != expectedSha1 {
		return fmt.Errorf("File SHA-1 does not match expected SHA-1. Expected: %s Actual: %s", expectedSha1, sha1)
	}
	return nil
}
