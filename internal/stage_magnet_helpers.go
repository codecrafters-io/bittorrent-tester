package internal

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"net"

	logger "github.com/codecrafters-io/tester-utils/logger"
	"github.com/jackpal/bencode-go"
)

type MagnetTestParams struct {
	TrackerAddress         string
	PeerPort               int
	PeerAddress            string
	PeersResponse          []byte
	ExpectedInfoHash       [20]byte
	ExpectedReservedBytes  []byte
	ExpectedPeerID         [20]byte
	MyMetadataExtensionID  uint8
	MagnetUrlEncoded       string
	MagnetLinkInfo         MagnetTestTorrentInfo
	Logger                 *logger.Logger
}

type MagnetTestTorrentInfo struct {
	Filename           string
	InfoHashStr        string
	FileLengthBytes    int
	PieceLengthBytes   int
	MetadataSizeBytes  int
	Bitfield           []byte
	PieceHashes        []string
	ExpectedSha1       string
}

var magnetTestTorrents = []MagnetTestTorrentInfo {
	{
		Filename: "magnet1.gif",
		InfoHashStr: "ad42ce8109f54c99613ce38f9b4d87e70f24a165",
		FileLengthBytes: 636505,
		MetadataSizeBytes: 132,
		PieceLengthBytes: 262144,
		Bitfield: []byte {224},
		PieceHashes: []string {
			"3b46a96d9bc3716d1b75da91e6d753a793ad1cef",
			"eda417cb5c1cdbf841125c412da0bec9db8301f3",
			"422f45b1052e2d45da3e2a6516e1bb1f1db00733",
		},
		ExpectedSha1: "dd0f88e853321f08cd1d45423152d09014082437",
	},
	{
		Filename: "magnet2.gif",
		InfoHashStr: "3f994a835e090238873498636b98a3e78d1c34ca",
		FileLengthBytes: 79752,
		MetadataSizeBytes: 91,
		PieceLengthBytes: 262144,
		PieceHashes: []string {
			"d78a7f55ddd89fef477bc49d938bc7e4d94094f1",
		},
		Bitfield: []byte {128},
		ExpectedSha1: "d78a7f55ddd89fef477bc49d938bc7e4d94094f1",
	},
	{
		Filename: "magnet3.gif",
		InfoHashStr: "c5fb9894bdaba464811b088d806bdd611ba490af",
		FileLengthBytes: 629944,
		MetadataSizeBytes: 132,
		PieceLengthBytes: 262144,
		PieceHashes: []string {
			"ca80fd83ffb34d6e1bbd26a8ef6d305827f1cd0a",
			"707fd7c657f6d636f0583466c3cfe134ddb2c08a",
			"47076d104d214c0052960ef767262649a8af0ea8",
		},
		Bitfield: []byte {224},
		ExpectedSha1: "b1807e3d7920a559df2a2f0f555a404dec66a63e",
	},
}

func (m *MagnetTestParams) toTrackerParams() TrackerParams {
	return TrackerParams {
		trackerAddress: m.TrackerAddress,
		peersResponse: m.PeersResponse,
		expectedInfoHash: m.ExpectedInfoHash,
		fileLengthBytes: m.MagnetLinkInfo.FileLengthBytes,
		logger: m.Logger,
		myMetadataExtensionID: m.MyMetadataExtensionID,
	}
}

func (m *MagnetTestParams) toPeerConnectionParams() PeerConnectionParams {
	return PeerConnectionParams {
		address: m.PeerAddress,
		myPeerID: m.ExpectedPeerID,
		infoHash: m.ExpectedInfoHash,
		expectedReservedBytes: m.ExpectedReservedBytes,
		myMetadataExtensionID: m.MyMetadataExtensionID,
		metadataSizeBytes: m.MagnetLinkInfo.MetadataSizeBytes,
		bitfield: m.MagnetLinkInfo.Bitfield,
		magnetLink: m.MagnetLinkInfo,
		logger: m.Logger,
	}
}

