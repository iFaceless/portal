package portal

import (
	"context"
	"fmt"
	"sync"
)

type Cacher interface {
	Set(ctx context.Context, key interface{}, value interface{}) error
	Get(ctx context.Context, key interface{}) (interface{}, error)
}

type ErrNil struct{}

func (e *ErrNil) Error() string {
	return "portal cache key not found."
}

type MapCache struct {
	c map[interface{}]interface{}
}

func NewMapCache() *MapCache {
	return &MapCache{
		c: make(map[interface{}]interface{}),
	}
}

var _ Cacher = (*MapCache)(nil)

func (m *MapCache) Set(_ context.Context, key, value interface{}) error {
	m.c[key] = value
	return nil
}

func (m *MapCache) Get(_ context.Context, key interface{}) (interface{}, error) {
	if v, ok := m.c[key]; ok {
		return v, nil
	}
	return nil, &ErrNil{}
}

const (
	cacheKeyTem = "%s#%s#%s"
)

var (
	DefaultCache    = NewMapCache()
	portalCache     Cacher
	isCacheDisabled = false
)

// SetCache enable cache strategy
func SetCache(c Cacher) {
	if c == nil {
		isCacheDisabled = true
		return
	}
	isCacheDisabled = false
	portalCache = c
}

// genCacheKey generate cache key
// rules: ReceiverName#MethodName#cacheID
// eg. meth:GetName UserSchema#GetName#0xc000498150,
// attr:Name UserModel#Name#0xc000498150
func genCacheKey(ctx context.Context, receiver interface{}, cacheObj interface{}, methodName string) *string {
	cacheID := defaultCacheID(cacheObj)

	ck := fmt.Sprintf(cacheKeyTem, structName(receiver), methodName, cacheID)
	return &ck
}

// defaultCacheID is the addr of src struct
func defaultCacheID(cacheObj interface{}) string {
	return fmt.Sprintf("%p", cacheObj)
}

type cachable interface {
	PortalDisableCache() bool
}

type cacheGroup struct {
	cache Cacher
	mu    sync.Mutex
	m     map[interface{}]*call
}

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

func newCacheGroup(cache Cacher) *cacheGroup {
	return &cacheGroup{
		cache: cache,
		m:     make(map[interface{}]*call),
	}
}

func (cg *cacheGroup) Valid() bool {
	return portalCache != nil && cg.cache != nil
}
