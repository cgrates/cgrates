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
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestSetFldPostCacheTTL(t *testing.T) {
	var1 := failedPostCache
	SetFailedPostCacheTTL(50 * time.Millisecond)
	var2 := failedPostCache
	if reflect.DeepEqual(var1, var2) {
		t.Error("Expecting to be different")
	}
}

func TestAddFldPost(t *testing.T) {
	SetFailedPostCacheTTL(5 * time.Second)
	AddFailedPost("", "path1", "format1", "module1", "1", make(map[string]interface{}))
	x, ok := failedPostCache.Get(utils.ConcatenatedKey("", "path1", "format1", "module1"))
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
		Format: "format1",
		module: "module1",
		Events: []interface{}{"1"},
		Opts:   make(map[string]interface{}),
	}
	if !reflect.DeepEqual(eOut, failedPost) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(failedPost))
	}
	AddFailedPost("", "path1", "format1", "module1", "2", make(map[string]interface{}))
	AddFailedPost("", "path2", "format2", "module2", "3", map[string]interface{}{utils.SQSQueueID: "qID"})
	x, ok = failedPostCache.Get(utils.ConcatenatedKey("", "path1", "format1", "module1"))
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
		Format: "format1",
		module: "module1",
		Events: []interface{}{"1", "2"},
		Opts:   make(map[string]interface{}),
	}
	if !reflect.DeepEqual(eOut, failedPost) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(failedPost))
	}
	x, ok = failedPostCache.Get(utils.ConcatenatedKey("", "path2", "format2", "module2", "qID"))
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
		Format: "format2",
		module: "module2",
		Events: []interface{}{"3"},
		Opts:   map[string]interface{}{utils.SQSQueueID: "qID"},
	}
	if !reflect.DeepEqual(eOut, failedPost) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(failedPost))
	}
}

func TestFilePath(t *testing.T) {
	exportEvent := &ExportEvents{}
	rcv := exportEvent.FilePath()
	if rcv[0] != '|' {
		t.Errorf("Expecting: '|', received: %+v", rcv[0])
	} else if rcv[8:] != ".gob" {
		t.Errorf("Expecting: '.gob', received: %+v", rcv[8:])
	}
	exportEvent = &ExportEvents{
		module: "module",
	}
	rcv = exportEvent.FilePath()
	if rcv[:7] != "module|" {
		t.Errorf("Expecting: 'module|', received: %+v", rcv[:7])
	} else if rcv[14:] != ".gob" {
		t.Errorf("Expecting: '.gob', received: %+v", rcv[14:])
	}

}

func TestSetModule(t *testing.T) {
	exportEvent := &ExportEvents{}
	eOut := &ExportEvents{
		module: "module",
	}
	exportEvent.SetModule("module")
	if !reflect.DeepEqual(eOut, exportEvent) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, exportEvent)
	}
}

func TestAddEvent(t *testing.T) {
	exportEvent := &ExportEvents{}
	eOut := &ExportEvents{Events: []interface{}{"event1"}}
	exportEvent.AddEvent("event1")
	if !reflect.DeepEqual(eOut, exportEvent) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, exportEvent)
	}
	exportEvent = &ExportEvents{}
	eOut = &ExportEvents{Events: []interface{}{"event1", "event2", "event3"}}
	exportEvent.AddEvent("event1")
	exportEvent.AddEvent("event2")
	exportEvent.AddEvent("event3")
	if !reflect.DeepEqual(eOut, exportEvent) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(exportEvent))
	}
}
