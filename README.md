# Seeker [![GoDoc](https://godoc.org/github.com/itsmontoya/seeker?status.svg)](https://godoc.org/github.com/itsmontoya/seeker) ![Status](https://img.shields.io/badge/status-beta-yellow.svg)

Seeker is a file-seeking library which focuses on lines

## Usage
``` go
package main

import (
	"bytes"
	"fmt"

	"github.com/itsmontoya/async/file"
	"github.com/itsmontoya/seeker"
)

func main() {
	var (
		f   *file.File
		err error
	)

	// Create a file to write to
	// Note: async/file is used for asynchronous disk I/O, can be swapped with os.File
	if f, err = file.Create("test.db"); err != nil {
		panic(err)
	}
	defer f.Close()

	// Write three lines
	f.Write([]byte("0\n"))
	f.Write([]byte("1\n"))
	f.Write([]byte("2\n"))

	s := seeker.New(f)

	// Seek to the beginning of the file
	if err = s.SeekToStart(); err != nil {
		panic(err)
	}

	// Read and print each line
	s.ReadLines(func(buf *bytes.Buffer) (end bool) {
		fmt.Println(buf.String())
		return
	})

	// Step back one line
	if err = s.PrevLine(); err != nil {
		panic(err)
	}

	// Will print only the last line
	s.ReadLines(func(buf *bytes.Buffer) (end bool) {
		fmt.Println(buf.String())
		return
	})
}
```
