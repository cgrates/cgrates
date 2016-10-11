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
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var mgoITDB *MongoStorage

func TestMGOitConnect(t *testing.T) {
	if !*testIntegration {
		return
	}
	var err error
	cdrsMongoCfgPath := path.Join(*dataDir, "conf", "samples", "cdrsv2mongo")
	mgoITCfg, err := config.NewCGRConfigFromFolder(cdrsMongoCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if mgoITDB, err = NewMongoStorage(mgoITCfg.StorDBHost, mgoITCfg.StorDBPort, mgoITCfg.StorDBName, mgoITCfg.StorDBUser, mgoITCfg.StorDBPass, nil, mgoITCfg.CacheConfig, mgoITCfg.LoadHistorySize); err != nil {
		t.Fatal(err)
	}
}

func TestMGOitFlush(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := mgoITDB.Flush(""); err != nil {
		t.Error(err)
	}
}

func TestMGOitSetReqFilterIndexes(t *testing.T) {
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
	if err := mgoITDB.SetReqFilterIndexes(utils.ResourceLimitsIndex, idxes); err != nil {
		t.Error(err)
	}
}

func TestMGOitGetReqFilterIndexes(t *testing.T) {
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
	if idxes, err := mgoITDB.GetReqFilterIndexes(utils.ResourceLimitsIndex); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, idxes) {
		t.Errorf("Expecting: %+v, received: %+v", eIdxes, idxes)
	}
	if _, err := mgoITDB.GetReqFilterIndexes("unknown_key"); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestMGOitMatchReqFilterIndex(t *testing.T) {
	if !*testIntegration {
		return
	}
	eMp := utils.StringMap{
		"RL1": true,
		"RL2": true,
	}
	if rcvMp, err := mgoITDB.MatchReqFilterIndex(utils.ResourceLimitsIndex, utils.ConcatenatedKey("Account", "1002")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, rcvMp) {
		t.Errorf("Expecting: %+v, received: %+v", eMp, rcvMp)
	}
	if _, err := mgoITDB.MatchReqFilterIndex(utils.ResourceLimitsIndex, utils.ConcatenatedKey("NonexistentField", "1002")); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}
