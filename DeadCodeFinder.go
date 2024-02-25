package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path"
	"sort"
	"strings"
	"sync"
)

func DeadCodeFinder(projectPath string) {

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

	cmd := exec.Command("find", path.Join(projectPath, "src"), "-type", "f")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	fileListString := out.String()

	fileList := strings.Split(fileListString, "\n")

	pageSet := NewSet()

	for _, page := range pages {
		for _, pageDep := range page.DependentFiles {
			pageSet.Add(pageDep)
		}
	}

	var fileCount = 0
	sort.Strings(fileList)
	for _, file := range fileList {
		if !pageSet.Contains(file) && !strings.Contains(file, ".test.") && !strings.Contains(file, ".md") && !strings.Contains(file, ".gitkeep") {
			fileCount += 1

			fmt.Println(strings.ReplaceAll(file, projectPath, ""))

		}
	}

	fmt.Printf("%d File(s) not used\n", fileCount)

}
