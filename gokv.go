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
	m     map[string]*Value
	mutex sync.RWMutex
}

func New() *KV {
	return &KV{
		m: make(map[string]*Value),
	}
}

func (kv *KV) Get(k string) interface{} {
	v, ok := kv.m[k]
	if !ok {
		return nil
	}

	if v.IsExpired() {
		kv.mutex.Lock()
		defer kv.mutex.Unlock()

		delete(kv.m, k)
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

	kv.mutex.Lock()
	defer kv.mutex.Unlock()

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

	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	kv.m[key] = value

	return key, true
}
