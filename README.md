# crawl [![Go Report Card](https://goreportcard.com/badge/github.com/bytemare/gonetmon)](https://goreportcard.com/report/github.com/bytemare/gonetmon)[![codebeat badge](https://codebeat.co/badges/7e86ba65-e7b9-4982-9996-6b42c0eb763e)](https://codebeat.co/projects/github-com-bytemare-dbmon-master)[![GolangCI](https://golangci.com/badges/github.com/bytemare/dbmon.svg)](https://golangci.com/r/github.com/bytemare/dbmon)[![GoDoc](https://godoc.org/github.com/bytemare/crawl?status.svg)](https://godoc.org/github.com/bytemare/crawl)
Simple web crawler with a single domain scope. Use it as a package or directly as an app.

The crawler scraps a page for links, follows them and scrapes them in the same fashion. 

You can launch the app with or without a timeout (in seconds), like this :

> go run app/crawl.go (-timeout=10) https://monzo.com

However the program was launched, you can interrupt it with ctrl+c.

## Features

- single domain scope
- parallel scrawling
- optional timeout
- avoid loops on already visited links
- usable as a package by calling Crawl(), CrawlAsync() and ScrapLinks() functions

## todo
- unit tests
- platform tests
- coverage