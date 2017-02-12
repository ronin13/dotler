package wire

import (
	"context"
	"fmt"
	"strings"
)

// NewNodeMapper returns a new instance of implementing NodeMapper interface.
func NewNodeMapper(numThreads int) NodeMapper {
	return &NodeMap{make(chan *stringPage, numThreads), make(chan *existsPage, numThreads)}
}

// RunLoop is the NodeMapper's map requests processing loop.
func (node *NodeMap) RunLoop(stopLoop context.Context) {

	pages := make(map[string]*Page)
	for {
		select {
		case <-stopLoop.Done():
			return
		case addPage := <-node.addChan:
			if _, exists := pages[addPage.key]; exists {
				//glog.Errorf("key %s already exists, value %s", addPage.key, val)
				addPage.Err <- fmt.Errorf("key exists")
				continue
			}
			pages[addPage.key] = addPage.value
			addPage.Err <- nil
		case checkPage := <-node.checkChan:
			if value, exists := pages[checkPage.key]; exists {
				checkPage.value <- value
			} else {
				checkPage.value <- nil
			}
		}
	}

}

// Needed for http/https sites to create smaller graph.
func httpStrip(input string) string {
	return strings.Split(input, "//")[1]
}

// Add method allows one to add new keys.
// Returns error.
func (node *NodeMap) Add(key string, value *Page) error {
	skey := httpStrip(key)
	sPage := &stringPage{skey, value, make(chan error, 1)}
	node.addChan <- sPage
	return <-sPage.Err
}

// Exists method allows to check and return the key.
func (node *NodeMap) Exists(key string) *Page {
	skey := httpStrip(key)
	sPage := &existsPage{key: skey, value: make(chan *Page, 1)}
	node.checkChan <- sPage
	return <-sPage.value
}
