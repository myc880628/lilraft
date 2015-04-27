package lilraft

import "net/url"

type Node interface {
	rpcAppendEntries()
}

type HttpNode struct {
	url *url.URL
}
