// Copyright 2017 Raghavendra Prabhu.
// Refer to LICENSE for more

// Core crawler to process page.

package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/PuerkitoBio/purell"
	"github.com/golang/glog"

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
func updateAttr(item *goquery.Selection, inPage *Page, attribTypes []string, reqChan chan *Page, nodes *NodeMap) error {

	var nPage *Page
	var err error
	var statTitle string
	var parsedURL *url.URL

	base := inPage.pageURL
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
				if _, exists := inPage.statList[parsedURL.String()]; !exists {
					statTitle = getStatTitle(parsedURL)
					inPage.statList[parsedURL.String()] = StatPage{
						staticURL: parsedURL,
						pageTitle: statTitle}
				}

			} else if parsedURL.Host == base.Host {
				parsedURL.RawQuery = ""
				parsedURL.Fragment = ""

				glog.Infof("Checking for %s", parsedURL.String())
				nPage = nodes.Exists(parsedURL.String())

				// Already processed
				if nPage != nil {
					glog.Infof("Page already exists %s", nPage.pageURL.String())
					updateOutLinksWithCard(parsedURL.String(), inPage, nPage)
				} else {
					// New discovery!

					// Title not known at this point
					nPage = &Page{pageURL: parsedURL}

					go writeToChan(nPage, reqChan)
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

func updateOutLinksWithCard(key string, iPage, nPage *Page) {

	if _, exists := iPage.outLinks[key]; exists {
		iPage.outLinks[key].card++
	} else {
		iPage.outLinks[key] = &PageWithCard{page: nPage, card: 1}
	}
}

// Get all links from a html page
// Tags checked: <a> <img> <script> <link>
// For attributes: href and src
// Updates Page structure with static and outside links.
// Uses goquery for parsing.
func getAllLinks(cancelParse context.Context, inPage *Page, reqChan chan *Page, nodes *NodeMap) chan bool {

	doneChan := make(chan bool, 1)

	go func() {
		// getContent has a timeout - clientTimeout
		body, err := getContent(inPage.pageURL)
		if err != nil {
			glog.Infof("Failed to crawl %s", inPage.pageURL.String())
			inPage.failCount++
			if inPage.failCount <= maxFetchFail {
				go writeToChan(inPage, reqChan)
			}
			doneChan <- false
			return
		}

		inPage.outLinks = make(map[string]*PageWithCard)
		inPage.statList = make(map[string]StatPage)
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
					glog.Infof("Skipping this - %s - page, probably bad", inPage.pageURL.String())
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
func crawl(cancelCrawl context.Context, inPage *Page, reqChan chan *Page, respChan chan *Page, waiter *sync.WaitGroup, nodes *NodeMap) {

	defer waiter.Done()
	if err := nodes.Add(inPage.pageURL.String(), inPage); err != nil {
		glog.Errorf("Possible duplicate addition %s", inPage.pageURL.String())
		return
	}

	glog.Infof("Processing page %s", inPage.pageURL.String())

	noParse, terminate := context.WithCancel(cancelCrawl)
	doneChan := getAllLinks(noParse, inPage, reqChan, nodes)

	for {
		select {
		case <-cancelCrawl.Done():
			terminate()
			atomic.AddUint64(&crawlCancelled, 1)
			glog.Infof("Cancelling crawling the page %s", inPage.pageURL.String())
			return
		case rval := <-doneChan:
			terminate()
			if rval == false {
				atomic.AddUint64(&crawlFail, 1)
				glog.Infof("Failed to crawl %s", inPage.pageURL.String())
				return
			}

			atomic.AddUint64(&crawlSuccess, 1)
			glog.Infof("Successfully crawled %s", inPage.pageURL.String())

			if genGraph {
				go writeToChan(inPage, respChan)
			}
			return
		case <-time.After(time.Second * time.Duration(crawlThreshold)):
			atomic.AddUint64(&crawlSkipped, 1)
			terminate()
			glog.Infof("This page %s is taking too long (> %d), skipping it!", inPage.pageURL.String(), crawlThreshold)
			return
		}
	}

}
