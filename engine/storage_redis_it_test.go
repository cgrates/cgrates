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
	"fmt"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var rds *RedisStorage
var err error

func TestRDSitConnectRedis(t *testing.T) {
	if !*testIntegration {
		return
	}
	cfg, _ := config.NewDefaultCGRConfig()
	rds, err = NewRedisStorage(fmt.Sprintf("%s:%s", cfg.TpDbHost, cfg.TpDbPort), 4, cfg.TpDbPass, cfg.DBDataEncoding, utils.REDIS_MAX_CONNS, nil, 1)
	if err != nil {
		t.Fatal("Could not connect to Redis", err.Error())
	}
}

func TestRDSitFlush(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := rds.Flush(""); err != nil {
		t.Error("Failed to Flush redis database", err.Error())
	}
	rds.PreloadRatingCache()
}

func TestRDSitSetGetDerivedCharges(t *testing.T) {
	if !*testIntegration {
		return
	}
	keyCharger1 := utils.ConcatenatedKey("*out", "cgrates.org", "call", "dan", "dan")
	charger1 := &utils.DerivedChargers{Chargers: []*utils.DerivedCharger{
		&utils.DerivedCharger{RunID: "extra1", RequestTypeField: "^prepaid", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "rif", SubjectField: "rif", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
		&utils.DerivedCharger{RunID: "extra2", RequestTypeField: "*default", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "ivo", SubjectField: "ivo", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
	}}
	if err := rds.SetDerivedChargers(keyCharger1, charger1, utils.NonTransactional); err != nil {
		t.Error("Error on setting DerivedChargers", err.Error())
	}
	// Retrieve from db
	if rcvCharger, err := rds.GetDerivedChargers(keyCharger1, true, utils.NonTransactional); err != nil {
		t.Error("Error when retrieving DerivedCHarger", err.Error())
	} else if !reflect.DeepEqual(rcvCharger, charger1) {
		t.Errorf("Expecting %v, received: %v", charger1, rcvCharger)
	}
	// Retrieve from cache
	if rcvCharger, err := rds.GetDerivedChargers(keyCharger1, false, utils.NonTransactional); err != nil {
		t.Error("Error when retrieving DerivedCHarger", err.Error())
	} else if !reflect.DeepEqual(rcvCharger, charger1) {
		t.Errorf("Expecting %v, received: %v", charger1, rcvCharger)
	}
}

func TestRDSitSetReqFilterIndexes(t *testing.T) {
	if !*testIntegration {
		return
	}
	idxes := map[string]map[string]utils.StringMap{
		"Account": map[string]utils.StringMap{
			"1001": utils.StringMap{
				"RL1": true,
			},
			"1002": utils.StringMap{
				"RL1": true,
				"RL2": true,
			},
			"dan": utils.StringMap{
				"RL2": true,
			},
		},
		"Subject": map[string]utils.StringMap{
			"dan": utils.StringMap{
				"RL2": true,
			},
		},
		utils.NOT_AVAILABLE: map[string]utils.StringMap{
			utils.NOT_AVAILABLE: utils.StringMap{
				"RL4": true,
				"RL5": true,
			},
		},
	}
	if err := rds.SetReqFilterIndexes(utils.ResourceLimitsIndex, idxes); err != nil {
		t.Error(err)
	}
}

func TestRDSitGetReqFilterIndexes(t *testing.T) {
	if !*testIntegration {
		return
	}
	eIdxes := map[string]map[string]utils.StringMap{
		"Account": map[string]utils.StringMap{
			"1001": utils.StringMap{
				"RL1": true,
			},
			"1002": utils.StringMap{
				"RL1": true,
				"RL2": true,
			},
			"dan": utils.StringMap{
				"RL2": true,
			},
		},
		"Subject": map[string]utils.StringMap{
			"dan": utils.StringMap{
				"RL2": true,
			},
		},
		utils.NOT_AVAILABLE: map[string]utils.StringMap{
			utils.NOT_AVAILABLE: utils.StringMap{
				"RL4": true,
				"RL5": true,
			},
		},
	}
	if idxes, err := rds.GetReqFilterIndexes(utils.ResourceLimitsIndex); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, idxes) {
		t.Errorf("Expecting: %+v, received: %+v", eIdxes, idxes)
	}
	if _, err := rds.GetReqFilterIndexes("unknown_key"); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestRDSitMatchReqFilterIndex(t *testing.T) {
	if !*testIntegration {
		return
	}
	eMp := utils.StringMap{
		"RL1": true,
		"RL2": true,
	}
	if rcvMp, err := rds.MatchReqFilterIndex(utils.ResourceLimitsIndex, utils.ConcatenatedKey("Account", "1002")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, rcvMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, rcvMp)
	}
	if _, err := rds.MatchReqFilterIndex(utils.ResourceLimitsIndex, utils.ConcatenatedKey("NonexistentField", "1002")); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}
