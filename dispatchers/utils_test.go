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

package dispatchers

import (
	"reflect"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestDParseStringSetMetaZero(t *testing.T) {
	stringTest := utils.MetaZero
	result := ParseStringSet(stringTest)
	expected := make(utils.StringSet)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDParseStringSet(t *testing.T) {
	stringTest := "testString"
	result := ParseStringSet(stringTest)
	expected := utils.NewStringSet(strings.Split(stringTest, utils.ANDSep))
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}
