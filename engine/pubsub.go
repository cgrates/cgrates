/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type SubscribeInfo struct {
	EventFilter string
	Transport   string
	Address     string
	LifeSpan    time.Duration
}

type CgrEvent map[string]string

func (ce CgrEvent) PassFilters(rsrFields utils.RSRFields) bool {
	for _, rsrFld := range rsrFields {
		if _, err := rsrFld.Parse(ce[rsrFld.Id]); err != nil {
			return false
		}
	}
	return true
}

type SubscriberData struct {
	ExpTime time.Time
	Filters utils.RSRFields
}

type PubSub struct {
	subscribers map[string]*SubscriberData
	ttlVerify   bool
	pubFunc     func(string, bool, []byte) ([]byte, error)
	mux         *sync.Mutex
	dm          *DataManager
}

func NewPubSub(dm *DataManager, ttlVerify bool) (*PubSub, error) {
	ps := &PubSub{
		ttlVerify:   ttlVerify,
		subscribers: make(map[string]*SubscriberData),
		pubFunc:     HttpJsonPost,
		mux:         &sync.Mutex{},
		dm:          dm,
	}
	// load subscribers
	if subs, err := dm.GetSubscribers(); err != nil {
		return nil, err
	} else {
		ps.subscribers = subs
	}
	for _, sData := range ps.subscribers {
		if err := sData.Filters.Compile(); err != nil { // Parse rules into regexp objects
			utils.Logger.Err(fmt.Sprintf("<PubSub> Error <%s> when parsing rules out of subscriber data: %+v", err.Error(), sData))
		}
	}
	return ps, nil
}

func (ps *PubSub) saveSubscriber(key string) {
	subData, found := ps.subscribers[key]
	if !found {
		return
	}
	if err := dm.SetSubscriber(key, subData); err != nil {
		utils.Logger.Err("<PubSub> Error saving subscriber: " + err.Error())
	}
}

func (ps *PubSub) removeSubscriber(key string) {
	if err := dm.RemoveSubscriber(key); err != nil {
		utils.Logger.Err("<PubSub> Error removing subscriber: " + err.Error())
	}
}

func (ps *PubSub) Subscribe(si SubscribeInfo, reply *string) error {
	ps.mux.Lock()
	defer ps.mux.Unlock()
	if si.Transport != utils.META_HTTP_POST {
		*reply = "Unsupported transport type"
		return errors.New(*reply)
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
	key := utils.InfieldJoin(si.Transport, si.Address)
	ps.subscribers[key] = &SubscriberData{
		ExpTime: expTime,
		Filters: rsr,
	}
	ps.saveSubscriber(key)
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
	key := utils.InfieldJoin(si.Transport, si.Address)
	delete(ps.subscribers, key)
	ps.removeSubscriber(key)
	*reply = utils.OK
	return nil
}

func (ps *PubSub) Publish(evt CgrEvent, reply *string) error {
	ps.mux.Lock()
	defer ps.mux.Unlock()
	evt["Timestamp"] = time.Now().Format(time.RFC3339Nano)
	for key, subData := range ps.subscribers {
		if !subData.ExpTime.IsZero() && subData.ExpTime.Before(time.Now()) {
			delete(ps.subscribers, key)
			ps.removeSubscriber(key)
			continue // subscription exevtred, do not send event
		}
		if subData.Filters == nil || !evt.PassFilters(subData.Filters) {
			continue // the event does not match the filters
		}
		split := utils.InfieldSplit(key)
		if len(split) != 2 {
			utils.Logger.Warning("<PubSub> Wrong transport;address pair: " + key)
			continue
		}
		transport := split[0]
		address := split[1]
		ttlVerify := ps.ttlVerify
		jsn, err := json.Marshal(evt)
		if err != nil {
			return err
		}
		switch transport {
		case utils.META_HTTP_POST:
			go func() {
				fib := utils.Fib()
				for i := 0; i < 5; i++ { // Loop so we can increase the success rate on best effort
					if _, err := ps.pubFunc(address, ttlVerify, jsn); err == nil {
						break // Success, no need to reinterate
					} else if i == 4 { // Last iteration, syslog the warning
						utils.Logger.Warning(fmt.Sprintf("<PubSub> Failed calling url: [%s], error: [%s], event type: %s", address, err.Error(), evt["EventName"]))
						break
					}
					time.Sleep(time.Duration(fib()) * time.Second)
				}
			}()
		}
	}
	*reply = utils.OK
	return nil
}

func (ps *PubSub) ShowSubscribers(in string, out *map[string]*SubscriberData) error {
	*out = ps.subscribers
	return nil
}

func (ps *PubSub) Call(serviceMethod string, args interface{}, reply interface{}) error {
	switch serviceMethod {
	case "PubSubV1.Subscribe":
		argsConverted, canConvert := args.(SubscribeInfo)
		if !canConvert {
			return rpcclient.ErrWrongArgsType
		}
		replyConverted, canConvert := reply.(*string)
		if !canConvert {
			return rpcclient.ErrWrongReplyType
		}
		return ps.Subscribe(argsConverted, replyConverted)
	case "PubSubV1.Unsubscribe":
		argsConverted, canConvert := args.(SubscribeInfo)
		if !canConvert {
			return rpcclient.ErrWrongArgsType
		}
		replyConverted, canConvert := reply.(*string)
		if !canConvert {
			return rpcclient.ErrWrongReplyType
		}
		return ps.Unsubscribe(argsConverted, replyConverted)
	case "PubSubV1.Publish":
		argsConverted, canConvert := args.(CgrEvent)
		if !canConvert {
			return rpcclient.ErrWrongArgsType
		}
		replyConverted, canConvert := reply.(*string)
		if !canConvert {
			return rpcclient.ErrWrongReplyType
		}
		return ps.Publish(argsConverted, replyConverted)
	case "PubSubV1.ShowSubscribers":
		argsConverted, canConvert := args.(string)
		if !canConvert {
			return rpcclient.ErrWrongArgsType
		}
		replyConverted, canConvert := reply.(*map[string]*SubscriberData)
		if !canConvert {
			return rpcclient.ErrWrongReplyType
		}
		return ps.ShowSubscribers(argsConverted, replyConverted)
	}
	return rpcclient.ErrUnsupporteServiceMethod
}

type ProxyPubSub struct {
	Client *rpcclient.RpcClient
}

func NewProxyPubSub(addr string, attempts, reconnects int, connectTimeout, replyTimeout time.Duration) (*ProxyPubSub, error) {
	client, err := rpcclient.NewRpcClient("tcp", addr, false, "", "", "", attempts, reconnects, connectTimeout, replyTimeout, utils.GOB, nil, false)
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
func (ps *ProxyPubSub) Publish(evt CgrEvent, reply *string) error {
	return ps.Client.Call("PubSubV1.Publish", evt, reply)
}

func (ps *ProxyPubSub) ShowSubscribers(in string, reply *map[string]*SubscriberData) error {
	return ps.Client.Call("PubSubV1.ShowSubscribers", in, reply)
}
