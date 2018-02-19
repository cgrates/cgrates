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
package utils

import (
	"errors"
	"reflect"
	"testing"
)

type myEv map[string]interface{}

func (ev myEv) AsCGRReply() (CGRReply, error) {
	return CGRReply(ev), nil
}

func TestCGRReplyNew(t *testing.T) {
	eCgrRply := CGRReply(map[string]interface{}{
		Error: "some",
	})
	if rpl, err := NewCGRReply(nil, errors.New("some")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCgrRply, rpl) {
		t.Errorf("Expecting: %+v, received: %+v", ToJSON(eCgrRply), ToJSON(rpl))
	}
	ev := myEv{
		"FirstLevel": map[string]interface{}{
			"SecondLevel": map[string]interface{}{
				"Fld1": "Val1",
			},
		},
	}
	eCgrRply = CGRReply(ev)
	eCgrRply[Error] = ""
	if rpl, err := NewCGRReply(CGRReplier(ev), nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCgrRply, rpl) {
		t.Errorf("Expecting: %+v, received: %+v", eCgrRply, rpl)
	}
}

func TestCGRReplyGetFieldAsString(t *testing.T) {
	ev := myEv{
		"FirstLevel": map[string]interface{}{
			"SecondLevel": map[string]interface{}{
				"ThirdLevel": map[string]interface{}{
					"Fld1": "Val1",
				},
			},
		},
		"AnotherFirstLevel": "ValAnotherFirstLevel",
	}
	cgrRply, _ := NewCGRReply(CGRReplier(ev), nil)
	if strVal, err := cgrRply.GetFieldAsString("*cgrReply>Error", ">"); err != nil {
		t.Error(err)
	} else if strVal != "" {
		t.Errorf("received: <%s>", strVal)
	}
	eVal := "Val1"
	if strVal, err := cgrRply.GetFieldAsString("*cgrReply>FirstLevel>SecondLevel>ThirdLevel>Fld1", ">"); err != nil {
		t.Error(err)
	} else if strVal != eVal {
		t.Errorf("expecting: <%s> received: <%s>", eVal, strVal)
	}
	eVal = "ValAnotherFirstLevel"
	if strVal, err := cgrRply.GetFieldAsString("*cgrReply>AnotherFirstLevel", ">"); err != nil {
		t.Error(err)
	} else if strVal != eVal {
		t.Errorf("expecting: <%s> received: <%s>", eVal, strVal)
	}
}
