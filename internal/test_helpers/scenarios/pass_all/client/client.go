package client

import (
	"bytes"
	"net/http"

	"fmt"
	"net"
	"time"

	"github.com/codecrafters-io/grep-starter-go/bencode"
	"github.com/codecrafters-io/grep-starter-go/handshake"
	"github.com/codecrafters-io/grep-starter-go/message"
	"github.com/codecrafters-io/grep-starter-go/parser"
	"github.com/codecrafters-io/grep-starter-go/peers"
	"github.com/codecrafters-io/grep-starter-go/torrent"
)

func HasPiece(bf []byte, index int) bool {
	byteIndex := index / 8
	offset := index % 8
	if byteIndex < 0 || byteIndex >= len(bf) {
		return false
	}
	return bf[byteIndex]>>uint(7-offset)&1 != 0
}

// SetPiece sets a bit in the bitfield
func SetPiece(bf []byte, index int) {
	byteIndex := index / 8
	offset := index % 8

	// silently discard invalid bounded index
	if byteIndex < 0 || byteIndex >= len(bf) {
		return
	}
	bf[byteIndex] |= 1 << uint(7-offset)
}

type Client struct {
	Conn      net.Conn
	Choked    bool
	Bitfield  []byte
	peer      string
	infoHash  [20]byte
	peerID    [20]byte
	Integrity int
	Handshake *handshake.Handshake
}

func RequestPeers(t *torrent.TorrentFile, peerID [20]byte, port uint16) ([]peers.Peer, error) {
	url, err := t.BuildTrackerURL(peerID, port)
	if err != nil {
		return nil, err
	}

	c := &http.Client{Timeout: 15 * time.Second}
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	/*
			// Create a new file to save the response body
			file, err := os.Create("sarpresponse.txt") // Change the filename as needed
			if err != nil {
				fmt.Println("Error:", err)
			}
			defer file.Close()

			// Copy the response body to the file
			_, err = io.Copy(file, resp.Body)
			if err != nil {
				fmt.Println("Error:", err)
			}
			fmt.Println("saved response")


		_, err = io.Copy(os.Stdout, resp.Body)
		if err != nil {
			fmt.Println("error printing", err)
			return nil, err
		}
	*/

	trackerResp := parser.BencodeTrackerResp{}
	err = bencode.Unmarshal(resp.Body, &trackerResp)
	if err != nil {
		return nil, err
	}

	return peers.Unmarshal([]byte(trackerResp.Peers))
}

func CompleteHandshake(conn net.Conn, infohash, peerID [20]byte, extensions []byte) (*handshake.Handshake, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	req := handshake.New(infohash, peerID, extensions)
	_, err := conn.Write(req.Serialize())
	if err != nil {
		fmt.Println("hmm1")
		return nil, err
	}

	res, err := handshake.Read(conn)
	if err != nil {
		fmt.Println("hmm2")
		return nil, err
	}
	//fmt.Printf("received handshake from peer: %x\n", res.PeerID)
	if !bytes.Equal(res.InfoHash[:], infohash[:]) {
		return nil, fmt.Errorf("Expected infohash %x but got %x", res.InfoHash, infohash)
	}
	return res, nil
}

func recvBitfield(conn net.Conn) ([]byte, error) {
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		err := fmt.Errorf("Expected bitfield but got %s", msg)
		return nil, err
	}
	if msg.ID != message.MsgBitfield {
		err := fmt.Errorf("Expected bitfield but got ID %d", msg.ID)
		return nil, err
	}

	return msg.Payload, nil
}

// New connects with a peer, completes a handshake, and receives a handshake
// returns an err if any of those fail.
func New(peer string, peerID, infoHash [20]byte, extensions []byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer, 3*time.Second)
	if err != nil {
		return nil, err
	}

	handshake, err := CompleteHandshake(conn, infoHash, peerID, extensions)
	if err != nil {
		conn.Close()
		return nil, err
	}

	bf, err := recvBitfield(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}
	// fmt.Println("received bitfield from peer")

	return &Client{
		Conn:     conn,
		Choked:   true,
		Bitfield: bf,
		peer:     peer,
		infoHash: infoHash,
		peerID:   peerID,
		Handshake: handshake,
	}, nil
}

// Read reads and consumes a message from the connection
func (c *Client) Read() (*message.Message, error) {
	msg, err := message.Read(c.Conn)
	return msg, err
}

// SendRequest sends a Request message to the peer
func (c *Client) SendRequest(index, begin, length int) error {
	//fmt.Printf("sending request for index: %d begin: %d length: %d\n", index, begin, length)
	req := message.FormatRequest(index, begin, length)
	_, err := c.Conn.Write(req.Serialize())
	return err
}

// SendInterested sends an Interested message to the peer
func (c *Client) SendInterested() error {
	//fmt.Println("sending interested")
	msg := message.Message{ID: message.MsgInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendNotInterested sends a NotInterested message to the peer
func (c *Client) SendNotInterested() error {
	msg := message.Message{ID: message.MsgNotInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendUnchoke sends an Unchoke message to the peer
func (c *Client) SendUnchoke() error {
	//fmt.Println("sending unchoke")
	msg := message.Message{ID: message.MsgUnchoke}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendHave sends a Have message to the peer
func (c *Client) SendHave(index int) error {
	msg := message.FormatHave(index)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendExtensionHandshake() error {
	req := message.FormatExtensionHandshake()
	serialized := req.Serialize()
	_, err := c.Conn.Write(serialized)
	return err
}

func (c *Client) SendMetadataRequest(extensionID uint8, index int) error {
	req := message.FormatMetadataExtensionMessage(extensionID, message.RequestMetadataExtensionMsgType, index)
	_, err := c.Conn.Write(req.Serialize())
	return err
}