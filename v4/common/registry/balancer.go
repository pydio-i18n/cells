package registry

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type Balancer struct {
	m map[string]Backend
}

type Backend struct {
	u *url.URL
	Alive bool
	mux sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func NewBalancer(r Registry) {

}
