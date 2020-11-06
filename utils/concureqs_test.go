/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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
	"reflect"
	"testing"
)

func TestConcureqsNewConReqs(t *testing.T) {
	expected := &ConcReqs{strategy: "test", aReqs: make(chan struct{}, 1)}
	received := NewConReqs(1, "test")
	if reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: %+v, received: %+v", expected, received)
	}
}

func TestConcureqsIsLimited(t *testing.T) {
	received := NewConReqs(1, "test").IsLimited()
	if received != true {
		t.Errorf("Expecting: true, received: %+v", received)
	}
}
