// Copyright 2017 Raghavendra Prabhu.
// Refer to LICENSE for more

// Package processor has logic
// responsible for generating graphviz graph
// concurrently when crawling is being done.
// Persisted only towards end.
// Gets the input from another channel to
// which the crawler writes Pages.
// Renders both Page nodes and Static nodes.
package processor

import (
	"github.com/awalterschulze/gographviz"
	"github.com/golang/glog"
	wire "github.com/ronin13/dotler/wire"

	"context"
	"fmt"
	"strconv"
)

// Adds a Page Node.
func (dot *dotPrinter) addNoteFromAttr(iPage *wire.Page) string {
	quotedURL := fmt.Sprintf("%q", iPage.PageURL.String())
	dot.cgraph.AddNode("dotler", quotedURL, map[string]string{
		"URL": quotedURL,
	})
	return quotedURL
}

// Adds a Static Node.
func (dot *dotPrinter) staticNodes(iPage wire.StatPage) string {
	quotedURL := fmt.Sprintf("%q", iPage.StaticURL.String())
	quotedTitle := fmt.Sprintf("%q", iPage.PageTitle)
	dot.cgraph.AddNode("dotler", quotedURL, map[string]string{
		"URL":     quotedTitle,
		"tooltip": quotedURL,
		"style":   "dashed",
	})
	return quotedURL
}

type dotPrinter struct {
	cgraph *gographviz.Escape
	result chan string
}

// NewPrinter returns a new instance implementing the GraphProcessor interface.
func NewPrinter() wire.GraphProcessor {
	dPrinter := new(dotPrinter)
	dPrinter.cgraph = gographviz.NewEscape()
	dPrinter.result = make(chan string, 1)
	dPrinter.cgraph.SetName("dotler")
	dPrinter.cgraph.SetDir(true)
	dPrinter.cgraph.SetStrict(true)
	return dPrinter
}

func (dot *dotPrinter) Result() chan string {
	return dot.result
}

// Runs till the inward channel closes, during shutdown.
// Runs in parallel with crawler, silently weaving the graph
// in background.
func (dot *dotPrinter) ProcessLoop(noPrint context.Context, inChan chan *wire.Page) {
	glog.Infoln("Starting the dot printer!")
	var addedURL, presURL string
	go func() {
		for {
			select {
			case iPage := <-inChan:
				if iPage != nil {
					presURL = dot.addNoteFromAttr(iPage)
					for _, oPage := range iPage.OutLinks {
						addedURL = dot.addNoteFromAttr(oPage.Page)
						dot.cgraph.AddEdge(presURL, addedURL, true, map[string]string{
							"label": strconv.Itoa(int(oPage.Card)),
						})
					}

					for _, sPage := range iPage.StatList {
						addedURL = dot.staticNodes(sPage)
						dot.cgraph.AddEdge(presURL, addedURL, true, map[string]string{
							"style": "dashed",
							"color": "blue",
						})
					}
				}

			case <-noPrint.Done():
				dot.result <- dot.cgraph.String()
				glog.Infoln("Halting the dot printer!")
				return
			}
		}
	}()
}
