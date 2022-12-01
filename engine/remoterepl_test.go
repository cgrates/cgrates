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
	"testing"

	"github.com/cgrates/cgrates/utils"
)

// unfinished
func TestUpdateReplicationFilters(t *testing.T) {
	UpdateReplicationFilters("objType", "objId", utils.EmptyString)
	UpdateReplicationFilters("objType", "objId", "connID")
}

// unfinished
// cover never stops
// func TestReplicateMultipleIDs(t *testing.T) {
// 	connMgr := NewConnManager(config.NewDefaultCGRConfig())
// 	ctx := context.Background()
// 	connId := []string{"connID"}
// 	objIds := []string{"ObjIds"}
// 	err := replicateMultipleIDs(ctx, connMgr, connId, false, "objType", objIds, "method", "args")
// 	t.Error(err)

// }
