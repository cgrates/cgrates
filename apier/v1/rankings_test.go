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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestNewRankingSv1(t *testing.T) {
	rankingS := &engine.RankingS{}
	rankingSvc := NewRankingSv1(rankingS)
	if rankingSvc == nil {
		t.Errorf("NewRankingSv1() returned nil")
	}
	if rankingSvc.rnkS != rankingS {
		t.Errorf("NewRankingSv1() did not correctly set rnkS field")
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
