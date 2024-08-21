package internal

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"

	logger "github.com/codecrafters-io/tester-utils/logger"
	"github.com/jackpal/bencode-go"
)

type ConnectionHandler func(net.Conn, PeerConnectionParams)

type PeerConnectionParams struct {
	address string
	myPeerID [20]byte
	infoHash [20]byte
	expectedReservedBytes []byte
	myMetadataExtensionID uint8
	metadataSizeBytes int
	bitfield []byte
	magnetLink MagnetTestTorrentInfo
	logger *logger.Logger
}

type TrackerParams struct {
	trackerAddress string
	peersResponse []byte
	expectedInfoHash [20]byte
	fileLengthBytes int
	logger *logger.Logger
	myMetadataExtensionID uint8
}

var samplePieceHashes = []string{
	"ddf33172599fda84f0a209a3034f79f0b8aa5e22",
	"795a618a1ee5275e952843b01a56ae4e142752ef",
	"cdae2ef532d611a46b2cf7b64d578c09b3ac0b6e",
	"098dadc0c19436f1927ea27b90eb18b1a2820a23",
	"8fa5355419886d9ec56e86cd7791343e9379de18",
	"1caeaceb15fd1134b1b4b21fad04125b227b4dcf",
	"fa586e20d579a4de76090e12bd0a3d9b1c539f3e",
	"aec2d7eb1db539c2a9d24d023fb916b79234b769",
}

type TestTorrentInfo struct {
	filename       string
	outputFilename string
	tracker        string
	infohash       string
	length         int64
	expectedSha1   string
	incorrectSha1  []string
}

