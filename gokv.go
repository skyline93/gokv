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
	mu   sync.RWMutex
}

func (l *List) Insert(key interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.insert(key)
}

func (l *List) insert(key interface{}) {
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

func (l *List) Get(key interface{}) *Node {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.get(key)
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
	l.mu.Lock()
	defer l.mu.Unlock()

	l.delete(key)
}

func (l *List) delete(key interface{}) {
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

func (l *List) ReSet(key interface{}) {
	l.delete(key)
	l.insert(key)
}

func (l *List) Head() (key interface{}) {
	l.mu.RLock()
	defer l.mu.Unlock()

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
	mu    sync.Mutex
	index *List
	kv    map[string]*Value
	len   int

	gcChan       chan interface{}
	delExpirChan chan string
	resetChan    chan string
}

func New(len int) *Cache {
	c := &Cache{
		index: new(List),
		kv:    make(map[string]*Value, len),
		len:   len,

		gcChan:       make(chan interface{}, 100),
		delExpirChan: make(chan string, 100),
		resetChan:    make(chan string, 100),
	}

	go func() {
		for {
			select {
			case g := <-c.gcChan:
				// TODO
				log.Printf("gc: %v\n", g)
			case key := <-c.delExpirChan:
				log.Printf("delete expired key: %v\n", key)
				c.Delete(key)
			case key := <-c.resetChan:
				c.index.ReSet(key)
			}
		}
	}()

	return c
}

func (c *Cache) collect() {
	k := c.index.Head()
	c.index.Delete(k)

	delete(c.kv, k.(string))

	c.gcChan <- k
}

func (c *Cache) delete(key string) {
	c.index.Delete(key)
	delete(c.kv, key)
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.delete(key)
}

func (c *Cache) set(key string, value interface{}, opts ...interface{}) {
	ttl := -1

	for i, opt := range opts {
		switch i {
		case 0:
			ttl = opt.(int)
		}
	}

	if len(c.kv) == c.len {
		c.collect()
	}

	c.kv[key] = NewValue(value, ttl)
	c.index.Insert(key)
}

func (c *Cache) Put(key string, value interface{}, opts ...interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.set(key, value, opts...)
}

func (c *Cache) PutWithKey(value interface{}, opts ...interface{}) (key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key = uuid.NewV4().String()

	c.set(key, value, opts...)
	return
}

func (c *Cache) Get(key string) interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()

	value, ok := c.kv[key]
	if !ok {
		return nil
	}

	if value.IsExpired() {
		c.delExpirChan <- key
		return nil
	}

	c.resetChan <- key
	return value.v
}
