package crawl

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
)

type crawler struct {
	domain  *url.URL
	visited map[string]bool
	pending map[string]bool
	todo    chan string
	results chan *result
}

type result struct {
	url   string
	links *[]string
}

// newCrawler returns an initialised crawler struct
func newCrawler(domain string) (*crawler, error) {

	dURL, err := url.Parse(domain)
	if err != nil {
		return nil, err
	}

	return &crawler{
		domain:  dURL,
		visited: make(map[string]bool),
		pending: make(map[string]bool),
		todo:    make(chan string, 100),
		results: make(chan *result, 100),
	}, nil
}

//  newResult returns an initialised result struct
func newResult(url string, links *[]string) *result {
	return &result{
		url:   url,
		links: links,
	}
}

// ScrapLinks returns the links found in the web page pointed to by url
func ScrapLinks(url string) ([]string, error) {

	// Retrieve page
	body, err := download(url)
	defer func() {
		if body != nil{
			_ = body.Close()
		}
	}()
	if err != nil {
		return nil, err
	}

	// Retrieve links
	return extractLinks(url, body), nil
}

// scraper retrieves a webpage, parses it for links, keeps only domain or relative links, sanitises them, an returns the result
func (c *crawler) scraper(url string) {

	// Scrap and retrieve links
	links, err := ScrapLinks(url)
	if err != nil {
		log.Errorf("Encountered error on page '%s' : %s", url, err)
		c.results <- newResult(url, nil)
		return
	}

	// Filter links by current domain
	links = c.filterDomain(links)

	// Enqueue results
	log.Infof("Found %d links on page %s\n", len(links), url)
	c.results <- newResult(url, &links)
}

// download retrieves the web page pointed to by the given url
func download(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		if resp != nil {
			_ = resp.Body.Close()
		}
		return nil, err
	}

	return resp.Body, nil
}

// filterDomain filters out links that are different from the crawler's scope
func (c *crawler) filterDomain(links []string) []string{
	n := 0
	for _, link := range links {
		linkURL, _ := url.Parse(link)
		if linkURL.Host == c.domain.Host {
			links[n] = link
			n++
		} else {
			//log.Infof("Filtering out element ", link)
		}
	}
	return links[:n]
}

// filterVisited filters out links that have already been visited
func (c *crawler) filterVisited(links *[]string) []string {

	filtered := make([]string, len(*links))
	// todo : modifying the slice in-place may be more efficient, setting a string to "" if don't keep
	//  it, and then only send to channel if it's non-""

	// filter out already encountered links
	for _, link := range *links {
		if _, ok := c.visited[link]; ok == false {
			// If value is not in map, we haven't visited it, thus keeping it
			filtered = append(filtered, link)
		}
	}

	return filtered
}

// handleResult treats the result of scraping a page for links
func (c *crawler) handleResult(result *result) {
	delete(c.pending, result.url)

	// If the download failed and links is nil
	if result.links == nil {
		// todo : handle pages that continuously fail on download (struct for each link with nb of retries)
		c.todo <- result.url
		return
	}

	// Change state from pending to visited
	c.visited[result.url] = true

	// Filter out already visited links
	filtered := c.filterVisited(result.links)
	log.Infof("Filtered out %d visited links.", len(*result.links) - len(filtered))

	// Add filtered list in queue of links to visit
	for _, link := range filtered {
		c.todo <- link
	}

	// Print out result
	fmt.Printf("Found %d unvisited links on page %s : %s\n", len(filtered), result.url, filtered)
}

// newTask triggers a new visit on a link
// todo : change that name
func (c *crawler) newTask(url string) {
	// Add to pending tasks
	c.pending[url] = true

	// Launch a worker goroutine on that link
	go c.scraper(url)
}

// crawl manages worker goroutines scraping pages and prints results
// todo : add a condition to quit when no more pages are to be visited
// todo : add a finer sync mechanism with workers when interrupting mid-request
// todo : keep a time tracker for stats
func crawl(domain string, syn *synchron) {
	defer syn.group.Done()

	c, err := newCrawler(domain)
	if err != nil {
		log.WithFields(log.Fields{"domain" : domain,}).Fatal(err)
	}
	c.todo <- c.domain.String()

loop:
	for {
		select {

		// Upon receiving a stop signal
		case <-syn.stopChan:
			log.Info("Stopping crawler.")
			close(c.todo)
			close(c.results)
			break loop

		// Upon receiving a resulting from a worker scraping a page
		case result := <-c.results:
			c.handleResult(result)

		// For every link that is left to visit in the queue
		case link := <-c.todo:
			log.Info("New task on " + link)
			c.newTask(link)
		}
	}
}
