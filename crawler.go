package crawl

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type crawler struct {
	domain         *url.URL
	requestTimeout time.Duration
	visited        map[string]bool
	pending        map[string]bool
	todo           chan string
	results        chan *result
	workerSync     sync.WaitGroup
	workerStop     chan struct{}
}

type result struct {
	url   string
	links *[]string
}

// newCrawler returns an initialised crawler struct
func newCrawler(domain string, timeout time.Duration) (*crawler, error) {

	dURL, err := url.Parse(domain)
	if err != nil {
		return nil, err
	}

	return &crawler{
		domain:         dURL,
		requestTimeout: timeout,
		visited:        make(map[string]bool),
		pending:        make(map[string]bool),
		todo:           make(chan string, 100),
		results:        make(chan *result, 100),
		workerSync:     sync.WaitGroup{},
		workerStop:     make(chan struct{}),
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
func ScrapLinks(url string, timeout time.Duration) ([]string, error) {

	// Retrieve page
	body, err := download(url, timeout)
	defer func() {
		if body != nil {
			_ = body.Close()
		}
	}()
	if err != nil {
		return nil, err
	}

	// Retrieve links
	return ExtractLinks(url, body), nil
}

// scraper retrieves a webpage, parses it for links, keeps only domain or relative links, sanitises them, an returns the result
func (c *crawler) scraper(url string) {
	defer c.workerSync.Done()

	// Result will hold the links on success, or send as is on error
	res := newResult(url, nil)

	// Scrap and retrieve links
	links, err := ScrapLinks(url, c.requestTimeout)
	if err != nil {
		log.Errorf("Encountered error on page '%s' : %s", url, err)
	} else {
		// Filter links by current domain
		links = c.filterHost(links)
		res.links = &links
	}

	// Don't send results if we're being asked to stop
	select {
	case <-c.workerStop:
		return

	// Enqueue results
	case c.results <- res:
	}
}

// download retrieves the web page pointed to by the given url
func download(url string, timeout time.Duration) (io.ReadCloser, error) {

	var client = &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		if resp != nil {
			_ = resp.Body.Close()
		}
		return nil, err
	}

	return resp.Body, nil
}

// filterHost filters out links that are different from the crawler's scope
func (c *crawler) filterHost(links []string) []string {
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

// filterLinks filters out links that have already been visited or are in pending treatment
func (c *crawler) filterLinks(links []string) []string {
	n := 0
	for _, link := range links {
		keep := true

		// Check pending links
		if _, ok := c.pending[link]; ok {
			keep = false
		}

		// Check visited links
		if _, ok := c.visited[link]; ok {
			keep = false
		}

		// Keep the link
		if keep {
			links[n] = link
			n++
		}
	}
	return links[:n]
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
	filtered := c.filterLinks(*result.links)

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
	c.workerSync.Add(1)
	go c.scraper(url)
}

// crawl manages worker goroutines scraping pages and prints results
// todo : add a condition to quit when no more pages are to be visited
// todo : add a finer sync mechanism with workers when interrupting mid-request
// todo : keep a time tracker for stats
func crawl(domain string, syn *synchron) {
	defer syn.group.Done()

	c, err := newCrawler(domain, 4*time.Second)
	if err != nil {
		log.WithFields(log.Fields{"domain": domain}).Fatal(err)
	}
	c.todo <- c.domain.String()

loop:
	for {
		select {

		// Upon receiving a stop signal
		case <-syn.stopChan:
			log.Info("Stopping crawler.")
			close(c.workerStop)
			c.workerSync.Wait()
			fmt.Printf("Crawler visited a total of %d links starting from %s\n", len(c.visited), domain)
			break loop

		// Upon receiving a resulting from a worker scraping a page
		case result := <-c.results:
			c.handleResult(result)

		// For every link that is left to visit in the queue
		case link := <-c.todo:
			c.newTask(link)
		}
	}
}
