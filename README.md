# crawl
Simple web crawler with a single domain scope. Use it as a package or directly as an app.

The crawler scraps a page for links, follows them and scrapes them in the same fashion. 

You can launch the app with or without a timer (in seconds), like this :

> go run app/crawl.go https://monzo.com (-timeout=10)

todo : 
- fix timer
- first argument as domain
- add conditions for stop case / no more links to follow
- fix todo
- orient logging either to file or stdout

enhancements :
- make the crawler report back to a channel
- 