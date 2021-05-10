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

func TestAMQPSetOpts(t *testing.T) {
	k := new(AMQPER)
	k.dialURL = "amqp://localhost:2013"
	expKafka := &AMQPER{
		dialURL: "amqp://localhost:2013",
		queueID: "cdrs",
		tag:     "new",
	}
	if k.setOpts(map[string]interface{}{"amqpQueueID": "cdrs", "consumerTag": "new"}); expKafka.dialURL != k.dialURL {
		t.Errorf("Expected: %s ,received: %s", expKafka.dialURL, k.dialURL)
	} else if expKafka.queueID != k.queueID {
		t.Errorf("Expected: %s ,received: %s", expKafka.queueID, k.queueID)
	} else if expKafka.tag != k.tag {
		t.Errorf("Expected: %s ,received: %s", expKafka.tag, k.tag)
	}
	k = new(AMQPER)
	k.dialURL = "amqp://localhost:2013"
	expKafka = &AMQPER{
		dialURL: "amqp://localhost:2013",
		queueID: "cgrates_cdrs",
		tag:     "cgrates",
	}
	if k.setOpts(map[string]interface{}{}); expKafka.dialURL != k.dialURL {
		t.Errorf("Expected: %s ,received: %s", expKafka.dialURL, k.dialURL)
	} else if expKafka.queueID != k.queueID {
		t.Errorf("Expected: %s ,received: %s", expKafka.queueID, k.queueID)
	} else if expKafka.tag != k.tag {
		t.Errorf("Expected: %s ,received: %s", expKafka.tag, k.tag)
	}
}
