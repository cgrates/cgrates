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
)

func TestAMQPSetURL(t *testing.T) {
	k := new(AMQPER)
	expKafka := &AMQPER{
		dialURL: "amqp://localhost:2013",
		queueID: "cdrs",
		tag:     "new",
	}
	url := "amqp://localhost:2013?queue_id=cdrs&consumer_tag=new"
	if err := k.setURL(url); err != nil {
		t.Fatal(err)
	} else if expKafka.dialURL != k.dialURL {
		t.Errorf("Expected: %s ,received: %s", expKafka.dialURL, k.dialURL)
	} else if expKafka.queueID != k.queueID {
		t.Errorf("Expected: %s ,received: %s", expKafka.queueID, k.queueID)
	} else if expKafka.tag != k.tag {
		t.Errorf("Expected: %s ,received: %s", expKafka.tag, k.tag)
	}
	k = new(AMQPER)
	expKafka = &AMQPER{
		dialURL: "amqp://localhost:2013",
		queueID: "cgrates_cdrs",
		tag:     "cgrates",
	}
	url = "amqp://localhost:2013"
	if err := k.setURL(url); err != nil {
		t.Fatal(err)
	} else if expKafka.dialURL != k.dialURL {
		t.Errorf("Expected: %s ,received: %s", expKafka.dialURL, k.dialURL)
	} else if expKafka.queueID != k.queueID {
		t.Errorf("Expected: %s ,received: %s", expKafka.queueID, k.queueID)
	} else if expKafka.tag != k.tag {
		t.Errorf("Expected: %s ,received: %s", expKafka.tag, k.tag)
	}
	k = new(AMQPER)
	expKafka = &AMQPER{
		dialURL: "amqp://localhost:2013",
		queueID: "cgrates",
		tag:     "cgrates",
	}
	if err := k.setURL("127.0.0.1:2013?queue_id=cdrs&consumer_tag=new"); err == nil {
		t.Errorf("Expected error received: %v", err)
	}
}
