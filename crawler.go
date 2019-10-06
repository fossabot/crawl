package crawl

import (
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
	pending        map[string]int
	failed         map[string]bool
	maxRetry       int
	todo           chan string
	results        chan *Result
	workerSync     sync.WaitGroup
	workerStop     chan struct{}
	output         chan<- *Result
}

type Result struct {
	Url   string
	Links *[]string
	err   error
}

// newCrawler returns an initialised crawler struct
func newCrawler(domain string, output chan<- *Result, timeout time.Duration, maxRetry int) (*crawler, error) {

	dURL, err := url.Parse(domain)
	if err != nil {
		return nil, err
	}

	return &crawler{
		domain:         dURL,
		requestTimeout: timeout,
		visited:        make(map[string]bool),
		pending:        make(map[string]int),
		failed:         make(map[string]bool),
		maxRetry:       maxRetry,
		todo:           make(chan string, 100),
		results:        make(chan *Result, 100),
		workerSync:     sync.WaitGroup{},
		workerStop:     make(chan struct{}),
		output:         output,
	}, nil
}

//  newResult returns an initialised Result struct
func newResult(url string, links *[]string) *Result {
	return &Result{
		Url:   url,
		Links: links,
		err:   nil,
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

// scraper retrieves a webpage, parses it for links, keeps only domain or relative links, sanitises them, an returns the Result
func (c *crawler) scraper(url string) {
	defer c.workerSync.Done()

	// Result will hold the links on success, or send as is on error
	res := newResult(url, nil)

	// Scrap and retrieve links
	links, err := ScrapLinks(url, c.requestTimeout)
	if err != nil {
		res.err = err
	} else {
		// Filter links by current domain
		links = c.filterHost(links)
		res.Links = &links
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

	log.Tracef("Attempting download of %s.", url)
	resp, err := client.Get(url)
	if err != nil {
		if resp != nil {
			_ = resp.Body.Close()
		}
		return nil, err
	}
	log.Tracef("Download of %s succeeded.", url)
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
			log.WithField("host", c.domain.Host).Tracef("Filtering out link to %s.", link)
		}
	}
	return links[:n]
}

// filterLinks filters out links that have already been visited or are in pending treatment
func (c *crawler) filterLinks(links []string) []string {
	n := 0
	// Only keep links that are neither pending or visited
	for _, link := range links {

		// If pending, skip
		if _, ok := c.pending[link]; ok {
			log.WithField("status", "pending").Tracef("Discarding %s.", link)
			continue
		}

		// If visited, skip
		if _, ok := c.visited[link]; ok {
			log.WithField("status", "pending").Tracef("Discarding %s.", link)
			continue
		}

		// Keep the link
		links[n] = link
		n++
	}
	return links[:n]
}

// handleResultError handles the error a Result has upon return of a link scraping attempt
func (c *crawler) handleResultError(res *Result) {
	log.WithField("url", res.Url).Tracef("Result returned with error : %s", res.err)

	// If we tried to much, mark it as failed
	if c.pending[res.Url] >= c.maxRetry {
		c.failed[res.Url] = true
		delete(c.pending, res.Url)
		log.Errorf("Discarding %d, page unreachable after %d attempts.\n", res.Url, c.maxRetry)
		return
	}

	// If we have not reached maximum retries, re-enqueue
	c.todo <- res.Url
	return
}

// handleResult treats the Result of scraping a page for links
func (c *crawler) handleResult(result *Result) {

	if result.err != nil {
		c.handleResultError(result)
		return
	}

	// Change state from pending to visited
	c.visited[result.Url] = true
	delete(c.pending, result.Url)

	// Filter out already visited links
	log.Tracef("Filtering links for %s.", result.Url)
	filtered := c.filterLinks(*result.Links)

	// Add filtered list in queue of links to visit
	for _, link := range filtered {
		c.todo <- link
	}

	// Log Result and send them to caller
	log.Infof("Found %d unvisited links on page %s : %s\n", len(filtered), result.Url, filtered)
	c.output <- result
}

// newTask triggers a new visit on a link
func (c *crawler) newTask(url string) {
	// Add to pending tasks
	c.pending[url]++

	// Launch a worker goroutine on that link
	c.workerSync.Add(1)
	go c.scraper(url)
}

// checkProgress verifies if there are pages left to scrap or being scraped. Returns false if not.
func (c *crawler) checkProgress() bool {
	return len(c.todo) != 0 || len(c.pending) != 0
}

// crawl manages worker goroutines scraping pages and prints results
func crawl(domain string, syn *synchron) {
	defer syn.group.Done()

	ticker := time.NewTicker(time.Second)
	c, err := newCrawler(domain, syn.results, 5*time.Second, 3)
	if err != nil {
		log.WithField("domain", domain).Error(err)
		goto quit
	}
	c.todo <- c.domain.String()

loop:
	for {
		select {

		// Upon receiving a stop signal
		case <-syn.stopChan:
			break loop

		// Upon receiving a resulting from a worker scraping a page
		case result := <-c.results:
			c.handleResult(result)

		// For every link that is left to visit in the queue
		case link := <-c.todo:
			c.newTask(link)

		// Every tick, verify if there are jobs or pending tasks left
		case <-ticker.C:
			if c.checkProgress() == false {
				log.Info("No links left to explore.")
				break loop
			}
		}
	}

	close(c.workerStop)
	log.Info("Stopping crawler.")
	c.workerSync.Wait()
	log.Infof("Visited %d links starting from %s\n", len(c.visited), domain)
quit:
	ticker.Stop()
	syn.sendQuitSignal()
	close(syn.results)
}
