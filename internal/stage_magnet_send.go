package internal

import (
    "errors"
    "net"

    "github.com/codecrafters-io/tester-utils/test_case_harness"
)

var handshakeChannel = make(chan bool)

func testMagnetSendExtendedHandshake(stageHarness *test_case_harness.TestCaseHarness) error {
    initRandom()

    logger := stageHarness.Logger
    executable := stageHarness.Executable

    magnetLink := randomMagnetLink()
    params, err := NewMagnetTestParams(magnetLink, logger)
    if err != nil {
        return err
    }

    go listenAndServeTrackerResponse(params.toTrackerParams())
    go waitAndHandlePeerConnection(params.toPeerConnectionParams(), handleSendExtensionHandshake)

    logger.Infof("Running ./your_bittorrent.sh magnet_handshake %q", params.MagnetUrlEncoded)
    result, err := executable.Run("magnet_handshake", params.MagnetUrlEncoded)
    if err != nil {
        return err
    }
    
    if err = assertExitCode(result, 0); err != nil {
        return err
    }

    success := <-handshakeChannel
    if success {
        return nil
    }

    return errors.New("extension handshake was not received")
}

func handleSendExtensionHandshake(conn net.Conn, params PeerConnectionParams) {
    defer conn.Close()
    logger := params.logger

    if err := receiveAndSendHandshake(conn, params); err != nil {
        return
    }

    if err := sendBitfieldMessage(conn, params.bitfield, logger); err != nil {
        return
    }

    if err := sendExtensionHandshake(conn, params.myMetadataExtensionID, params.metadataSizeBytes, logger); err != nil {
        return
    }

    if _, err := receiveAndAssertExtensionHandshake(conn, logger); err != nil {
        return
    }

    handshakeChannel <- true
}
