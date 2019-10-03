package main

import (
	"flag"
	"github.com/bytemare/crawl"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

func main() {

	// Define and parse command line arguments
	timeout := flag.Int("timeout", 0, "crawling time, in seconds. 0 or none is infinite.")
	domain := flag.String("domain", "", "crawling scope / target domain")
	flag.Parse()

	// Launch crawler
	err := crawl.Crawl(*domain, time.Duration(*timeout)*time.Second)

	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
	os.Exit(0)
}
