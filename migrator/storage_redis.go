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

//Stats methods
//get
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

//AttributeProfile methods
//get
func (v1rs *redisMigrator) getV1AttributeProfile() (v1attrPrf *v1AttributeProfile, err error) {
	var v1attr *v1AttributeProfile
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(context.TODO(), utils.AttributeProfilePrefix)
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
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(context.TODO(), utils.ThresholdProfilePrefix)
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
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(context.TODO(), utils.ThresholdProfilePrefix)
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

//AttributeProfile methods
//get
func (v1rs *redisMigrator) getV2AttributeProfile() (v2attrPrf *v2AttributeProfile, err error) {
	var v2attr *v2AttributeProfile
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(context.TODO(), utils.AttributeProfilePrefix)
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
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(context.TODO(), utils.AttributeProfilePrefix)
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
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(context.TODO(), utils.AttributeProfilePrefix)
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

func (v1rs *redisMigrator) getV5AttributeProfile() (v6attr *v6AttributeProfile, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(context.TODO(), utils.AttributeProfilePrefix)
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
		if err := v1rs.rds.Marshaler().Unmarshal(strVal, &v6attr); err != nil {
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
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(context.TODO(), SupplierProfilePrefix)
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

func (v1rs *redisMigrator) getV1DispatcherProfile() (v1chrPrf *engine.DispatcherProfile, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.rds.GetKeysForPrefix(context.TODO(), utils.DispatcherProfilePrefix)
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
