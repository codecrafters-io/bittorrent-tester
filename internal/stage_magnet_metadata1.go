package internal

import (
    "bytes"
    "errors"
    "fmt"
    "net"
    "time"

    logger "github.com/codecrafters-io/tester-utils/logger"
    "github.com/codecrafters-io/tester-utils/test_case_harness"
    "github.com/jackpal/bencode-go"
)

var metadataRequestChannel = make(chan bool)

func testMagnetRequestMetadata(stageHarness *test_case_harness.TestCaseHarness) error {
    logger := stageHarness.Logger
    executable := stageHarness.Executable

    magnetLink := randomMagnetLink()
    params, err := NewMagnetTestParams(magnetLink, logger)
    if err != nil {
        return err
    }
    go listenAndServeTrackerResponse(params.toTrackerParams())
    go waitAndHandlePeerConnection(params.toPeerConnectionParams(), handleMetadataRequest)

    logger.Infof("Running ./your_bittorrent.sh magnet_info %q", params.MagnetUrlEncoded)
    result, err := executable.Run("magnet_info", params.MagnetUrlEncoded)
    if err != nil {
        return err
    }

    if err = assertExitCode(result, 0); err != nil {
        return err
    }

    success := <-metadataRequestChannel
    if success {
        return nil
    }

    return errors.New("metadata request not received")
}

func handleMetadataRequest(conn net.Conn, params PeerConnectionParams) {
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

    // Send in case other party is waiting for this to terminate
    sendMetadataResponse(conn, theirMetadataExtensionID, params.magnetLink, logger)

    metadataRequestChannel <- true
}

func readMetadataRequest(conn net.Conn, logger *logger.Logger) (err error) {
    defer logOnExit(logger, &err)

    logger.Debugln("Waiting to receive metadata request")
    msg, err := readMessage(conn, logger)
    if err != nil {
        return fmt.Errorf("error reading message: %v", err.Error())
    }

    if msg.ID != MsgExtended {
        return fmt.Errorf("incorrect message ID, expected=%d, actual=%d", MsgExtended, msg.ID)
    }

    logger.Debugf("Received payload: %s", string(msg.Payload))

    // First byte of the payload is extension message id
    // Rest of payload will be a bencoded dictionary like: d8:msg_typei0e5:piecei0ee
    decoded, err := bencode.Decode(bytes.NewReader(msg.Payload[1:]))
    if err != nil {
        return fmt.Errorf("error decoding metadata request message payload: %v", err)
    }

    dict, ok := decoded.(map[string]interface{})
    if !ok {
        return errors.New("expected dictionary with string keys not found in payload")
    }

    messageType, exists := dict["msg_type"]
    if !exists {
        return errors.New("expected msg_type key not found in dictionary")
    }

    messageTypeInt, ok := messageType.(int64)
    if !ok {
        return errors.New("expected msg_type to be an integer")
    }

    if messageTypeInt != 0 {
        return fmt.Errorf("expected msg_type key with value=0, actual value=%v", messageType)
    }

    pieceIndex, exists := dict["piece"]
    if !exists {
        return errors.New("expected piece key not found in dictionary")
    }

    pieceIndexInt, ok := pieceIndex.(int64)
    if !ok {
        return errors.New("expected value for piece key to be an integer")
    }

    if pieceIndexInt != 0 {
        return fmt.Errorf("expected piece key with value=0, actual value=%v", pieceIndex)
    }

    return nil
}

func sendMetadataResponse(conn net.Conn, metadataExtensionID uint8, magnetLink MagnetTestTorrentInfo, logger *logger.Logger) error {
    defer conn.SetDeadline(time.Time{}) // Disable the deadline

    logger.Debugln("Sending metadata response")

    info := TorrentFileInfo{
        Name: magnetLink.Filename,
        Length: magnetLink.FileLengthBytes,
        Pieces: toPiecesStr(magnetLink.PieceHashes),
        PieceLength: magnetLink.PieceLengthBytes,
    }
    m, err := createMetadataDataMessage(metadataExtensionID, magnetLink.MetadataSizeBytes, 0, info, logger)
    if err != nil {
        return err
    }

    bytes := m.Serialize()

    conn.SetDeadline(time.Now().Add(3 * time.Second))
    _, err = conn.Write(bytes)
    if err != nil {
        return fmt.Errorf("error sending metadata response: %v", err)
    }
    return nil
}