// Copyright 2017 Raghavendra Prabhu.
// Refer to LICENSE for more

// Helper functions used.
package dotler

import (
	//"github.com/golang/glog"
	wire "github.com/ronin13/dotler/wire"

	"log"
	"net/url"
	"strings"
)

func panicCrawl(err error) {
	if err != nil {
		log.Fatalf("dotler has come to a halt due to %s", err)
	}
}

func writeToChan(iPage *wire.Page, inChan chan *wire.Page) {
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
