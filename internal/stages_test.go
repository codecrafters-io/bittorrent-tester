package internal

import (
	"testing"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

func TestStages(t *testing.T) {
	testCases := map[string]tester_utils.TesterOutputTestCase{
		"bencoded_string": {
			StageName:           "bencode-string",
			CodePath:            "./test_helpers/scenarios/init/failure",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/init/failure",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
	}

	tester_utils.TestTesterOutput(t, testerDefinition, testCases)
}

func normalizeTesterOutput(testerOutput []byte) []byte {
	return testerOutput
}
