package internal

import (
	testerutils "github.com/codecrafters-io/tester-utils"
)

var testerDefinition = testerutils.TesterDefinition{
	AntiCheatStages:    []testerutils.Stage{},
	ExecutableFileName: "your_bittorrent.sh",
	Stages: []testerutils.Stage{
		{
			Number:                  1,
			Slug:                    "bencode-string",
			Title:                   "Decode bencoded strings",
			TestFunc:                testBencodeString,
			ShouldRunPreviousStages: true,
		},
	},
}
