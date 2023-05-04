package resolver

import (
	"math/rand"
	"net/url"
	"time"
)

type RandomResolver struct {
	urls []*url.URL
}

func (r *RandomResolver) Add(address *url.URL) error {
	r.urls = append(r.urls, address)
	return nil
}

func (r *RandomResolver) Get() (*url.URL, bool) {
	randG := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.urls[randG.Intn(len(r.urls))], true
}

func (r *RandomResolver) Delete(address *url.URL) error {
	idx := 0
	for idx < len(r.urls) && r.urls[idx] != address {
		idx++
	}
	r.urls = append(r.urls[:idx], r.urls[idx+1:]...)
	return nil
}

func (r *RandomResolver) Len() int {
	return len(r.urls)
}

func RandomResolverFunc() func() resolver {
	return func() resolver {
		return &RandomResolver{
			urls: make([]*url.URL, 0),
		}
	}
}
