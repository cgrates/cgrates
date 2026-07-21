// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

func TestNewKafkaEEParsesOpts(t *testing.T) {
	cfg := &config.EventExporterCfg{
		ExportPath: "127.0.0.1:9092",
		Attempts:   10,
		Opts: &config.EventExporterOpts{
			KafkaTopic: utils.StringPointer("cdr_billing"),
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
