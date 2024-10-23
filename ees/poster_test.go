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

package ees

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/segmentio/kafka-go"
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
		AMQPQueueID:      utils.StringPointer("q1"),
		AMQPExchange:     utils.StringPointer("E1"),
		AMQPRoutingKey:   utils.StringPointer("CGRCDR"),
		AMQPExchangeType: utils.StringPointer("fanout"),
	}
	amqp.parseOpts(opts)
	if !reflect.DeepEqual(expected, amqp) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(expected), utils.ToJSON(amqp))
	}
}

func TestKafkaParseURL(t *testing.T) {
	cfg := &config.EventExporterCfg{
		ExportPath: "127.0.0.1:9092",
		Attempts:   10,
		Opts: &config.EventExporterOpts{
			KafkaTopic: utils.StringPointer("cdr_billing"),
		},
	}
	want := &KafkaEE{
		cfg:  cfg,
		reqs: newConcReq(0),
		writer: &kafka.Writer{
			Addr:        kafka.TCP("127.0.0.1:9092"),
			Topic:       "cdr_billing",
			MaxAttempts: 1,
			Transport:   &kafka.Transport{},
		},
	}
	got, err := NewKafkaEE(cfg, nil)
	if err != nil {
		t.Fatalf("NewKafkaEE() failed unexpectedly: %v", err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("NewKafkaEE() = %+v, want %+v", got, want)
	}
}