var testTorrents = []TestTorrentInfo{
	{
		filename:       "codercat.gif.torrent",
		outputFilename: "codercat.gif",
		tracker:        "http://bittorrent-test-tracker.codecrafters.io/announce",
		infohash:       "c77829d2a77d6516f88cd7a3de1a26abcbfab0db",
		length:         2994120,
		expectedSha1:   "89d5dcbb92d31f040f8fe42b559fd6ec7f4d83a5",
		incorrectSha1: []string{
			"9fe338a7333eaf95533895e0340df633f0985462",
			"38dd803e5272c58ef1edeb808d160e9b02c208bd",
			"4ae8b721254a1e9363ad8f2458d184075b18ff50",
			"c90f0c6574de821be40cf741357dafb99429ca70",
			"1a390773ed42171e63e083ed1fffb860604fbba4",
			"d53e7626e60c1acc1d3f1269955e99e3122c19ee",
			"d2087fec76212da342e864351950c92a363dfedc",
			"d4c98d43da1038fe79d2351232e68910b123e603",
			"cb46db2fadb568903716c5d1f007546347da68e4",
			"b4be41970720fd343cfe2aefe427f4711486d44c",
			"a89ff934345bba1c6fff9fdec529407a8487ea6e",
			"d35e869ac9e7a49cb63c50e6f95cf174e2288fe6",
			"f46ae8bb99685dfb1ceaa73eb8699f214bc3917a",
			"54d46efe1684a3704eb6371d8b303c1734d55aca",
			"6b54d951b7eb715002aef478846741735f4ed6ca",
			"ef8e41743f4c52aedc99852dcf6ee06adc80fba9",
			"b41b700691c17202f3f1c8bbad88db8a594f9406",
			"3c8fa675dd60c4b714b8092d08eca65fbf1c8224",
			"c9620ee5d267ef0d12ac629fff6d2a08cae16e04",
			"bacb060807196ddf9a608c02a9b51d06f1f5630b",
			"b9f6595abc73c2a9d7e4cd8d58b9a6b0daf3a2c2",
			"ddff12f2cd2b8d7b17eb5ddddf723203f57a4583",
			"751c1cce843365f2c0243353dd37c663fc136a38",
		},
	},
	{
		filename:       "congratulations.gif.torrent",
		outputFilename: "congratulations.gif",
		tracker:        "http://bittorrent-test-tracker.codecrafters.io/announce",
		infohash:       "1cad4a486798d952614c394eb15e75bec587fd08",
		length:         820892,
		expectedSha1:   "fe3cc9002bc84c4776c3a962a717244c9cb962c0",
		incorrectSha1: []string{
			"a51f27f7098ad52b23d43b7377a038061dc86d82",
			"17f995b0a08a9546fa21248a7b3dd7b88900550b",
			"0475ae5ced66e355114367ca3531513bc741d584",
			"f3b7a5af63de420c195bff73ef9a1eb6adbe316c",
			"25c979353bd0a94561234dc98c21e79fb2d3f36a",
			"4dc9c94a9814dffef84eccb30d1f402a93c967b4",
			"ed63aa89af68e20c7daa8958fb2e6549ad9d46b9",
			"f7dbfd0ea39c8867ec41a43cdae8f94ecf041c5c",
			"c50cbe4589fd7cbb13f3702ec95415e5672cc16b",
			"8ce1b44f4cb957c82a09e1ea9816240cf75858cb",
			"448875e4899066578afc755a5896a255b83b8765",
			"f83036e6fe380a97a93e47d916cf344ad1137f01",
			"e7df0e8ee5f434bbafa3ea1899c833f07b68e34f",
			"06ecfa7284cd8e6d23cb1c79bf74e039c1ca46d0",
			"151da9dc8374787dd823a2e0de40ee48027c5619",
			"31b6f8167791cdab14c742b24bcfd625f088afaa",
			"364fda25722abb30ef7260f8d9f1dcedd6fe442a",
			"9ffc716aa95846cf8b4f550b83b8717dad473f2c",
			"77288778c94881d13307446fdc4671853c36d3e4",
			"6af4d48cf003ad80849cffe25745417651d57944",
			"88c0eedcf9bdbaf9093e138703b9c217f0176d07",
			"1fa4e4e1cf5842f20ec3a5c8799b65fef08338df",
			"adceada73d45d1e32ca861c83d2ba6f247643c15",
		},
	},
	{
		filename:       "itsworking.gif.torrent",
		outputFilename: "itsworking.gif",
		tracker:        "http://bittorrent-test-tracker.codecrafters.io/announce",
		infohash:       "70edcac2611a8829ebf467a6849f5d8408d9d8f4",
		length:         2549700,
		expectedSha1:   "683e899db9d7a38e50eb87874b62f7fdd0c14c9c",
		incorrectSha1: []string{
			"031ca6e6c27062f87935fad1cb4ceec94f4b7342",
			"ef1fb68a9ce29987375d4a9d27f70d3f39466292",
			"f580128989c01fba318fa6079712d2a6e5629d79",
			"18d3c1d13cbda0b453c7d80d3b0af0b5b693109c",
			"8de41c8ec28703aaa57a5de4fa0cdb3a63809a9e",
			"df7ee75594cbf87f5c01f5b8e806fc378c71c881",
			"31fc5c2997edef37e5b6c92faba6324aa930315c",
			"d2aa57b57d425815be6021ac30d4a6fb30bd2914",
			"c0096df565c3c5d8f7ded6b1ddad01531b1507bd",
			"294673ce148426c79afede0a77497f9df0e797ab",
			"c6566181dd51e709a1125c329008f144e00e011e",
			"f7aafeb3f9e71bcdb886c80d2e79e09989f31eac",
			"79fee8553cb38e3c8f012b8c5a34ac8a1ceba2a6",
			"a17b52dd5dafbeb6407b115a7ad8f72247ed0ddb",
			"cb832682ddc645b95cf5926cce3d30cd54375584",
			"46118c1764e8c162f7cf11fa219412b4739fd3db",
			"9e8634490c61f49d1a1027de12f8f87e4be9a577",
			"05ef950cdbed43ea342e892e2ea392ad1df985fd",
			"b9750268e247dda0ae9e470ff009449e239910d1",
			"900b74c76c3a9ab3fe3e7baf81bad6b184ac348f",
			"89fad97dbde4097e4498c32d31da435b5c7ab95b",
			"c36b30d886a1c164ecf9c3a31019f8983525ed6f",
			"562de033449e4e5f09c755c7a54e0ff1ee4f0f75",
		},
	},
}

func copyTorrent(tempDir string, torrentFilename string) error {
	destinationPath := path.Join(tempDir, torrentFilename)
	sourcePath := getTorrentPath(torrentFilename)
	if err := copyFile(sourcePath, destinationPath); err != nil {
		return err
	}
	return nil
}

func randomTorrent() TestTorrentInfo {
	return testTorrents[rand.Intn(len(testTorrents))]
}

func randomMagnetLink() MagnetTestTorrentInfo {
	return magnetTestTorrents[rand.Intn(len(magnetTestTorrents))]
}

func getTorrentPath(filename string) string {
	return path.Join(getProjectDir(), "torrents", filename)
}

func getResponsePath(filename string) string {
	return path.Join(getProjectDir(), "response", filename)
}

func getProjectDir() string {
	return os.Getenv("TESTER_DIR")
}

func findFreePort() (int, error) {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer lis.Close()

	addr := lis.Addr()
	return addr.(*net.TCPAddr).Port, nil
}

func copyFile(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)

	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}

