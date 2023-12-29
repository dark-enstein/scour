package utils

import (
	"io/fs"
	"log"
	"os"
)

// IsSocket checks if the provided path is a Unix socket.
func IsSocket(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Printf("Error opening file: %s\n", err.Error())
	}
	return fileInfo.Mode().Type() == fs.ModeSocket
}
