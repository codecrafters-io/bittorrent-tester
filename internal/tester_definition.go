package internal

import (
	"time"

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
		{
			Number:                  2,
			Slug:                    "bencode-int",
			Title:                   "Decode bencoded integers",
			TestFunc:                testBencodeInt,
			ShouldRunPreviousStages: true,
		},
		{
			Number:                  3,
			Slug:                    "bencode-list",
			Title:                   "Decode bencoded lists",
			TestFunc:                testBencodeList,
			ShouldRunPreviousStages: true,
		},
		{
			Number:                  4,
			Slug:                    "bencode-dict",
			Title:                   "Decode bencoded dictionaries",
			TestFunc:                testBencodeDict,
			ShouldRunPreviousStages: true,
		},
		{
			Number:                  5,
			Slug:                    "parse-torrent",
			Title:                   "Parse torrent file",
			TestFunc:                testParseTorrent,
			ShouldRunPreviousStages: true,
		},
		{
			Number:                  6,
			Slug:                    "infohash",
			Title:                   "Calculate info hash",
			TestFunc:                testInfoHash,
			ShouldRunPreviousStages: true,
		},
		{
			Number:                  7,
			Slug:                    "peers",
			Title:                   "Discover peers",
			TestFunc:                testDiscoverPeers,
			ShouldRunPreviousStages: true,
		},
		{
			Number:                  8,
			Slug:                    "handshake",
			Title:                   "Peer handshake",
			TestFunc:                testHandshake,
			ShouldRunPreviousStages: true,
		},
		{
			Number:                  9,
			Slug:                    "hashes",
			Title:                   "Piece hashes",
			TestFunc:                testPieceHashes,
			ShouldRunPreviousStages: true,
		},
		{
			Number:                  10,
			Slug:                    "dl-piece",
			Title:                   "Download a piece",
			TestFunc:                testDownloadPiece,
			ShouldRunPreviousStages: true,
		},
		{
			Number:                  11,
			Slug:                    "dl-file",
			Title:                   "Download the whole file",
			TestFunc:                testDownloadFile,
			ShouldRunPreviousStages: true,
			Timeout:                 90 * time.Second,
		},
	},
}
