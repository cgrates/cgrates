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

package console

import (
	"reflect"
	"strings"
	"testing"

	v1 "github.com/cgrates/cgrates/apier/v1"

	"github.com/cgrates/cgrates/utils"
)

func TestCmdCacheGroupItemId(t *testing.T) {
	// commands map is initiated in init function
	command := commands["cache_group_item_ids"]
	// verify if ApierSv1 object has method on it
	m, ok := reflect.TypeOf(new(v1.CacheSv1)).MethodByName(strings.Split(command.RpcMethod(), utils.NestingSep)[1])
	if !ok {
		t.Fatal("method not found")
	}
	if m.Type.NumIn() != 3 { // ApierSv1 is consider and we expect 3 inputs
		t.Fatalf("invalid number of input parameters ")
	}
	// verify the type of input parameter
	if ok := m.Type.In(1).AssignableTo(reflect.TypeOf(command.RpcParams(true))); !ok {
		t.Fatalf("cannot assign input parameter")
	}
	// verify the type of output parameter
	if ok := m.Type.In(2).AssignableTo(reflect.TypeOf(command.RpcResult())); !ok {
		t.Fatalf("cannot assign output parameter")
	}
	// for coverage purpose
	if err := command.PostprocessRpcParams(); err != nil {
		t.Fatal(err)
	}
}
