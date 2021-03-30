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

	"github.com/cgrates/cgrates/utils"
)

func TestModelsTimingMdlTableName(t *testing.T) {
	testStruct := TimingMdl{}
	exp := utils.TBLTPTimings
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", exp, result)
	}
}

func TestModelsDestinationMdlTableName(t *testing.T) {
	testStruct := DestinationMdl{}
	exp := utils.TBLTPDestinations
	result := testStruct.TableName()
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("\nExpected <%+v>,\nreceived <%+v>", exp, result)
	}
}
