package wire

import (
	"context"
	"fmt"
	gmap "github.com/ronin13/goimutmap"
	"strings"
)

// NewNodeMapper returns a new instance of implementing NodeMapper interface.
func NewNodeMapper(ctx context.Context) (NodeMapper, context.CancelFunc) {

	mapper, cFunc := gmap.NewcontextMapper(ctx)
	return &NodeMap{mapper}, cFunc
}

// Needed for http/https sites to create smaller graph.
func httpStrip(input string) string {
	return strings.Split(input, "//")[1]
}

// Add method allows one to add new keys.
// Returns error.
func (node *NodeMap) Add(key string, value *Page) error {
	skey := httpStrip(key)
	already := node.ContextMapper.Add(skey, value)
	if already != nil {
		return fmt.Errorf("Key %s already existed", key)
	}
	return nil
}

// Exists method allows to check and return the key.
func (node *NodeMap) Exists(key string) *Page {
	skey := httpStrip(key)
	if page, exists := node.ContextMapper.Exists(skey); exists {
		retPage, ok := page.(*Page)
		if ok {
			return retPage
		} else {
			panic("Stored value is not a *Page")
		}
	}
	return nil
}
