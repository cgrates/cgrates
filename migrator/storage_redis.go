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
	"strings"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type redisMigrator struct {
	dm       *engine.DataManager
	rds      *engine.RedisStorage
	dataKeys []string
	qryIdx   *int
}

var (
	reverseAliasesPrefix = "rls_"
)

func newRedisMigrator(dm *engine.DataManager) (rM *redisMigrator) {
	return &redisMigrator{
		dm:  dm,
		rds: dm.DataDB().(*engine.RedisStorage),
	}
}

func (v1rs *redisMigrator) close() {
	v1rs.rds.Close()
}

func (v1rs *redisMigrator) DataManager() *engine.DataManager {
	return v1rs.dm
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
		var strVal []byte
		if err := v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
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
	if err = v1rs.rds.Cmd(nil, "SET", key, string(bit)); err != nil {
		return err
	}
	return
}

//rem
func (v1rs *redisMigrator) remV1Account(id string) (err error) {
	key := v1AccountDBPrefix + id
	return v1rs.rds.Cmd(nil, "DEL", key)
}

//V2
//get
func (v1rs *redisMigrator) getv2Account() (v2Acnt *v2Account, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.AccountPrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err := v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
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
	key := utils.AccountPrefix + x.ID
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd(nil, "SET", key, string(bit)); err != nil {
		return err
	}
	return
}

//rem
func (v1rs *redisMigrator) remV2Account(id string) (err error) {
	key := utils.AccountPrefix + id
	return v1rs.rds.Cmd(nil, "DEL", key)
}

//ActionPlans methods
//get
func (v1rs *redisMigrator) getV1ActionPlans() (v1aps *v1ActionPlans, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.ActionPlanPrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err := v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
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
	key := utils.ActionPlanPrefix + (*x)[0].Id
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd(nil, "SET", key, string(bit)); err != nil {
		return err
	}
	return
}

//rem
func (v1rs *redisMigrator) remV1ActionPlans(x *v1ActionPlans) (err error) {
	key := utils.ActionPlanPrefix + (*x)[0].Id
	return v1rs.rds.Cmd(nil, "DEL", key)
}

//Actions methods
//get
func (v1rs *redisMigrator) getV1Actions() (v1acs *v1Actions, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.ActionPrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
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
	key := utils.ActionPrefix + (*x)[0].Id
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd(nil, "SET", key, string(bit)); err != nil {
		return err
	}
	return
}

//rem
func (v1rs *redisMigrator) remV1Actions(x v1Actions) (err error) {
	key := utils.ActionPrefix + x[0].Id
	return v1rs.rds.Cmd(nil, "DEL", key)

}

//ActionTriggers methods
//get
func (v1rs *redisMigrator) getV1ActionTriggers() (v1acts *v1ActionTriggers, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.ActionTriggerPrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
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
	key := utils.ActionTriggerPrefix + (*x)[0].Id
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd(nil, "SET", key, string(bit)); err != nil {
		return err
	}
	return
}

//rem
func (v1rs *redisMigrator) remV1ActionTriggers(x *v1ActionTriggers) (err error) {
	key := utils.ActionTriggerPrefix + (*x)[0].Id
	return v1rs.rds.Cmd(nil, "DEL", key)
}

//SharedGroup methods
//get
func (v1rs *redisMigrator) getV1SharedGroup() (v1sg *v1SharedGroup, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.SharedGroupPrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
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
	key := utils.SharedGroupPrefix + x.Id
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd(nil, "SET", key, string(bit)); err != nil {
		return err
	}
	return
}

//Stats methods
//get
func (v1rs *redisMigrator) getV1Stats() (v1st *v1Stat, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.CDRsStatsPrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
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

func (v1rs *redisMigrator) getV3Stats() (v1st *engine.StatQueueProfile, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.StatQueueProfilePrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
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
	key := utils.CDRsStatsPrefix + x.Id
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd(nil, "SET", key, string(bit)); err != nil {
		return err
	}
	return
}

//get
func (v1rs *redisMigrator) getV2Stats() (v2 *engine.StatQueue, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.StatQueuePrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
			return nil, err
		}
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, &v2); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return v2, nil
}

//set
func (v1rs *redisMigrator) setV2Stats(v2 *engine.StatQueue) (err error) {
	key := utils.StatQueuePrefix + v2.ID
	bit, err := v1rs.rds.Marshaler().Marshal(v2)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd(nil, "SET", key, string(bit)); err != nil {
		return err
	}
	return
}

