// Copyright 2017 Raghavendra Prabhu.
// Refer to LICENSE for more

// Helper functions used.
package main

import (
	"github.com/golang/glog"

	"net/url"
	"strings"
)

var nodeMap = &NodeMap{pages: make(map[string]*Page)}

// Update a map with write locking of a RW lock.
// Contention here should be low.
// Hence, the RW Lock.
func updNodeMap(key string, iPage *Page) {
	nodeMap.Lock()
	nodeMap.pages[key] = iPage
	nodeMap.Unlock()
}

func panicCrawl(err error) {
	if err != nil {
		glog.Fatalf("dotler has come to a halt due to %s", err)
	}
}

func writeToChan(iPage *Page, inChan chan *Page) {
	// To prevent panic from closed channel during shutdown
	// Yes, there are other safeguards, but real world is not perfect :)
	defer func() { recover() }()

	inChan <- iPage
}

// For static assets, get the title as last component
// Example: http://abcd.com/qq.js returns qq.js
func getStatTitle(url *url.URL) string {
	comps := strings.Split(url.Path, "/")
	lastone := comps[len(comps)-1]
	return lastone
}
