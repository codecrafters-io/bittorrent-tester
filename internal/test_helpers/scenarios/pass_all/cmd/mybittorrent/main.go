package main

import (
	// Uncomment this line to pass the first stage
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/grep-starter-go/bencode"
	"github.com/codecrafters-io/grep-starter-go/client"
	"github.com/codecrafters-io/grep-starter-go/magnet"
	"github.com/codecrafters-io/grep-starter-go/p2p"
	"github.com/codecrafters-io/grep-starter-go/parser"
	"github.com/codecrafters-io/grep-starter-go/peers"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	//fmt.Println("Logs from your program will appear here!")

	command := os.Args[1]

	switch command {
	case "decode":
		Stage_init()
	case "info":
		Stage_infohash()
	case "peers":
		Stage_tracker_get()
	case "handshake":
		Stage_handshake()
	case "download_piece":
		Stage_dl_piece()
	case "download":
		Stage_dl_file()
	case "magnet_parse":
		Stage_magnet_parse()
	case "magnet_handshake":
		Stage_magnet_handshake(false)
	case "magnet_info":
		Stage_magnet_handshake(true)
	case "magnet_download_piece":
		Stage_magnet_dl_piece()
	case "magnet_download":
		Stage_magnet_dl_file()
	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

func Stage_init() {
	data := os.Args[2]
	//fmt.Println("data", data)
	sarp3, err := bencode.Decode(strings.NewReader(data))
	if err != nil {
		fmt.Println("Error decoding", err)
		return
	}
	jsonBytes, err := json.Marshal(sarp3)
	if err != nil {
		fmt.Println("Error encoding JSON", err)
		return
	}
	fmt.Println(string(jsonBytes))
}

func Stage_infohash() {
	fileName := os.Args[2]
	torrentFile, err := parser.Open(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening file2: %v", err)
		return
	}
	fmt.Println("Tracker URL:", torrentFile.Announce)
	fmt.Println("Length:", torrentFile.Length)
	fmt.Println("Info Hash:", fmt.Sprintf("%x", torrentFile.InfoHash))
	fmt.Println("Piece Length:", torrentFile.PieceLength)
	fmt.Println("Piece Hashes:")
	for _, hash := range torrentFile.PieceHashes {
		fmt.Printf("%x\n", hash)
	}
}

func Stage_tracker_get() {
	fileName := os.Args[2]
	torrentFile, err := parser.Open(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening file2: %v", err)
		return
	}
	//fmt.Println("Tracker URL:", torrentFile.Announce)
	var peerID [20]byte
	_, err = rand.Read(peerID[:])
	if err != nil {
		fmt.Println("error generating peer id")
		return
	}

	peers, error := client.RequestPeers(&torrentFile, peerID, peers.Port)
	if error != nil {
		fmt.Println("error connecting", error)
		return
	}
	peer_str := make([]string, len(peers))
	for i, peer := range peers {
		// Process the element, e.g., convert to uppercase
		peer_str[i] = peer.String()
		fmt.Println(peer_str[i])
	}
}

func Stage_handshake() {
	//./your_bittorrent.sh bitfield test.torrent ip:port
	fileName := os.Args[2]
	peer := os.Args[3]
	torrentFile, err := parser.Open(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening file2: %v", err)
		return
	}

	var peerID [20]byte
	_, err = rand.Read(peerID[:])
	if err != nil {
		fmt.Println("error generating peer id")
		return
	}
	//fmt.Printf("my peer id: %x\n", peerID)
	/*
		peers, error := client.RequestPeers(&torrentFile, peerID, peers.Port)
		if error != nil {
			fmt.Println("error connecting", error)
		}

		m.Seed(time.Now().UnixNano())
		random := m.Intn(len(peers))
		first := peers[random]
		fmt.Printf("random: %d peer: %s\n", random, first)
	*/

	p2p.TalkToPeer(torrentFile, peer, peerID, torrentFile.InfoHash)
}

func Stage_dl_piece() {
	outputPath := os.Args[3]
	torrentPath := os.Args[4]
	pieceIndexStr := os.Args[5]

	pieceIndex, err := strconv.Atoi(pieceIndexStr)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	torrentFile, err := parser.Open(torrentPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening file2: %v", err)
		return
	}

	myPeerID := generateMyPeerID()
	err = p2p.DownloadToFile(&torrentFile, outputPath, pieceIndex, nil, myPeerID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error downloading file: %v", err)
		return
	}
	// fmt.Printf("Piece %s downloaded to %s.", pieceIndexStr, outputPath)
}

func Stage_dl_file() {
	outputPath := os.Args[3]
	torrentPath := os.Args[4]
	torrentFile, err := parser.Open(torrentPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening file2: %v", err)
		return
	}

	myPeerID := generateMyPeerID()
	err = p2p.DownloadToFile(&torrentFile, outputPath, -1, nil, myPeerID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error downloading file: %v", err)
		return
	}
	// fmt.Printf("Downloaded %s to %s.", torrentPath, outputPath)
}

func Stage_magnet_parse() {
	magnetURL := os.Args[2]
	link, err := magnet.Parse(magnetURL)
	if err != nil {
		fmt.Println("Error", err)
	}

	for _, tracker := range(link.Trackers) {
		fmt.Printf("Tracker URL: %s\n", tracker)
	}
	fmt.Printf("Info Hash: %s\n", link.InfoHash)
}

func Stage_magnet_handshake(shouldSendMetadata bool) {
	magnetUrl := os.Args[2]
	
	myPeerID := generateMyPeerID()
	peers, err := p2p.FetchPeers(magnetUrl, myPeerID)
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	peer := peers[0].String()
	myTorrent, err := p2p.FetchTorrentMetadata(magnetUrl, peer, myPeerID, shouldSendMetadata)
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	if shouldSendMetadata {
		fmt.Printf("Tracker URL: %s\n", myTorrent.Announce)
		fmt.Printf("Length: %d\n", myTorrent.Length)
		fmt.Printf("Info Hash: %x\n", myTorrent.InfoHash)
		fmt.Printf("Piece Length: %d\n", myTorrent.PieceLength)
		fmt.Printf("Piece Hashes:\n")
		for _, hash := range(myTorrent.PieceHashes) {
			fmt.Printf("%x\n", hash)
		}
	}
}

func generateMyPeerID() [20]byte {
	myid := []byte("00112233445566778899")
	var peerID [20]byte
	copy(peerID[:], myid)
	return peerID
}

func Stage_magnet_dl_piece() {
	outputPath := os.Args[3]
	magnetUrl := os.Args[4]
	pieceIndexStr := os.Args[5]

	pieceIndex, err := strconv.Atoi(pieceIndexStr)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	myPeerID := generateMyPeerID()
	peers, err := p2p.FetchPeers(magnetUrl, myPeerID)
	if err != nil {
		fmt.Println("Error", err)
		return
	}
	
	peer := peers[0].String()
	torrentFile, err := p2p.FetchTorrentMetadata(magnetUrl, peer, myPeerID, true)
	if err != nil {
		return
	}

	err = p2p.DownloadToFile(torrentFile, outputPath, pieceIndex, peers, myPeerID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error downloading file: %v", err)
		return
	}
	fmt.Printf("Piece %s downloaded to %s.", pieceIndexStr, outputPath)
}

func Stage_magnet_dl_file() {
	outputPath := os.Args[3]
	magnetUrl := os.Args[4]

	myPeerID := generateMyPeerID()
	peers, err := p2p.FetchPeers(magnetUrl, myPeerID)
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	peer := peers[0].String()
	torrentFile, err := p2p.FetchTorrentMetadata(magnetUrl, peer, myPeerID, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error downloading file: %v", err)
		return
	}

	err = p2p.DownloadToFile(torrentFile, outputPath, -1, peers, myPeerID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error downloading file: %v", err)
		return
	}
	fmt.Printf("Downloaded to %s.", outputPath)
}