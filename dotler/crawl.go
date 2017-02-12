// Copyright 2017 Raghavendra Prabhu.
// Refer to LICENSE for more

// Core crawler to process page.

package dotler

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/PuerkitoBio/purell"
	"github.com/golang/glog"
	wire "github.com/ronin13/dotler/wire"

	"context"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Iterates over attributes, parses the page,
// gets URLs from same domain, gets static assets
// sends new links onto reqChan.
func updateAttr(item *goquery.Selection, inPage *wire.Page, attribTypes []string, reqChan chan *wire.Page, nodes wire.NodeMapper) error {

	var nPage *wire.Page
	var err error
	var statTitle string
	var parsedURL *url.URL

	base := inPage.PageURL
	for _, attribs := range attribTypes {
		if link, exists := item.Attr(attribs); exists {
			// We skip data URIs
			if strings.Contains(link, "data:") {
				glog.Infoln("Skipping data uri")
				return nil
			}
			// Normalizing links.
			link, err = purell.NormalizeURLString(link, purell.FlagsUsuallySafeGreedy)

			if err != nil {
				glog.Infof("Failed to normalize %s with error %s", link, err)
				return err
			}
			parsedURL, err = url.Parse(link)
			if err != nil {
				glog.Infof("Failed to parse %s with error %s", link, err)
				return err
			}
			if !parsedURL.IsAbs() {
				parsedURL = base.ResolveReference(parsedURL)
			}
			parsedURL.RawQuery = ""
			parsedURL.Fragment = ""
			if isStatic(parsedURL.String()) {
				if _, exists := inPage.StatList[parsedURL.String()]; !exists {
					statTitle = getStatTitle(parsedURL)
					inPage.StatList[parsedURL.String()] = wire.StatPage{
						StaticURL: parsedURL,
						PageTitle: statTitle}
				}

			} else if parsedURL.Host == base.Host {

				nPage = nodes.Exists(parsedURL.String())

				// Already processed
				if nPage != nil {
					updateOutLinksWithCard(parsedURL.String(), inPage, nPage)
				} else {
					// New discovery!

					// Title not known at this point
					nPage = &wire.Page{PageURL: parsedURL}

					//TODO: go writeToChan?
					writeToChan(nPage, reqChan)
					updateOutLinksWithCard(parsedURL.String(), inPage, nPage)
				}
			} else {
				// Very verbose!
				if glog.V(2) {
					glog.Infof("Skipping %s", link)
				}
			}
		}
	}
	return nil
}

func updateOutLinksWithCard(key string, iPage, nPage *wire.Page) {

	if _, exists := iPage.OutLinks[key]; exists {
		iPage.OutLinks[key].Card++
	} else {
		iPage.OutLinks[key] = &wire.PageWithCard{Page: nPage, Card: 1}
	}
}

// Get all links from a html page
// Tags checked: <a> <img> <script> <link>
// For attributes: href and src
// Updates Page structure with static and outside links.
// Uses goquery for parsing.
func getAllLinks(cancelParse context.Context, inPage *wire.Page, reqChan chan *wire.Page, nodes wire.NodeMapper) chan bool {

	doneChan := make(chan bool, 1)

	go func() {
		// getContent has a timeout - clientTimeout
		body, err := getContent(inPage.PageURL)
		if err != nil {
			glog.Infof("Failed to crawl %s", inPage.PageURL.String())
			inPage.FailCount++
			if inPage.FailCount <= maxFetchFail {
				//TODO: go writeToChan?
				writeToChan(inPage, reqChan)
			}
			doneChan <- false
			return
		}

		inPage.OutLinks = make(map[string]*wire.PageWithCard)
		inPage.StatList = make(map[string]wire.StatPage)
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
		panicCrawl(err)

		successful := true

		doc.Find("a, img, script, link, source").EachWithBreak(func(i int, item *goquery.Selection) bool {
			select {
			case <-cancelParse.Done():
				glog.Infof("Cancelling further processing here")
				successful = false
				return false
			default:
				err = updateAttr(item, inPage, []string{"href", "src"}, reqChan, nodes)
				if err != nil {
					glog.Infof("Skipping this - %s - page, probably bad", inPage.PageURL.String())
					successful = false
					return false
				}
			}
			return true
		})

		doneChan <- successful

	}()
	return doneChan

}

// Core crawl function called from main.
// Uses getAllLinks for actual processing.
// updNodeMap to ensure nodes processed successfully are not revisited.
// Gets two channels - reqChan and respChan.
// Sends reqChan downwards for further parse + load.
// Uses respChan for graph rendering.
// Also has a timeout of crawlThreshold.
// Uses a new child context noParse - used to terminate parsing.
func Crawl(cancelCrawl context.Context, inPage *wire.Page, reqChan chan *wire.Page, respChan chan *wire.Page, waiter *sync.WaitGroup, nodes wire.NodeMapper) {

	defer waiter.Done()
	if err := nodes.Add(inPage.PageURL.String(), inPage); err != nil {
		if glog.V(2) {
			glog.Errorf("Possible duplicate addition %s", inPage.PageURL.String())
		}
		return
	}

	glog.Infof("Processing page %s", inPage.PageURL.String())

	noParse, terminate := context.WithCancel(cancelCrawl)
	doneChan := getAllLinks(noParse, inPage, reqChan, nodes)

	for {
		select {
		case <-cancelCrawl.Done():
			terminate()
			atomic.AddUint64(&crawlCancelled, 1)
			glog.Infof("Cancelling crawling the page %s", inPage.PageURL.String())
			return
		case rval := <-doneChan:
			terminate()
			if rval == false {
				atomic.AddUint64(&crawlFail, 1)
				glog.Infof("Failed to crawl %s", inPage.PageURL.String())
				return
			}

			atomic.AddUint64(&crawlSuccess, 1)
			glog.Infof("Successfully crawled %s", inPage.PageURL.String())

			if genGraph {
				//TODO: go writeToChan?
				writeToChan(inPage, respChan)
			}
			return
		case <-time.After(time.Second * time.Duration(crawlThreshold)):
			atomic.AddUint64(&crawlSkipped, 1)
			terminate()
			glog.Infof("This page %s is taking too long (> %d), skipping it!", inPage.PageURL.String(), crawlThreshold)
			return
		}
	}

}