func calculateSHA1(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	hashBytes := hash.Sum(nil)
	return hex.EncodeToString(hashBytes), nil
}

func listenAndServeTrackerResponse(p TrackerParams) {
	logger := p.logger
	mux := http.NewServeMux()
	mux.HandleFunc("/announce", func(w http.ResponseWriter, r *http.Request) {
		serveTrackerResponse(w, r, p.peersResponse, p.expectedInfoHash, p.fileLengthBytes, p.logger)
	})
	
	// Redirect /announce/ to /announce while preserving query parameters 
	mux.HandleFunc("/announce/", func(w http.ResponseWriter, r *http.Request) {
		parsedURL, err := url.Parse(r.URL.String())
		if err != nil {
			logger.Errorf("Error parsing URL: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		parsedURL.Path = "/announce"
		http.Redirect(w, r, parsedURL.String(), http.StatusMovedPermanently)
	})

	logger.Debugf("Tracker started on address %s...\n", p.trackerAddress)
	err := http.ListenAndServe(p.trackerAddress, mux)
	if err != nil {
		logger.Errorf("Error: %s", err)
	}
}

func serveTrackerResponse(w http.ResponseWriter, r *http.Request, responseContent []byte, expectedInfoHash [20]byte, fileLengthBytes int, logger *logger.Logger) {
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
		logger.Errorf("left parameter needs to be a numeric value, received: %s", left)
		w.Write([]byte("d14:failure reason31:failed to parse parameter: lefte"))
		return
	} else if leftNumber == 0 {
		logger.Errorf("left parameter needs to be greater than zero to receive peers, received: %s", left)
		w.Write([]byte("d8:completei4e10:incompletei0e8:intervali60e12:min intervali60ee"))
		return
	} else if leftNumber > fileLengthBytes {
		logger.Errorf("left parameter needs to be less than or equal to file length (%d bytes), received: %s", fileLengthBytes, left)
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
		logger.Errorln("info_hash correct length, but does not match expected value. It needs to be SHA-1 of the bencoded info dictionary from the torrent file")
		w.Write([]byte("d14:failure reason25:provided invalid infohashe"))
		return
	}

	w.Write(responseContent)
}

func createPeersResponse(peerIP string, peerPort int) []byte {
	peerBytes := make([]byte, 6)
	peerIPAddress := net.ParseIP(peerIP).To4()

	copy(peerBytes[:4], peerIPAddress)
	peerBytes[4] = byte(peerPort >> 8)
	peerBytes[5] = byte(peerPort)

	response := map[string]interface{}{
		"complete":    1,
		"incomplete":  0,
		"mininterval": 1800,
		"peers":       peerBytes,
	}

	var buf bytes.Buffer
	if err := bencode.Marshal(&buf, response); err != nil {
		fmt.Println("Error encoding bencoded response:", err)
		return nil
	}
	return buf.Bytes()
}

func receiveAndSendHandshake(conn net.Conn, peer PeerConnectionParams) (err error) {
	defer logOnExit(peer.logger, &err)

	logger := peer.logger
	handshake, err := readHandshake(conn, logger)
	if err != nil {
		return fmt.Errorf("error reading handshake: %s", err)
	}
	
	if !bytes.Equal(handshake.Reserved[:], peer.expectedReservedBytes) {
		return fmt.Errorf("did you send reserved bytes? expected bytes: %v but received: %v", peer.expectedReservedBytes, handshake.Reserved)
	}

	if !bytes.Equal(handshake.InfoHash[:], peer.infoHash[:]) {
		return fmt.Errorf("expected infohash %x but got %x", peer.infoHash, handshake.InfoHash)
	}

	logger.Debugf("Received handshake: [infohash: %x, peer_id: %x]\n", handshake.InfoHash, handshake.PeerID)
	logger.Debugf("Sending back handshake with peer_id: %x", peer.myPeerID)
	
	var reservedBytes [8]byte 
	copy(reservedBytes[:], peer.expectedReservedBytes)

	err = sendHandshake(conn, reservedBytes, handshake.InfoHash, peer.myPeerID)
	if err != nil {
		return err
	}
	return nil
}

func waitAndHandlePeerConnection(p PeerConnectionParams, handler ConnectionHandler) {
	logger := p.logger
	logger.Debugf("Peer listening on address: %s", p.address)
	listener, err := net.Listen("tcp", p.address)
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
		handler(conn, p)
	}
}

func randomHash() ([20]byte, error) {
	var hash [20]byte
	if _, err := rand.Read(hash[:]); err != nil {
		return [20]byte{}, err
	}
	return hash, nil
}