package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	Cache // Remove me after realization.

	capacity int
	queue    List
	items    map[Key]*ListItem
}

type cacheValue struct {
	key   Key
	value interface{}
}

func (cache *lruCache) Set(key Key, value interface{}) bool {
	if _, ok := cache.items[key]; ok {
		cache.queue.MoveToFront(cache.items[key])
		cache.items[key].Value = cacheValue{key, value}
		return true
	}
	if cache.queue.Len() == cache.capacity {
		delete(cache.items, cache.queue.Back().Value.(cacheValue).key)
		cache.queue.Remove(cache.queue.Back())
	}
	cache.items[key] = cache.queue.PushFront(cacheValue{key, value})
	return false
}

func (cache *lruCache) Get(key Key) (interface{}, bool) {
	if value, ok := cache.items[key]; ok {
		cache.queue.MoveToFront(cache.items[key])
		return value.Value.(cacheValue).value, true
	}
	return nil, false
}

func (cache *lruCache) Clear() {
	cache.queue = NewList()
	cache.items = make(map[Key]*ListItem, cache.capacity)
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}
