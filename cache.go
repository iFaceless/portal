package portal

import (
	"context"
	"fmt"
	"reflect"

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

const (
	cacheKeyTem     = "%s#%s#%s"
	cacheIDMethName = "CacheID"
)

const DefaultLRUSize = 65536

var (
	DefaultCache    = NewLRUCache(DefaultLRUSize)
	PortalCache     Cache
	IsCacheDisabled bool
)

// SetCache enable cache strategy
func SetCache(c Cache) {
	IsCacheDisabled = false
	PortalCache = c
}

// GlobalDisableCache disable cache strategy globally
func GlobalDisableCache() {
	IsCacheDisabled = true
}

// genCacheKey generate cache key
// rules: ReceiverName#MethodName#cacheID
// eg. meth:GetName UserSchema#GetName#0xc000498150,
// attr:Name UserModel#Name#0xc000498150
func genCacheKey(ctx context.Context, receiver interface{}, cacheObj interface{}, methodName string) *string {
	var cacheID string
	var ok bool

	// if src's CacheID is not defined, use default cacheID
	ret, err := invokeMethodOfAnyType(ctx, cacheObj, cacheIDMethName)
	if err != nil {
		cacheID = defaultCacheID(cacheObj)
	} else {
		cacheID, ok = ret.(string)
		if !ok {
			logger.Warnf("'%s.%s' return value must be a string", reflect.TypeOf(cacheObj), cacheIDMethName)
			return nil
		}
	}

	ck := fmt.Sprintf(cacheKeyTem, structName(receiver), methodName, cacheID)
	return &ck
}

// defaultCacheID is the addr of src struct
func defaultCacheID(cacheObj interface{}) string {
	return fmt.Sprintf("%p", cacheObj)
}
