// Copyright 2017 Raghavendra Prabhu.
// Refer to LICENSE for more

// Core dotler file. Entrypoint.
package main

import (
	"github.com/awalterschulze/gographviz"
	"github.com/golang/glog"

	"context"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	// MAXWORKERS is a internal constant for channel capacity, should be enough mostly.
	MAXWORKERS = 100
	// STATICTYPES defines extensions we consider to be 'static' when processing a page.
	STATICTYPES = `\.(jpg|gif|bmp|jpeg|png|svg|mp3|mp4|flv|js|css|webm|ogg|flac|wav|ico)$`
)

var (
	wg                                      sync.WaitGroup // To make sure we don't leave any handlers behind!
	rootURL                                 string
	genImage                                bool
	genGraph                                bool
	numThreads                              int
	graphFormat                             string
	showProg                                string
	clientTimeout, idleTime, crawlThreshold uint
	domain                                  string
	termChannel                             = make(chan struct{}, 2)
	maxFetchFail                            uint
	crawlSuccess                            uint64
	crawlFail                               uint64
	crawlSkipped                            uint64
	crawlCancelled                          uint64
)

var crawlGraph = gographviz.NewEscape()

// Signal handler!
// a) SIGTERM/SIGINT - gracefully shuts down the server.
func handleSignal(schannel chan os.Signal) {
	for {
		signl := <-schannel
		switch signl {
		case syscall.SIGTERM:
		case syscall.SIGINT:
			termChannel <- struct{}{}
			glog.Infoln("Time to leave and cleanup!")
			return
		}
	}
}

func printStats() {
	glog.Infoln("Crawl statistics")
	glog.Infoln("===========================================")
	var statsFinal uint64

	statsFinal = atomic.LoadUint64(&crawlSuccess)
	glog.Infof("Successfully crawled URLs %d", statsFinal)

	statsFinal = atomic.LoadUint64(&crawlSkipped)
	glog.Infof("Skipped URLs %d", statsFinal)

	statsFinal = atomic.LoadUint64(&crawlFail)
	glog.Infof("Failed URLs %d", statsFinal)

	statsFinal = atomic.LoadUint64(&crawlCancelled)
	glog.Infof("Cancelled URLs %d", statsFinal)

	glog.Infoln("===========================================")
}

// Main function with deferred processing in case of return with code.
// Basic functions such as signal processing. setup and main loop.
// Exits during shutdown or when idle time is reached, and then waits.
// The main loop queries the reqChan and dispatches crawl function repeatedly.
func startCrawl(startURL string) int {
	var err error
	var parsedURL *url.URL
	var endTime int64

	if numThreads > 0 {
		runtime.GOMAXPROCS(numThreads)
	} else {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	crawlGraph.SetName("dotler")
	crawlGraph.SetDir(true)
	crawlGraph.SetStrict(true)

	if showProg != "" {
		glog.Infoln("Turning on gen-image")
		genImage = true
	}
	if genImage {
		genGraph = true
		if _, err = exec.LookPath("dot"); err != nil {
			glog.Infoln("Need dot (from graphviz) in PATH for image generation")
			return 2
		}
	}

	parentContext := context.Background()
	noCrawl, terminate := context.WithCancel(parentContext)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	defer wg.Wait()

	reqChan := make(chan *Page, MAXWORKERS)
	dotChan := make(chan *Page, MAXWORKERS)

	parsedURL, err = url.Parse(startURL)
	if err != nil {
		glog.Fatalf("Failed in parsing root url %s", err)
	}
	reqChan <- &Page{pageURL: parsedURL}
	if genGraph {
		wg.Add(1)
		go dotPrinter(dotChan, &wg)
	}
	go handleSignal(sigs)

	startTime := time.Now().Unix()
	glog.Infof("Starting crawl for %s at %s", startURL, time.Now().String())

	for {
		select {
		case inPage := <-reqChan:
			wg.Add(1)
			go crawl(noCrawl, inPage, reqChan, dotChan, &wg)
		case <-time.After(time.Second * time.Duration(idleTime)):
			glog.Infoln("Idle timeout reached, bye!")
			endTime = time.Now().Unix() - int64(idleTime)
			termChannel <- struct{}{}
		case <-termChannel:
			terminate()
			close(dotChan)
			close(reqChan)
			wg.Wait()

			if endTime == 0 {
				endTime = time.Now().Unix()
			}
			glog.Infof("Crawling %s took %d seconds", startURL, endTime-startTime)
			glog.Flush()
			dotString := crawlGraph.String()

			err = ioutil.WriteFile("dotler.dot", []byte(dotString), 0644)
			panicCrawl(err)
			glog.Infof("We are done, phew!, persisting graph to dotler.dot\n")

			printStats()

			if genImage {
				postProcess()
			}

			return 0
		}
	}
	// Avoiding 'unreachable code' from go vet.
	// Was added just to be a failsafe.
	// return 0
}

func main() {
	parseFlags()
	os.Exit(startCrawl(rootURL))
}
