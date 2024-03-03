package utils

import (
	"fmt"
	"io/fs"
	"log"
	"os"
)

// IsSocket checks if the provided path is a Unix socket.
func IsSocket(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Printf("error opening file: %s: %s\n", path, err.Error())
		return false, fmt.Errorf("error opening file: %s: %s\n", path, err.Error())
	}
	return fileInfo.Mode().Type() == fs.ModeSocket, nil
}
