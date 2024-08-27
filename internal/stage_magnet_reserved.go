package internal

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testMagnetReserved(stageHarness *test_case_harness.TestCaseHarness) error {
    initRandom()

    logger := stageHarness.Logger
    executable := stageHarness.Executable

    magnetLink := randomMagnetLink()
    params, err := NewMagnetTestParams(magnetLink, logger)
    if err != nil {
        return err
    }

    go listenAndServeTrackerResponse(params.toTrackerParams())
    go waitAndHandlePeerConnection(params.toPeerConnectionParams(), handleReservedBytes)

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

    return nil
}

func handleReservedBytes(conn net.Conn, params PeerConnectionParams) {
    defer conn.Close()

    if err := receiveAndSendHandshake(conn, params); err != nil {
        return
    }
}