package internal

import (
	"fmt"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

// Example from the grep course
func testInit(stageHarness *tester_utils.StageHarness) error {
	// should exit with 0: echo "dog" | grep -E "d"
	stageHarness.Logger.Infof("$ echo \"%s\" | ./script.sh -E \"%s\"", "dog", "d")

	result, err := stageHarness.Executable.RunWithStdin([]byte("dog"), "-E", "d")
	if err != nil {
		return err
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("expected exit code %v, got %v", 0, result.ExitCode)
	}

	stageHarness.Logger.Successf("✓ Received exit code %d.", 0)

	// should exit with 1: echo "dog" | grep -E "d"
	stageHarness.Logger.Infof("$ echo \"%s\" | ./script.sh -E \"%s\"", "dog", "f")

	result, err = stageHarness.Executable.RunWithStdin([]byte("dog"), "-E", "f")
	if err != nil {
		return err
	}

	if result.ExitCode != 1 {
		return fmt.Errorf("expected exit code %v, got %v", 1, result.ExitCode)
	}

	stageHarness.Logger.Successf("✓ Received exit code %d.", 1)
	return nil
}
