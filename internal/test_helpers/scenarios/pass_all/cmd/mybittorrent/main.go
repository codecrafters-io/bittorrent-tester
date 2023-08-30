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
	"github.com/codecrafters-io/grep-starter-go/p2p"
	"github.com/codecrafters-io/grep-starter-go/parser"
	"github.com/codecrafters-io/grep-starter-go/peers"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	//fmt.Println("Logs from your program will appear here!")

	command := os.Args[1]

	if command == "decode" {
		// Uncomment this block to pass the first stage
		//
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
	} else if command == "info" {
		Stage_infohash()
	} else if command == "peers" {
		Stage_tracker_get()
	} else if command == "handshake" {
		Stage_handshake()
	} else if command == "download_piece" {
		Stage_dl_piece()
	} else if command == "download" {
		Stage_dl_file()
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
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

	err = p2p.DownloadToFile(&torrentFile, outputPath, pieceIndex)
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

	err = p2p.DownloadToFile(&torrentFile, outputPath, -1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error downloading file: %v", err)
		return
	}
	// fmt.Printf("Downloaded %s to %s.", torrentPath, outputPath)
}
