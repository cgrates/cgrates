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
	amqp := &AMQPPoster{
		dialURL: "amqp://guest:guest@localhost:5672/?heartbeat=5",
	}
	expected := &AMQPPoster{
		dialURL:      "amqp://guest:guest@localhost:5672/?heartbeat=5",
		queueID:      "q1",
		exchange:     "E1",
		exchangeType: "fanout",
		routingKey:   "CGRCDR",
	}
	opts := map[string]interface{}{
		utils.AMQPQueueID:      "q1",
		utils.AMQPExchange:     "E1",
		utils.AMQPRoutingKey:   "CGRCDR",
		utils.AMQPExchangeType: "fanout",
	}
	amqp.parseOpts(opts)
	if !reflect.DeepEqual(expected, amqp) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(expected), utils.ToJSON(amqp))
	}
}

func TestKafkaParseURL(t *testing.T) {
	u := "127.0.0.1:9092"
	exp := &KafkaPoster{
		dialURL:  "127.0.0.1:9092",
		topic:    "cdr_billing",
		attempts: 10,
	}
	if kfk := NewKafkaPoster(u, 10, map[string]interface{}{utils.KafkaTopic: "cdr_billing"}); !reflect.DeepEqual(exp, kfk) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(kfk))
	}
	u = "localhost:9092"
	exp = &KafkaPoster{
		dialURL:  "localhost:9092",
		topic:    "cdr_billing",
		attempts: 10,
	}
	if kfk := NewKafkaPoster(u, 10, map[string]interface{}{utils.KafkaTopic: "cdr_billing"}); !reflect.DeepEqual(exp, kfk) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(kfk))
	}
}
