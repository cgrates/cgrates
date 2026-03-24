/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package ees

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestAMQPeeParseURL(t *testing.T) {
	amqp := &AMQPee{
		cfg: &config.EventExporterCfg{ExportPath: "amqp://guest:guest@localhost:5672/?heartbeat=5"},
	}
	expected := &AMQPee{
		cfg:          &config.EventExporterCfg{ExportPath: "amqp://guest:guest@localhost:5672/?heartbeat=5"},
		queueID:      "q1",
		exchange:     "E1",
		exchangeType: "fanout",
		routingKey:   "CGRCDR",
	}
	opts := &config.EventExporterOpts{
		AMQP: &config.AMQPOpts{
			QueueID:      utils.StringPointer("q1"),
			Exchange:     utils.StringPointer("E1"),
			RoutingKey:   utils.StringPointer("CGRCDR"),
			ExchangeType: utils.StringPointer("fanout"),
		}}
	amqp.parseOpts(opts)
	if !reflect.DeepEqual(expected, amqp) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(expected), utils.ToJSON(amqp))
	}
}

func TestNewKafkaEEParsesOpts(t *testing.T) {
	cfg := &config.EventExporterCfg{
		ExportPath: "127.0.0.1:9092",
		Attempts:   10,
		Opts: &config.EventExporterOpts{
			Kafka: &config.KafkaOpts{
				Topic: utils.StringPointer("cdr_billing"),
			},
		},
	}
	got, err := NewKafkaEE(cfg, nil)
	if err != nil {
		t.Fatalf("NewKafkaEE() failed unexpectedly: %v", err)
	}
	defer got.Close()
	if got.Cfg() != cfg {
		t.Error("NewKafkaEE() config mismatch")
	}
	if got.timeout != defaultKafkaTimeout {
		t.Errorf("NewKafkaEE() timeout = %v, want %v", got.timeout, defaultKafkaTimeout)
	}
}
