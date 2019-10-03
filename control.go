package crawl

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// synchron holds the synchronisation tools and parameters
type synchron struct {
	timeout     time.Duration
	group       sync.WaitGroup
	timeoutChan chan time.Duration
	stopChan    chan struct{}
}

// newSynchron returns an initialised synchron struct
func newSynchron(timeout time.Duration, nbParties int) *synchron {
	s := &synchron{
		timeout:     timeout,
		group:       sync.WaitGroup{},
		timeoutChan: make(chan time.Duration),
		stopChan:    make(chan struct{}),
	}

	s.group.Add(nbParties)
	return s
}

// timer implements a timeout (should be called as a goroutine)
func timer(syn *synchron) {
	defer syn.group.Done()

	if syn.timeout <= 0 {
		return
	}

loop:
	for {
		select {

		// Quit if keyboard interruption
		case <-syn.stopChan:
			break loop

		// When timeout is reached, inform of timeout, send signal, and quit
		case <-time.After(syn.timeout):
			syn.timeoutChan <- syn.timeout
			// Send interrupt signal to ourselves, intercepted by signalHandler
			p, _ := os.FindProcess(os.Getpid())
			_ = p.Signal(os.Interrupt)
			break loop
		}
	}
}

// signalHandler is called as a goroutine to intercept signals and stop the program
func signalHandler(syn *synchron) {
	defer syn.group.Done()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	for s := range sig {

		if len(syn.timeoutChan) > 0 {
			log.Infof("Timing out after %d seconds. Shutting down.", <-syn.timeoutChan)
		} else {
			log.Infof("Crawler received signal : ", s.String())
			if syn.timeout <= 0 {
				syn.stopChan <- struct{}{} // for timer
			}
		}

		syn.stopChan <- struct{}{} // for crawler
		break
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

// Crawl implements the crawler with a control frame (timeout and/or interruption)
func Crawl(domain string, timeout time.Duration) (err error) {

	if err = validateInput(domain, timeout); err != nil {
		return err
	}

	syn := newSynchron(timeout, 3)

	log.Info("Starting web crawler. You can interrupt the program any time with ctrl+c or ctrl+z.")

	go signalHandler(syn)
	go timer(syn)
	go crawl(domain, syn)

	syn.group.Wait()
	close(syn.timeoutChan)
	close(syn.stopChan)

	log.Info("Crawler now shutting down.")

	return err
}
