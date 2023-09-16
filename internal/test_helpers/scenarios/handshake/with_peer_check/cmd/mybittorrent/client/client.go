package client

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/bencode"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/metafile"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/peer"
)

const PEER_ID = "00112233445566778899"
const PORT = 6881
const COMPACT = "1"
const PROTOCOL = "BitTorrent protocol"
const PROTOCOL_LENGTH = len(PROTOCOL)
const RETRIES = 3

type Client struct {
	id         string
	port       int
	compact    string
	uploaded   int
	downloaded int
	left       int
	peers      []*peer.Peer
	metafile   *metafile.Metafile
}

func NewClient(meta *metafile.Metafile) *Client {
	client := &Client{
		id:       PEER_ID,
		port:     PORT,
		left:     meta.Length,
		compact:  COMPACT,
		metafile: meta,
	}

	return client
}

func (c *Client) requestTracker() (string, error) {
	u, err := url.Parse(c.metafile.Tracker)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("info_hash", c.metafile.HashedInfo)
	q.Set("peer_id", c.id)
	q.Set("port", fmt.Sprint(c.port))
	q.Set("uploaded", fmt.Sprint(c.uploaded))
	q.Set("downloaded", fmt.Sprint(c.downloaded))
	q.Set("left", fmt.Sprint(c.left))
	q.Set("compact", c.compact)
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (c *Client) UpdatePeers() error {
	response, err := c.requestTracker()
	oldPeers := c.peers
	c.peers = []*peer.Peer{}

	if err != nil {
		return err
	}
	decodedResponse, _, err := bencode.DecodeBencode(response)
	if err != nil {
		return err
	}
	parsedResponse, ok := decodedResponse.(map[string]interface{})
	if !ok {
		return fmt.Errorf("Response is invalid %s", response)
	}

	var peersInterface interface{}
	var exists bool
	for i := 0; i < RETRIES; i++ {
		peersInterface, exists = parsedResponse["peers"]
		if !exists {
			continue
		}
	}

	if !exists {
		return fmt.Errorf("Tracker didn't return Peers list")
	}
	peersStr, ok := peersInterface.(string)
	if !ok {
		return fmt.Errorf("Peers are not properly encoded")
	}
	peers := []byte(peersStr)
	for i := 0; i < len(peers); i += peer.BytesLength {
		newPeer, err := peer.NewPeer(peers[i : i+peer.BytesLength])
		if err != nil {
			return err
		}
		c.peers = append(c.peers, newPeer)
	}

	if len(c.peers) == 0 {
		c.peers = oldPeers
		return fmt.Errorf("Tracker found zero peers")
	}

	return nil
}

func (c *Client) hasPeer(peerAddr string) bool {
	for _, peer := range c.peers {
		if peer.String() == peerAddr {
			return true
		}
	}
	return false
}

func (c *Client) Handshake(peerAddr string) (string, error) {
	if !c.hasPeer(peerAddr) {
		var peers []string
		for _, peer := range c.peers {
			peers = append(peers, peer.String())
		}
		return "", fmt.Errorf("Address %s doesn't match any peers %#v", peerAddr, peers)
	}
	peerAddress, err := net.ResolveTCPAddr("tcp", peerAddr)
	if err != nil {
		return "", fmt.Errorf("failed to resolve peer address: %w", err)
	}
	conn, err := net.DialTCP("tcp", nil, peerAddress)
	if err != nil {
		return "", fmt.Errorf("failed to dial peer: %w", err)
	}
	buf := bytes.NewBuffer([]byte{})
	buf.WriteString(PROTOCOL)
	buf.Write(make([]byte, 8))
	buf.Write([]byte(c.metafile.HashedInfo))
	buf.Write([]byte(PEER_ID))

	buffer := buf.Bytes()

	newBuffer := append([]byte{19}, buffer...)

	_, err = conn.Write(newBuffer)
	if err != nil {
		return "", fmt.Errorf("Error writing to connection")
	}

	response := make([]byte, 68)
	_, err = io.ReadFull(conn, response)
	if err != nil {
		return "", fmt.Errorf("Error reading response: %s", err.Error())
	}

	if response[0] != byte(PROTOCOL_LENGTH) {
		return "", fmt.Errorf("byte 0 of response is not %d", PROTOCOL_LENGTH)
	}

	if string(response[1:20]) != PROTOCOL {
		return "", fmt.Errorf("bytes 1:20 of response %s not %s", string(response[1:20]), PROTOCOL)
	}

	if string(response[28:48]) != c.metafile.HashedInfo {
		return "", fmt.Errorf("Hash returned in the response doesn't match")
	}

	peerId := response[48:]

	return fmt.Sprintf("%x", peerId), nil
}

func (c *Client) PrintPeers() {
	for _, peer := range c.peers {
		fmt.Printf("%s\n", peer)
	}

}
