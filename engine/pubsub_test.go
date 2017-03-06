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
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestSubscribe(t *testing.T) {
	ps, err := NewPubSub(dataStorage, false)
	if err != nil {
		t.Error(err)
	}
	var r string
	if err := ps.Subscribe(SubscribeInfo{
		EventFilter: "EventName/test",
		Transport:   utils.META_HTTP_POST,
		Address:     "url",
		LifeSpan:    time.Second,
	}, &r); err != nil {
		t.Error("Error subscribing: ", err)
	}
	if subData, exists := ps.subscribers[utils.InfieldJoin(utils.META_HTTP_POST, "url")]; !exists || subData.ExpTime.IsZero() {
		t.Error("Error adding subscriber: ", ps.subscribers)
	}
}

func TestSubscribeSave(t *testing.T) {
	ps, err := NewPubSub(dataStorage, false)
	if err != nil {
		t.Error(err)
	}
	var r string
	if err := ps.Subscribe(SubscribeInfo{
		EventFilter: "EventName/test",
		Transport:   utils.META_HTTP_POST,
		Address:     "url",
		LifeSpan:    time.Second,
	}, &r); err != nil {
		t.Error("Error subscribing: ", err)
	}
	subs, err := dataStorage.GetSubscribers()
	if err != nil || len(subs) != 1 {
		t.Error("Error saving subscribers: ", err, subs)
	}
}

func TestSubscribeNoTransport(t *testing.T) {
	ps, err := NewPubSub(dataStorage, false)
	if err != nil {
		t.Error(err)
	}
	var r string
	if err := ps.Subscribe(SubscribeInfo{
		EventFilter: "EventName/test",
		Transport:   "test",
		Address:     "url",
		LifeSpan:    time.Second,
	}, &r); err == nil {
		t.Error("Error subscribing error: ", err)
	}
}

func TestSubscribeNoExpire(t *testing.T) {
	ps, err := NewPubSub(dataStorage, false)
	if err != nil {
		t.Error(err)
	}
	var r string
	if err := ps.Subscribe(SubscribeInfo{
		EventFilter: "EventName/test",
		Transport:   utils.META_HTTP_POST,
		Address:     "url",
		LifeSpan:    0,
	}, &r); err != nil {
		t.Error("Error subscribing: ", err)
	}
	if subData, exists := ps.subscribers[utils.InfieldJoin(utils.META_HTTP_POST, "url")]; !exists || !subData.ExpTime.IsZero() {
		t.Error("Error adding no expire subscriber: ", ps.subscribers)
	}
}

func TestUnsubscribe(t *testing.T) {
	ps, err := NewPubSub(dataStorage, false)
	if err != nil {
		t.Error(err)
	}
	var r string
	if err := ps.Subscribe(SubscribeInfo{
		EventFilter: "EventName/test",
		Transport:   utils.META_HTTP_POST,
		Address:     "url",
		LifeSpan:    time.Second,
	}, &r); err != nil {
		t.Error("Error subscribing: ", err)
	}
	if err := ps.Unsubscribe(SubscribeInfo{
		EventFilter: "EventName/test",
		Transport:   utils.META_HTTP_POST,
		Address:     "url",
	}, &r); err != nil {
		t.Error("Error unsubscribing: ", err)
	}
	if _, exists := ps.subscribers[utils.InfieldJoin(utils.META_HTTP_POST, "url")]; exists {
		t.Error("Error adding subscriber: ", ps.subscribers)
	}
}

func TestUnsubscribeSave(t *testing.T) {
	ps, err := NewPubSub(dataStorage, false)
	if err != nil {
		t.Error(err)
	}
	var r string
	if err := ps.Subscribe(SubscribeInfo{
		EventFilter: "EventName/test",
		Transport:   utils.META_HTTP_POST,
		Address:     "url",
		LifeSpan:    time.Second,
	}, &r); err != nil {
		t.Error("Error subscribing: ", err)
	}
	if err := ps.Unsubscribe(SubscribeInfo{
		EventFilter: "EventName/test",
		Transport:   utils.META_HTTP_POST,
		Address:     "url",
	}, &r); err != nil {
		t.Error("Error unsubscribing: ", err)
	}
	subs, err := dataStorage.GetSubscribers()
	if err != nil || len(subs) != 0 {
		t.Error("Error saving subscribers: ", err, subs)
	}
}

