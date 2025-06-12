package internal

import (
	"fmt"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testParseMagnetLink(stageHarness *test_case_harness.TestCaseHarness) error {
	link := randomMagnetLink()
	trackerUrl := "http://bittorrent-test-tracker.codecrafters.io/announce"
	urlEncoded := "magnet:?xt=urn:btih:" + link.InfoHashStr + "&dn=" + link.Filename + "&tr=http%3A%2F%2Fbittorrent-test-tracker.codecrafters.io%2Fannounce"

	logger := stageHarness.Logger
	executable := stageHarness.Executable

	logger.Infof("Running ./your_bittorrent.sh magnet_parse %q", urlEncoded)
	result, err := executable.Run("magnet_parse", urlEncoded)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	expected := fmt.Sprintf("Info Hash: %s\n", link.InfoHashStr)
	if err = assertStdoutContains(result, expected); err != nil {
		return err
	}

	logger.Successln("✓ Info Hash is correct.")

	trackerStr := fmt.Sprintf("Tracker URL: %s\n", trackerUrl)
	if err = assertStdoutContains(result, trackerStr); err != nil {
		return err
	}

	logger.Successln("✓ Tracker URL is correct.")
	return nil
}
