# GoKV
GoKV is a key-value storage in the memory, it is simple and fast.

## Requirements
Go1.18+

## Installation
```bash
go get -u github.com/skyline93/gokv
```

## Example
```go
package main

import (
	"github.com/skyline93/gokv"
)

func main(){
	cache := gokv.New(5)
	cache.Put("key1", "value1")
	key2 := cache.PutWithKey("value2")   // auto generate key

	v1 := cache.Get("key1")
	v2 := cache.Get(key2)	
}
```

```go
package main

import (
	"time"
	
	"github.com/skyline93/gokv"
)

func main(){
	cache := gokv.New(5)
	cache.Put("key1", "value1", 2)        // value expires after 2 second
	key2 := cache.PutWithKey("value2", 5) // value expires after 5 second


	time.Sleep(time.Second * 2)
	v1 := cache.Get("key1")                // v1 is nil

	time.Sleep(time.Second * 3)
	v2 := cache.Get(key2)                  // v2 is nil	
}
```
