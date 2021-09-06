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
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func TestDynamicDPnewDynamicDP(t *testing.T) {

	expDDP := &dynamicDP{
		resConns:  []string{"conn1"},
		stsConns:  []string{"conn2"},
		actsConns: []string{"conn3"},
		tenant:    "cgrates.org",
		initialDP: utils.StringSet{
			"test": struct{}{},
		},
		cache: utils.MapStorage{},
		ctx:   context.Background(),
	}

	if rcv := newDynamicDP(context.Background(), []string{"conn1"}, []string{"conn2"},
		[]string{"conn3"}, "cgrates.org",
		utils.StringSet{"test": struct{}{}}); !reflect.DeepEqual(rcv, expDDP) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expDDP), utils.ToJSON(rcv))
	}
}

// func TestDynamicDPString(t *testing.T) {
// 	dDP := &dynamicDP{
// 		tenant: "cgrates.org",
// 		initialDP: utils.StringSet{
// 			"test1": struct{}{},
// 			"test2": struct{}{},
// 		},
// 		cache: utils.MapStorage{},
// 		ctx:   context.Background(),
// 	}

// 	exp := `["test1","test2"]`
// 	rcv := dDP.String()
// 	if rcv != exp {
// 		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
// 	}
// }
