package main

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
)

func DependencyCheck(projectPath string) {

	if len(os.Args) < 4 {
		fmt.Println("No files passed in")
		os.Exit(1)
	}

	pages := ReadDirectory(fmt.Sprintf("%s/pages", projectPath))

	semaphore := make(chan struct{}, MAX_CONCURRENCY)

	var wg sync.WaitGroup

	for i := 0; i < len(pages); i++ {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(p *Page) {
			defer func() {
				<-semaphore
				wg.Done()
			}()
			p.FileDepTree()
		}(&pages[i])
	}

	wg.Wait()

	changedFiles := os.Args[3:]

	for _, file := range changedFiles {
		_, err := os.Stat(path.Join(projectPath, file))

		if err != nil {
			fmt.Println("Error finding file ", path.Join(projectPath, file))

			os.Exit(1)
		}
	}

	pageSet := NewSet()
	for _, changedFile := range changedFiles {
		for _, page := range pages {
			if page.IsFileADep(path.Join(projectPath, changedFile)) {
				pageSet.Add(page.PagePath)
			}
		}

	}

	usedPageList := pageSet.GetKeys()
	for _, page := range usedPageList {
		fmt.Println(strings.ReplaceAll(page, fmt.Sprintf("%s/pages", projectPath), ""))
	}

	fmt.Printf("%d Page(s) used\n", len(usedPageList))

}
