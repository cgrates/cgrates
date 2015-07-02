package engine

import (
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestSubscribe(t *testing.T) {
	ps := NewPubSub(accountingStorage, false)
	var r string
	if err := ps.Subscribe(SubscribeInfo{
		EventName: "test",
		Transport: utils.META_HTTP_POST,
		Address:   "url",
		LifeSpan:  time.Second,
	}, &r); err != nil {
		t.Error("Error subscribing: ", err)
	}
	if expTime, exists := ps.subscribers["test"][utils.InfieldJoin(utils.META_HTTP_POST, "url")]; !exists || expTime.IsZero() {
		t.Error("Error adding subscriber: ", ps.subscribers)
	}
}

func TestSubscribeSave(t *testing.T) {
	ps := NewPubSub(accountingStorage, false)
	var r string
	if err := ps.Subscribe(SubscribeInfo{
		EventName: "test",
		Transport: utils.META_HTTP_POST,
		Address:   "url",
		LifeSpan:  time.Second,
	}, &r); err != nil {
		t.Error("Error subscribing: ", err)
	}
	subs, err := accountingStorage.GetPubSubSubscribers()
	if err != nil || len(subs["test"]) != 1 {
		t.Error("Error saving subscribers: ", err, subs)
	}
}

func TestSubscribeNoTransport(t *testing.T) {
	ps := NewPubSub(accountingStorage, false)
	var r string
	if err := ps.Subscribe(SubscribeInfo{
		EventName: "test",
		Transport: "test",
		Address:   "url",
		LifeSpan:  time.Second,
	}, &r); err == nil {
		t.Error("Error subscribing error: ", err)
	}
}

func TestSubscribeNoExpire(t *testing.T) {
	ps := NewPubSub(accountingStorage, false)
	var r string
	if err := ps.Subscribe(SubscribeInfo{
		EventName: "test",
		Transport: utils.META_HTTP_POST,
		Address:   "url",
		LifeSpan:  0,
	}, &r); err != nil {
		t.Error("Error subscribing: ", err)
	}
	if expTime, exists := ps.subscribers["test"][utils.InfieldJoin(utils.META_HTTP_POST, "url")]; !exists || !expTime.IsZero() {
		t.Error("Error adding no expire subscriber: ", ps.subscribers)
	}
}

func TestUnsubscribe(t *testing.T) {
	ps := NewPubSub(accountingStorage, false)
	var r string
	if err := ps.Subscribe(SubscribeInfo{
		EventName: "test",
		Transport: utils.META_HTTP_POST,
		Address:   "url",
		LifeSpan:  time.Second,
	}, &r); err != nil {
		t.Error("Error subscribing: ", err)
	}
	if err := ps.Unsubscribe(SubscribeInfo{
		EventName: "test",
		Transport: utils.META_HTTP_POST,
		Address:   "url",
	}, &r); err != nil {
		t.Error("Error unsubscribing: ", err)
	}
	if _, exists := ps.subscribers["test"]["url"]; exists {
		t.Error("Error adding subscriber: ", ps.subscribers)
	}
}

func TestUnsubscribeSave(t *testing.T) {
	ps := NewPubSub(accountingStorage, false)
	var r string
	if err := ps.Subscribe(SubscribeInfo{
		EventName: "test",
		Transport: utils.META_HTTP_POST,
		Address:   "url",
		LifeSpan:  time.Second,
	}, &r); err != nil {
		t.Error("Error subscribing: ", err)
	}
	if err := ps.Unsubscribe(SubscribeInfo{
		EventName: "test",
		Transport: utils.META_HTTP_POST,
		Address:   "url",
	}, &r); err != nil {
		t.Error("Error unsubscribing: ", err)
	}
	subs, err := accountingStorage.GetPubSubSubscribers()
	if err != nil || len(subs["test"]) != 0 {
		t.Error("Error saving subscribers: ", err, subs)
	}
}

func TestPublish(t *testing.T) {
	ps := NewPubSub(accountingStorage, true)
	ps.pubFunc = func(url string, ttl bool, obj interface{}) ([]byte, error) {
		obj.(map[string]string)["called"] = url
		return nil, nil
	}
	var r string
	if err := ps.Subscribe(SubscribeInfo{
		EventName: "test",
		Transport: utils.META_HTTP_POST,
		Address:   "url",
		LifeSpan:  time.Second,
	}, &r); err != nil {
		t.Error("Error subscribing: ", err)
	}
	m := make(map[string]string)
	m["EventName"] = "test"
	if err := ps.Publish(PublishInfo{
		Event: m,
	}, &r); err != nil {
		t.Error("Error publishing: ", err)
	}
	for i := 0; i < 1000; i++ { // wait for the theread to populate map
		if len(m) == 1 {
			time.Sleep(time.Microsecond)
		} else {
			break
		}
	}
	if r, exists := m["called"]; !exists || r != "url" {
		t.Error("Error calling publish function: ", m)
	}
}

func TestPublishExpired(t *testing.T) {
	ps := NewPubSub(accountingStorage, true)
	ps.pubFunc = func(url string, ttl bool, obj interface{}) ([]byte, error) {
		m := obj.(map[string]string)
		m["called"] = "yes"
		return nil, nil
	}
	var r string
	if err := ps.Subscribe(SubscribeInfo{
		EventName: "test",
		Transport: utils.META_HTTP_POST,
		Address:   "url",
		LifeSpan:  1,
	}, &r); err != nil {
		t.Error("Error subscribing: ", err)
	}
	if err := ps.Publish(PublishInfo{
		Event: map[string]string{"EventName": "test"},
	}, &r); err != nil {
		t.Error("Error publishing: ", err)
	}
	if len(ps.subscribers["test"]) != 0 {
		t.Error("Error removing expired subscribers: ", ps.subscribers)
	}
}

func TestPublishExpiredSave(t *testing.T) {
	ps := NewPubSub(accountingStorage, true)
	ps.pubFunc = func(url string, ttl bool, obj interface{}) ([]byte, error) {
		m := obj.(map[string]string)
		m["called"] = "yes"
		return nil, nil
	}
	var r string
	if err := ps.Subscribe(SubscribeInfo{
		EventName: "test",
		Transport: utils.META_HTTP_POST,
		Address:   "url",
		LifeSpan:  1,
	}, &r); err != nil {
		t.Error("Error subscribing: ", err)
	}
	subs, err := accountingStorage.GetPubSubSubscribers()
	if err != nil || len(subs["test"]) != 1 {
		t.Error("Error saving subscribers: ", err, subs)
	}
	if err := ps.Publish(PublishInfo{
		Event: map[string]string{"EventName": "test"},
	}, &r); err != nil {
		t.Error("Error publishing: ", err)
	}
	subs, err = accountingStorage.GetPubSubSubscribers()
	if err != nil || len(subs["test"]) != 0 {
		t.Error("Error saving subscribers: ", err, subs)
	}
}
