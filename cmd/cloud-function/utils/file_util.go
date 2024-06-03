package utils

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	bytesize "github.com/inhies/go-bytesize"
)

const (
	sourceSizeLimitB               = (int64)(512 * bytesize.MB)
	pathIsNotDirectoryErrorMessage = "the path provided is not a directory"
	sizeLimitExceedsErrorMessage   = "uncompressed deployment is bigger than maximum allowed size of 512Mb"
)

type FileUtil interface {
	ArchiveDirectoryContentIntoZip(dir string) (string, error)
	UploadFileToSignedURL(ctx context.Context, uploadURL, file string) error
	CleanUp(zipFile string) error
}

type OsFileUtil struct{}

// CleanUp removes the directory of the zip provided with all its contect.
func (fileUtil OsFileUtil) CleanUp(zipFile string) error {
	dir, _ := filepath.Split(zipFile)
	return os.RemoveAll(dir)
}

// ArchiveDirectoryContentIntoZip validates the directory, archives its contect and
// returns the file name of the created zip.
func (fileUtil OsFileUtil) ArchiveDirectoryContentIntoZip(dir string) (string, error) {
	err := validateSourceDirectory(dir)
	if err != nil {
		return "", err
	}
	tmpDir, err := os.MkdirTemp(".", "function-tmp-")
	if err != nil {
		return "", err
	}
	tmpZip, err := os.CreateTemp(tmpDir, "function-*.zip")
	if err != nil {
		return "", err
	}
	defer tmpZip.Close()
	zipFileName := tmpZip.Name()
	err = archiveDirectoryIntoZip(zipFileName, dir)
	if err != nil {
		return "", err
	}
	return zipFileName, nil
}

// UploadFileToSignedURL uploads the file from the specified path to the provided URL.
func (fileUtil OsFileUtil) UploadFileToSignedURL(ctx context.Context, uploadURL, file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	fileUploadReq, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, f)
	if err != nil {
		return err
	}
	fileUploadReq.Header.Add("Content-Type", "application/zip")
	httpClient := &http.Client{}
	resp, err := httpClient.Do(fileUploadReq)
	if err != nil {
		return err
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	bodyString := string(bodyBytes)
	defer resp.Body.Close()
	// upload status code != Success (2xx).
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("failed to upload the function source code to signed url: %s, status: %v, details: %s", uploadURL, resp.StatusCode, bodyString)
	}
	return nil
}

// validateSourceDirectory checks that the path provided is a directory and
// its contect is less than allowed size.
func validateSourceDirectory(dir string) error {
	stat, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf(pathIsNotDirectoryErrorMessage)
	}
	var sizeB int64
	err = filepath.WalkDir(dir, func(_ string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := dirEntry.Info()
		if err != nil {
			return err
		}
		if !info.IsDir() {
			sizeB += info.Size()
		}
		if sizeB > sourceSizeLimitB {
			return fmt.Errorf(sizeLimitExceedsErrorMessage)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// archiveDirectoryIntoZip puts the provided directory content into the
// specified archive.
func archiveDirectoryIntoZip(zipFileName, dir string) error {
	zipFile, err := os.OpenFile(zipFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	w := zip.NewWriter(zipFile)
	walkFunc := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if dir == path || info.IsDir() {
			return nil
		}
		zipPath := strings.Replace(path, dir, "", 1)
		zipPath = strings.TrimPrefix(zipPath, string(filepath.Separator))
		zipPath = filepath.ToSlash(zipPath)
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		zipFile, err := w.Create(zipPath)
		if err != nil {
			return err
		}
		_, err = io.Copy(zipFile, f)
		if err != nil {
			return err
		}
		return nil
	}
	err = filepath.WalkDir(dir, walkFunc)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return nil
}
