package portal

import (
	"context"
	"fmt"

	"github.com/bluele/gcache"
)

type Cache interface {
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

var _ Cache = (*LRUCache)(nil)

func (lru *LRUCache) Set(_ context.Context, key, value interface{}) error {
	return lru.c.Set(key, value)
}

func (lru *LRUCache) Get(_ context.Context, key interface{}) (interface{}, error) {
	return lru.c.Get(key)
}

const DefaultLRUSize = 65536

var DefaultCache = NewLRUCache(DefaultLRUSize)
var PortalCache Cache
var IsCacheDisabled bool

func SetCache(c Cache) {
	IsCacheDisabled = false
	PortalCache = c
}

func DisableCache() {
	IsCacheDisabled = true
}

const cacheKeyTem = "%s#%s#%s"

// genCacheKey generate cache key
// rules: ReceiverName#MethodName#cacheObj_PointerAddr
// eg. meth:GetName UserSchema#GetName#0xc000498150,
// attr:Name UserModel#Name#0xc000498150
func genCacheKey(ctx context.Context, receiver interface{}, cacheObj interface{}, methodName string) *string {
	cacheID := fmt.Sprintf("%p", cacheObj)
	ck := fmt.Sprintf(cacheKeyTem, structName(receiver), methodName, cacheID)
	return &ck
}
