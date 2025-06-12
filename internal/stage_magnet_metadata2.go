package internal

import (
    "fmt"
    "net"

    "github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testMagnetSendMetadata(stageHarness *test_case_harness.TestCaseHarness) error {

    logger := stageHarness.Logger
    executable := stageHarness.Executable

    magnetLink := randomMagnetLink()
    params, err := NewMagnetTestParams(magnetLink, logger)
    if err != nil {
        return err
    }

    go listenAndServeTrackerResponse(params.toTrackerParams())
    go waitAndHandlePeerConnection(params.toPeerConnectionParams(), handleSendMetadata)

    logger.Infof("Running ./your_bittorrent.sh magnet_info %q", params.MagnetUrlEncoded)
    result, err := executable.Run("magnet_info", params.MagnetUrlEncoded)
    if err != nil {
        return err
    }

    if err = assertExitCode(result, 0); err != nil {
        return err
    }

    expected := fmt.Sprintf("Tracker URL: http://%s/announce", params.TrackerAddress)
    if err = assertStdoutContains(result, expected); err != nil {
        return err
    }

    logger.Successln("✓ Tracker URL is correct.")

    expected = fmt.Sprintf("Length: %d", params.MagnetLinkInfo.FileLengthBytes)
    if err = assertStdoutContains(result, expected); err != nil {
        return err
    }

    logger.Successln("✓ Length is correct.")

    expected = fmt.Sprintf("Info Hash: %s", params.MagnetLinkInfo.InfoHashStr)
    if err = assertStdoutContains(result, expected); err != nil {
        return err
    }

    logger.Successln("✓ Info Hash is correct.")

    expected = fmt.Sprintf("Piece Length: %d", params.MagnetLinkInfo.PieceLengthBytes)
    if err = assertStdoutContains(result, expected); err != nil {
        return err
    }

    logger.Successln("✓ Piece Length is correct.")

    pieceHashes := params.MagnetLinkInfo.PieceHashes
    for _, pieceHash := range pieceHashes {
        if err = assertStdoutContains(result, pieceHash); err != nil {
            return err
        }
    }

    logger.Successln("✓ Piece Hashes are correct.")

    return nil
}

func handleSendMetadata(conn net.Conn, params PeerConnectionParams) {
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

    theirMetadataExtensionID, err := receiveAndAssertExtensionHandshake(conn, logger)
    if err != nil {
        return
    }

    if err := readMetadataRequest(conn, logger); err != nil {
        return
    }

    if err := sendMetadataResponse(conn, theirMetadataExtensionID, params.magnetLink, logger); err != nil {
        logger.Errorln(err.Error())
        return
    }
}

