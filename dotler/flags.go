// Copyright 2017 Raghavendra Prabhu.
// Refer to LICENSE for more

// Package dotler 's flag parsing.
package dotler

import (
	"flag"
)

// ParseFlags provides parsing of all the flags.
// Usage of ./dotler:
//  -alsologtostderr
//        log to standard error as well as files
//  -format string
//        Format of generated image (default "svg")
//  -gen-graph
//        Generate a graphviz graph (default true)
//  -gen-image
//        Generate an image of sitemap (implies gen-graph)
//  -display-prog string
//        If not empty, program to show the image (implies gen-graph and gen-image), chromium etc.
//  -log_backtrace_at value
//        when logging hits line file:N, emit a stack trace
//  -log_dir string
//        If non-empty, write log files in this directory
//  -logtostderr
//        log to standard error instead of files
//  -max-crawl uint
//        Timeout in seconds to scrape and process a single page (default 10)
//  -max-threads int
//        Number of goroutines, defaults to NumCPU
//  -retry uint
//        Number of failures to tolerate if http fetch fails (default 2)
//  -stderrthreshold value
//        logs at or above this threshold go to stderr
//  -timeout uint
//        Timeout in seconds (default 60)
//  -url string
//        Url to crawl (default "http://www.wnohang.net/")
// -v value
//        log level for V logs
//  -vmodule value
//      comma-separated list of pattern=N settings for file-filtered logging
func ParseFlags() {
	flag.StringVar(&RootURL, "url", "http://www.wnohang.net/", "Url to crawl")
	flag.UintVar(&ClientTimeout, "timeout", 60, "Timeout in seconds")
	flag.UintVar(&maxFetchFail, "retry", 2, "Number of failures to tolerate if http fetch fails")
	flag.UintVar(&crawlThreshold, "max-crawl", 10, "Timeout in seconds to scrape and process a single page")
	flag.IntVar(&numThreads, "max-threads", 0, "Number of goroutines, defaults to NumCPU")

	flag.BoolVar(&genImage, "gen-image", false, "Generate an image of sitemap (implies gen-graph), default false")
	flag.BoolVar(&genGraph, "gen-graph", true, "Generate a graphviz graph")
	flag.StringVar(&showProg, "display-prog", "", "If not empty, program to display the image (implies gen-graph and gen-image), chromium etc.")
	flag.StringVar(&graphFormat, "format", "svg", "Format of generated image")

	flag.Lookup("alsologtostderr").Value.Set("true")
	flag.Parse()
}
