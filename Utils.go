package main

import (
	"path/filepath"
	"strings"
)

func GetFileDirPath(filePath string) string {
	splitPath := strings.Split(filePath, "/")
	return strings.Join(splitPath[:len(splitPath)-1], "/")

}

func CleanFilePath(filePath, newFilePath string) string {
	return filepath.Clean(filePath + "/" + newFilePath)
}