func TestPublishExpired(t *testing.T) {
	ps, err := NewPubSub(dataStorage, true)
	if err != nil {
		t.Error(err)
	}
	ps.pubFunc = func(url string, ttl bool, obj []byte) ([]byte, error) {
		return nil, nil
	}
	var r string
	if err := ps.Subscribe(SubscribeInfo{
		EventFilter: "EventName/test",
		Transport:   utils.META_HTTP_POST,
		Address:     "url",
		LifeSpan:    1,
	}, &r); err != nil {
		t.Error("Error subscribing: ", err)
	}
	if err := ps.Publish(map[string]string{"EventFilter": "test"}, &r); err != nil {
		t.Error("Error publishing: ", err)
	}
	if len(ps.subscribers) != 0 {
		t.Error("Error removing expired subscribers: ", ps.subscribers)
	}
}

func TestPublishExpiredSave(t *testing.T) {
	ps, err := NewPubSub(dataStorage, true)
	if err != nil {
		t.Error(err)
	}
	ps.pubFunc = func(url string, ttl bool, obj []byte) ([]byte, error) {
		return nil, nil
	}
	var r string
	if err := ps.Subscribe(SubscribeInfo{
		EventFilter: "EventName/test",
		Transport:   utils.META_HTTP_POST,
		Address:     "url",
		LifeSpan:    1,
	}, &r); err != nil {
		t.Error("Error subscribing: ", err)
	}
	subs, err := dataStorage.GetSubscribers()
	if err != nil || len(subs) != 1 {
		t.Error("Error saving subscribers: ", err, subs)
	}
	if err := ps.Publish(map[string]string{"EventFilter": "test"}, &r); err != nil {
		t.Error("Error publishing: ", err)
	}
	subs, err = dataStorage.GetSubscribers()
	if err != nil || len(subs) != 0 {
		t.Error("Error saving subscribers: ", err, subs)
	}
}

func TestCgrEventPassFilters(t *testing.T) {
	ev := CgrEvent{"EventName": "TEST_EVENT", "Header1": "Value1", "Header2": "Value2"}
	if !ev.PassFilters(utils.ParseRSRFieldsMustCompile("EventName(TEST_EVENT)", utils.INFIELD_SEP)) {
		t.Error("Not passing filter")
	}
	if ev.PassFilters(utils.ParseRSRFieldsMustCompile("EventName(DUMMY_EVENT)", utils.INFIELD_SEP)) {
		t.Error("Passing filter")
	}
	if !ev.PassFilters(utils.ParseRSRFieldsMustCompile("^EventName::TEST_EVENT(TEST_EVENT)", utils.INFIELD_SEP)) {
		t.Error("Not passing filter")
	}
	if !ev.PassFilters(utils.ParseRSRFieldsMustCompile("^EventName::DUMMY", utils.INFIELD_SEP)) { // Should pass since we have no filter defined
		t.Error("Not passing no filter")
	}
	if !ev.PassFilters(utils.ParseRSRFieldsMustCompile("~EventName:s/^(.*)_/$1/(TEST)", utils.INFIELD_SEP)) {
		t.Error("Not passing filter")
	}
	if !ev.PassFilters(utils.ParseRSRFieldsMustCompile("~EventName:s/^(\\w*)_/$1/:s/^(\\w)(\\w)(\\w)(\\w)/$1$3$4/(TST)", utils.INFIELD_SEP)) {
		t.Error("Not passing filter")
	}
	if !ev.PassFilters(utils.ParseRSRFieldsMustCompile("EventName(TEST_EVENT);Header1(Value1)", utils.INFIELD_SEP)) {
		t.Error("Not passing filter")
	}
	if ev.PassFilters(utils.ParseRSRFieldsMustCompile("EventName(TEST_EVENT);Header1(Value2)", utils.INFIELD_SEP)) {
		t.Error("Passing filter")
	}
	if !ev.PassFilters(utils.ParseRSRFieldsMustCompile("EventName(TEST_EVENT);~Header1:s/(\\d)/$1/(1)", utils.INFIELD_SEP)) {
		t.Error("Not passing filter")
	}
	if ev.PassFilters(utils.ParseRSRFieldsMustCompile("EventName(TEST_EVENT);~Header1:s/(\\d)/$1/(2)", utils.INFIELD_SEP)) {
		t.Error("Passing filter")
	}
}
