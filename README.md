# crawl
Simple web crawler with a single domain scope. Use it as a package or directly as an app.

The crawler scraps a page for links, follows them and scrapes them in the same fashion. 

You can launch the app with or without a timeout (in seconds), like this :

> go run app/crawl.go https://monzo.com (-timeout=10)

## Features

- single domain scope
- parallel scrawling
- avoid loops on already visited links
- usable as a package by calling Crawl() and extractLinks() functions

## todo 
- fix timer
- first argument as domain
- add conditions for stop case / no more links to follow
- fix todo
- orient logging level and either to file or stdout

enhancements :
- make the crawler report back to a caller provided channel