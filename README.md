# crawl
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