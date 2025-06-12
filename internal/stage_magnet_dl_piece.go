package internal

import (
    "fmt"
    "os"
    "path"
    "strings"

    "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testMagnetDownloadPiece(stageHarness *test_case_harness.TestCaseHarness) error {
    type MagnetLinkPieceTest struct {
        MagnetURL string
        PieceIndex int
        PieceHash string
        PieceLength int64
    }

    var magnetLinkPieceTests = [][]MagnetLinkPieceTest{
        {
            {
                MagnetURL: "magnet:?xt=urn:btih:ad42ce8109f54c99613ce38f9b4d87e70f24a165&dn=magnet1.gif&tr=http%3A%2F%2Fbittorrent-test-tracker.codecrafters.io%2Fannounce",
                PieceIndex:      1,
                PieceHash:       "eda417cb5c1cdbf841125c412da0bec9db8301f3",
                PieceLength:     262144,
            },
            {
                MagnetURL: "magnet:?xt=urn:btih:ad42ce8109f54c99613ce38f9b4d87e70f24a165&dn=magnet1.gif&tr=http%3A%2F%2Fbittorrent-test-tracker.codecrafters.io%2Fannounce",
                PieceIndex:      2,
                PieceHash:       "422f45b1052e2d45da3e2a6516e1bb1f1db00733",
                PieceLength:     112217,
            },
        },
        {
            {
                MagnetURL: "magnet:?xt=urn:btih:c5fb9894bdaba464811b088d806bdd611ba490af&dn=magnet3.gif&tr=http%3A%2F%2Fbittorrent-test-tracker.codecrafters.io%2Fannounce",
                PieceIndex:      0,
                PieceHash:       "ca80fd83ffb34d6e1bbd26a8ef6d305827f1cd0a",
                PieceLength:     262144,
            },
            {
                MagnetURL: "magnet:?xt=urn:btih:c5fb9894bdaba464811b088d806bdd611ba490af&dn=magnet3.gif&tr=http%3A%2F%2Fbittorrent-test-tracker.codecrafters.io%2Fannounce",
                PieceIndex:      2,
                PieceHash:       "47076d104d214c0052960ef767262649a8af0ea8",
                PieceLength:     105656,
            },
        },
        {
            {
                MagnetURL: "magnet:?xt=urn:btih:3f994a835e090238873498636b98a3e78d1c34ca&dn=magnet2.gif&tr=http%3A%2F%2Fbittorrent-test-tracker.codecrafters.io%2Fannounce",
                PieceIndex:      0,
                PieceHash:       "d78a7f55ddd89fef477bc49d938bc7e4d94094f1",
                PieceLength:     79752,
            },
            {
                MagnetURL: "magnet:?xt=urn:btih:ad42ce8109f54c99613ce38f9b4d87e70f24a165&dn=magnet1.gif&tr=http%3A%2F%2Fbittorrent-test-tracker.codecrafters.io%2Fannounce",
                PieceIndex:      0,
                PieceHash:       "3b46a96d9bc3716d1b75da91e6d753a793ad1cef",
                PieceLength:     262144,
            },
        },
    }

    randomIndex := random.RandomInt(0, len(magnetLinkPieceTests))
    tests := magnetLinkPieceTests[randomIndex]

    logger := stageHarness.Logger
    executable := stageHarness.Executable

    tempDir, err := os.MkdirTemp("", "torrents")
    if err != nil {
        logger.Errorln("Couldn't create temp directory")
        return err
    }

    for _, t := range tests {
        expectedFilename := fmt.Sprintf("piece-%d", t.PieceIndex)
        downloadedFilePath := path.Join(tempDir, expectedFilename)

        logger.Infof("Running ./your_bittorrent.sh magnet_download_piece -o %s %q %d", downloadedFilePath, t.MagnetURL, t.PieceIndex)
        result, err := executable.Run("magnet_download_piece", "-o", downloadedFilePath, t.MagnetURL, fmt.Sprintf("%d", t.PieceIndex))
        resultStr := string(result.Stdout)

        if strings.Contains(resultStr, "Connection reset by peer") || strings.Contains(resultStr, "EOF") {
            logger.Infoln("Connection reset by peer or EOF type of errors can be transient issues (try again), but they can also happen when peers receive unexpected messages (sending a REQUEST message with an incorrect offset/length, block size larger than 16 kb etc.)")
            if t.PieceLength != 262144 {
                logger.Infof("Last piece of the file can be less than piece length specified in torrent metadata. For instance, the length of this piece is %d", t.PieceLength)
            }
        }

        if err != nil {
            return err
        }

        if err = assertExitCode(result, 0); err != nil {
            return err
        }

        if err = assertFileSize(downloadedFilePath, t.PieceLength); err != nil {
            return err
        }

        logger.Successln("✓ Piece size is correct.")

        if err = assertFileSHA1(downloadedFilePath, t.PieceHash); err != nil {
            return err
        }

        logger.Successln("✓ Piece SHA-1 is correct.")
    }

    return nil
}