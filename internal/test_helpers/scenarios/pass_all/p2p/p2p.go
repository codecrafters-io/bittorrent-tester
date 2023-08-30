package p2p

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"net"
	"os"

	"time"

	"github.com/codecrafters-io/grep-starter-go/client"
	"github.com/codecrafters-io/grep-starter-go/message"
	"github.com/codecrafters-io/grep-starter-go/peers"
	"github.com/codecrafters-io/grep-starter-go/torrent"
)

// Torrent holds data required to download a torrent from a list of peers
type Torrent struct {
	Peers       []peers.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}

type pieceProgress struct {
	index      int
	client     *client.Client
	buf        []byte
	downloaded int
	requested  int
	backlog    int
}

const MaxBlockSizeKb = 16 * 1024
const MaxBacklogSize = 5

func TalkToPeer(torrentFile torrent.TorrentFile, peer string, peerID [20]byte, infoHash [20]byte) {
	conn, err := net.DialTimeout("tcp", peer, 3*time.Second)
	if err != nil {
		fmt.Println("error", err)
		return
	}

	handshake, err := client.CompleteHandshake(conn, infoHash, peerID)
	if err != nil {
		fmt.Println("error", err)
		conn.Close()
		return
	}
	fmt.Printf("Peer ID: %x\n", handshake.PeerID)
	/*
		for i := 0; i < len(torrentFile.PieceHashes); i++ {
			fmt.Printf("Piece #%d: %t\n", i, client.HasPiece(c.Bitfield, i))
		}
	*/
	//fmt.Println("bitfield", c.Bitfield)
}

// DownloadToFile downloads a torrent and writes it to a file
func DownloadToFile(t *torrent.TorrentFile, savePath string, pieceIndex int) error {
	var peerID [20]byte
	_, err := rand.Read(peerID[:])
	if err != nil {
		return err
	}

	//fmt.Printf("my peer id: %x\n", peerID)
	peers, err := client.RequestPeers(t, peerID, peers.Port)
	if err != nil {
		return err
	}
	// fmt.Println("# of peers", len(peers))
	// fmt.Println("peers:", peers)
	/*
		m.Seed(time.Now().UnixNano())
		random := m.Intn(len(peers))
		first := peers[random]
		fmt.Printf("random: %d peer: %s\n", random, first)
	*/

	torrent := Torrent{
		Peers:       peers, // TODO: Remove
		PeerID:      peerID,
		InfoHash:    t.InfoHash,
		PieceHashes: t.PieceHashes,
		PieceLength: t.PieceLength,
		Length:      t.Length,
		Name:        t.Name,
	}

	var buf []byte
	if pieceIndex == -1 {
		buf, err = torrent.Download()
	} else {
		buf, err = torrent.DownloadPiece(pieceIndex)
	}

	if err != nil {
		return err
	}

	outFile, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer outFile.Close()
	_, err = outFile.Write(buf)
	if err != nil {
		return err
	}

	return nil
}

func (t *Torrent) calculateBoundsForPiece(index int) (begin int, end int) {
	begin = index * t.PieceLength
	end = begin + t.PieceLength
	if end > t.Length {
		end = t.Length
	}
	return begin, end
}

func (t *Torrent) CalculatePieceSize(index int) int {
	begin, end := t.calculateBoundsForPiece(index)
	return end - begin
}

func checkIntegrity(expectedHash [20]byte, buf []byte, index int) error {
	hash := sha1.Sum(buf)
	// fmt.Printf("found hash: %x expected hash: %x\n", hash, expectedHash)
	if !bytes.Equal(hash[:], expectedHash[:]) {
		return fmt.Errorf("index %d failed integrity check", index)
	}
	return nil
}

