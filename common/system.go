package common

import (
	"os"
	"path/filepath"
)

func WorkDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}

	return dir
}

func IsFileExist(path string) bool {
	if !filepath.IsAbs(path) {
		path = filepath.Join(WorkDir(), path)
	}

	_, err := os.Stat(path)
	switch {
	case os.IsNotExist(err):
		return false
	case err != nil:
		// This condition means that we unable know if the file exists and should return an error.
		// For now, we return false to keep the function simple
		return false
	default:
		return true
	}
}
