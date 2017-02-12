// Copyright 2017 Raghavendra Prabhu.
// Refer to LICENSE for more

// All the structs we use for crawler.
package wire

import (
	"context"
	"net/url"
)

// StatPage maintains
// - pageTitle: Title of page
// - staticURL: URL of page.
type StatPage struct {
	PageTitle string
	StaticURL *url.URL
}

// PageWithCard is a struct which encapsulates a Page with its cardinality.
// A page can have multiple links to another single page
// card here is cardinality - number of links to that page.
type PageWithCard struct {
	Page *Page
	Card uint
}

// Page maintains:
// - statList: a map of URL to StatPage
// - outLinks: a map of URL to Page
// - pageURL:  URL structure
// - failCount: number of times this page is tried
type Page struct {
	StatList  map[string]StatPage
	OutLinks  map[string]*PageWithCard
	PageURL   *url.URL
	FailCount uint
}

type stringPage struct {
	key   string
	value *Page
	Err   chan error
}

type existsPage struct {
	key   string
	value chan *Page
}

// A NodeMap which is protected by RWMutex.
// Used to ensure we don't process a page twice.
type NodeMap struct {
	//pages map[string]*Page - not exposed.
	addChan   chan *stringPage
	checkChan chan *existsPage
}

type GraphProcessor interface {
	ProcessLoop(context.Context, chan *Page)
	Result() chan string
}

type NodeMapper interface {
	Exists(string) *Page
	Add(string, *Page) error
	RunLoop(context.Context)
}
