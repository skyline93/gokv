package gokv

import (
	"fmt"
	"log"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
)

type Node struct {
	next *Node
	prev *Node

	key interface{}
}

type List struct {
	head *Node
	tail *Node
}

func (l *List) Insert(key interface{}) {
	node := &Node{key: key}

	if l.head == nil {
		l.head = node
		l.tail = node
		return
	}

	t := l.tail
	t.next = node
	node.prev = t
	l.tail = node
}

func (l *List) get(key interface{}) *Node {
	n := l.head

	if n.key == key {
		return n
	}

	for n.next != nil {
		if n.key == key {
			return n
		}

		n = n.next
	}

	return nil
}

func (l *List) Delete(key interface{}) {
	n := l.get(key)

	if n == nil {
		return
	}

	if n.prev == nil && n.next == nil {
		l.head = nil
		l.tail = nil
		return
	}

	if n.prev == nil {
		l.head = n.next
		n.next.prev = nil
		return
	}

	if n.next == nil {
		l.tail = n.prev
		n.prev.next = nil
		return
	}

	next := n.next
	prev := n.prev

	next.prev = prev
	prev.next = next
}

func (l *List) Head() (key interface{}) {
	if l.head == nil {
		return nil
	}

	return l.head.key
}

func (l *List) All() {
	n := l.head

	if n == nil {
		return
	}

	fmt.Printf("n: %s ", n.key)
	for n.next != nil {
		n = n.next
		fmt.Printf("n: %s ", n.key)
	}
}

type Value struct {
	v interface{}

	ttl        int
	createTime time.Time
}

func NewValue(v interface{}, ttl int) *Value {
	return &Value{
		v:          v,
		ttl:        ttl,
		createTime: time.Now(),
	}
}

func (v *Value) expiration() time.Time {
	if v.ttl < 0 {
		return v.createTime.Add(time.Second * time.Duration(999999999))
	}

	return v.createTime.Add(time.Second * time.Duration(v.ttl))
}

func (v *Value) IsExpired() bool {
	return time.Now().After(v.expiration())
}

type Cache struct {
	sync.Mutex
	index *List
	kv    map[string]*Value
	len   int

	gcChan chan interface{}
}

func New(len int) *Cache {
	c := &Cache{
		index: new(List),
		kv:    make(map[string]*Value, len),
		len:   len,

		gcChan: make(chan interface{}, 100),
	}

	go func() {
		for {
			select {
			case g := <-c.gcChan:
				// TODO
				log.Printf("gc: %v\n", g)
			}
		}
	}()

	return c
}

func (c *Cache) collect() {
	k := c.index.Head()
	c.index.Delete(k)

	g := struct {
		K string
		V interface{}
	}{K: k.(string), V: c.kv[k.(string)]}

	delete(c.kv, k.(string))

	c.gcChan <- g
}

func (c *Cache) reset(key string) {
	c.index.Delete(key)
	c.index.Insert(key)
}

func (c *Cache) delete(key string) {
	c.index.Delete(key)
	delete(c.kv, key)
}

func (c *Cache) Delete(key string) {
	c.Lock()
	defer c.Unlock()

	c.delete(key)
}

func (c *Cache) put(key string, value interface{}, opts ...interface{}) {
	ttl := -1

	for i, opt := range opts {
		switch i {
		case 0:
			ttl = opt.(int)
		}
	}

	c.Lock()
	defer c.Unlock()

	if len(c.kv) == c.len {
		c.collect()
	}

	c.kv[key] = NewValue(value, ttl)
	c.index.Insert(key)
}

func (c *Cache) Put(key string, value interface{}, opts ...interface{}) {
	c.put(key, value, opts...)
}

func (c *Cache) PutWithKey(value interface{}, opts ...interface{}) (key string) {
	key = uuid.NewV4().String()

	c.put(key, value, opts...)
	return
}

func (c *Cache) Get(key string) interface{} {
	c.Lock()
	defer c.Unlock()

	value, ok := c.kv[key]
	if !ok {
		return nil
	}

	if value.IsExpired() {
		c.delete(key)
		return nil
	}

	c.reset(key)
	return value.v
}

func (c *Cache) Update(key string, value interface{}) (ok bool) {
	c.Lock()
	defer c.Unlock()

	v, ok := c.kv[key]
	if !ok {
		return false
	}

	if v.IsExpired() {
		c.delete(key)
		return false
	}

	v.v = value
	c.reset(key)
	return true
}
