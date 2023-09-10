package internal

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strconv"

	tester_utils "github.com/codecrafters-io/tester-utils"
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

func testDiscoverPeers(stageHarness *tester_utils.StageHarness) error {
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

	go listenAndServePeersResponse(address, destinationPath, expectedInfoHash, fileLengthBytes, logger)

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

func listenAndServePeersResponse(address string, responseFilePath string, expectedInfoHash [20]byte, fileLengthBytes int, logger *tester_utils.Logger) {
	http.HandleFunc("/announce/", func(w http.ResponseWriter, r *http.Request) {
		serveTrackerResponse(w, r, responseFilePath, expectedInfoHash, fileLengthBytes, logger)
	})

	logger.Debugf("Server started on port %s...\n", address)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		logger.Errorf("Error: %s", err)
	}
}

func serveTrackerResponse(w http.ResponseWriter, r *http.Request, responseFilePath string, expectedInfoHash [20]byte, fileLengthBytes int, logger *tester_utils.Logger) {
	if r.Method != "GET" {
		logger.Errorln("HTTP method GET expected")
		http.Error(w, "HTTP method GET expected", http.StatusMethodNotAllowed)
		return
	}
	queryParams := r.URL.Query()
	left := queryParams.Get("left")
	if left == "" {
		logger.Errorln("Required parameter missing: left")
		w.Write([]byte("d14:failure reason31:failed to parse parameter: lefte"))
		return
	}
	leftNumber, err := strconv.Atoi(left)
	if err != nil {
		logger.Errorf("left needs to be a numeric value, received: %s", left)
		w.Write([]byte("d14:failure reason31:failed to parse parameter: lefte"))
		return
	} else if leftNumber > fileLengthBytes {
		logger.Errorf("left needs to be less than or equal to file length (%d bytes), received: %s", fileLengthBytes, left)
		w.Write([]byte("d14:failure reason27:provided invalid left valuee"))
		return
	}

	port := queryParams.Get("port")
	if port == "" {
		logger.Errorln("Required parameter missing: port")
		w.Write([]byte("d14:failure reason31:failed to parse parameter: porte"))
		return
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 0 || portNumber > 65536 {
		logger.Errorf("port needs to be between 0 and 65536, received: %s", port)
		w.Write([]byte("d14:failure reason31:failed to parse parameter: porte"))
		return
	}

	downloaded := queryParams.Get("downloaded")
	if downloaded == "" {
		logger.Errorln("Required parameter missing: downloaded")
		w.Write([]byte("d14:failure reason37:failed to parse parameter: downloadedede"))
		return
	}
	if _, err := strconv.Atoi(downloaded); err != nil {
		logger.Errorf("downloaded needs to be a numeric value, received: %s", downloaded)
		w.Write([]byte("d14:failure reason37:failed to parse parameter: downloadedede"))
		return
	}

	uploaded := queryParams.Get("uploaded")
	if uploaded == "" {
		logger.Errorln("Required parameter missing: uploaded")
		w.Write([]byte("d14:failure reason35:failed to parse parameter: uploadede"))
		return
	}
	if _, err := strconv.Atoi(uploaded); err != nil {
		logger.Errorf("uploaded needs to be a numeric value, received: %s", uploaded)
		w.Write([]byte("d14:failure reason35:failed to parse parameter: uploadede"))
		return
	}

	if queryParams.Get("compact") == "" {
		logger.Errorln("Required parameter missing: compact")
		w.Write([]byte("d14:failure reason34:failed to parse parameter: compacte"))
		return
	} else if queryParams.Get("compact") != "1" {
		logger.Errorln("compact parameter value needs to be 1 for compact representation of peer list")
		w.Write([]byte("d14:failure reason34:failed to parse parameter: compacte"))
		return
	}

	peerId := queryParams.Get("peer_id")
	if peerId == "" {
		logger.Errorln("Required parameter missing: peer_id")
		w.Write([]byte("d14:failure reason34:failed to parse parameter: peer_ide"))
		return
	} else if len(peerId) != 20 {
		logger.Errorln("peer_id needs to be a string of length 20")
		w.Write([]byte("d14:failure reason31:failed to provide valid peer_ide"))
		return
	}

	infoHash := queryParams.Get("info_hash")
	if infoHash == "" {
		logger.Errorln("Required parameter missing: info_hash")
		w.Write([]byte("d14:failure reason31:no info_hash parameter suppliede"))
		return
	}
	if len(infoHash) == 40 {
		logger.Errorln("info_hash needs to be 20 bytes long, don't use hexadecimal")
		w.Write([]byte("d14:failure reason25:provided invalid infohashe"))
		return
	}
	if len(infoHash) != 20 {
		logger.Errorf("info_hash needs to be 20 bytes long, found: %d", len(infoHash))
		w.Write([]byte("d14:failure reason25:provided invalid infohashe"))
		return
	}

	receivedHash := []byte(infoHash)

	if !bytes.Equal(receivedHash[:], expectedInfoHash[:]) {
		logger.Errorln("info_hash correct length, but does not match expected value. It needs to be SHA-1 hash of the bencoded form of the info value from the metainfo file")
		w.Write([]byte("d14:failure reason25:provided invalid infohashe"))
		return
	}

	file, err := os.Open(responseFilePath)
	if err != nil {
		http.Error(w, "Internal server error, file not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	content, err := os.ReadFile(responseFilePath)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Write(content)
}
