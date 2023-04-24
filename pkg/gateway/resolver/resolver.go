package resolver

import (
	"errors"
	"net/url"
	"sync"
)

type resolver interface {
	Add(address *url.URL) error
	Get() (*url.URL, bool)
	Delete(address *url.URL) error
	Len() int
}

type Resolver struct {
	dns             map[string]resolver
	resolverNewFunc func() resolver
	rwmutex         sync.RWMutex
}

func NewResolver(newFunc func() resolver) *Resolver {
	r := &Resolver{
		dns:             make(map[string]resolver),
		resolverNewFunc: newFunc,
		rwmutex:         sync.RWMutex{},
	}
	return r
}

func (r *Resolver) Add(funcName string, address *url.URL) error {
	r.rwmutex.Lock()
	defer r.rwmutex.Unlock()
	if _, ok := r.dns[funcName]; !ok {
		r.dns[funcName] = r.resolverNewFunc()
	}
	return r.dns[funcName].Add(address)
}

func (r *Resolver) Get(funcName string) (*url.URL, bool) {
	r.rwmutex.RLock()
	defer r.rwmutex.RUnlock()
	if rr, ok := r.dns[funcName]; !ok {
		return nil, false
	} else {
		return rr.Get()
	}
}

func (r *Resolver) Delete(funcName string, address *url.URL) error {
	r.rwmutex.Lock()
	defer r.rwmutex.Unlock()
	if rr, ok := r.dns[funcName]; !ok {
		return errors.New("需要删除的函数不存在")
	} else {
		return rr.Delete(address)
	}
}

func (r *Resolver) Len(funcName string) int {
	r.rwmutex.RLock()
	defer r.rwmutex.RUnlock()
	if rr, ok := r.dns[funcName]; !ok {
		return 0
	} else {
		return rr.Len()
	}
}
