package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"sync"
	"time"
	"path"
)

func listFiles(root string) []string {
	fileSystem := os.DirFS(root)
	fileList := make([]string, 0, 1024)

	fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		fileList = append(fileList, path)
		return nil
	})
	fmt.Println(fileList)
	return fileList
}

func main() {
	files := listFiles(os.Args[1])
	start := time.Now()
	var wg sync.WaitGroup
	for _, fPath := range files {
		fPath := fPath
		wg.Add(1)
		go func() {
			defer wg.Done()
			f, err := os.Open(path.Join(os.Args[1], fPath))
			if err != nil {
				log.Fatal(err)
			}
			io.Copy(io.Discard, f)
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)
	log.Printf("time taken: %f", elapsed.Seconds())
}
