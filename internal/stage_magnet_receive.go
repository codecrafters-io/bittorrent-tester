package internal

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testMagnetReceiveExtendedHandshake(stageHarness *test_case_harness.TestCaseHarness) error {
	initRandom()

	logger := stageHarness.Logger
	executable := stageHarness.Executable

	magnetLink := randomMagnetLink()
	params, err := NewMagnetTestParams(magnetLink, logger)
	if err != nil {
		return err
	}

	go listenAndServeTrackerResponse(params.toTrackerParams())
	go waitAndHandlePeerConnection(params.toPeerConnectionParams(), handleReceiveExtensionHandshake)

	logger.Infof("Running ./your_bittorrent.sh magnet_handshake %q", params.MagnetUrlEncoded)
	result, err := executable.Run("magnet_handshake", params.MagnetUrlEncoded)
	if err != nil {
		return err
	}
	
	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	expected := fmt.Sprintf("Peer ID: %x\n", params.ExpectedPeerID)

	if err = assertStdoutContains(result, expected); err != nil {
		return err
	}
	
	expected = fmt.Sprintf("Peer Metadata Extension ID: %d\n", params.MyMetadataExtensionID)

	if err = assertStdoutContains(result, expected); err != nil {
		return err
	}

	return nil
}

func handleReceiveExtensionHandshake(conn net.Conn, p PeerConnectionParams) {
	defer conn.Close()
	logger := p.logger

	if err := receiveAndSendHandshake(conn, p); err != nil {
		return
	}

	if err := sendBitfieldMessage(conn, p.bitfield, logger); err != nil {
		return
	}

	if err := sendExtensionHandshake(conn, p.myMetadataExtensionID, p.metadataSizeBytes, logger); err != nil {
		return
	}
}