package internal

import (
	"fmt"
	"math/rand"
	"os"
	"path"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

type DiscoverPeersTestCase struct {
	filename string
	expected []string
}

var discoverPeersResponses = []DiscoverPeersTestCase{
	{
		filename: "response.txt",
		expected: []string{
			"188.119.61.177:6881",
			"71.224.0.29:51414",
			"62.153.208.98:3652",
			"37.48.74.20:44697",
			"82.149.227.229:6890",
			"195.90.215.221:45682",
			"66.55.206.70:60000",
			"69.53.20.159:60000",
			"216.195.129.27:60000",
		},
	},
	{
		filename: "response1.txt",
		expected: []string{
			"106.72.196.0:41485",
			"188.119.61.177:6881",
			"2.7.245.20:51413",
			"71.224.0.29:51414",
			"37.48.74.20:44697",
			"82.149.227.229:6890",
			"72.175.28.2:58966",
			"45.67.229.74:60007",
			"195.90.215.221:45682",
			"66.55.206.70:60000",
			"69.53.20.159:60000",
			"216.195.129.27:60000",
		},
	},
	{
		filename: "response2.txt",
		expected: []string{
			"188.119.61.177:6881",
			"185.107.13.235:54542",
			"88.99.2.101:6881",
		},
	},
}

func randomResponse() DiscoverPeersTestCase {
	return discoverPeersResponses[rand.Intn(len(discoverPeersResponses))]
}

func testDiscoverPeers(stageHarness *test_case_harness.TestCaseHarness) error {
	initRandom()

	logger := stageHarness.Logger
	executable := stageHarness.Executable

	tempDir, err := os.MkdirTemp("", "torrents")
	if err != nil {
		return err
	}

	// Copy sample response for HTTP server
	response := randomResponse()
	destinationPath := path.Join(tempDir, response.filename)
	sourcePath := getResponsePath(response.filename)
	if err = copyFile(sourcePath, destinationPath); err != nil {
		logger.Errorf("Couldn't copy sample response: %s", err)
		return err
	}

	port, err := findFreePort()
	if err != nil {
		logger.Errorf("Error finding free port: %s", err)
		return err
	}

	torrentFileName := "test.torrent"
	torrentFilePath := path.Join(tempDir, torrentFileName)

	address := fmt.Sprintf("127.0.0.1:%d", port)
	pieceLengthBytes := 256 * 1024
	fileLengthBytes := pieceLengthBytes * len(samplePieceHashes)
	torrent := TorrentFile{
		Announce: fmt.Sprintf("http://%s/announce", address),
		Info: TorrentFileInfo{
			Name:        "faketorrent.iso",
			Length:      fileLengthBytes,
			Pieces:      toPiecesStr(samplePieceHashes),
			PieceLength: pieceLengthBytes,
		},
	}
	expectedInfoHash, err := torrent.writeToFile(torrentFilePath)
	if err != nil {
		logger.Errorf("Couldn't write torrent file: %s", err)
		return err
	}

	peersResponse, err := os.ReadFile(destinationPath)
	if err != nil {
		return err
	}

	go listenAndServePeersResponse(address, peersResponse, expectedInfoHash, fileLengthBytes, logger)

	logger.Infof("Running ./your_bittorrent.sh peers %s", torrentFilePath)
	result, err := executable.Run("peers", torrentFilePath)

	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	for _, ip := range response.expected {
		if err = assertStdoutContains(result, ip); err != nil {
			return err
		}
	}

	return nil
}
