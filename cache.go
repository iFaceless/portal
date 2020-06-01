package portal

import (
	"context"
	"fmt"

	"github.com/bluele/gcache"
)

type Cacher interface {
	Set(ctx context.Context, key interface{}, value interface{}) error
	Get(ctx context.Context, key interface{}) (interface{}, error)
}

type LRUCache struct {
	c gcache.Cache
}

func NewLRUCache(size int) *LRUCache {
	return &LRUCache{
		c: gcache.New(size).LRU().Build(),
	}
}

var _ Cacher = (*LRUCache)(nil)

func (lru *LRUCache) Set(_ context.Context, key, value interface{}) error {
	return lru.c.Set(key, value)
}

func (lru *LRUCache) Get(_ context.Context, key interface{}) (interface{}, error) {
	return lru.c.Get(key)
}

const (
	cacheKeyTem    = "%s#%s#%s"
	defaultLRUSize = 8192
)

var (
	DefaultCache    = NewLRUCache(defaultLRUSize)
	portalCache     Cacher
	isCacheDisabled bool
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
