package internal

import (
	"os"
	"regexp"
	"testing"

	tester_utils_testing "github.com/codecrafters-io/tester-utils/testing"
)

func TestStages(t *testing.T) {
	os.Setenv("CODECRAFTERS_RANDOM_SEED", "1234567890")

	testCases := map[string]tester_utils_testing.TesterOutputTestCase{
		"bencoded_string_failure": {
			UntilStageSlug:      "ns2",
			CodePath:            "./test_helpers/scenarios/init/failure",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/init/failure",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"bencoded_string_success": {
			UntilStageSlug:      "ns2",
			CodePath:            "./test_helpers/scenarios/init/success",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/init/success",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"handshake_with_peer_check": {
			UntilStageSlug:      "ca4",
			CodePath:            "./test_helpers/scenarios/handshake/with_peer_check",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/handshake/with_peer_check",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"pass_all": {
			UntilStageSlug:      "hw0",
			CodePath:            "./test_helpers/scenarios/pass_all",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/pass_all",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
	}

	tester_utils_testing.TestTesterOutput(t, testerDefinition, testCases)
}

func normalizeTesterOutput(testerOutput []byte) []byte {
	replacements := map[string][]*regexp.Regexp{
		"Running ./your_bittorrent.sh <truncated>": {regexp.MustCompile("Running ./your_bittorrent.sh .*")},
		"127.0.0.1:xxxx": {regexp.MustCompile("127.0.0.1:\\d+")},
	}

	for replacement, regexes := range replacements {
		for _, regex := range regexes {
			testerOutput = regex.ReplaceAll(testerOutput, []byte(replacement))
		}
	}

	return testerOutput
}
