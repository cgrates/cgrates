/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or56
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
	//"time"
	// "github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var mapEv = MapEvent(map[string]interface{}{
	"test1": nil,
	"test2": 42,
	"test3": 42.3,
	"test4": true,
	"test5": "test",
})

func TestMapEventNewMapEvent(t *testing.T) {
	expected := MapEvent(make(map[string]interface{}))
	rply := NewMapEvent(nil)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	mp := map[string]interface{}{
		"test1": nil,
		"test2": 42,
		"test3": 42.3,
		"test4": true,
		"test5": "test",
	}
	expected = MapEvent(mp)
	rply = NewMapEvent(mp)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestMapEventString(t *testing.T) {
	me := NewMapEvent(nil)
	expected := utils.ToJSON(me)
	rply := me.String()
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = utils.ToJSON(mapEv)
	rply = mapEv.String()
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestMapEventHasField(t *testing.T) {
	me := NewMapEvent(nil)
	expected := false
	rply := me.HasField("test1")
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = true
	rply = mapEv.HasField("test2")
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = false
	rply = mapEv.HasField("test7")
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestMapEventGetString(t *testing.T) {
	if _, err := mapEv.GetString("test"); err != utils.ErrNotFound {
		t.Errorf("Expected: %+v, received: %+v", utils.ErrNotFound, err)
	}
	strI, err := mapEv.GetString("test2")
	if strI != "42" || err != nil {
		t.Errorf("Expecting %+v, received: %+v", "42", strI)
	}
	strI, err = mapEv.GetString("test1")
	if strI != "null" || err != nil {
		t.Errorf("Expecting %+v, received: %+v", "null", strI)
	}
}

func TestMapEventGetStringIgnoreErrors(t *testing.T) {
	if s := mapEv.GetStringIgnoreErrors("test"); s != "" {
		t.Errorf("Expected: %+v, received: %+v", "", s)
	}
	strI := mapEv.GetStringIgnoreErrors("test2")
	if strI != "42" {
		t.Errorf("Expecting %+v, received: %+v", "42", strI)
	}
	strI = mapEv.GetStringIgnoreErrors("test1")
	if strI != "null" {
		t.Errorf("Expecting %+v, received: %+v", "null", strI)
	}
}
