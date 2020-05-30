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

var DefaultCache Cache

func SetCache(c Cache) {
	DefaultCache = c
}

const cacheKeyTem = "%s#%s#%s"

// genCacheKey generate cache key
// rules: ReceiverName#MethodName#CacheID
// eg. meth:GetName UserSchema#GetName#123,
// attr:Name UserModel#Name#123
func genCacheKey(ctx context.Context, receiver interface{}, cacheObj interface{}, methodName string) *string {
	ret, err := invokeMethodOfAnyType(ctx, cacheObj, "CacheID", nil)
	if err != nil {
		return nil
	}
	cacheID, ok := ret.(string)
	if !ok {
		return nil
	}

	ck := fmt.Sprintf(cacheKeyTem, structName(receiver), methodName, cacheID)
	return &ck
}
