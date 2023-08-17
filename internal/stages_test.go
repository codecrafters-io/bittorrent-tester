package internal

import (
	"testing"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

func TestStages(t *testing.T) {
	testCases := map[string]tester_utils.TesterOutputTestCase{
		"bencoded_string_failure": {
			StageName:           "bencode-string",
			CodePath:            "./test_helpers/scenarios/init/failure",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/init/failure",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"bencoded_string_success": {
			StageName:           "bencode-string",
			CodePath:            "./test_helpers/scenarios/init/success",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/init/success",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
	}

	tester_utils.TestTesterOutput(t, testerDefinition, testCases)
}

func normalizeTesterOutput(testerOutput []byte) []byte {
	return testerOutput
}
