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
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestSetFldPostCacheTTL(t *testing.T) {
	var1 := failedPostCache
	SetFailedPostCacheTTL(50*time.Millisecond, false)
	var2 := failedPostCache
	if reflect.DeepEqual(var1, var2) {
		t.Error("Expecting to be different")
	}
}

func TestAddFldPost(t *testing.T) {
	SetFailedPostCacheTTL(5*time.Second, false)
	AddFailedPost("", "path1", "format1", "1", &config.EventExporterOpts{
		AMQP:  &config.AMQPOpts{},
		Els:   &config.ElsOpts{},
		AWS:   &config.AWSOpts{},
		NATS:  &config.NATSOpts{},
		Kafka: &config.KafkaOpts{},
		RPC:   &config.RPCOpts{},
		SQL:   &config.SQLOpts{},
	})
	x, ok := failedPostCache.Get(utils.ConcatenatedKey("", "path1", "format1"))
	if !ok {
		t.Error("Error reading from cache")
	}
	if x == nil {
		t.Error("Received an empty element")
	}

	failedPost, canCast := x.(*ExportEvents)
	if !canCast {
		t.Error("Error when casting")
	}
	eOut := &ExportEvents{
		Path:   "path1",
		Type:   "format1",
		Events: []any{"1"},
		Opts: &config.EventExporterOpts{
			AMQP:  &config.AMQPOpts{},
			Els:   &config.ElsOpts{},
			AWS:   &config.AWSOpts{},
			NATS:  &config.NATSOpts{},
			Kafka: &config.KafkaOpts{},
			RPC:   &config.RPCOpts{},
			SQL:   &config.SQLOpts{},
		},
	}
	if !reflect.DeepEqual(eOut, failedPost) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(failedPost))
	}
	AddFailedPost("", "path1", "format1", "2", &config.EventExporterOpts{
		AMQP:  &config.AMQPOpts{},
		Els:   &config.ElsOpts{},
		AWS:   &config.AWSOpts{},
		NATS:  &config.NATSOpts{},
		Kafka: &config.KafkaOpts{},
		RPC:   &config.RPCOpts{},
		SQL:   &config.SQLOpts{},
	})
	AddFailedPost("", "path2", "format2", "3", &config.EventExporterOpts{
		AWS: &config.AWSOpts{
			SQSQueueID: utils.StringPointer("qID"),
		},
		NATS:  &config.NATSOpts{},
		Kafka: &config.KafkaOpts{},
		RPC:   &config.RPCOpts{},
		AMQP:  &config.AMQPOpts{},
		Els:   &config.ElsOpts{},
		SQL:   &config.SQLOpts{},
	})
	x, ok = failedPostCache.Get(utils.ConcatenatedKey("", "path1", "format1"))
	if !ok {
		t.Error("Error reading from cache")
	}
	if x == nil {
		t.Error("Received an empty element")
	}
	failedPost, canCast = x.(*ExportEvents)
	if !canCast {
		t.Error("Error when casting")
	}
	eOut = &ExportEvents{
		Path:   "path1",
		Type:   "format1",
		Events: []any{"1", "2"},
		Opts: &config.EventExporterOpts{
			AMQP:  &config.AMQPOpts{},
			Els:   &config.ElsOpts{},
			AWS:   &config.AWSOpts{},
			NATS:  &config.NATSOpts{},
			Kafka: &config.KafkaOpts{},
			RPC:   &config.RPCOpts{},
			SQL:   &config.SQLOpts{},
		},
	}
	if !reflect.DeepEqual(eOut, failedPost) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(failedPost))
	}
	x, ok = failedPostCache.Get(utils.ConcatenatedKey("", "path2", "format2", "qID"))
	if !ok {
		t.Error("Error reading from cache")
	}
	if x == nil {
		t.Error("Received an empty element")
	}
	failedPost, canCast = x.(*ExportEvents)
	if !canCast {
		t.Error("Error when casting")
	}
	eOut = &ExportEvents{
		Path:   "path2",
		Type:   "format2",
		Events: []any{"3"},
		Opts: &config.EventExporterOpts{
			Els:   &config.ElsOpts{},
			NATS:  &config.NATSOpts{},
			SQL:   &config.SQLOpts{},
			AMQP:  &config.AMQPOpts{},
			RPC:   &config.RPCOpts{},
			Kafka: &config.KafkaOpts{},
			AWS: &config.AWSOpts{
				SQSQueueID: utils.StringPointer("qID")},
		},
	}
	if !reflect.DeepEqual(eOut, failedPost) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(failedPost))
	}
	for _, id := range failedPostCache.GetItemIDs("") {
		failedPostCache.Set(id, nil, nil)
	}
}

func TestFilePath(t *testing.T) {
	exportEvent := &ExportEvents{}
	rcv := exportEvent.FilePath()
	if !strings.HasSuffix(rcv, ".gob") || !strings.HasPrefix(rcv, utils.EEs) {
		t.Errorf("Unexpected fileName: %q", rcv) // EEs|sha1.gob
	}
	exportEvent = &ExportEvents{}
	rcv = exportEvent.FilePath()
	if !strings.HasSuffix(rcv, ".gob") || !strings.HasPrefix(rcv, utils.EEs) {
		t.Errorf("Unexpected fileName: %q", rcv) // EEs|sha1.gob
	}

}

func TestAddEvent(t *testing.T) {
	exportEvent := &ExportEvents{}
	eOut := &ExportEvents{Events: []any{"event1"}}
	exportEvent.AddEvent("event1")
	if !reflect.DeepEqual(eOut, exportEvent) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, exportEvent)
	}
	exportEvent = &ExportEvents{}
	eOut = &ExportEvents{Events: []any{"event1", "event2", "event3"}}
	exportEvent.AddEvent("event1")
	exportEvent.AddEvent("event2")
	exportEvent.AddEvent("event3")
	if !reflect.DeepEqual(eOut, exportEvent) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(exportEvent))
	}
}
