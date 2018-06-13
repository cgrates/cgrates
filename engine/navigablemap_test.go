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
	"errors"
	"strings"
	"testing"
)

func TestNavMapGetFieldAsString(t *testing.T) {
	nM := &NavigableMap{
		data: map[string]interface{}{
			"FirstLevel": map[string]interface{}{
				"SecondLevel": map[string]interface{}{
					"ThirdLevel": map[string]interface{}{
						"Fld1": "Val1",
					},
				},
			},
			"AnotherFirstLevel": "ValAnotherFirstLevel",
		},
	}
	eVal := "Val1"
	if strVal, err := nM.FieldAsString(
		strings.Split("FirstLevel>SecondLevel>ThirdLevel>Fld1", ">")); err != nil {
		t.Error(err)
	} else if strVal != eVal {
		t.Errorf("expecting: <%s> received: <%s>", eVal, strVal)
	}
	eVal = "ValAnotherFirstLevel"
	if strVal, err := nM.FieldAsString(
		strings.Split("AnotherFirstLevel", ">")); err != nil {
		t.Error(err)
	} else if strVal != eVal {
		t.Errorf("expecting: <%s> received: <%s>", eVal, strVal)
	}
	fPath := "NonExisting>AnotherFirstLevel"
	if _, err := nM.FieldAsString(strings.Split(fPath, ">")); err.Error() !=
		errors.New("no map at path: <NonExisting>").Error() {
		t.Error(err)
	}
}

type myEv map[string]interface{}

func (ev myEv) AsNavigableMap() (*NavigableMap, error) {
	return NewNavigableMap(ev), nil
}
