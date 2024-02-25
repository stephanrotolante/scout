package main

import (
	"flag"
	"fmt"
	"os"
	"path"
)

func ReadDirectory(dirPath string) []Page {

	var pages []Page
	// Open the directory
	d, err := os.Open(dirPath)
	if err != nil {
		fmt.Println("Error opening directory:", err)
		return pages
	}
	defer d.Close()

	// Read directory contents
	fileInfos, err := d.Readdir(-1)
	if err != nil {
		fmt.Println("Error reading directory contents:", err)
		return pages
	}

	for _, fi := range fileInfos {
		if fi.IsDir() {
			pages = append(pages, ReadDirectory(path.Join(dirPath, fi.Name()))...)
		} else {
			pages = append(pages, Page{
				PageName: fi.Name(),
				PagePath: path.Join(dirPath, fi.Name()),
			})
		}
	}

	return pages
}

const MAX_CONCURRENCY = 8

func main() {
	location := flag.String("p", "", "Location to next project")

	task := flag.String("t", "dependency-check", "Task to perfrom\nAvailable options are deadcode-check or dependency-check")

	flag.Parse()

	switch *task {
	case "deadcode-check":
		DeadCodeFinder(*location)
	case "dependency-check":
		DependencyCheck(*location)

	}

}
