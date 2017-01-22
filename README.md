

# Dotler

## Introduction
Dotler is a web crawler and graph generator. The graph hereforth refers
to a graphviz graph, an internal graph also exists but it is internal.
The crawling and generation are done concurrently. Much like a pipeline,
one goroutine crawls and dispatches it to another which generates the
graph.

It also crawls only within its domain, and 
maintains a map of any static assets from a page.

Algorithm is as follows:
(Skipping implementation details)

### Main (dotler.go)

1) Get the root url.
2) Send to request Channel.
3) Iterate over request Channel and dispatch urls from it to crawl goroutine until
    i) idle timeout - needed since crawling cannot be considered to halt deterministically. Hence, if there are no new urls in request Channel after idle timeout, we leave.
    ii) SIGINT / SIGTERM are sent.
4) During shutdown, we persist the graphviz to a .dot file, generate SVG if required and show the image as well.

### crawl goroutine (crawler.go)

1) Update a map - nodeMap - (with RW lock) to ensure we don't duplicate process page.
2) Get all links on the page, goquery is used for this purpose.
    i) Unless the link is already processed, send it to request Channel if it is part of same domain.
    ii) Update this pages map of static assets and links originating from this page.
    iii) Follow HTML tags are considered:
        a) a, img, script, link, source
            - for attributes: href, src

3) After a page is successfully processed, send it over to dotprinter goroutine over a separate channel.

### dotprinter goroutine (printer.go)

1) Get new page from channel.
2) Generate new graphviz node and outlinks and links to static assets, creating new nodes if necessary.

### General implementation details.

1) Channels are used for communication between crawling goroutines (which are many and capped at MAXWORKERS at a time), dotprinter (which is one) and main function.
    - Also used for signalling shutdown and completion of work and with contexts.
2) Contexts with Cancel are used to cancel, timeout channels and timeouts are also used wherever necessary.
3) For the nodeMap, the contention is low since it is only added to and never updated again. Hence, a RW lock is used.
4) Each crawling goroutine itself uses a timeout - crawlThreshold - if a page is taking too long.
5) Statistics are printed during shutdown.
6) A signal handler (for shutdown) sends signal on termChannel which does cleanup, persist graph among other things.
7) Each page is tried maxFetchFail (default 2) times in case of failure.
8) Wait groups are used to wait on goroutines

### Core data structure

```
type Page struct {
	statList  map[string]StatPage
	outLinks  map[string]*PageWithCard
	pageURL   *url.URL
	failCount uint
}
```


This is the structure passed around in channels for request and graph processing.

1) statList is a map of static assets from this page.
2) outLinks is a map of Page links from this page. PageWithCard is a struct with cardinality of links to it.
3) pageURL is the url of the page.
4) failCount is the number of times this page crawling can fail.

## Races
Every attempt has been made to eliminate race conditions.

```
make race 
```

runs with race detector.

## Example

### Generating image.
```

./dotler  -idle 15 -gen-image
I0122 19:09:34.653313   24825 crawl.go:150] Processing page http://www.wnohang.net/
I0122 19:09:34.653338   24825 printer.go:44] Starting the dot printer!
I0122 19:09:34.730647   24825 crawl.go:150] Processing page http://www.wnohang.net/pages/conferences
I0122 19:09:34.730818   24825 crawl.go:150] Processing page http://www.wnohang.net/pages/about
I0122 19:09:34.731176   24825 crawl.go:150] Processing page http://www.wnohang.net/pages/consult
I0122 19:09:34.731468   24825 crawl.go:150] Processing page http://www.wnohang.net/pages/code
I0122 19:09:34.731532   24825 crawl.go:172] Successfully crawled http://www.wnohang.net/
I0122 19:09:34.731699   24825 crawl.go:150] Processing page http://www.wnohang.net/pages/contact
I0122 19:09:34.839630   24825 crawl.go:172] Successfully crawled http://www.wnohang.net/pages/about
I0122 19:09:34.843923   24825 crawl.go:172] Successfully crawled http://www.wnohang.net/pages/conferences
I0122 19:09:34.855408   24825 crawl.go:172] Successfully crawled http://www.wnohang.net/pages/code
I0122 19:09:34.873383   24825 crawl.go:150] Processing page http://www.wnohang.net/pages/Yelp
I0122 19:09:34.874024   24825 crawl.go:172] Successfully crawled http://www.wnohang.net/pages/contact
I0122 19:09:34.891400   24825 crawl.go:172] Successfully crawled http://www.wnohang.net/pages/Yelp
I0122 19:09:34.930642   24825 crawl.go:172] Successfully crawled http://www.wnohang.net/pages/consult
I0122 19:09:49.873476   24825 dotler.go:145] Idle timeout reached, bye!
I0122 19:09:49.874124   24825 dotler.go:157] We are done, phew!, persisting graph to dotler.dot
I0122 19:09:49.874184   24825 dotler.go:68] Crawl statistics
I0122 19:09:49.874194   24825 dotler.go:69] ===========================================
I0122 19:09:49.874200   24825 dotler.go:73] Successfully crawled URLs 7
I0122 19:09:49.874206   24825 dotler.go:76] Skipped URLs 0
I0122 19:09:49.874210   24825 dotler.go:79] Failed URLs 0
I0122 19:09:49.874214   24825 dotler.go:82] Cancelled URLs 0
I0122 19:09:49.874217   24825 dotler.go:84] ===========================================
I0122 19:09:49.874222   24825 postprocess.go:20] Generating svg from dot file

```

### Showing image

```
./dotler  -idle 15 -max-crawl 30 -gen-image -image-prog='chromium'
```

### With url

```
/dotler -url 'https://gobyexample.com/'  -idle 10
```

#### Note
The SVGs generated have clickable graph nodes.


## Testing

```
make test
```

## Benchmarking

```
make bench
```

## Build 

```
make dotler
```

## Sample SVG
A sample svg generated is present as sameple.svg. The numbers of edges is link cardinality (number of links from one page to another).
Blue dotted line is to link to static assets (dashed)
Nodes themselves are oval with URL inside them. They can be clicked upon to lead to the actual page.


## How fast is this
- It runs quite fast. :)

- In the logs, look for a line like:
 " Crawling http://blog.golang.org took 8 seconds "

- Typically for small sites like http://wnohang.net it will show 0 or 1.

- Note that this is only the crawling time (minus idle time) excluding any post processing time to persist the graph, generate SVG or show the image.

- Benchmarks (from make bench) include the entire time.

## Credits
-  github.com/PuerkitoBio/goquery 
-  github.com/PuerkitoBio/purell
-  github.com/golang/glog
-  github.com/Masterminds/glide
-  github.com/awalterschulze/gographviz
