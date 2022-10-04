package gokv

import (
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
)

type Value struct {
	V          interface{}
	Ttl        int
	createTime time.Time
}

func NewValue(v interface{}) *Value {
	return &Value{
		V:          v,
		Ttl:        60 * 60 * 2,
		createTime: time.Now(),
	}
}

func (v *Value) expiration() time.Time {
	return v.createTime.Add(time.Second * time.Duration(v.Ttl))
}

func (v *Value) IsExpired() bool {
	return time.Now().After(v.expiration())
}

type KV struct {
	m map[string]*Value
	sync.RWMutex
}

func New() *KV {
	return &KV{
		m: make(map[string]*Value),
	}
}

func (kv *KV) Get(k string) interface{} {
	kv.RLock()
	v, ok := kv.m[k]
	kv.RUnlock()

	if !ok {
		return nil
	}

	if v.IsExpired() {
		kv.Lock()
		delete(kv.m, k)
		kv.Unlock()

		return nil
	}

	return v.V
}

func (kv *KV) Put(k string, v interface{}, args ...interface{}) bool {
	value := NewValue(v)
	for i, v := range args {
		switch i {
		case 0:
			ttl, ok := v.(int)
			if !ok {
				return false
			}
			value.Ttl = ttl
		default:
			return false
		}
	}

	kv.Lock()
	defer kv.Unlock()

	kv.m[k] = value

	return true
}

func (kv *KV) PutWithUuid(v interface{}, args ...interface{}) (key string, ok bool) {
	value := NewValue(v)
	for i, v := range args {
		switch i {
		case 0:
			ttl, ok := v.(int)
			if !ok {
				return "", false
			}
			value.Ttl = ttl
		default:
			return "", false
		}
	}

	key = uuid.NewV4().String()

	kv.Lock()
	defer kv.Unlock()

	kv.m[key] = value

	return key, true
}
