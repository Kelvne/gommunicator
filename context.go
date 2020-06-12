package gommunicator

import "sync"

// Context main context on actions
type Context struct {
	Request *DataTransactionRequest

	store map[string]interface{}
	lock  sync.RWMutex
}

// Set sets a new value on the context store
func (ctx *Context) Set(key string, val interface{}) {
	ctx.lock.Lock()
	defer ctx.lock.Unlock()

	if ctx.store == nil {
		ctx.store = make(map[string]interface{})
	}

	ctx.store[key] = val
}

// Get retrieves a value from param store
func (ctx *Context) Get(key string) interface{} {
	ctx.lock.RLock()
	defer ctx.lock.RUnlock()
	return ctx.store[key]
}
