package history

import (
	"net/rpc"
)

type ProxyStore struct {
	client *rpc.Client
}

func NewProxyStore(addr string) (*ProxyStore, error) {
	client, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &ProxyStore{client: client}, nil
}

func (ps *ProxyStore) Record(key string, obj interface{}) error {
	if err := ps.client.Call("Store.Record", key, obj); err != nil {
		return err
	}
	return nil
}
