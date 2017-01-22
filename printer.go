// Copyright 2017 Raghavendra Prabhu.
// Refer to LICENSE for more

// Responsible for generating graphviz graph
// concurrently when crawling is being done.
// Persisted only towards end.
// Gets the input from another channel to
// which the crawler writes Pages.
// Renders both Page nodes and Static nodes.
package main

import (
	"github.com/golang/glog"

	"fmt"
	"strconv"
	"sync"
)

// Adds a Page Node.
func addNoteFromAttr(iPage *Page) string {
	quotedURL := fmt.Sprintf("%q", iPage.pageURL.String())
	crawlGraph.AddNode("dotler", quotedURL, map[string]string{
		"URL": quotedURL,
	})
	return quotedURL
}

// Adds a Static Node.
func staticNodes(iPage StatPage) string {
	quotedURL := fmt.Sprintf("%q", iPage.staticURL.String())
	quotedTitle := fmt.Sprintf("%q", iPage.pageTitle)
	crawlGraph.AddNode("dotler", quotedURL, map[string]string{
		"URL":     quotedTitle,
		"tooltip": quotedURL,
		"style":   "dashed",
	})
	return quotedURL
}

// Runs till the inward channel closes, during shutdown.
// Runs in parallel with crawler, silently weaving the graph
// in background.
func dotPrinter(inChan chan *Page, waiter *sync.WaitGroup) {
	defer waiter.Done()
	glog.Infoln("Starting the dot printer!")
	var addedURL, presURL string
	for iPage := range inChan {
		presURL = addNoteFromAttr(iPage)
		for _, oPage := range iPage.outLinks {
			addedURL = addNoteFromAttr(oPage.page)
			crawlGraph.AddEdge(presURL, addedURL, true, map[string]string{
				"label": strconv.Itoa(int(oPage.card)),
			})
		}

		for _, sPage := range iPage.statList {
			addedURL = staticNodes(sPage)
			crawlGraph.AddEdge(presURL, addedURL, true, map[string]string{
				"style": "dashed",
				"color": "blue",
			})
		}
	}
}
