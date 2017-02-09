package dotler_test

import (
	"context"
	"flag"
	"github.com/ronin13/dotler/dotler"
	"net/url"
	"sync"
	"testing"
)

type NodeMap struct {
}

func (node *NodeMap) Add(key string, value *dotler.Page) error {
	return nil
}

func (node *NodeMap) Exists(key string) *dotler.Page {
	return nil
}

func (node *NodeMap) RunLoop(stopLoop context.Context) {
	return
}

func TestDotlerCrawl(t *testing.T) {

	nd := &NodeMap{}

	flag.Lookup("alsologtostderr").Value.Set("false")
	var wg sync.WaitGroup
	var reqChan, dotChan chan *dotler.Page
	testURLs := []struct {
		turl      string
		linkCount int
	}{
		{"http://www.wnohang.net/pages/about/", 8},
		{"http://gneuron.freehostia.com/", 4},
	}

	for _, urls := range testURLs {

		reqChan = make(chan *dotler.Page, dotler.MAXWORKERS)
		dotChan = make(chan *dotler.Page, dotler.MAXWORKERS)
		parsedURL, _ := url.Parse(urls.turl)
		samplePage := &dotler.Page{PageURL: parsedURL}
		wg.Add(1)
		dotler.Crawl(context.Background(), samplePage, reqChan, dotChan, &wg, nd)
		wg.Wait()
		if len(reqChan) != urls.linkCount {
			t.Fatalf("Failed to crawl %s: %d %d", urls.turl, len(reqChan), urls.linkCount)
		}
	}
}

func BenchmarkCrawl(b *testing.B) {

	nd := &NodeMap{}
	flag.Lookup("alsologtostderr").Value.Set("false")
	reqChan := make([]chan *dotler.Page, b.N)
	dotChan := make([]chan *dotler.Page, b.N)
	bURL := "http://www.wnohang.net/pages/about/"
	for n := 0; n < b.N; n++ {

		var wg sync.WaitGroup
		reqChan[n] = make(chan *dotler.Page, dotler.MAXWORKERS)
		dotChan[n] = make(chan *dotler.Page, dotler.MAXWORKERS)
		parsedURL, _ := url.Parse(bURL)
		samplePage := &dotler.Page{PageURL: parsedURL}
		wg.Add(1)
		dotler.Crawl(context.Background(), samplePage, reqChan[n], dotChan[n], &wg, nd)
		wg.Wait()
	}
}
