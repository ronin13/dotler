// Copyright 2017 Raghavendra Prabhu.
// Refer to LICENSE for more

// Helper functions used.
package main

import (
	"github.com/golang/glog"

	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
)

func (node *NodeMap) RunLoop(stopLoop context.Context) {

	pages := make(map[string]*Page)
	log.Println("Starting loop")
	for {
		select {
		case <-stopLoop.Done():
			log.Println("Exiting the loop")
			return
		case addPage := <-node.addChan:
			if val, exists := pages[addPage.key]; exists {
				//glog.Errorf("Key %s already exists, value %s", addPage.key, val)
				addPage.err <- fmt.Errorf("Key exists")
				continue
			}
			pages[addPage.key] = addPage.value
			addPage.err <- nil
		case checkPage := <-node.checkChan:
			if value, exists := pages[checkPage.key]; exists {
				checkPage.value <- value
			} else {
				checkPage.value <- nil
			}
		}
	}

}

func (node *NodeMap) Add(key string, value *Page) error {
	sPage := &stringPage{key, value, make(chan error, 1)}
	node.addChan <- sPage
	return <-sPage.err
}

func (node *NodeMap) Exists(key string) *Page {
	sPage := &existsPage{key: key, value: make(chan *Page, 1)}
	node.checkChan <- sPage
	return <-sPage.value
}

func panicCrawl(err error) {
	if err != nil {
		log.Fatalf("dotler has come to a halt due to %s", err)
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
