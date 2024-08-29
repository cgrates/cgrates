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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestRankingSv1Ping(t *testing.T) {
	sa := &RankingSv1{}
	ctx := context.Background()
	ign := &utils.CGREvent{}
	var reply string
	err := sa.Ping(ctx, ign, &reply)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if reply != utils.Pong {
		t.Errorf("expected reply to be %v, got %v", utils.Pong, reply)
	}
}
func TestNewRankingSv1(t *testing.T) {
	rankingSvc := NewRankingSv1()
	if rankingSvc == nil {
		t.Errorf("NewRankingSv1() returned nil")
	}
}

func TestRemoveRankingProfile(t *testing.T) {
	dataManager := &engine.DataManager{}
	apierSv1 := &APIerSv1{
		DataManager: dataManager,
	}
	args := &utils.TenantIDWithAPIOpts{
		APIOpts: map[string]interface{}{
			utils.CacheOpt: "cacheOptValue",
		},
	}
	var reply string
	err := apierSv1.RemoveRankingProfile(nil, args, &reply)
	if err == nil {
		t.Fatalf("RemoveRankingProfile() returned an error: %v", err)
	}
	if reply == utils.OK {
		t.Errorf("RemoveRankingProfile() returned reply = %v, want %v", reply, utils.OK)
	}

}

func TestRalsPing(t *testing.T) {
	rsv1 := &RALsV1{}
	ctx := context.Background()
	var ign *utils.CGREvent
	var reply string
	err := rsv1.Ping(ctx, ign, &reply)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expectedReply := utils.Pong
	if reply != expectedReply {
		t.Errorf("Expected reply %v, got %v", expectedReply, reply)
	}
}

func TestRalsCall(t *testing.T) {
	rsv1 := &RALsV1{}
	ctx := context.Background()
	serviceMethod := "TestServiceMethod"
	args := "TestArgs"
	var reply string
	err := rsv1.Call(ctx, serviceMethod, args, &reply)
	if err == nil {
		t.Errorf("UNSUPPORTED_SERVICE_METHOD")
	}
	expectedReply := "response"
	if reply == expectedReply {
		t.Errorf("Expected reply %v, got %v", expectedReply, reply)
	}
}
