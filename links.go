package crawl

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"io"
	"net/url"
)

// extractLinks returns a slice of all links from an http.Get response body like reader object.
// Links won't contain queries or fragments
// It does not close the reader.
func extractLinks(origin string, body io.Reader) []string {
	tokenz := html.NewTokenizer(body)

	// This map is an intermediary container for found links, avoiding duplicates
	links := make(map[string]bool)

	for {
		ttype := tokenz.Next()

		if ttype == html.ErrorToken {
			break
		}

		token := tokenz.Token()
		if ttype == html.StartTagToken && token.Data == "a" {
			// If it's an anchor, try get the link
			if link, err := extractLink(origin, token); link != "" {
				links[link] = true
			} else {
				log.Error("Error in token '%s' : %s.", token.String(), err)
			}
		}
	}

	return mapToSlice(links)
}

// extractLink tries to return the link inside the token
func extractLink(origin string, token html.Token) (string, error) {
	// get href value
	for _, a := range token.Attr {
		if a.Key == "href" {
			return rel2abs(origin, a.Val)
		}
	}

	return "", errors.New("no href found in token")
}

// rel2abs returns the absolute url if the given link is relative to origin
func rel2abs(origin string, link string) (string, error) {
	log.Infof("Fixing %s from %s", link, origin)

	u, err := url.Parse(link)
	if err != nil {
		msg := fmt.Sprintf("Couldn't parse %s : %s", link, err)
		return "", errors.New(msg)
	}

	base, err := url.Parse(origin)
	if err != nil {
		msg := fmt.Sprintf("Couldn't parse %s : %s", origin, err)
		return "", errors.New(msg)
	}

	u = base.ResolveReference(u)

	log.Infof("Fixed %s to %s", link, u.Path)

	// extracting Path strips away the query and fragment
	return u.Path, nil
}

// mapToSlice returns a slice of strings containing the map's keys
func mapToSlice(_map map[string]bool) []string {
	// Extract the keys from map into a slice
	keys := make([]string, len(_map))
	i := 0
	for k := range _map {
		keys[i] = k
		i++
	}
	return keys
}
