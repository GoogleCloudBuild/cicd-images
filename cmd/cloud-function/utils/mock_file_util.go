package utils

import (
	"context"
)

type MockFileUtil struct {
	ZipFileName      string
	ZipCreationError error
	UploadError      error
}

func (fileUtil MockFileUtil) ArchiveDirectoryContentIntoZip(_ string) (string, error) {
	return fileUtil.ZipFileName, fileUtil.ZipCreationError
}

func (fileUtil MockFileUtil) UploadFileToSignedURL(_ context.Context, _, _ string) error {
	return fileUtil.UploadError
}

func (fileUtil MockFileUtil) CleanUp(_ string) error {
	return nil
}
