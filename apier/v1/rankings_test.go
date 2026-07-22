// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
