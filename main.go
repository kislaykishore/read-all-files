package main

import (
	"context"
	"golang.org/x/sync/semaphore"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

var sem = semaphore.NewWeighted(128)

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
	return fileList
}

func main() {
	files := listFiles(os.Args[1])
	start := time.Now()
	var wg sync.WaitGroup
	ctx := context.Background()
	for _, fPath := range files {
		fPath := fPath
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem.Acquire(ctx, 1)
			f, err := os.Open(path.Join(os.Args[1], fPath))
			if err != nil {
				log.Fatal(err)
			}
			io.Copy(io.Discard, f)
			sem.Release(1)
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)
	log.Printf("time taken: %f", elapsed.Seconds())
}
