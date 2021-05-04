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

package ers

import (
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestKafkasetOpts(t *testing.T) {
	k := new(KafkaER)
	k.dialURL = "localhost:2013"
	expKafka := &KafkaER{
		dialURL: "localhost:2013",
		topic:   "cdrs",
		groupID: "new",
		maxWait: time.Second,
	}
	if err := k.setOpts(map[string]interface{}{
		utils.KafkaTopic:   "cdrs",
		utils.KafkaGroupID: "new",
		utils.KafkaMaxWait: "1s",
	}); err != nil {
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
	k.dialURL = "localhost:2013"
	expKafka = &KafkaER{
		dialURL: "localhost:2013",
		topic:   "cgrates",
		groupID: "cgrates",
		maxWait: time.Millisecond,
	}
	if err := k.setOpts(map[string]interface{}{}); err != nil {
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
	k.dialURL = "127.0.0.1:2013"
	expKafka = &KafkaER{
		dialURL: "127.0.0.1:2013",
		topic:   "cdrs",
		groupID: "new",
		maxWait: time.Second,
	}
	if err := k.setOpts(map[string]interface{}{
		utils.KafkaTopic:   "cdrs",
		utils.KafkaGroupID: "new",
		utils.KafkaMaxWait: "1s",
	}); err != nil {
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
