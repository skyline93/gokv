package gokv_test

import (
	"log"
	"sync"
	"testing"
	"time"

	"github.com/skyline93/gokv"
)

func TestBanch(t *testing.T) {
	cache := gokv.New(10)

	for {
		t.Log("concurrent put")
		concurrent_put(0, cache)
		time.Sleep(time.Second * 5)
	}
}

func concurrent_put(n int, cache *gokv.Cache) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)

		log.Print("gogogo")
		go func(key int) {
			defer wg.Done()

			cache.PutWithKey(key)
		}(i)
	}

	wg.Wait()
}
