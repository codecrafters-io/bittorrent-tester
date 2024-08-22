package internal

import (
	"fmt"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testParseMagnetLink(stageHarness *test_case_harness.TestCaseHarness) error {
	type MagnetLink struct {
		UrlEncoded string
		Trackers []string
		InfoHash string
	}

	magnetLinks := []MagnetLink {
		{
		   UrlEncoded: "magnet:?xt=urn:btih:c77829d2a77d6516F88cd7a3de1a26abcbfab0db&dn=codercat.gif&tr=http%3A%2F%2Fbittorrent-test-tracker.codecrafters.io%2Fannounce",
		   Trackers: []string { "http://bittorrent-test-tracker.codecrafters.io/announce" },
		   InfoHash: "c77829d2a77d6516F88cd7a3de1a26abcbfab0db",
	   },
	   {
		UrlEncoded: "magnet:?xt=urn:btih:c77829d2a77d6516F88cd7a3de1a26abcbfab0db",
		Trackers: []string { ""},
		InfoHash: "c77829d2a77d6516F88cd7a3de1a26abcbfab0db",
	   },
	}

	logger := stageHarness.Logger
	executable := stageHarness.Executable

	for _, link := range(magnetLinks) {
		logger.Infof("Running ./your_bittorrent.sh magnet_parse %s", escape(link.UrlEncoded))
		result, err := executable.Run("magnet_parse", link.UrlEncoded)
		if err != nil {
			return err
		}
	
		if err = assertExitCode(result, 0); err != nil {
			return err
		}

		expected := fmt.Sprintf("Info Hash: %s\n", link.InfoHash)
		if err = assertStdoutContains(result, expected); err != nil {
			return err
		}

		for _, tracker := range(link.Trackers) {
			if len(tracker) > 0 {
				trackerStr := fmt.Sprintf("Tracker URL: %s\n", tracker)
				if err = assertStdoutContains(result, trackerStr); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
