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
package v1

import (
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func TestSchedulerSv1Ping(t *testing.T) {

	scheduler := &SchedulerSv1{}

	ctx := context.Background()
	ign := &utils.CGREvent{}
	reply := ""

	err := scheduler.Ping(ctx, ign, &reply)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if reply != utils.Pong {
		t.Errorf("Expected reply to be %v, got %v", utils.Pong, reply)
	}
}

func TestSchedulerSv1Call(t *testing.T) {
	scheduler := &SchedulerSv1{}
	ctx := context.Background()
	serviceMethod := "ServiceMethod"
	args := "Args"
	var reply string
	err := scheduler.Call(ctx, serviceMethod, args, &reply)
	if err == nil {
		t.Fatalf("Expected error")
	}

}
