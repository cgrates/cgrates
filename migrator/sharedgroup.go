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
	"fmt"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"gopkg.in/mgo.v2/bson"
)

type v1SharedGroup struct {
	Id                string
	AccountParameters map[string]*engine.SharingParameters
	MemberIds         []string
}

func (m *Migrator) migrateSharedGroups() (err error) {
	switch m.dataDBType {
	case utils.REDIS:
		var sgv1keys []string
		sgv1keys, err = m.dataDB.GetKeysForPrefix(utils.SHARED_GROUP_PREFIX)
		if err != nil {
			return
		}
		for _, sgv1key := range sgv1keys {
			v1sg, err := m.getv1SharedGroupFromDB(sgv1key)
			if err != nil {
				return err
			}
			sg := v1sg.AsSharedGroup()
			if err = m.dataDB.SetSharedGroup(sg, utils.NonTransactional); err != nil {
				return err
			}
		}
		// All done, update version wtih current one
		vrs := engine.Versions{utils.SHARED_GROUP_PREFIX: engine.CurrentStorDBVersions()[utils.SHARED_GROUP_PREFIX]}
		if err = m.dataDB.SetVersions(vrs, false); err != nil {
			return utils.NewCGRError(utils.Migrator,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error: <%s> when updating SharedGroup version into dataDB", err.Error()))
		}
		return
	case utils.MONGO:
		dataDB := m.dataDB.(*engine.MongoStorage)
		mgoDB := dataDB.DB()
		defer mgoDB.Session.Close()
		var v1sg v1SharedGroup
		iter := mgoDB.C(utils.SHARED_GROUP_PREFIX).Find(nil).Iter()
		for iter.Next(&v1sg) {
			sg := v1sg.AsSharedGroup()
			if err = m.dataDB.SetSharedGroup(sg, utils.NonTransactional); err != nil {
				return err
			}
		}
		// All done, update version wtih current one
		vrs := engine.Versions{utils.SHARED_GROUP_PREFIX: engine.CurrentStorDBVersions()[utils.SHARED_GROUP_PREFIX]}
		if err = m.dataDB.SetVersions(vrs, false); err != nil {
			return utils.NewCGRError(utils.Migrator,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error: <%s> when updating SharedGroup version into dataDB", err.Error()))
		}
		return
	default:
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			utils.UnsupportedDB,
			fmt.Sprintf("error: unsupported: <%s> for migrateSharedGroups method", m.dataDBType))
	}
}

func (m *Migrator) getv1SharedGroupFromDB(key string) (*v1SharedGroup, error) {
	switch m.dataDBType {
	case utils.REDIS:
		dataDB := m.dataDB.(*engine.RedisStorage)
		if strVal, err := dataDB.Cmd("GET", key).Bytes(); err != nil {
			return nil, err
		} else {
			v1SG := &v1SharedGroup{Id: key}
			if err := m.mrshlr.Unmarshal(strVal, v1SG); err != nil {
				return nil, err
			}
			return v1SG, nil
		}
	case utils.MONGO:
		dataDB := m.dataDB.(*engine.MongoStorage)
		mgoDB := dataDB.DB()
		defer mgoDB.Session.Close()
		v1SG := new(v1SharedGroup)
		if err := mgoDB.C(utils.SHARED_GROUP_PREFIX).Find(bson.M{"id": key}).One(v1SG); err != nil {
			return nil, err
		}
		return v1SG, nil
	default:
		return nil, utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			utils.UnsupportedDB,
			fmt.Sprintf("error: unsupported: <%s> for getv1SharedGroupFromDB method", m.dataDBType))
	}
}

func (v1SG v1SharedGroup) AsSharedGroup() (sg *engine.SharedGroup) {
	sg = &engine.SharedGroup{
		Id:                v1SG.Id,
		AccountParameters: v1SG.AccountParameters,
		MemberIds:         make(utils.StringMap),
	}
	for _, accID := range v1SG.MemberIds {
		sg.MemberIds[accID] = true
	}
	return
}
