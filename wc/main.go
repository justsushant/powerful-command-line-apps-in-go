package main

import (
	"bufio"		// reads text
	"fmt"		// prints formatted output
	"io"		// provides io.Reader interface
	"os"		// uses os resources
	"flag"		// manage command line flags
)

func main() {
	// defining boolean flag "-l" to count lines
	lines := flag.Bool("l", false, "Count lines")
	// defining boolean flag "-b" to count bytes
	bytes := flag.Bool("b", false, "Count bytes")
	// defining string flag "-f" to accept files
	fileName := flag.String("f", "", "Accept file(s)")

	// parses all the flags
	flag.Parse()

	// calling the count function and checking error
	count, err := count(os.Stdin, *fileName, *lines, *bytes)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	fmt.Println(count)
}

func count(r io.Reader, fileName string, countLines, countBytes bool) (int, error) {
	// reads text from reader r
	scanner := bufio.NewScanner(r)

	// if fileName is supplied
	if fileName != "" {
		file, err := os.Open(fileName)
		if err != nil {
			return 0, err
		}

		scanner = bufio.NewScanner(file)
	}
	
	// defines scanner split type to scan words (default is set to ScanLines)
	// takes a function which splits according to the function passed in parameter
	if countBytes {
		scanner.Split(bufio.ScanBytes)
	} else if !countLines {
		scanner.Split(bufio.ScanWords)
	}

	// declares the counter
	wc := 0

	// increments wc for every word/line token scanned
	for scanner.Scan() {
		wc++
	}

	return wc, nil
}