package main

import (
	"log"
	"sync"
	"thegoods.biz/httpbuf"
	"time"
)

type cache struct {
	cache map[string]*httpbuf.Buffer
	mu    sync.RWMutex
}

func (c *cache) get(path string) (buf *httpbuf.Buffer, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	buf, ok = c.cache[path]
	return
}

func (c *cache) set(path string, val *httpbuf.Buffer) {
	c.mu.Lock()
	defer c.mu.Unlock()

	log.Println("caching:", path)
	c.cache[path] = val

	//set up the expiration
	time.AfterFunc(time.Minute*5, func() {
		c.del(path)
	})
}

func (c *cache) del(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	log.Println("cache expired:", path)
	delete(c.cache, path)
}

var app_cache = cache{
	cache: make(map[string]*httpbuf.Buffer),
}
