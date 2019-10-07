package crawl_test

import (
	"github.com/bytemare/crawl"
	"testing"
	"time"
)

func TestFetchLinks(t *testing.T){
	// With empty url
	url := ""
	timeout := 0
	output, err := crawl.FetchLinks(url, time.Duration(timeout))
	if err == nil || output != nil {
		t.Error("StreamLinks returned without error, but url is empty.")
	}

	// With invalid url name
	url = "bytema.re"
	output, err = crawl.FetchLinks(url, time.Duration(timeout))
	if err == nil || output != nil {
		t.Errorf("StreamLinks returned without error, but url is invalid. URL : %s.", url)
	}

	// With valid domain but invalid timeout
	url = "https://bytema.re"
	timeout = -10
	output, err = crawl.FetchLinks(url, time.Duration(timeout))
	if err == nil || output != nil {
		t.Errorf("StreamLinks returned without error, but timeout is invalid. URL : %d.", timeout)
	}

	errMsg := "StreamLinks returned with error, but url and timeout are valid. URL : %s, timeout : %d."

	// With valid domain name and 0 timeout
	timeout = 0
	output, err = crawl.FetchLinks(url, time.Duration(timeout))
	if err != nil || output == nil {
		t.Errorf(errMsg, url, timeout)
	}

	// With valid domain name and low timeout
	timeout = 2
	output, err = crawl.FetchLinks(url, time.Duration(timeout))
	if err != nil || output == nil {
		t.Errorf(errMsg, url, timeout)
	}

	// With valid domain name and high timeout
	timeout = 10
	output, err = crawl.FetchLinks(url, time.Duration(timeout))
	if err != nil || output == nil {
		t.Errorf(errMsg, url, timeout)
	}
}