//Action  methods
//get
func (v1rs *redisMigrator) getV2ActionTrigger() (v2at *v2ActionTrigger, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.ActionTriggerPrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
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
	key := utils.ActionTriggerPrefix + x.ID
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd(nil, "SET", key, string(bit)); err != nil {
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
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
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
	if err = v1rs.rds.Cmd(nil, "SET", key, string(bit)); err != nil {
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
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
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

func (v1rs *redisMigrator) getV3ThresholdProfile() (v2T *engine.ThresholdProfile, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.ThresholdProfilePrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
			return nil, err
		}
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, &v2T); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return v2T, nil
}

//set
func (v1rs *redisMigrator) setV2ThresholdProfile(x *v2Threshold) (err error) {
	key := utils.ThresholdProfilePrefix + utils.ConcatenatedKey(x.Tenant, x.ID)
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd(nil, "SET", key, string(bit)); err != nil {
		return err
	}
	return
}

//rem
func (v1rs *redisMigrator) remV2ThresholdProfile(tenant, id string) (err error) {
	key := utils.ThresholdProfilePrefix + utils.ConcatenatedKey(tenant, id)
	return v1rs.rds.Cmd(nil, "DEL", key)
}

//ThresholdProfile methods
//get
func (v1rs *redisMigrator) getV1Alias() (v1a *v1Alias, err error) {
	v1a = &v1Alias{Values: make(v1AliasValues, 0)}
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(AliasesPrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		key := v1rs.dataKeys[*v1rs.qryIdx]
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", key); err != nil {
			return nil, err
		}
		v1a.SetId(strings.TrimPrefix(key, AliasesPrefix))
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, &v1a.Values); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return v1a, nil
}

//set
func (v1rs *redisMigrator) setV1Alias(al *v1Alias) (err error) {
	var result []byte
	result, err = v1rs.rds.Marshaler().Marshal(al.Values)
	if err != nil {
		return
	}
	key := AliasesPrefix + al.GetId()
	if err = v1rs.rds.Cmd(nil, "SET", key, string(result)); err != nil {
		return
	}
	return
}

//rem
func (v1rs *redisMigrator) remV1Alias(key string) (err error) {

	// get alias for values list

	var values []byte
	if err = v1rs.rds.Cmd(&values, "GET",
		AliasesPrefix+key); err != nil {
		return
	} else if len(values) == 0 {
		err = utils.ErrNotFound
		return
	}
	al := &v1Alias{Values: make(v1AliasValues, 0)}
	al.SetId(key)
	if err = v1rs.rds.Marshaler().Unmarshal(values, &al.Values); err != nil {
		return err
	}

	err = v1rs.rds.Cmd(nil, "DEL", AliasesPrefix+key)
	if err != nil {
		return err
	}
	for _, value := range al.Values {
		tmpKey := utils.ConcatenatedKey(al.GetId(), value.DestinationId)
		for target, pairs := range value.Pairs {
			for _, alias := range pairs {
				revID := alias + target + al.Context
				err = v1rs.rds.Cmd(nil, "SREM", reverseAliasesPrefix+revID, tmpKey)
				if err != nil {
					return err
				}
			}
		}
	}
	return

	return v1rs.rds.Cmd(nil, "DEL", key)
}

// User methods
//get
func (v1rs *redisMigrator) getV1User() (v1u *v1UserProfile, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.UsersPrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
			return nil, err
		}
		v1u = new(v1UserProfile)
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, v1u); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
		return v1u, nil
	}
	v1rs.qryIdx = nil
	return nil, utils.ErrNoMoreData
}

//set
func (v1rs *redisMigrator) setV1User(us *v1UserProfile) (err error) {
	bit, err := v1rs.rds.Marshaler().Marshal(us)
	if err != nil {
		return err
	}
	return v1rs.rds.Cmd(nil, "SET", utils.UsersPrefix+us.GetId(), string(bit))
}

//rem
func (v1rs *redisMigrator) remV1User(key string) (err error) {
	return v1rs.rds.Cmd(nil, "DEL", utils.UsersPrefix+key)
}

// DerivedChargers methods
//get
func (v1rs *redisMigrator) getV1DerivedChargers() (v1d *v1DerivedChargersWithKey, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.DerivedChargersPrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
			return nil, err
		}
		v1d = new(v1DerivedChargersWithKey)
		v1d.Key = strings.TrimPrefix(v1rs.dataKeys[*v1rs.qryIdx], utils.DerivedChargersPrefix)
		v1d.Value = new(v1DerivedChargers)
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, v1d.Value); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
		return v1d, nil
	}
	v1rs.qryIdx = nil
	return nil, utils.ErrNoMoreData
}

//set
func (v1rs *redisMigrator) setV1DerivedChargers(dc *v1DerivedChargersWithKey) (err error) {
	if dc == nil || len(dc.Value.Chargers) == 0 {
		return v1rs.remV1DerivedChargers(dc.Key)
	}
	bit, err := v1rs.rds.Marshaler().Marshal(dc.Value)
	if err != nil {
		return err
	}
	return v1rs.rds.Cmd(nil, "SET", utils.DerivedChargersPrefix+dc.Key, string(bit))
}

//rem
func (v1rs *redisMigrator) remV1DerivedChargers(key string) (err error) {
	return v1rs.rds.Cmd(nil, "DEL", utils.DerivedChargersPrefix+key)
}

//AttributeProfile methods
//get
func (v1rs *redisMigrator) getV2AttributeProfile() (v2attrPrf *v2AttributeProfile, err error) {
	var v2attr *v2AttributeProfile
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.AttributeProfilePrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
			return nil, err
		}
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, &v2attr); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return v2attr, nil
}

