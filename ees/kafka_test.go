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
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	kafka "github.com/segmentio/kafka-go"
)

func TestKafkaEEConnect(t *testing.T) {
	kafkaEE := &KafkaEE{
		writer: nil,
		cfg:    &config.EventExporterCfg{},
		dc:     &utils.ExporterMetrics{},
		reqs:   &concReq{},
	}
	err := kafkaEE.Connect()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestKafkaEE_Cfg(t *testing.T) {
	expectedCfg := &config.EventExporterCfg{}
	kafkaEE := &KafkaEE{
		cfg: expectedCfg,
	}
	result := kafkaEE.Cfg()
	if result != expectedCfg {
		t.Errorf("expected %v, got %v", expectedCfg, result)
	}
}

func TestKafkaEEGetMetrics(t *testing.T) {
	safeMapStorage := &utils.ExporterMetrics{}
	kafkaEE := &KafkaEE{
		dc: safeMapStorage,
	}
	result := kafkaEE.GetMetrics()
	if result != safeMapStorage {
		t.Errorf("expected %v, got %v", safeMapStorage, result)
	}
}

func TestKafkaEEClose(t *testing.T) {
	writer := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "test-topic",
		Balancer: &kafka.LeastBytes{},
	}
	kafkaEE := &KafkaEE{
		writer: writer,
	}
	err := kafkaEE.Close()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestKafkaEE_ExportEvent(t *testing.T) {
	writer := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "test-topic",
		Balancer: &kafka.LeastBytes{},
	}
	kafkaEE := &KafkaEE{
		writer: writer,
		reqs:   &concReq{},
	}
	content := []byte("test message")
	key := "test-key"
	err := kafkaEE.ExportEvent(content, key)
	if err == nil {
		t.Errorf("expected no error, got %v", err)
	}
}
