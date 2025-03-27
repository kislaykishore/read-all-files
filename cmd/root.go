/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"path"
	"sync/atomic"

	"context"
	"io"
	"io/fs"
	"log"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/semaphore"
)

const (
	dirFlagName         = "dir"
	concurrencyFlagName = "concurrency"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ReadAllFiles",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := cmd.Flags().GetString(dirFlagName)
		if err != nil {
			return err
		}
		concurrency, err := cmd.Flags().GetInt64(concurrencyFlagName)
		if err != nil {
			return err
		}
		var readBytes atomic.Int64
		files := listFiles(p)
		start := time.Now()
		var wg sync.WaitGroup
		ctx := context.Background()
		sem := semaphore.NewWeighted(concurrency)
		for _, fPath := range files {
			fPath := fPath
			wg.Add(1)
			go func() {
				defer wg.Done()
				sem.Acquire(ctx, 1)
				f, err := os.Open(path.Join(p, fPath))
				if err != nil {
					log.Fatal(err)
				}
				n, _ := io.Copy(io.Discard, f)
				readBytes.Add(n)
				sem.Release(1)
			}()
		}
		wg.Wait()
		elapsed := time.Since(start)
		log.Printf("time taken: %f", elapsed.Seconds())
		log.Printf("bandwidth: %f (GiB/s)", float64(readBytes.Load())*1.0/(1024*1024*1024*elapsed.Seconds()))
		return nil
	},
}

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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ReadAllFiles.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().String(dirFlagName, "", "The directory for which read-performance needs to be evaluated")
	rootCmd.Flags().Int64(concurrencyFlagName, 1, "The number of concurrent calls.")
}
