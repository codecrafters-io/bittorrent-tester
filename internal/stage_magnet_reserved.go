package internal

import (
	"fmt"
	"net"
	"time"

	logger "github.com/codecrafters-io/tester-utils/logger"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testMagnetReserved(stageHarness *test_case_harness.TestCaseHarness) error {

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

func closeConnection(conn net.Conn, logger *logger.Logger) {
    logger.Debugln("Closing connection")
    // Wait in case other party wants to send extra data
    time.Sleep(1 * time.Second)
    err := conn.Close()
    if err != nil {
        logger.Debugf("Error closing connection: %v", err)
    }
}

func handleReservedBytes(conn net.Conn, p PeerConnectionParams) {
    defer closeConnection(conn, p.logger)

    if err := receiveAndSendHandshake(conn, p); err != nil {
        return
    }

    if err := sendBitfieldMessage(conn, p.bitfield, p.logger); err != nil {
        return
    }

    if err := sendExtensionHandshake(conn, p.myMetadataExtensionID, p.metadataSizeBytes, p.logger); err != nil {
        return
    }
}
