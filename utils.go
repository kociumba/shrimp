package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func ExpandPath(path string) (string, error) {
	path = os.ExpandEnv(path)

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for %s: %w", path, err)
	}

	clean := filepath.Clean(absPath)

	return clean, nil
}
