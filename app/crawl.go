package main

import (
	"flag"
	"fmt"
	"github.com/bytemare/crawl"
	"os"
	"path/filepath"
	"time"
)

func main() {

	// Define and parse command line arguments
	timeout := flag.Int("timeout", 0, "crawling time, in seconds. 0 or none is infinite.")

	if len(os.Args) < 2 {
		fmt.Printf("Expecting at least an url as entry point. e.g. './%s https://bytema.re'\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	domain := os.Args[1]

	// Launch crawler
	err := crawl.Crawl(domain, time.Duration(*timeout)*time.Second)

	if err != nil {
		fmt.Printf("Error : %s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
