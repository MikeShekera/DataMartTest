package cache

import (
	"container/list"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

const outdatedDeleteFrequency = 2 * time.Second

type Cache struct {
	capacity int
	order    *list.List
	cacheMap map[interface{}]*valueWrapper
	mu       *sync.RWMutex
}

type valueWrapper struct {
	val          interface{}
	endTimepoint time.Time
	elemPointer  *list.Element
}

func main() {
	cache, err := InitializeCache(3)
	if err != nil {
		log.Fatal(err)
	}
	expectedLen := 0

	elems := rand.Intn(10)
	i := 0
	for i < elems {
		cache.AddWithTTL(i, i, 2)
		i++
	}
	time.Sleep(3 * time.Second)
	result := cache.Len()
	if result != expectedLen {
		fmt.Printf("Expected %d, but got %d", expectedLen, result)
	}
}

func InitializeCache(size int) (Cache, error) {
	if size < 0 {
		return Cache{}, fmt.Errorf("Capacity must be more than zero")
	}
	c := Cache{
		capacity: size,
		order:    list.New(),
		cacheMap: make(map[interface{}]*valueWrapper, size),
		mu:       &sync.RWMutex{},
	}

	go func() {
		for {
			<-time.After(outdatedDeleteFrequency)
			c.cleanupOutdated()
		}
	}()

	return c, nil
}

func (c *Cache) Cap() int {
	return c.capacity
}

func (c *Cache) Len() int {
	c.cleanupOutdated()
	return len(c.cacheMap)
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	clear(c.cacheMap)
	c.order.Init()
}

func (c *Cache) Add(key, value any) {
	c.innerAdd(key, value, time.Time{})
}

func (c *Cache) AddWithTTL(key, value any, ttl time.Duration) {
	c.innerAdd(key, value, time.Now().Add(ttl))
}

func (c *Cache) innerAdd(key, value any, endTime time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if v, ok := c.cacheMap[key]; !ok {
		if len(c.cacheMap) < c.capacity {
			elem := c.order.PushFront(key)
			c.cacheMap[key] = &valueWrapper{val: value, elemPointer: elem, endTimepoint: endTime}
		} else {
			oldest := c.order.Back()
			c.order.Remove(oldest)
			newElem := c.order.PushFront(key)
			delete(c.cacheMap, oldest.Value)
			c.cacheMap[key] = &valueWrapper{val: value, elemPointer: newElem, endTimepoint: endTime}
		}
	} else {
		c.order.MoveToFront(v.elemPointer)
		v.endTimepoint = endTime
	}
}

func (c *Cache) Get(key any) (value any, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if val, ok := c.cacheMap[key]; ok {
		if !val.endTimepoint.IsZero() {
			if val.endTimepoint.Before(time.Now()) {
				c.Remove(key)
				return nil, false
			}
		}
		c.order.MoveToFront(val.elemPointer)
		return val.val, true
	}
	return nil, false
}

func (c *Cache) Remove(key any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if v, ok := c.cacheMap[key]; ok {
		c.order.Remove(v.elemPointer)
		delete(c.cacheMap, key)
	} else {
		fmt.Println("no such elem in cache")
	}
}

func (c *Cache) cleanupOutdated() {
	for k, v := range c.cacheMap {
		if !v.endTimepoint.IsZero() && v.endTimepoint.Before(time.Now()) {
			c.Remove(k)
		}
	}
}

type ICache interface {
	Cap() int
	Len() int
	Clear()
	Add(key, value any)
	AddWithTTL(key, value any, ttl time.Duration)
	Get(key any) (value any, ok bool)
	Remove(key any)
}
