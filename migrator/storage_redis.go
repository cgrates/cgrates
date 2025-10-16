/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package migrator

import (
	"github.com/cgrates/birpc/context"
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
	var rdsDB *engine.RedisStorage
	for _, dbInf := range dm.DataDB() {
		var canCast bool
		if rdsDB, canCast = dbInf.(*engine.RedisStorage); canCast {
			return &redisMigrator{
				dm:  dm,
				rds: rdsDB,
			}
		}
	}
	return nil
}

func (v1rs *redisMigrator) close() {
	v1rs.rds.Close()
}

func (v1rs *redisMigrator) DataManager() *engine.DataManager {
	return v1rs.dm
}

// Stats methods
// get
func (v1rs *redisMigrator) getV1Stats() (v1st *v1Stat, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(context.TODO(), utils.CDRsStatsPrefix)
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
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(context.TODO(), utils.StatQueueProfilePrefix)
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

// set
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

// get
func (v1rs *redisMigrator) getV2Stats() (v2 *engine.StatQueue, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(context.TODO(), utils.StatQueuePrefix)
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

// set
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

// Filter Methods
// get
func (v1rs *redisMigrator) getV1Filter() (v1Fltr *v1Filter, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(context.TODO(), utils.FilterPrefix)
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
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(context.TODO(), utils.FilterPrefix)
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

// set
func (v1rs *redisMigrator) setV1Filter(x *v1Filter) (err error) {
	key := utils.FilterPrefix + utils.ConcatenatedKey(x.Tenant, x.ID)
	bit, err := v1rs.rds.Marshaler().Marshal(x)
	if err != nil {
		return err
	}
	return v1rs.rds.Cmd(nil, "SET", key, string(bit))
}

// rem
func (v1rs *redisMigrator) remV1Filter(tenant, id string) (err error) {
	key := utils.FilterPrefix + utils.ConcatenatedKey(tenant, id)
	return v1rs.rds.Cmd(nil, "DEL", key)
}

func (v1rs *redisMigrator) getV1ChargerProfile() (v1chrPrf *utils.ChargerProfile, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(context.TODO(), utils.ChargerProfilePrefix)
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

func (v1rs *redisMigrator) getV1RouteProfile() (v1chrPrf *utils.RouteProfile, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(context.TODO(), utils.RouteProfilePrefix)
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
