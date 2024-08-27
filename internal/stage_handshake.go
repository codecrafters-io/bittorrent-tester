package internal

import (
	"fmt"
	"net"
	"os"
	"path"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testHandshake(stageHarness *test_case_harness.TestCaseHarness) error {
	initRandom()

	logger := stageHarness.Logger
	executable := stageHarness.Executable

	tempDir, err := os.MkdirTemp("", "torrents")
	if err != nil {
		return err
	}

	peerPort, err := findFreePort()
	if err != nil {
		logger.Errorf("Couldn't find free port: %s", err)
		return err
	}
	peerAddress := fmt.Sprintf("127.0.0.1:%d", peerPort)

	trackerPort, err := findFreePort()
	if err != nil {
		logger.Errorf("Couldn't find free port: %s", err)
		return err
	}

	trackerAddress := fmt.Sprintf("127.0.0.1:%d", trackerPort)
	trackerURL := fmt.Sprintf("http://%s/announce", trackerAddress)
	pieceLengthBytes := 1024 * 256
	fileLengthBytes := pieceLengthBytes * len(samplePieceHashes)
	torrent := TorrentFile{
		Announce: trackerURL,
		Info: TorrentFileInfo{
			Name:        "fakefilename.iso",
			Length:      fileLengthBytes,
			Pieces:      toPiecesStr(samplePieceHashes),
			PieceLength: pieceLengthBytes,
		},
	}

	torrentFilename := "test.torrent"
	torrentFilePath := path.Join(tempDir, torrentFilename)
	infoHash, err := torrent.writeToFile(torrentFilePath)
	if err != nil {
		logger.Errorf("Error writing torrent file: %s", err)
		return err
	}

	expectedPeerID, err := randomHash()
	if err != nil {
		return err
	}

	expectedReservedBytes := [][]byte{
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 16, 0, 0},
	}

	peersResponse := createPeersResponse("127.0.0.1", peerPort)

	go listenAndServeTrackerResponse(
		TrackerParams {
			trackerAddress: trackerAddress,
			peersResponse: peersResponse,
			expectedInfoHash: infoHash,
			fileLengthBytes: fileLengthBytes,
			logger: logger,
	})

	go waitAndHandlePeerConnection(
		PeerConnectionParams {
			address: peerAddress,
			myPeerID: expectedPeerID,
			infoHash: infoHash,
			expectedReservedBytes: expectedReservedBytes,
			logger: logger,
		},
		handleHandshake,
	)

	logger.Infof("Running ./%s handshake %s %s", path.Base(executable.Path), torrentFilePath, peerAddress)
	result, err := executable.Run("handshake", torrentFilePath, peerAddress)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	expected := fmt.Sprintf("Peer ID: %x\n", expectedPeerID)

	if err = assertStdoutContains(result, expected); err != nil {
		return err
	}

	return nil
}

func handleHandshake(conn net.Conn, params PeerConnectionParams) {
	defer conn.Close()

	err := receiveAndSendHandshake(conn, params)
	if err != nil {
		return
	}
}
