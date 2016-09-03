package util

import (
	"os"
)

// GetFileSize returns the size in bytes of the specified file
func GetFileSize(filePath string) (int64, error) {
	fi, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}
