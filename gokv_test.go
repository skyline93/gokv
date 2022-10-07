package gokv

import (
	"testing"
	"time"
)

func TestPut(t *testing.T) {
	cache := New(3)
	cache.Put("1", "a")
	cache.Put("2", "b")
	cache.Put("3", "c")

	if v := cache.Get("1"); v != "a" {
		t.FailNow()
	}

	if cache.index.tail.key != "1" {
		t.FailNow()
	}

	if v := cache.Get("2"); v != "b" {
		t.FailNow()
	}

	if cache.index.tail.key != "2" {
		t.FailNow()
	}

	if v := cache.Get("3"); v != "c" {
		t.FailNow()
	}

	cache.Put("4", "d")
	if cache.index.head.key != "2" {
		t.FailNow()
	}
}

func TestPutWithKey(t *testing.T) {
	cache := New(2)

	k := cache.PutWithKey("a")
	if k == "" {
		t.FailNow()
	}
}

func TestTTL(t *testing.T) {
	cache := New(2)

	cache.Put("1", "a", 2)
	time.Sleep(time.Second * 1)

	if v := cache.Get("1"); v == nil {
		t.FailNow()
	}

	time.Sleep(time.Second * 1)
	if v := cache.Get("1"); v != nil {
		t.FailNow()
	}
}
