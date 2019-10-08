# crawl

The crawler scraps a page for links, follows them and scrapes them in the same fashion.

You can launch the app with or without a timeout (in seconds), like this :

```go
go run app/crawl.go (-timeout=10) https://monzo.com
```

However the program was launched, you can interrupt it with ctrl+c.

## Features

- single domain scope
- parallel scrawling
- optional timeout
- scraps queries and fragments from url
- avoid loops on already visited links
- usable as a package by calling FetchLinks(), StreamLinks() and ScrapLinks() functions
- logs to file in JSON for log aggregation
