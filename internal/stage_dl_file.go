package internal

import (
	"path"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

func testDownloadFile(stageHarness *tester_utils.StageHarness) error {
	torrentFile := "alpine-minirootfs-3.18.3-aarch64.tar.gz.torrent"
	expectedFilename := "alpine-minirootfs-3.18.3-aarch64.tar.gz"
	expectedSha1 := "e3fae2149ee08c75ced032561ea8ee5bd9b01d14"
	expectedFileSize := int64(3201634)

	logger := stageHarness.Logger
	executable := stageHarness.Executable

	tempDir, err := createTempDir(executable)
	if err != nil {
		logger.Errorf("Couldn't create temp directory")
		return err
	}

	if err := copyTorrent(tempDir, torrentFile); err != nil {
		logger.Errorf("Couldn't copy torrent file")
		return err
	}

	logger.Debugf("Running ./your_bittorrent.sh download -o %s %s", expectedFilename, torrentFile)
	result, err := executable.Run("download", "-o", expectedFilename, torrentFile)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	downloadedFilePath := path.Join(tempDir, expectedFilename)
	if err = assertFileSize(downloadedFilePath, expectedFileSize); err != nil {
		return err
	}

	if err = assertFileSHA1(downloadedFilePath, expectedSha1); err != nil {
		return err
	}

	return nil
}

/*
// Uncomment to test using private tracker.
// Make sure there's an image in this repo named codecrafters.jpeg
func testDownloadFile(stageHarness *tester_utils.StageHarness) error {
	fileName := "codecrafters.jpeg"
	imgFilePath := path.Join(getProjectDir(), fileName)

	logger := stageHarness.Logger
	executable := stageHarness.Executable

	tempDir, err := createTempDir(executable)
	if err != nil {
		logger.Errorf("Couldn't create temp directory")
		return err
	}

	torrentFile := "test.torrent"
	torrentFilePath := path.Join(tempDir, torrentFile)
	pieceLengthBytes := 256 * 1024
	expectedSha1, err := calculateSHA1(imgFilePath)
	if err != nil {
		return err
	}
	piecesStr, err := createPiecesStrFromFile(imgFilePath, pieceLengthBytes)
	if err != nil {
		return err
	}
	fileLengthBytes, err := getFileSizeBytes(imgFilePath)
	if err != nil {
		return err
	}

	torrent := TorrentFile{
		Announce: "http://bittorrent-test-tracker.codecrafters.io/announce",
		Info: TorrentFileInfo{
			Name:        fileName,
			Length:      int(fileLengthBytes),
			Pieces:      piecesStr,
			PieceLength: pieceLengthBytes,
		},
	}
	_, err = torrent.writeToFile(torrentFilePath)
	if err != nil {
		logger.Errorf("Couldn't write torrent file", err)
		return err
	}

	logger.Debugf("Running ./your_bittorrent.sh download -o %s %s", fileName, torrentFile)
	result, err := executable.Run("download", "-o", fileName, torrentFile)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	downloadedFilePath := path.Join(tempDir, fileName)
	if err = assertFileSize(downloadedFilePath, fileLengthBytes); err != nil {
		return err
	}

	if err = assertFileSHA1(downloadedFilePath, expectedSha1); err != nil {
		return err
	}

	return nil
}
*/