func (t *Torrent) DownloadPiece(pieceIndex int) ([]byte, error) {
	// fmt.Println("starting download for", t.Name)
	numPieces := 1
	workQueue := make(chan *pieceWork, numPieces)
	resultsQueue := make(chan *pieceResult)
	length := t.CalculatePieceSize(pieceIndex)
	hash := t.PieceHashes[pieceIndex]
	// fmt.Printf("adding piece with i: %d len: %d hash: %x\n", pieceIndex, length, hash)
	workQueue <- &pieceWork{pieceIndex, hash, length}

	go t.startDownloadWorker(t.Peers[2], workQueue, resultsQueue)

	buf := make([]byte, length)
	finishedPieces := 0
	for finishedPieces < numPieces {
		result := <-resultsQueue
		copy(buf[0:length], result.buf)
		finishedPieces++

		// percent := float64(finishedPieces) / float64(numPieces) * 100
		// numWorkers := runtime.NumGoroutine() - 1
		// fmt.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, result.index, numWorkers)
	}
	close(workQueue)
	return buf, nil
}
func (t *Torrent) Download() ([]byte, error) {
	// fmt.Println("starting download for", t.Name)
	numPieces := len(t.PieceHashes)
	workQueue := make(chan *pieceWork, numPieces)
	resultsQueue := make(chan *pieceResult)
	for index, hash := range t.PieceHashes {
		length := t.CalculatePieceSize(index)
		// fmt.Printf("adding piece with i: %d len: %d hash: %x\n", index, length, hash)
		workQueue <- &pieceWork{index, hash, length}
	}
	for _, peer := range t.Peers {
		go t.startDownloadWorker(peer, workQueue, resultsQueue)
	}
	// TODO: this can be too large
	buf := make([]byte, t.Length)
	finishedPieces := 0
	for finishedPieces < numPieces {
		result := <-resultsQueue
		begin, end := t.calculateBoundsForPiece(result.index)
		copy(buf[begin:end], result.buf)
		finishedPieces++

		// percent := float64(finishedPieces) / float64(numPieces) * 100
		// numWorkers := runtime.NumGoroutine() - 1
		// fmt.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, result.index, numWorkers)
	}
	close(workQueue)
	return buf, nil
}

func (state *pieceProgress) readMessage() error {
	msg, err := state.client.Read() // this call blocks
	if err != nil {
		return err
	}

	if msg == nil { // keep-alive
		return nil
	}

	switch msg.ID {
	case message.MsgUnchoke:
		// fmt.Println("received unchoke message")
		state.client.Choked = false
	case message.MsgChoke:
		// fmt.Println("received choke message")
		state.client.Choked = true
	case message.MsgHave:
		index, err := message.ParseHave(msg)
		if err != nil {
			return err
		}
		client.SetPiece(state.client.Bitfield, index)
	case message.MsgPiece:
		// fmt.Println("received piece message")
		n, err := message.ParsePiece(state.index, state.buf, msg)
		if err != nil {
			return err
		}
		state.downloaded += n
		state.backlog--
	}
	return nil
}

func attemptDownloadPiece(c *client.Client, pw *pieceWork) ([]byte, error) {
	state := pieceProgress{
		index:  pw.index,
		client: c,
		buf:    make([]byte, pw.length),
	}

	// Setting a deadline helps get unresponsive peers unstuck.
	// 30 seconds is more than enough time to download a 262 KB piece
	c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer c.Conn.SetDeadline(time.Time{}) // Disable the deadline

	for state.downloaded < pw.length {
		// If unchoked, send requests until we have enough unfulfilled requests
		if !state.client.Choked {
			for state.backlog < MaxBacklogSize && state.requested < pw.length {
				blockSizeKb := MaxBlockSizeKb
				// Last block might be shorter than the typical block
				if pw.length-state.requested < blockSizeKb {
					blockSizeKb = pw.length - state.requested
				}

				err := c.SendRequest(pw.index, state.requested, blockSizeKb)
				if err != nil {
					return nil, err
				}
				state.backlog++
				state.requested += blockSizeKb
			}
		} else {
			// fmt.Println("can't download because client is in choked state")
		}

		// fmt.Println("will read message")
		err := state.readMessage()
		if err != nil {
			return nil, err
		}
	}

	return state.buf, nil
}

func (t *Torrent) startDownloadWorker(peer peers.Peer, workQueue chan *pieceWork, resultQueue chan *pieceResult) {
	// fmt.Println("trying to start download worker")
	c, err := client.New(peer.String(), t.PeerID, t.InfoHash)
	if err != nil {
		fmt.Errorf("error connecting to peer", err)
		return
	}
	// fmt.Println("handshake complete with peer:", peer)

	// fmt.Println("will send unchoke")
	err = c.SendUnchoke()
	if err != nil {
		fmt.Errorf("error sending unchoke", err)
		return
	}

	// fmt.Println("will send interested")
	err = c.SendInterested()
	if err != nil {
		fmt.Errorf("error sending interested", err)
		return
	}

	for pw := range workQueue {
		if !client.HasPiece(c.Bitfield, pw.index) {
			workQueue <- pw // Put piece back on the queue
			continue
		}

		// Download the piece
		buf, err := attemptDownloadPiece(c, pw)
		if err != nil {
			fmt.Println("Exiting", err)
			workQueue <- pw // Put piece back on the queue
			return
		}

		err = checkIntegrity(pw.hash, buf, pw.index)
		if err != nil {
			fmt.Printf("Piece #%d failed integrity check\n", pw.index)
			c.Integrity++
			if c.Integrity >= 3 {
				fmt.Printf("Client failed integrity too many times, closing connection")
				return
			}
			workQueue <- pw // Put piece back on the queue
			continue
		}

		c.SendHave(pw.index)
		resultQueue <- &pieceResult{pw.index, buf}
	}
}
