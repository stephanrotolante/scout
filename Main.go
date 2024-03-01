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

var (
	Location  = ""
	Task      = ""
	Alias     = ""
	AliasPath = ""
)

func init() {
	flag.StringVar(&Location, "p", "", "Location to next project")

	flag.StringVar(&Task, "t", "dependency-check", "Task to perfrom\nAvailable options are deadcode-check or dependency-check")

	flag.StringVar(&AliasPath, "ap", "", "Alias path")
	flag.StringVar(&Alias, "a", "", "Alias pattern")

	flag.Parse()

}

func main() {

	switch Task {
	case "deadcode-check":
		DeadCodeFinder(Location)
	case "dependency-check":
		DependencyCheck(Location)

	}

}
