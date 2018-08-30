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

package migrator

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type redisMigrator struct {
	dm       *engine.DataManager
	rds      *engine.RedisStorage
	dataKeys []string
	qryIdx   *int
}

func newRedisMigrator(dm *engine.DataManager) (rM *redisMigrator) {
	return &redisMigrator{
		dm:  dm,
		rds: dm.DataDB().(*engine.RedisStorage),
	}
}

func (rdsMig *redisMigrator) DataManager() *engine.DataManager {
	return rdsMig.dm
}

//Account methods
//V1
//get
func (v1rs *redisMigrator) getv1Account() (v1Acnt *v1Account, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(v1AccountDBPrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		strVal, err := v1rs.rds.Cmd("GET", v1rs.dataKeys[*v1rs.qryIdx]).Bytes()
		if err != nil {
			return nil, err
		}
		v1Acnt = &v1Account{Id: v1rs.dataKeys[*v1rs.qryIdx]}
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, v1Acnt); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return v1Acnt, nil
}

//set
func (v1rs *redisMigrator) setV1Account(x *v1Account) (err error) {
	key := v1AccountDBPrefix + x.Id
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd("SET", key, bit).Err; err != nil {
		return err
	}
	return
}

//rem
func (v1rs *redisMigrator) remV1Account(id string) (err error) {
	key := v1AccountDBPrefix + id
	return v1rs.rds.Cmd("DEL", key).Err
}

//V2
//get
func (v1rs *redisMigrator) getv2Account() (v2Acnt *v2Account, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.ACCOUNT_PREFIX)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		strVal, err := v1rs.rds.Cmd("GET", v1rs.dataKeys[*v1rs.qryIdx]).Bytes()
		if err != nil {
			return nil, err
		}
		v2Acnt = &v2Account{ID: v1rs.dataKeys[*v1rs.qryIdx]}
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, v2Acnt); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return v2Acnt, nil
}

//set
func (v1rs *redisMigrator) setV2Account(x *v2Account) (err error) {
	key := utils.ACCOUNT_PREFIX + x.ID
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd("SET", key, bit).Err; err != nil {
		return err
	}
	return
}

//rem
func (v1rs *redisMigrator) remV2Account(id string) (err error) {
	key := utils.ACCOUNT_PREFIX + id
	return v1rs.rds.Cmd("DEL", key).Err
}

//ActionPlans methods
//get
func (v1rs *redisMigrator) getV1ActionPlans() (v1aps *v1ActionPlans, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.ACTION_PLAN_PREFIX)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNotFound
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		strVal, err := v1rs.rds.Cmd("GET", v1rs.dataKeys[*v1rs.qryIdx]).Bytes()
		if err != nil {
			return nil, err
		}
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, &v1aps); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return v1aps, nil
}

//set
func (v1rs *redisMigrator) setV1ActionPlans(x *v1ActionPlans) (err error) {
	key := utils.ACTION_PLAN_PREFIX + (*x)[0].Id
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd("SET", key, bit).Err; err != nil {
		return err
	}
	return
}

//Actions methods
//get
func (v1rs *redisMigrator) getV1Actions() (v1acs *v1Actions, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.ACTION_PREFIX)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNotFound
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		strVal, err := v1rs.rds.Cmd("GET", v1rs.dataKeys[*v1rs.qryIdx]).Bytes()
		if err != nil {
			return nil, err
		}
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, &v1acs); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return v1acs, nil
}

//set
func (v1rs *redisMigrator) setV1Actions(x *v1Actions) (err error) {
	key := utils.ACTION_PREFIX + (*x)[0].Id
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd("SET", key, bit).Err; err != nil {
		return err
	}
	return
}

//ActionTriggers methods
//get
func (v1rs *redisMigrator) getV1ActionTriggers() (v1acts *v1ActionTriggers, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.ACTION_TRIGGER_PREFIX)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNotFound
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		strVal, err := v1rs.rds.Cmd("GET", v1rs.dataKeys[*v1rs.qryIdx]).Bytes()
		if err != nil {
			return nil, err
		}
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, &v1acts); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return v1acts, nil
}

