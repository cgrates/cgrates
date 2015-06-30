package pubsub

import (
	"fmt"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type SubscribeInfo struct {
	EventType    string
	PostUrl      string
	LiveDuration time.Duration
}

type PublishInfo struct {
	EventType string
	Event     map[string]string
}

type PublishSubscriber interface {
	Subscribe(SubscribeInfo, *string) error
	Unsubscribe(SubscribeInfo, *string) error
	Publish(PublishInfo, *string) error
}

type PubSub struct {
	subscribers map[string]map[string]time.Time
	conf        *config.CGRConfig
	pubFunc     func(string, bool, interface{}) ([]byte, error)
}

func NewPubSub(conf *config.CGRConfig) *PubSub {
	return &PubSub{
		conf:        conf,
		subscribers: make(map[string]map[string]time.Time),
		pubFunc:     utils.HttpJsonPost,
	}
}

func (ps *PubSub) Subscribe(si SubscribeInfo, reply *string) error {
	if ps.subscribers[si.EventType] == nil {
		ps.subscribers[si.EventType] = make(map[string]time.Time)
	}
	var expTime time.Time
	if si.LiveDuration > 0 {
		expTime = time.Now().Add(si.LiveDuration)
	}
	ps.subscribers[si.EventType][si.PostUrl] = expTime
	*reply = utils.OK
	return nil
}

func (ps *PubSub) Unsubscribe(si SubscribeInfo, reply *string) error {
	delete(ps.subscribers[si.EventType], si.PostUrl)
	*reply = utils.OK
	return nil
}

func (ps *PubSub) Publish(pi PublishInfo, reply *string) error {
	subs := ps.subscribers[pi.EventType]
	for postURL, expTime := range subs {
		if !expTime.IsZero() && expTime.Before(time.Now()) {
			delete(subs, postURL)
			continue // subscription expired, do not send event
		}
		url := postURL
		go func() {
			delay := utils.Fib()
			for i := 0; i < 5; i++ { // Loop so we can increase the success rate on best effort
				if _, err := ps.pubFunc(url, ps.conf.HttpSkipTlsVerify, pi.Event); err == nil {
					break // Success, no need to reinterate
				} else if i == 4 { // Last iteration, syslog the warning
					engine.Logger.Warning(fmt.Sprintf("<PubSub> WARNING: Failed calling url: [%s], error: [%s], event type: %s", url, err.Error(), pi.EventType))
					break
				}
				time.Sleep(delay())
			}
		}()
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

func (ps *ProxyPubSub) Subscribe(sqID string, values *map[string]float64) error {
	return ps.Client.Call("PubSub.Subscribe", sqID, values)
}
