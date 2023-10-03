package internal

import (
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
)

var testerDefinition = testerutils.TesterDefinition{
	AntiCheatTestCases:    []testerutils.TestCase{},
	ExecutableFileName: "your_bittorrent.sh",
	TestCases: []testerutils.TestCase{
		{
			Slug:                    "bencode-string",
			TestFunc:                testBencodeString,
		},
		{
			Slug:                    "bencode-int",
			TestFunc:                testBencodeInt,
		},
		{
			Slug:                    "bencode-list",
			TestFunc:                testBencodeList,
		},
		{
			Slug:                    "bencode-dict",
			TestFunc:                testBencodeDict,
		},
		{
			Slug:                    "parse-torrent",
			TestFunc:                testParseTorrent,
		},
		{
			Slug:                    "infohash",
			TestFunc:                testInfoHash,
		},
		{
			Slug:                    "hashes",
			TestFunc:                testPieceHashes,
		},
		{
			Slug:                    "peers",
			TestFunc:                testDiscoverPeers,
		},
		{
			Slug:                    "handshake",
			TestFunc:                testHandshake,
		},
		{
			Slug:                    "dl-piece",
			TestFunc:                testDownloadPiece,
		},
		{
			Slug:                    "dl-file",
			TestFunc:                testDownloadFile,
			Timeout:                 90 * time.Second,
		},
	},
}
