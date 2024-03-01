package main

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

func GetFileDirPath(filePath string) string {
	splitPath := strings.Split(filePath, string(os.PathSeparator))
	return strings.Join(splitPath[:len(splitPath)-1], string(os.PathSeparator))

}

func CleanFilePath(filePath, newFilePath string) string {
	return filepath.Clean(path.Join(filePath, newFilePath))
}
