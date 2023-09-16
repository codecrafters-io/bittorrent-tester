package internal

import (
	"crypto/sha1"
	"fmt"

	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/bencode"
)

func HashInfo(info map[string]interface{}) string {
	encoded, err := bencode.EncodeBencode(info)
	if err != nil {
		fmt.Printf("Not a valid torrent file: %s", err.Error())
	}
	hasher := sha1.New()
	hasher.Write([]byte(encoded))
	return string(hasher.Sum(nil))
}
