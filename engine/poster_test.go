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
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestAMQPPosterParseURL(t *testing.T) {
	amqp := &AMQPPoster{}
	expected := &AMQPPoster{
		dialURL:      "amqp://guest:guest@localhost:5672/?heartbeat=5",
		queueID:      "q1",
		exchange:     "E1",
		exchangeType: "fanout",
		routingKey:   "CGRCDR",
	}
	dialURL := "amqp://guest:guest@localhost:5672/?queue_id=q1&exchange=E1&routing_key=CGRCDR&heartbeat=5&exchange_type=fanout"
	if err := amqp.parseURL(dialURL); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, amqp) {
		t.Errorf("Expected: %s ,recived: %s", utils.ToJSON(expected), utils.ToJSON(amqp))
	}
}
