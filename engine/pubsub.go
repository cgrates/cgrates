package engine

import (
	"errors"
	"fmt"
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
	ShowSubscribers(string, *map[string]map[string]*SubscriberData) error
}

type SubscriberData struct {
	ExpTime time.Time
	Filters utils.RSRFields
}

type PubSub struct {
	subscribers map[string]map[string]*SubscriberData
	ttlVerify   bool
	pubFunc     func(string, bool, interface{}) ([]byte, error)
	mux         *sync.Mutex
	accountDb   AccountingStorage
}

func NewPubSub(accountDb AccountingStorage, ttlVerify bool) *PubSub {
	ps := &PubSub{
		ttlVerify:   ttlVerify,
		subscribers: make(map[string]map[string]*SubscriberData),
		pubFunc:     utils.HttpJsonPost,
		mux:         &sync.Mutex{},
		accountDb:   accountDb,
	}
	// load subscribers
	if subs, err := accountDb.GetPubSubSubscribers(); err == nil {
		ps.subscribers = subs
	}
	return ps
}

func (ps *PubSub) saveSubscribers(key string) {
	if key != "" {
		if _, found := ps.subscribers[key]; !found {
			return
		}
		if err := accountingStorage.SetPubSubSubscribers(key, ps.subscribers[key]); err != nil {
			Logger.Err("<PubSub> Error saving subscribers: " + err.Error())
		}
	} else { // save all
		for key, valueMap := range ps.subscribers {
			if err := accountingStorage.SetPubSubSubscribers(key, valueMap); err != nil {
				Logger.Err("<PubSub> Error saving subscribers: " + err.Error())
			}
		}
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
		ps.subscribers[si.EventName] = make(map[string]*SubscriberData)
	}
	var expTime time.Time
	if si.LifeSpan > 0 {
		expTime = time.Now().Add(si.LifeSpan)
	}
	rsr, err := utils.ParseRSRFields(si.EventFilter, utils.INFIELD_SEP)
	if err != nil {
		*reply = err.Error()
		return err
	}
	ps.subscribers[si.EventName][utils.InfieldJoin(si.Transport, si.Address)] = &SubscriberData{
		ExpTime: expTime,
		Filters: rsr,
	}
	ps.saveSubscribers(si.EventName)
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
	ps.saveSubscribers(si.EventName)
	*reply = utils.OK
	return nil
}

func (ps *PubSub) Publish(pi PublishInfo, reply *string) error {
	ps.mux.Lock()
	defer ps.mux.Unlock()
	subs := ps.subscribers[pi.Event["EventName"]]
	for transportAddress, subData := range subs {
		split := utils.InfieldSplit(transportAddress)
		if len(split) != 2 {
			Logger.Warning("<PubSub> Wrong transport;address pair: " + transportAddress)
			continue
		}
		transport := split[0]
		address := split[1]
		if !subData.ExpTime.IsZero() && subData.ExpTime.Before(time.Now()) {
			delete(subs, transportAddress)
			ps.saveSubscribers(pi.Event["EventName"])
			continue // subscription expired, do not send event
		}
		switch transport {
		case utils.META_HTTP_POST:
			go func() {
				delay := utils.Fib()
				for i := 0; i < 5; i++ { // Loop so we can increase the success rate on best effort
					if _, err := ps.pubFunc(address, ps.ttlVerify, pi.Event); err == nil {
						break // Success, no need to reinterate
					} else if i == 4 { // Last iteration, syslog the warning
						Logger.Warning(fmt.Sprintf("<PubSub> Failed calling url: [%s], error: [%s], event type: %s", address, err.Error(), pi.Event["EventName"]))
						break
					}
					time.Sleep(delay())
				}
			}()
		}
	}
	*reply = utils.OK
	return nil
}

func (ps *PubSub) ShowSubscribers(in string, out *map[string]map[string]*SubscriberData) error {
	*out = ps.subscribers
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
	return ps.Client.Call("PubSubV1.Subscribe", si, reply)
}
func (ps *ProxyPubSub) Unsubscribe(si SubscribeInfo, reply *string) error {
	return ps.Client.Call("PubSubV1.Unsubscribe", si, reply)
}
func (ps *ProxyPubSub) Publish(pi PublishInfo, reply *string) error {
	return ps.Client.Call("PubSubV1.Publish", pi, reply)
}

func (ps *ProxyPubSub) ShowSubscribers(in string, reply *map[string]map[string]*SubscriberData) error {
	return ps.Client.Call("PubSubV1.ShowSubscribers", in, reply)
}
