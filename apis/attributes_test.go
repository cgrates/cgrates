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

package apis

import (
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func BenchmarkCallAPIerRPC(b *testing.B) {
	as := NewAttributeSv1(nil)
	ctx := context.Background()
	args := &utils.CGREvent{}
	var reply string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		as.Call(ctx, utils.AttributeSv1Ping, args, &reply)
	}
}

func BenchmarkCallAsRPCService(b *testing.B) {
	as := NewAttributeSv1(nil)
	ctx := context.Background()
	args := &utils.CGREvent{}
	var reply string
	b.ResetTimer()
	srv, _ := birpc.NewService(as, "", false)
	for i := 0; i < b.N; i++ {
		srv.Call(ctx, utils.AttributeSv1Ping, args, &reply)
	}
}