//set
func (v1rs *redisMigrator) setV2AttributeProfile(x *v2AttributeProfile) (err error) {
	key := utils.AttributeProfilePrefix + utils.ConcatenatedKey(x.Tenant, x.ID)
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd(nil, "SET", key, string(bit)); err != nil {
		return err
	}
	return
}

//rem
func (v1rs *redisMigrator) remV2AttributeProfile(tenant, id string) (err error) {
	key := utils.AttributeProfilePrefix + utils.ConcatenatedKey(tenant, id)
	return v1rs.rds.Cmd(nil, "DEL", key)
}

//AttributeProfile methods
//get
func (v1rs *redisMigrator) getV3AttributeProfile() (v3attrPrf *v3AttributeProfile, err error) {
	var v3attr *v3AttributeProfile
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.AttributeProfilePrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
			return nil, err
		}
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, &v3attr); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return v3attr, nil
}

//set
func (v1rs *redisMigrator) setV3AttributeProfile(x *v3AttributeProfile) (err error) {
	key := utils.AttributeProfilePrefix + utils.ConcatenatedKey(x.Tenant, x.ID)
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd(nil, "SET", key, string(bit)); err != nil {
		return err
	}
	return
}

//rem
func (v1rs *redisMigrator) remV3AttributeProfile(tenant, id string) (err error) {
	key := utils.AttributeProfilePrefix + utils.ConcatenatedKey(tenant, id)
	return v1rs.rds.Cmd(nil, "DEL", key)
}

//AttributeProfile methods
//get
func (v1rs *redisMigrator) getV4AttributeProfile() (v3attrPrf *v4AttributeProfile, err error) {
	var v4attr *v4AttributeProfile
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.AttributeProfilePrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
			return nil, err
		}
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, &v4attr); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return v4attr, nil
}

func (v1rs *redisMigrator) getV5AttributeProfile() (v5attr *engine.AttributeProfile, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.AttributeProfilePrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
			return nil, err
		}
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, &v5attr); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return
}

//set
func (v1rs *redisMigrator) setV4AttributeProfile(x *v4AttributeProfile) (err error) {
	key := utils.AttributeProfilePrefix + utils.ConcatenatedKey(x.Tenant, x.ID)
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd(nil, "SET", key, string(bit)); err != nil {
		return err
	}
	return
}

//rem
func (v1rs *redisMigrator) remV4AttributeProfile(tenant, id string) (err error) {
	key := utils.AttributeProfilePrefix + utils.ConcatenatedKey(tenant, id)
	return v1rs.rds.Cmd(nil, "DEL", key)
}

// Filter Methods
//get
func (v1rs *redisMigrator) getV1Filter() (v1Fltr *v1Filter, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.FilterPrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
			return nil, err
		}
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, &v1Fltr); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return
}

func (v1rs *redisMigrator) getV4Filter() (v4Fltr *engine.Filter, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.FilterPrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
			return nil, err
		}
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, &v4Fltr); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return
}

//set
func (v1rs *redisMigrator) setV1Filter(x *v1Filter) (err error) {
	key := utils.FilterPrefix + utils.ConcatenatedKey(x.Tenant, x.ID)
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	return v1rs.rds.Cmd(nil, "SET", key, string(bit))
}

//rem
func (v1rs *redisMigrator) remV1Filter(tenant, id string) (err error) {
	key := utils.FilterPrefix + utils.ConcatenatedKey(tenant, id)
	return v1rs.rds.Cmd(nil, "DEL", key)
}

// SupplierMethods
func (v1rs *redisMigrator) getSupplier() (spl *SupplierProfile, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(SupplierProfilePrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
			return nil, err
		}
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, &spl); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return
}

//set
func (v1rs *redisMigrator) setSupplier(spl *SupplierProfile) (err error) {
	key := SupplierProfilePrefix + utils.ConcatenatedKey(spl.Tenant, spl.ID)
	bit, err := v1rs.rds.Marshaler().Marshal(spl)
	if err != nil {
		return err
	}
	if err = v1rs.rds.Cmd(nil, "SET", key, string(bit)); err != nil {
		return err
	}
	return
}

//rem
func (v1rs *redisMigrator) remSupplier(tenant, id string) (err error) {
	key := SupplierProfilePrefix + utils.ConcatenatedKey(tenant, id)
	return v1rs.rds.Cmd(nil, "DEL", key)
}

func (v1rs *redisMigrator) getV1ChargerProfile() (v1chrPrf *engine.ChargerProfile, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.ChargerProfilePrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
			return nil, err
		}
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, &v1chrPrf); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return
}

func (v1rs *redisMigrator) getV1DispatcherProfile() (v1chrPrf *engine.DispatcherProfile, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.DispatcherProfilePrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
			return nil, err
		}
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, &v1chrPrf); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return
}

func (v1rs *redisMigrator) getV1RouteProfile() (v1chrPrf *engine.RouteProfile, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(utils.RouteProfilePrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNoMoreData
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		var strVal []byte
		if err = v1rs.rds.Cmd(&strVal, "GET", v1rs.dataKeys[*v1rs.qryIdx]); err != nil {
			return nil, err
		}
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, &v1chrPrf); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return
}
