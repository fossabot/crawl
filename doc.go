/*
Package crawl is a simple web crawler with single domain scope.
It can be limited with a timeout and interrupted with SIGINT or SIGTERM.

The FetchLinks and StreamLinks functions start at a given url, scrap the content of the corresponding web page for links
to the same host, and crawl the whole web site through all encountered links.
The return values can be used for a site map.

Some precautions have been taken to prevent infinite loops, like stripping queries and fragments off urls.

A sample program calling the package is given in the project repository.
*/
package crawl
