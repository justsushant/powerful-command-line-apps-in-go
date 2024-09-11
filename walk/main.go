package main

import (
	"fmt"
	"io"
	"flag"
	"os"
	"path/filepath"
	"log"
)

type config struct {
	ext string // extension to filter out
	size int64 // min file size
	list bool // list files
	del bool // delete files
	wLog io.Writer	// log destination writer
	archive string // archive directory
}

func main() {
	// parsing command line flags
	root := flag.String("root", ".",  "Root directory to start")
	logFile := flag.String("log", "", "Log deletes to this file")

	// action options
	list := flag.Bool("list", false, "List files only")
	del := flag.Bool("del", false, "Delete files")
	archive := flag.String("archive", "", "Archive directory")

	// filter options
	ext := flag.String("ext", "", "File extension to filter out")
	size := flag.Int64("size", 0, "Minimum file size")

	flag.Parse()

	var (
		f = os.Stdout
	)

	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_APPEND| os.O_CREATE| os.O_RDWR, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer f.Close()
	}

	c := config {
		ext : *ext,
		size : *size,
		list : *list,
		del : *del,
		wLog: f,
		archive: *archive,
	}

	if err := run(*root, os.Stdout, c); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(root string, out io.Writer, cfg config) error {
	delLogger := log.New(cfg.wLog, "DELETED FILE: ", log.LstdFlags)

	return filepath.Walk(root, 
		func (path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if filterOut(path, cfg.ext, cfg.size, info) {
				return nil
			}

			// list was explicitly set, don't do anything else
			if cfg.list {
				return listFile(path, out)
			}

			// Archive files and continue if successful
			if cfg.archive != "" {
				if err := archiveFile(cfg.archive, root, path); err != nil {
					return err
				}
			}

			// delete files
			if cfg.del {
				return delFile(path, delLogger)
			}

			// list is the default option if nothing else was set
			return listFile(path, out)
		})	
}