func NewMagnetTestParams(magnetLink MagnetTestTorrentInfo, logger *logger.Logger) (*MagnetTestParams, error) {
	params := MagnetTestParams{}

	peerPort, err := findFreePort()
	if err != nil {
		return nil, fmt.Errorf("couldn't find free port: %s", err)
	}
	params.PeerPort = peerPort
	params.PeerAddress = fmt.Sprintf("127.0.0.1:%d", peerPort)
	params.PeersResponse = createPeersResponse("127.0.0.1", peerPort)

	trackerPort, err := findFreePort()
	if err != nil {
		return nil, fmt.Errorf("couldn't find free port: %s", err)
	}
	trackerAddress :=  fmt.Sprintf("127.0.0.1:%d", trackerPort)
	params.TrackerAddress = trackerAddress
	
	infoHashStr := magnetLink.InfoHashStr
	params.MagnetUrlEncoded = "magnet:?xt=urn:btih:" + infoHashStr + "&dn=" + magnetLink.Filename + "&tr=http%3A%2F%2F" + trackerAddress + "%2Fannounce"

	infoHash, err := decodeInfoHash(infoHashStr)
	if err != nil {
		return nil, fmt.Errorf("error decoding infohash: %v", err)
	}
	params.ExpectedInfoHash = infoHash

	expectedPeerID, err := randomHash()
	if err != nil {
		return nil, fmt.Errorf("error generating random peer id: %v", err)
	}
	params.ExpectedPeerID = expectedPeerID
	params.ExpectedReservedBytes = []byte{0, 0, 0, 0, 0, 16, 0, 0}
	params.MyMetadataExtensionID = uint8(rand.Intn(255) + 1)
	params.MagnetLinkInfo = magnetLink
	params.Logger = logger
	return &params, nil
}

func decodeInfoHash(infoHashStr string) ([20]byte, error) {
	var infoHash [20]byte
	decodedBytes, err := hex.DecodeString(infoHashStr)
	if err != nil {
		return infoHash, err
	}
	copy(infoHash[:], decodedBytes)
	return infoHash, nil
}

func receiveAndAssertExtensionHandshake(conn net.Conn, logger *logger.Logger) (id uint8, err error) {
	defer logOnExit(logger, &err)

	msg, err := receiveExtensionHandshake(conn, logger)
	if err != nil {
		return 0, fmt.Errorf("error receiving extension handshake: %v", err)
	}

	metadataExtensionID, err := assertExtensionHandshake(msg, logger)
	if err != nil {
		return 0, err
	}

	return metadataExtensionID, nil
}

func receiveExtensionHandshake(conn net.Conn, logger *logger.Logger) (*Message, error) {
	logger.Debugln("Waiting to receive extension handshake message")
	msg, err := readMessage(conn)
	if err != nil {
		return nil, err
	}
	logger.Infof("Received extension handshake with payload: %s", string(msg.Payload))
	return msg, nil
}

func assertExtensionHandshake(msg *Message, logger *logger.Logger) (uint8, error) {
	if msg.ID != MsgExtended {
		return 0, fmt.Errorf("expected message id: %d, actual: %d", MsgExtended, msg.ID)
	}

	if len(msg.Payload) < 2 {
		return 0, fmt.Errorf("expecting a larger payload size than %d", len(msg.Payload))
	}

	if msg.Payload[0] != 0 {
		return 0, fmt.Errorf("expected extension handshake message id: %d, actual: %d. First byte of payload indicates extension message id and it needs to be zero for extension handshake", 0, msg.Payload[0])
	}

	metadataExtensionID, err := extractMetadataExtensionID(msg, logger)
	if err != nil {
		return 0, err
	}
	
	return metadataExtensionID, nil
}

func extractMetadataExtensionID(msg *Message, logger *logger.Logger) (id uint8, err error) {
	defer logOnExit(logger, &err)

	logger.Debugln("Checking metadata extension id received")

	handshake, err := bencode.Decode(bytes.NewReader(msg.Payload[1:]))
	if err != nil {
		return 0, fmt.Errorf("error decoding bencoded dictionary in message payload starting at payload index 1, error message: %s", err)
	}
	dict, ok := handshake.(map[string]interface{})
	if !ok {
		return 0, errors.New("bencoded dictionary missing or wrong type in payload, expected a dictionary with string keys")
	}
	inner, exists := dict["m"]
	if !exists {
		return 0, errors.New("dictionary under key m is missing or wrong type")
	}

	innerDict, ok := inner.(map[string]interface{})
	if !ok {
		return 0, errors.New("dictionary under key m is of wrong type, expected a dictionary with string keys")
	}
	value, exists := innerDict["ut_metadata"]
	if exists {
		theirMetadataExtensionID, ok := value.(int64)
		if !ok {
			return 0, errors.New("value for ut_metadata needs to be an integer, it's wrong type")
		}
		if theirMetadataExtensionID <= 0 {
			return 0, errors.New("value for ut_metadata needs to be greater than zero")
		}
		theirMetadataExtensionIDUint8, err := safeINT64toUINT8(theirMetadataExtensionID)
		if err != nil {
			return 0, err
		}
		return theirMetadataExtensionIDUint8, nil
	} else {
		return 0, errors.New("ut_metadata key is missing in dictionary under key m during extension handshake")
	}
}

func safeINT64toUINT8(i any) (uint8, error) {
	if value, ok := i.(int64); ok {
        if value < 0 || value > 255 {
			return 0, fmt.Errorf("number out of range for uint8")
        } else {
            return uint8(value), nil
        }
	} else {
		return 0, fmt.Errorf("expected int64, received different type")
	}
}
