package crawl_test

import (
	"github.com/bytemare/crawl"
	"testing"
	"time"
)

type Test struct {
	url     string
	timeout time.Duration
	errMsg  string
}

// TestFetchLinksFail tests cases where FetchLinks is supposed to fail and/or return an error
func TestFetchLinksFail(t *testing.T) {
	failing := []Test{
		{"", 0 * time.Second, "StreamLinks returned without error, but url is empty."},
		{"bytema.re", 0 * time.Second, "StreamLinks returned without error, but url is invalid."},
		{"https://bytema.re", -10 * time.Second, "StreamLinks returned without error, but timeout is invalid."},
	}

	for _, test := range failing {
		output, err := crawl.FetchLinks(test.url, test.timeout)
		if err == nil || output != nil {
			t.Errorf("%s URL : %s, timeout %d.", test.errMsg, test.url, test.timeout)
		}
	}
}

// TestFetchLinksSuccess tests cases where FetchLinks is supposed to succeed
func TestFetchLinksSuccess(t *testing.T) {
	var succeed = []Test{
		{"https://bytema.re", 10 * time.Second, ""},
		{"https://bytema.re", 250 * time.Millisecond, ""},
		{"https://bytema.re", 0 * time.Second, ""},
	}
	errMsg := "StreamLinks returned with error, but url and timeout are valid. URL : %s, timeout : %0.3fs."

	for _, test := range succeed {
		output, err := crawl.FetchLinks(test.url, test.timeout)
		if err != nil || output == nil {
			t.Errorf(errMsg, test.url, test.timeout.Seconds())
		}
	}
}