//set
func (v1rs *redisMigrator) setV1ActionTriggers(x *v1ActionTriggers) (err error) {
	key := utils.ACTION_TRIGGER_PREFIX + (*x)[0].Id
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd("SET", key, bit).Err; err != nil {
		return err
	}
	return
}

//SharedGroup methods
//get
func (v1rs *redisMigrator) getV1SharedGroup() (v1sg *v1SharedGroup, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.SHARED_GROUP_PREFIX)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNotFound
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		strVal, err := v1rs.rds.Cmd("GET", v1rs.dataKeys[*v1rs.qryIdx]).Bytes()
		if err != nil {
			return nil, err
		}
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, &v1sg); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return v1sg, nil
}

//set
func (v1rs *redisMigrator) setV1SharedGroup(x *v1SharedGroup) (err error) {
	key := utils.SHARED_GROUP_PREFIX + x.Id
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd("SET", key, bit).Err; err != nil {
		return err
	}
	return
}

//Stats methods
//get
func (v1rs *redisMigrator) getV1Stats() (v1st *v1Stat, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.CDR_STATS_PREFIX)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNotFound
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		strVal, err := v1rs.rds.Cmd("GET", v1rs.dataKeys[*v1rs.qryIdx]).Bytes()
		if err != nil {
			return nil, err
		}
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, &v1st); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return v1st, nil
}

//set
func (v1rs *redisMigrator) setV1Stats(x *v1Stat) (err error) {
	key := utils.CDR_STATS_PREFIX + x.Id
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd("SET", key, bit).Err; err != nil {
		return err
	}
	return
}

//Action  methods
//get
func (v1rs *redisMigrator) getV2ActionTrigger() (v2at *v2ActionTrigger, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.ACTION_TRIGGER_PREFIX)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNotFound
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		strVal, err := v1rs.rds.Cmd("GET", v1rs.dataKeys[*v1rs.qryIdx]).Bytes()
		if err != nil {
			return nil, err
		}
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, &v2at); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return v2at, nil
}

//set
func (v1rs *redisMigrator) setV2ActionTrigger(x *v2ActionTrigger) (err error) {
	key := utils.ACTION_TRIGGER_PREFIX + x.ID
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd("SET", key, bit).Err; err != nil {
		return err
	}
	return
}

//AttributeProfile methods
//get
func (v1rs *redisMigrator) getV1AttributeProfile() (v1attrPrf *v1AttributeProfile, err error) {
	var v1attr *v1AttributeProfile
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.AttributeProfilePrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNotFound
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		strVal, err := v1rs.rds.Cmd("GET", v1rs.dataKeys[*v1rs.qryIdx]).Bytes()
		if err != nil {
			return nil, err
		}
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, &v1attr); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return v1attr, nil
}

//set
func (v1rs *redisMigrator) setV1AttributeProfile(x *v1AttributeProfile) (err error) {
	key := utils.AttributeProfilePrefix + utils.ConcatenatedKey(x.Tenant, x.ID)
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd("SET", key, bit).Err; err != nil {
		return err
	}
	return
}

//ThresholdProfile methods
//get
func (v1rs *redisMigrator) getV2ThresholdProfile() (v2T *v2Threshold, err error) {
	var v2Th *v2Threshold
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.ThresholdProfilePrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNotFound
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		strVal, err := v1rs.rds.Cmd("GET", v1rs.dataKeys[*v1rs.qryIdx]).Bytes()
		if err != nil {
			return nil, err
		}
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, &v2Th); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return v2Th, nil
}

//set
func (v1rs *redisMigrator) setV2ThresholdProfile(x *v2Threshold) (err error) {
	key := utils.ThresholdProfilePrefix + utils.ConcatenatedKey(x.Tenant, x.ID)
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd("SET", key, bit).Err; err != nil {
		return err
	}
	return
}

//rem
func (v1rs *redisMigrator) remV2ThresholdProfile(tenant, id string) (err error) {
	key := utils.ThresholdProfilePrefix + utils.ConcatenatedKey(tenant, id)
	return v1rs.rds.Cmd("DEL", key).Err
}
