package internal

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path"

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

	tempDir, err := createTempDir(executable)
	if err != nil {
		logger.Errorf("Couldn't create temp directory")
		return err
	}

	// Copy sample response for HTTP server
	response := randomResponse()
	destinationPath := path.Join(tempDir, response.filename)
	sourcePath := getResponsePath(response.filename)
	if err = copyFile(sourcePath, destinationPath); err != nil {
		logger.Errorf("Couldn't copy sample response", err)
		return err
	}

	port, err := findFreePort()
	if err != nil {
		logger.Errorf("Error finding free port", err)
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
		logger.Errorf("Couldn't write torrent file", err)
		return err
	}

	go listenAndServePeersResponse(address, destinationPath, expectedInfoHash, logger)

	logger.Debugf("Running ./your_bittorrent.sh peers %s", torrentFileName)
	result, err := executable.Run("peers", torrentFileName)
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

func listenAndServePeersResponse(address string, responseFilePath string, expectedInfoHash [20]byte, logger *tester_utils.Logger) {
	http.HandleFunc("/announce/", func(w http.ResponseWriter, r *http.Request) {
		serveTrackerResponse(w, r, responseFilePath, expectedInfoHash, logger)
	})

	logger.Debugf("Server started on port %s...\n", address)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		logger.Errorf("Error:", err)
	}
}

func serveTrackerResponse(w http.ResponseWriter, r *http.Request, responseFilePath string, expectedInfoHash [20]byte, logger *tester_utils.Logger) {
	if r.Method != "GET" {
		logger.Errorf("HTTP method GET expected")
		http.Error(w, "HTTP method GET expected", http.StatusBadRequest)
		return
	}
	queryParams := r.URL.Query()
	if queryParams.Get("left") == "" {
		logger.Errorf("Required parameter missing: left")
		http.Error(w, "Required parameter missing: left", http.StatusBadRequest)
		return
	}
	infoHash := queryParams.Get("info_hash")
	if infoHash == "" {
		logger.Errorf("Required parameter missing: info_hash")
		http.Error(w, "Required parameter missing: info_hash", http.StatusBadRequest)
		return
	}
	if len(infoHash) == 40 {
		logger.Errorf("info_hash needs to be 20 bytes long, don't use hexadecimal")
		http.Error(w, "info_hash needs to be 20 bytes long, don't use hexadecimal", http.StatusBadRequest)
		return
	}
	if len(infoHash) != 20 {
		logger.Errorf("info_hash needs to be 20 bytes long, found: %d", len(infoHash))
		http.Error(w, "info_hash needs to be 20 bytes long", http.StatusBadRequest)
		return
	}

	receivedHash := []byte(infoHash)

	if !bytes.Equal(receivedHash[:], expectedInfoHash[:]) {
		logger.Errorf("info_hash correct length, but does not match expected value. It needs to be SHA-1 hash of the bencoded form of the info value from the metainfo file")
		http.Error(w, "info_hash correct length, but does not match expected value", http.StatusBadRequest)
		return
	}

	file, err := os.Open(responseFilePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	content, err := os.ReadFile(responseFilePath)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-bittorrent")
	w.Write(content)
}
