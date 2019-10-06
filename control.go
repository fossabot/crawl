package crawl

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	log                      = logrus.New()
	logFilePerms os.FileMode = 0666
	logFilePath              = "./log-crawler.log"
)

// log2File switches logging to be output to file only
func log2File(logFile string) {
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, logFilePerms)
	if err == nil {
		log.Out = file
	} else {
		log.Info("Failed to log to file, using default stderr")
	}
	log.SetFormatter(&logrus.JSONFormatter{})
}

// timer implements a timeout (should be called as a goroutine)
func timer(syn *synchron) {
	defer syn.group.Done()

	if syn.timeout <= 0 {
		log.Info("No value assigned for timeout. Timer will not run.")
		return
	}

loop:
	for {
		select {

		// Quit if keyboard interruption
		case <-syn.stopChan:
			log.Trace("Timer received stop message. Stopping Timer.")
			break loop

		// When timeout is reached, inform of timeout, send signal, and quit
		case <-time.After(syn.timeout):
			log.Infof("Timing out after %d seconds.", syn.timeout/time.Second)
			syn.sendQuitSignal()
			break loop
		}
	}
}

// signalHandler is called as a goroutine to intercept signals and stop the program
func signalHandler(syn *synchron) {
	defer syn.group.Done()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	_ = <- sig
	log.Trace("signalHandler received a signal. Stopping.")
	if syn.checkout() {
		syn.stopChan <- struct{}{} // for timer
		syn.stopChan <- struct{}{} // for crawler
	}
}

// validateInput returns whether input is valid and can be worked with
func validateInput(domain string, timeout time.Duration) error {

	// We can't crawl without a target domain
	if domain == "" {
		return errors.New("if you want to crawl something, please specify the target domain as argument")
	}

	// Check whether domain is of valid form
	if _, err := url.ParseRequestURI(domain); err != nil {
		msg := fmt.Sprintf("Invalid url : you must specify a valid target domain/url to crawl : %s.", err)
		return errors.New(msg)
	}

	// Invalid timeout values are handled later, but let's not the user mess with us
	if timeout < 0 {
		msg := fmt.Sprintf("Invalid timeout value '%d' : you must specify a valid timeout in [0;+yourpatience[ in seconds.", timeout)
		return errors.New(msg)
	}

	return nil
}

// startCrawling launches the goroutines that constitute the crawler implementation.
func startCrawling(domain string, syn *synchron) {
	go signalHandler(syn)
	go timer(syn)
	go crawl(domain, syn)

	syn.group.Wait()
	close(syn.stopChan)

	log.WithField("url", domain).Info("Shutting down.")
}

// StreamLinks returns a channel on which it will report links as they come during the crawling.
// The timeout parameter allows for a time frame to crawl for.
func StreamLinks(domain string, timeout time.Duration) (outputChan chan *Result, err error) {

	if err = validateInput(domain, timeout); err != nil {
		return nil, err
	}

	syn := newSynchron(timeout, 3)
	log.Infof("Starting web crawler. You can interrupt the program any time with ctrl+c. Logging to %s.", logFilePath)
	log2File(logFilePath)

	go startCrawling(domain, syn)

	return syn.results, nil
}

// FetchLinks returns all visited links on crawling until exhaustion or up until timeout is reached.
func FetchLinks(domain string, timeout time.Duration) ([]string, error) {
	results, err := StreamLinks(domain, timeout)
	if err != nil {
		return nil, err
	}
	links := make([]string, 100) // todo : trade-off here, look if we really need that

	for res := range results {
		links = append(links, *res.Links...)
	}

	return links, nil
}
