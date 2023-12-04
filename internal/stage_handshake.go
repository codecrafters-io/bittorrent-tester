package internal

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"os"
	"path"

	tester_utils "github.com/codecrafters-io/tester-utils"
	logger "github.com/codecrafters-io/tester-utils/logger"
)

func testHandshake(stageHarness *tester_utils.StageHarness) error {
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

	peersResponse := createPeersResponse("127.0.0.1", peerPort)
	go listenAndServePeersResponse(trackerAddress, peersResponse, infoHash, fileLengthBytes, logger)
	go waitAndHandlePeerConnection(peerAddress, expectedPeerID, infoHash, logger)

	logger.Infof("Running ./your_bittorrent.sh handshake %s %s", torrentFilePath, peerAddress)
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

func randomHash() ([20]byte, error) {
	var hash [20]byte
	if _, err := rand.Read(hash[:]); err != nil {
		return [20]byte{}, err
	}
	return hash, nil
}

func waitAndHandlePeerConnection(address string, myPeerID [20]byte, infoHash [20]byte, logger *logger.Logger) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Errorf("Error: %s", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Errorf("Error accepting connection: %s", err)
		}
		logger.Debugf("Waiting for handshake message")
		handleConnection(conn, myPeerID, infoHash, logger)
	}
}

func handleConnection(conn net.Conn, myPeerID [20]byte, infoHash [20]byte, logger *logger.Logger) {
	defer conn.Close()

	handshake, err := readHandshake(conn, logger)
	if err != nil {
		logger.Errorf("error reading handshake: %s", err)
		return
	}
	if !bytes.Equal(handshake.InfoHash[:], infoHash[:]) {
		logger.Errorf("expected infohash %x but got %x", infoHash, handshake.InfoHash)
		return
	}

	logger.Debugf("Received handshake: [infohash: %x, peer_id: %x]\n", handshake.InfoHash, handshake.PeerID)
	logger.Debugf("Sending back handshake with peer_id: %x", myPeerID)
	sendHandshake(conn, handshake.InfoHash, myPeerID)
}
