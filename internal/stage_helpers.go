package internal

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"math/rand"
	"net"
	"os"
	"path"

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
	filename       string
	outputFilename string
	tracker        string
	infohash       string
	length         int64
	expectedSha1   string
}

var testTorrents = []TestTorrentInfo{
	{
		filename:       "codercat.gif.torrent",
		outputFilename: "codercat.gif",
		tracker:        "http://bittorrent-test-tracker.codecrafters.io/announce",
		infohash:       "c77829d2a77d6516f88cd7a3de1a26abcbfab0db",
		length:         2994120,
		expectedSha1:   "89d5dcbb92d31f040f8fe42b559fd6ec7f4d83a5",
	},
	{
		filename:       "congratulations.gif.torrent",
		outputFilename: "congratulations.gif",
		tracker:        "http://bittorrent-test-tracker.codecrafters.io/announce",
		infohash:       "1cad4a486798d952614c394eb15e75bec587fd08",
		length:         820892,
		expectedSha1:   "fe3cc9002bc84c4776c3a962a717244c9cb962c0",
	},
	{
		filename:       "itsworking.gif.torrent",
		outputFilename: "itsworking.gif",
		tracker:        "http://bittorrent-test-tracker.codecrafters.io/announce",
		infohash:       "70edcac2611a8829ebf467a6849f5d8408d9d8f4",
		length:         2549700,
		expectedSha1:   "683e899db9d7a38e50eb87874b62f7fdd0c14c9c",
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

func getFileSizeBytes(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return -1, err
	}
	fileSize := fileInfo.Size()
	return fileSize, nil
}
