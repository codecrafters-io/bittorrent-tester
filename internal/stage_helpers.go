package internal

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"math/rand"
	"net"
	"os"
	"path"
	"time"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

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
	filename string
	tracker  string
	length   string
	infohash string
}

var testTorrents = []TestTorrentInfo{
	{
		filename: "test.torrent",
		tracker:  "http://linuxtracker.org:2710/00000000000000000000000000000000/announce",
		length:   "406847488",
		infohash: "6d4795dee70aeb88e03e5336ca7c9fcf0a1e206d",
	},
	{
		filename: "debian-9.1.0-amd64-netinst.iso.torrent",
		tracker:  "http://bttracker.debian.org:6969/announce",
		length:   "304087040",
		infohash: "fd5fdf21aef4505451861da97aa39000ed852988",
	},
	{
		filename: "debian-10.8.0-amd64-netinst.iso.torrent",
		tracker:  "http://bttracker.debian.org:6969/announce",
		length:   "352321536",
		infohash: "4090c3c2a394a49974dfbbf2ce7ad0db3cdeddd7",
	},
	{
		filename: "alpine-minirootfs-3.18.3-aarch64.tar.gz.torrent",
		tracker:  "http://tracker.openbittorrent.com:80/announce",
		length:   "3201634",
		infohash: "0307870997cb1ae741431cfc54f3b2dcf52e2c80",
	},
}

func createTempDir(executable *tester_utils.Executable) (string, error) {
	tempDir, err := os.MkdirTemp("", "worktree")
	if err != nil {
		return "", err
	}
	executable.WorkingDir = tempDir
	return tempDir, nil
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
	rand.Seed(time.Now().UnixNano())
	return testTorrents[rand.Intn(len(testTorrents))]
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
