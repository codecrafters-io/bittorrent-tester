package internal

import (
    "time"

    "github.com/codecrafters-io/tester-utils/tester_definition"
)

var testerDefinition = tester_definition.TesterDefinition{
    AntiCheatTestCases:       []tester_definition.TestCase{},
    ExecutableFileName:       "your_program.sh",
    LegacyExecutableFileName: "your_bittorrent.sh",
    TestCases: []tester_definition.TestCase{
        {
            Slug: "ns2",
            TestFunc: testBencodeString,
        },
        {
            Slug: "eb4",
            TestFunc: testBencodeInt,
        },
        {
            Slug: "ah1",
            TestFunc: testBencodeList,
        },
        {
            Slug: "mn6",
            TestFunc: testBencodeDict,
        },
        {
            Slug: "ow9",
            TestFunc: testParseTorrent,
        },
        {
            Slug: "rb2",
            TestFunc: testInfoHash,
        },
        {
            Slug: "bf7",
            TestFunc: testPieceHashes,
        },
        {
            Slug: "fi9",
            TestFunc: testDiscoverPeers,
        },
        {
            Slug: "ca4",
            TestFunc: testHandshake,
        },
        {
            Slug: "nd2",
            TestFunc: testDownloadPiece,
        },
        {
            Slug: "jv8",
            TestFunc: testDownloadFile,
            Timeout:  90 * time.Second,
        },
        {
            Slug: "hw0",
            TestFunc: testParseMagnetLink,
        },
        {
            Slug: "pk2",
            TestFunc: testMagnetReserved,
        },
        {
            Slug: "xi4",
            TestFunc: testMagnetSendExtendedHandshake,
        },
        {
            Slug: "jk6",
            TestFunc: testMagnetReceiveExtendedHandshake,
        },
        {
            Slug: "ns5",
            TestFunc: testMagnetRequestMetadata,
        },
        {
            Slug: "zh1",
            TestFunc: testMagnetSendMetadata,
        },
    },
}
