package pubsub

import (
	"errors"
	"sync"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type SubscribeInfo struct {
	EventName   string
	EventFilter string
	Transport   string
	Address     string
	LifeSpan    time.Duration
}

type PublishInfo struct {
	Event map[string]string
}

type PublisherSubscriber interface {
	Subscribe(SubscribeInfo, *string) error
	Unsubscribe(SubscribeInfo, *string) error
	Publish(PublishInfo, *string) error
}

type PubSub struct {
	subscribers map[string]map[string]time.Time
	ttlVerify   bool
	pubFunc     func(string, bool, interface{}) ([]byte, error)
	mux         *sync.Mutex
}

func NewPubSub(ttlVerify bool) *PubSub {
	return &PubSub{
		ttlVerify:   ttlVerify,
		subscribers: make(map[string]map[string]time.Time),
		pubFunc:     utils.HttpJsonPost,
		mux:         &sync.Mutex{},
	}
}

func (ps *PubSub) Subscribe(si SubscribeInfo, reply *string) error {
	ps.mux.Lock()
	defer ps.mux.Unlock()
	if si.Transport != utils.META_HTTP_POST {
		*reply = "Unsupported transport type"
		return errors.New(*reply)
	}
	if ps.subscribers[si.EventName] == nil {
		ps.subscribers[si.EventName] = make(map[string]time.Time)
	}
	var expTime time.Time
	if si.LifeSpan > 0 {
		expTime = time.Now().Add(si.LifeSpan)
	}
	ps.subscribers[si.EventName][utils.InfieldJoin(si.Transport, si.Address)] = expTime
	*reply = utils.OK
	return nil
}

func (ps *PubSub) Unsubscribe(si SubscribeInfo, reply *string) error {
	ps.mux.Lock()
	defer ps.mux.Unlock()
	if si.Transport != utils.META_HTTP_POST {
		*reply = "Unsupported transport type"
		return errors.New(*reply)
	}
	delete(ps.subscribers[si.EventName], utils.InfieldJoin(si.Transport, si.Address))
	*reply = utils.OK
	return nil
}

func (ps *PubSub) Publish(pi PublishInfo, reply *string) error {
	ps.mux.Lock()
	defer ps.mux.Unlock()
	subs := ps.subscribers[pi.Event["EventName"]]
	for transport_address, expTime := range subs {
		split := utils.InfieldSplit(transport_address)
		if len(split) != 2 {
			continue
		}
		transport := split[0]
		address := split[1]
		if !expTime.IsZero() && expTime.Before(time.Now()) {
			delete(subs, transport_address)
			continue // subscription expired, do not send event
		}
		switch transport {
		case utils.META_HTTP_POST:
			go func() {
				delay := utils.Fib()
				for i := 0; i < 5; i++ { // Loop so we can increase the success rate on best effort
					if _, err := ps.pubFunc(address, ps.ttlVerify, pi.Event); err == nil {
						break // Success, no need to reinterate
					}
					time.Sleep(delay())
				}
			}()
		}
	}
	*reply = utils.OK
	return nil
}

type ProxyPubSub struct {
	Client *rpcclient.RpcClient
}

func NewProxyPubSub(addr string, reconnects int) (*ProxyPubSub, error) {
	client, err := rpcclient.NewRpcClient("tcp", addr, reconnects, utils.GOB)
	if err != nil {
		return nil, err
	}
	return &ProxyPubSub{Client: client}, nil
}

func (ps *ProxyPubSub) Subscribe(si SubscribeInfo, reply *string) error {
	return ps.Client.Call("PubSub.Subscribe", si, reply)
}
func (ps *ProxyPubSub) Unsubscribe(si SubscribeInfo, reply *string) error {
	return ps.Client.Call("PubSub.Unsubscribe", si, reply)
}
func (ps *ProxyPubSub) Publish(pi PublishInfo, reply *string) error {
	return ps.Client.Call("PubSub.Publish", pi, reply)
}
