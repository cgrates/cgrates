// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ers

import (
	"testing"
	"time"
)

func TestKafkaSetURL(t *testing.T) {
	k := new(KafkaER)
	expKafka := &KafkaER{
		dialURL: "localhost:2013",
		topic:   "cdrs",
		groupID: "new",
		maxWait: time.Second,
	}
	url := "localhost:2013?topic=cdrs&group_id=new&max_wait=1s"
	if err := k.setURL(url); err != nil {
		t.Fatal(err)
	} else if expKafka.dialURL != k.dialURL {
		t.Errorf("Expected: %s ,received: %s", expKafka.dialURL, k.dialURL)
	} else if expKafka.topic != k.topic {
		t.Errorf("Expected: %s ,received: %s", expKafka.topic, k.topic)
	} else if expKafka.groupID != k.groupID {
		t.Errorf("Expected: %s ,received: %s", expKafka.groupID, k.groupID)
	} else if expKafka.maxWait != k.maxWait {
		t.Errorf("Expected: %s ,received: %s", expKafka.maxWait, k.maxWait)
	}
	k = new(KafkaER)
	expKafka = &KafkaER{
		dialURL: "localhost:2013",
		topic:   "cgrates",
		groupID: "cgrates",
		maxWait: time.Millisecond,
	}
	url = "localhost:2013"
	if err := k.setURL(url); err != nil {
		t.Fatal(err)
	} else if expKafka.dialURL != k.dialURL {
		t.Errorf("Expected: %s ,received: %s", expKafka.dialURL, k.dialURL)
	} else if expKafka.topic != k.topic {
		t.Errorf("Expected: %s ,received: %s", expKafka.topic, k.topic)
	} else if expKafka.groupID != k.groupID {
		t.Errorf("Expected: %s ,received: %s", expKafka.groupID, k.groupID)
	} else if expKafka.maxWait != k.maxWait {
		t.Errorf("Expected: %s ,received: %s", expKafka.maxWait, k.maxWait)
	}
	k = new(KafkaER)
	expKafka = &KafkaER{
		dialURL: "localhost:2013",
		topic:   "cgrates",
		groupID: "cgrates",
		maxWait: time.Millisecond,
	}
	if err := k.setURL("127.0.0.1?%"); err == nil {
		t.Errorf("Expected error received: %v", err)
	}

	k = new(KafkaER)
	expKafka = &KafkaER{
		dialURL: "127.0.0.1:2013",
		topic:   "cdrs",
		groupID: "new",
		maxWait: time.Second,
	}
	url = "127.0.0.1:2013?topic=cdrs&group_id=new&max_wait=1s"
	if err := k.setURL(url); err != nil {
		t.Fatal(err)
	} else if expKafka.dialURL != k.dialURL {
		t.Errorf("Expected: %s ,received: %s", expKafka.dialURL, k.dialURL)
	} else if expKafka.topic != k.topic {
		t.Errorf("Expected: %s ,received: %s", expKafka.topic, k.topic)
	} else if expKafka.groupID != k.groupID {
		t.Errorf("Expected: %s ,received: %s", expKafka.groupID, k.groupID)
	} else if expKafka.maxWait != k.maxWait {
		t.Errorf("Expected: %s ,received: %s", expKafka.maxWait, k.maxWait)
	}
}
