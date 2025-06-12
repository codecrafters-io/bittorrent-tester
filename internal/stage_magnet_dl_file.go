package internal

import (
    "os"
    "path"

    "github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testMagnetDownloadFile(stageHarness *test_case_harness.TestCaseHarness) error {

    logger := stageHarness.Logger
    executable := stageHarness.Executable

    t := randomMagnetLink()

    tempDir, err := os.MkdirTemp("", "torrents")
    if err != nil {
        logger.Errorln("Couldn't create temp directory")
        return err
    }
    downloadedFilePath := path.Join(tempDir, t.Filename)
    magnetUrl := "magnet:?xt=urn:btih:" + t.InfoHashStr + "&dn=" + t.Filename + "&tr=http%3A%2F%2Fbittorrent-test-tracker.codecrafters.io%2Fannounce"

    logger.Infof("Running ./your_bittorrent.sh magnet_download -o %s %q", downloadedFilePath, magnetUrl)
    result, err := executable.Run("magnet_download", "-o", downloadedFilePath, magnetUrl)
    if err != nil {
        return err
    }

    if err = assertExitCode(result, 0); err != nil {
        return err
    }

    if err = assertFileSize(downloadedFilePath, int64(t.FileLengthBytes)); err != nil {
        return err
    }

    logger.Successln("✓ File size is correct.")

    if err = assertFileSHA1(downloadedFilePath, t.ExpectedSha1); err != nil {
        return err
    }

    logger.Successln("✓ File SHA-1 is correct.")

    return nil
}