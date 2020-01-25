package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

var (
	addr = flag.String("addr", "127.0.0.1:8080", "listen address")
)

func main() {
	flag.Usage = printUsage
	flag.Parse()
	if flag.NArg() == 0 {
		printUsage()
		os.Exit(2)
	}
	library := flag.Arg(0)

	s, err := newServer(library)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", s.rootHandler)
	http.Handle("/photos/", http.StripPrefix("/photos/", http.HandlerFunc(s.photosHandler)))

	fs := http.FileServer(http.Dir(library))
	http.Handle("/masters/", fs)
	http.Handle("/Masters/", fs) // since fs is case-sensitive

	log.Printf("listening on %q\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "usage: %s [options] <*.photoslibrary>\n", os.Args[0])
	flag.PrintDefaults()
}
