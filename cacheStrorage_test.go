package cache

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestCache_Add(t *testing.T) {
	cache, _ := InitializeCache(3)
	expected := 3

	elems := rand.Intn(10)
	i := 0
	for i < elems {
		cache.Add(i, i)
		i++
	}
	result := cache.Len()
	if result != expected {
		fmt.Printf("Expected %d, but got %d", expected, result)
	}
}

func TestCache_OverwriteOldest(t *testing.T) {
	cache, _ := InitializeCache(3)
	expected := 15

	testValues := []int{1, 3, 5, 5, 5}
	for i, v := range testValues {
		cache.Add(i, v)
	}
	result := 0
	for k, _ := range cache.cacheMap {
		val, _ := cache.Get(k)
		result += val.(int)
	}

	if result != expected {
		fmt.Printf("Expected %d, but got %d", expected, result)
	}
}

func TestCache_AddWithTTL(t *testing.T) {
	cache, _ := InitializeCache(3)
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

func TestCache_GetOutdated(t *testing.T) {
	cache, _ := InitializeCache(3)
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

func TestCache_GetNotExisting(t *testing.T) {
	cache, _ := InitializeCache(3)

	elems := rand.Intn(10)
	i := 0
	for i < elems {
		cache.Add(i, i)
		i++
	}
	_, ok := cache.Get(100)
	if ok {
		fmt.Printf("Expected %t, but got %t", !ok, ok)
	}
}

func TestCache_ConcurrentAdd(t *testing.T) {
	cache, _ := InitializeCache(50)
	expected := 50

	for i := range 500 {
		go cache.Add(i, i)
		i++
	}
	time.Sleep(2 * time.Second)
	result := cache.Len()
	if result != expected {
		fmt.Printf("Expected %d, but got %d", expected, result)
	}
}

func TestCache_Clear(t *testing.T) {
	cache, _ := InitializeCache(50)
	expected := 0

	for i := range 500 {
		go cache.Add(i, i)
		i++
	}
	time.Sleep(2 * time.Second)
	cache.Clear()
	result := cache.Len()
	if result != expected {
		fmt.Printf("Expected %d, but got %d", expected, result)
	}
}

func TestCache_NegativeCapacityCreation(t *testing.T) {
	_, result := InitializeCache(-5)

	if result == nil {
		fmt.Printf("Expected error, but got %d", result)
	}
}

func TestCache_GetCapacity(t *testing.T) {
	capacity := 10
	cache, _ := InitializeCache(capacity)
	result := cache.Cap()
	if result != capacity {
		fmt.Printf("Expected %d, but got %d", capacity, result)
	}
}
