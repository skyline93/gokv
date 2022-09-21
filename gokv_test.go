package gokv_test

import (
	"testing"

	"github.com/skyline93/gokv"
)

func TestPutAndGet(t *testing.T) {
	kv := gokv.New()

	tests := []struct {
		arg1     string
		arg2     interface{}
		expected interface{}
	}{
		{"key1", 123, 123},
		{"key1", "123", "123"},
		{"key2", struct{ v1 string }{"v1"}, struct{ v1 string }{"v1"}},
	}

	for _, i := range tests {
		if ok := kv.Put(i.arg1, i.arg2); !ok {
			t.Fatal("put error")
		}

		output := kv.Get(i.arg1)
		if output != i.expected {
			t.Errorf("Output %q not equal to expected %q", output, i.expected)
		}
	}
}

func TestPutWithUuidAndGet(t *testing.T) {
	kv := gokv.New()

	tests := []struct {
		arg      interface{}
		expected interface{}
	}{
		{123, 123},
		{"123", "123"},
		{struct{ v1 string }{"v1"}, struct{ v1 string }{"v1"}},
	}

	for _, i := range tests {
		k, ok := kv.PutWithUuid(i.arg)
		if !ok {
			t.Fatal("put error")
		}

		output := kv.Get(k)
		if output != i.expected {
			t.Errorf("Output %q not equal to expected %q", output, i.expected)
		}
	}
}
