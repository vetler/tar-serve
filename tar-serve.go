package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func usage() {
	fmt.Println("usage: tar-serve <filename.[tar.gz/bz2|tgz>")
}

func main() {
	if len(os.Args[1:]) != 1 {
		usage()
		os.Exit(-1)
	}

	filename := os.Args[1]
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Println("No such file or directory: %s", filename)
		return
	}

	if strings.HasSuffix(filename, "tgz") || strings.HasSuffix(filename, "tar.gz") {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		http.HandleFunc("/", archiveHandler(file))
		http.ListenAndServe("localhost:4000", nil)
	} else {
		fmt.Println("File type not supported")
	}
}

func archiveHandler(file *os.File) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		file.Seek(0, 0)
		zip, err := gzip.NewReader(file)
		if err != nil {
			log.Fatal(err)
		}
		defer zip.Close()

		archive := tar.NewReader(zip)

		fmt.Fprintln(w, "<!DOCTYPE html>")

		if r.URL.String() == "/" {
			for header, err := archive.Next(); err == nil; header, err = archive.Next() {
				io.WriteString(w, "<li><a href=\""+header.Name+"\">"+header.Name+"</a>")
			}
		} else {
			for header, err := archive.Next(); err == nil; header, err = archive.Next() {
				if header.Name == r.URL.String()[1:] {
					bytes := make([]byte, header.FileInfo().Size())
					archive.Read(bytes)
					w.Write(bytes)
				}
			}
		}
	}
}

func fatal(err error) {
	fmt.Println(err)
	os.Exit(-1)
}
