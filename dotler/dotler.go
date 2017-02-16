// Copyright 2017 Raghavendra Prabhu.
// Refer to LICENSE for more

// Package dotler entrypoint.
package dotler

import (
	"github.com/golang/glog"
	processor "github.com/ronin13/dotler/processor"
	wire "github.com/ronin13/dotler/wire"

	"context"
	"fmt"
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
	// Also add rss|xml|atom
	STATICTYPES = `\.(jpg|gif|bmp|jpeg|png|svg|mp3|mp4|flv|js|css|webm|ogg|flac|wav|ico|atom|rss|xml)$`
)

var (
	// RootURL is the base URL to crawl from.
	RootURL     string
	genImage    bool
	genGraph    bool
	numThreads  int
	graphFormat string
	showProg    string

	// ClientTimeout is the http timeout.
	ClientTimeout    uint
	crawlThreshold   uint
	domain           string
	termChannel      chan struct{}
	reqChan, dotChan chan *wire.Page
	maxFetchFail     uint
	crawlSuccess     uint64
	crawlFail        uint64
	crawlSkipped     uint64
	crawlCancelled   uint64
)

// Signal handler!
// a) SIGTERM/SIGINT - gracefully shuts down the server.
func handleSignal(schannel chan os.Signal) {
	for {
		signl := <-schannel
		switch signl {
		case syscall.SIGTERM, syscall.SIGINT:
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

func setup() {

	if numThreads > 0 {
		runtime.GOMAXPROCS(numThreads)
	} else {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	if showProg != "" {
		glog.Infoln("Turning on gen-image")
		genImage = true
	}
	if genImage {
		genGraph = true
		if _, err := exec.LookPath("dot"); err != nil {
			glog.Infoln("Need dot (from graphviz) in PATH for image generation")
			os.Exit(2)
		}
	}
}

// StartCrawl is the main function with deferred processing in case of return with code.
// Basic functions such as signal processing. setup and main loop.
// Exits during shutdown or when idle time is reached, and then waits.
// The main loop queries the reqChan and dispatches crawl function repeatedly.
func StartCrawl(startURL string) int {
	var err error
	var parsedURL *url.URL
	var endTime int64
	var once sync.Once
	var wg sync.WaitGroup
	var nodeMap wire.NodeMapper

	var printerChan wire.GraphProcessor

	crawlDone := make(chan struct{}, 2)
	reqChan = make(chan *wire.Page, MAXWORKERS)
	termChannel = make(chan struct{}, 2)

	setup()

	parentContext := context.Background()
	noCrawl, terminate := context.WithCancel(parentContext)

	nodeMap, cFunc := wire.NewNodeMapper(parentContext)
	defer cFunc()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	defer wg.Wait()

	parsedURL, err = url.Parse(startURL)
	if err != nil {
		panic(fmt.Sprintf("Failed in parsing root url %s", err))
	}
	reqChan <- &wire.Page{PageURL: parsedURL}

	if genGraph {
		dotChan = make(chan *wire.Page, MAXWORKERS)
		printerChan = processor.NewPrinter()
		printerChan.ProcessLoop(noCrawl, dotChan)
	}
	go handleSignal(sigs)

	waitChan := make(chan struct{})
	go func() {
		for {
			<-waitChan
			wg.Wait()
			if len(reqChan) > 0 {
				continue
			} else {
				crawlDone <- struct{}{}
				return
			}
		}
	}()

	startTime := time.Now().Unix()
	glog.Infof("Starting crawl for %s at %s", startURL, time.Now().String())

	extStatus := make(chan int, 1)

	go func() {
		<-termChannel
		var dotString string

		status := 0

		if crawlDone != nil {
			crawlDone <- struct{}{}
		}
		terminate()
		if genGraph {
			dotString = <-printerChan.Result()
			close(dotChan)
		}
		// This is safe.
		wg.Wait()

		glog.Flush()

		if genGraph {
			err = ioutil.WriteFile("dotler.dot", []byte(dotString), 0644)
			panicCrawl(err)
			glog.Infof("We are done, phew!, persisting graph to dotler.dot\n")
		}

		printStats()

		if genImage {
			status = postProcess(dotString)
		}
		extStatus <- status

	}()

	for {
		select {
		case inPage := <-reqChan:
			if inPage != nil {
				wg.Add(1)
				go Crawl(noCrawl, inPage, reqChan, dotChan, &wg, nodeMap)
				once.Do(func() { close(waitChan) })
			}
		case <-crawlDone:
			endTime = time.Now().Unix()
			glog.Infof("Crawling %s took %d seconds", startURL, endTime-startTime)
			reqChan = nil
			crawlDone = nil
			termChannel <- struct{}{}

		}

		if reqChan == nil && crawlDone == nil {
			break
		}
	}
	return <-extStatus

}
