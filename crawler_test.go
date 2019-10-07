package crawl_test

import (
	"github.com/bytemare/crawl"
	"testing"
	"time"
)

type Test struct {
	url	string
	timeout	time.Duration
	errMsg	string
}

var failing = []Test{
	{"", 0, "StreamLinks returned without error, but url is empty."},
	{"bytema.re", 0, "StreamLinks returned without error, but url is invalid."},
	{"https://bytema.re", time.Duration(-10), "StreamLinks returned without error, but timeout is invalid."},
}

// TestFetchLinksFail tests cases where FetchLinks is supposed to fail and/or return an error
func TestFetchLinksFail(t *testing.T){
	for _, test := range failing {
		output, err := crawl.FetchLinks(test.url, test.timeout)
		if err == nil || output != nil {
			t.Errorf("%s URL : %s, timeout %d.", test.errMsg, test.url, test.timeout)
		}
	}
}

var errMsg = "StreamLinks returned with error, but url and timeout are valid. URL : %s, timeout : %d."
var succeed = []Test{
	{"https://bytema.re", 0, ""},
	{"https://bytema.re", 2, ""},
	{"https://bytema.re", 10, ""},
}

// TestFetchLinksSuccess tests cases where FetchLinks is supposed to succeed
func TestFetchLinksSuccess(t *testing.T) {
	for _, test := range succeed {
		output, err := crawl.FetchLinks(test.url, test.timeout)
		if err != nil || output == nil {
			t.Errorf(errMsg, test.url, test.timeout)
		}
	}
}