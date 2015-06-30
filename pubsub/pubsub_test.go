package pubsub

import (
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
)

func TestSubscribe(t *testing.T) {
	ps := NewPubSub(nil)
	var r string
	if err := ps.Subscribe(SubscribeInfo{
		EventType:    "test",
		PostUrl:      "url",
		LiveDuration: time.Second,
	}, &r); err != nil {
		t.Error("Error subscribing: ", err)
	}
	if expTime, exists := ps.subscribers["test"]["url"]; !exists || expTime.IsZero() {
		t.Error("Error adding subscriber: ", ps.subscribers)
	}
}

func TestSubscribeNoExpire(t *testing.T) {
	ps := NewPubSub(nil)
	var r string
	if err := ps.Subscribe(SubscribeInfo{
		EventType:    "test",
		PostUrl:      "url",
		LiveDuration: 0,
	}, &r); err != nil {
		t.Error("Error subscribing: ", err)
	}
	if expTime, exists := ps.subscribers["test"]["url"]; !exists || !expTime.IsZero() {
		t.Error("Error adding no expire subscriber: ", ps.subscribers)
	}
}

func TestUnsubscribe(t *testing.T) {
	ps := NewPubSub(nil)
	var r string
	if err := ps.Subscribe(SubscribeInfo{
		EventType:    "test",
		PostUrl:      "url",
		LiveDuration: time.Second,
	}, &r); err != nil {
		t.Error("Error subscribing: ", err)
	}
	if err := ps.Unsubscribe(SubscribeInfo{
		EventType: "test",
		PostUrl:   "url",
	}, &r); err != nil {
		t.Error("Error unsubscribing: ", err)
	}
	if _, exists := ps.subscribers["test"]["url"]; exists {
		t.Error("Error adding subscriber: ", ps.subscribers)
	}
}

func TestPublish(t *testing.T) {
	ps := NewPubSub(&config.CGRConfig{HttpSkipTlsVerify: true})
	ps.pubFunc = func(url string, ttl bool, obj interface{}) ([]byte, error) {
		obj.(map[string]string)["called"] = "yes"
		return nil, nil
	}
	var r string
	if err := ps.Subscribe(SubscribeInfo{
		EventType:    "test",
		PostUrl:      "url",
		LiveDuration: time.Second,
	}, &r); err != nil {
		t.Error("Error subscribing: ", err)
	}
	m := make(map[string]string)
	if err := ps.Publish(PublishInfo{
		EventType: "test",
		Event:     m,
	}, &r); err != nil {
		t.Error("Error publishing: ", err)
	}
	for i := 0; i < 1000; i++ { // wait for the theread to populate map
		if len(m) == 0 {
			time.Sleep(time.Microsecond)
		} else {
			break
		}
	}
	if r, exists := m["called"]; !exists || r != "yes" {
		t.Error("Error calling publish function: ", m)
	}
}

func TestPublishExpired(t *testing.T) {
	ps := NewPubSub(&config.CGRConfig{HttpSkipTlsVerify: true})
	ps.pubFunc = func(url string, ttl bool, obj interface{}) ([]byte, error) {
		m := obj.(map[string]string)
		m["called"] = "yes"
		return nil, nil
	}
	var r string
	if err := ps.Subscribe(SubscribeInfo{
		EventType:    "test",
		PostUrl:      "url",
		LiveDuration: 1,
	}, &r); err != nil {
		t.Error("Error subscribing: ", err)
	}
	if err := ps.Publish(PublishInfo{
		EventType: "test",
		Event:     nil,
	}, &r); err != nil {
		t.Error("Error publishing: ", err)
	}
	if len(ps.subscribers["test"]) != 0 {
		t.Error("Error removing expired subscribers: ", ps.subscribers)
	}
}
