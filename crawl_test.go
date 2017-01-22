package main

import (
	"context"
	"flag"
	"net/url"
	"sync"
	"testing"
)

func TestDotlerCrawl(t *testing.T) {

	flag.Lookup("alsologtostderr").Value.Set("false")
	var wg sync.WaitGroup
	var reqChan, dotChan chan *Page
	testURLs := []struct {
		turl      string
		linkCount int
	}{
		{"http://www.wnohang.net/pages/about/", 7},
		{"http://gneuron.freehostia.com/", 4},
	}

	for _, urls := range testURLs {
		reqChan = make(chan *Page, MAXWORKERS)
		dotChan = make(chan *Page, MAXWORKERS)
		parsedURL, _ := url.Parse(urls.turl)
		samplePage := &Page{pageURL: parsedURL}
		wg.Add(1)
		crawl(context.Background(), samplePage, reqChan, dotChan, &wg)
		wg.Wait()
		if len(reqChan) != urls.linkCount {
			t.Fatalf("Failed to crawl %s", urls.turl)
		}
	}
}

func BenchmarkCrawl(b *testing.B) {
	flag.Lookup("alsologtostderr").Value.Set("false")
	reqChan := make([]chan *Page, b.N)
	dotChan := make([]chan *Page, b.N)
	bUrl := "http://www.wnohang.net/pages/about/"
	for n := 0; n < b.N; n++ {
		var wg sync.WaitGroup
		reqChan[n] = make(chan *Page, MAXWORKERS)
		dotChan[n] = make(chan *Page, MAXWORKERS)
		parsedURL, _ := url.Parse(bUrl)
		samplePage := &Page{pageURL: parsedURL}
		wg.Add(1)
		crawl(context.Background(), samplePage, reqChan[n], dotChan[n], &wg)
		wg.Wait()
		if len(reqChan[n]) != 7 {
			b.Fatal("Crawl benchmark failed")
		}
	}
}
