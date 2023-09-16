package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/bencode"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/client"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/metafile"
)

func main() {
	command := os.Args[1]
	if command == "peers" {
		filename := os.Args[2]
		meta, err := metafile.NewMetafile(filename)
		if err != nil {
			println(fmt.Sprintf("Problem with the metafile %s: %s", filename, err.Error()))
			return
		}

		self := client.NewClient(meta)
		for i := 0; i < 3; i++ {
			err = self.UpdatePeers()
			if err != nil {
				time.Sleep(3 * time.Second)
				continue
			}
			err = nil
		}

		if err != nil {
			println(err.Error())
			return
		}

		self.PrintPeers()
	} else if command == "info" {
		filename := os.Args[2]
		meta, err := metafile.NewMetafile(filename)
		if err != nil {
			println(fmt.Sprintf("Problem with the metafile %s: %s", filename, err.Error()))
			return
		}
		fmt.Printf("Tracker URL: %s\n", meta.Tracker)
		fmt.Printf("Length: %d\n", meta.Length)
		fmt.Printf("Info Hash: %x\n", meta.HashedInfo)
		fmt.Printf("Piece Length: %d\n", meta.PieceLength)
		fmt.Printf("Piece hashes:\n")
		for _, hash := range meta.PiecesHash {
			fmt.Print(hash)
		}
	} else if command == "decode" {
		bencodedValue := os.Args[2]

		decoded, _, err := bencode.DecodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, err := json.Marshal(decoded)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(jsonOutput))
	} else if command == "handshake" {
		filename := os.Args[2]
		peer := os.Args[3]
		meta, err := metafile.NewMetafile(filename)
		if err != nil {
			println(fmt.Sprintf("Problem with the metafile %s: %s", filename, err.Error()))
			return
		}
		self := client.NewClient(meta)
		err = self.UpdatePeers()
		if err != nil {
			fmt.Println(err)
			return
		}
		peerId, err := self.Handshake(peer)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Peer ID: %s\n", peerId)
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

func DecodeBencode(s string) {
	panic("unimplemented")
}
