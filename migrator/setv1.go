// /*
// Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
// Copyright (C) ITsysCOM GmbH

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>
// */
package migrator

// import (
// 	"fmt"

// 	"github.com/cgrates/cgrates/engine"
// 	"github.com/cgrates/cgrates/utils"
// 	"gopkg.in/mgo.v2/bson"
// )

// type AcKeyValue struct {
// 	Key   string
// 	Value v1Actions
// }
// type AtKeyValue struct {
// 	Key   string
// 	Value v1ActionPlans
// }

// func (m *Migrator) SetV1onOldRedis(key string, bl []byte) (err error) {
// 	dataDB := m.oldDataDB.(*engine.RedisStorage)
// 	if err = dataDB.Cmd("SET", key, bl).Err; err != nil {
// 		return err
// 	}
// 	return
// }

// func (m *Migrator) SetV1onRedis(key string, bl []byte) (err error) {
// 	dataDB := m.dataDB.(*engine.RedisStorage)
// 	if err = dataDB.Cmd("SET", key, bl).Err; err != nil {
// 		return err
// 	}
// 	return
// }

// func (m *Migrator) SetV1onMongoAccount( x *v1Account) (err error) {
// 	if err := m.oldDataDB.session.DB().C("userbalances").Insert(x); err != nil {
// 		return err
// 	}
// 	return
// }

// func (m *Migrator) SetV1onMongoAction(key string, x *v1Actions) (err error) {
// 	if err := m.oldDataDB.session.DB().C("actions").Insert(&AcKeyValue{key, *x}); err != nil {
// 		return err
// 	}
// 	return
// }

// func (m *Migrator) SetV1onMongoActionPlan(key string, x *v1ActionPlans) (err error) {
// 	if err := m.oldDataDB.session.DB().C("actiontimings").Insert(&AtKeyValue{key, *x}); err != nil {
// 		return err
// 	}
// 	return
// }

// func (m *Migrator) SetV1onMongoActionTrigger(pref string, x *v1ActionTriggers) (err error) {
// 	if err := m.oldDataDB.session.DB().C(pref).Insert(x); err != nil {
// 		return err
// 	}
// 	return
// }

// func (m *Migrator) SetV1onMongoSharedGroup(pref string, x *v1SharedGroup) (err error) {
// 	if err := m.oldDataDB.session.DB().C(pref).Insert(x); err != nil {
// 		return err
// 	}
// 	return
// }
// func (m *Migrator) DropV1Colection(pref string) (err error) {
// 	if err := m.oldDataDB.session.DB().C(pref).DropCollection(); err != nil {
// 		return err
// 	}
// 	return
// }

// func (m *Migrator) getV1AccountFromDB(key string) (*v1Account, error) {
// 	switch m.oldDataDBType {
// 	case utils.REDIS:
// 		dataDB := m.oldDataDB.(*engine.RedisStorage)
// 		if strVal, err := dataDB.Cmd("GET", key).Bytes(); err != nil {
// 			return nil, err
// 		} else {
// 			v1Acnt := &v1Account{Id: key}
// 			if err := m.mrshlr.Unmarshal(strVal, v1Acnt); err != nil {
// 				return nil, err
// 			}
// 			return v1Acnt, nil
// 		}
// 	case utils.MONGO:
// 		dataDB := m.oldDataDB.(*engine.MongoStorage)
// 		mgoDB := dataDB.DB()
// 		defer mgoDB.Session.Close()
// 		v1Acnt := new(v1Account)
// 		if err := mgoDB.C(v1AccountTBL).Find(bson.M{"id": key}).One(v1Acnt); err != nil {
// 			return nil, err
// 		}
// 		return v1Acnt, nil
// 	default:
// 		return nil, utils.NewCGRError(utils.Migrator,
// 			utils.ServerErrorCaps,
// 			utils.UnsupportedDB,
// 			fmt.Sprintf("error: unsupported: <%s> for getV1AccountFromDB method", m.oldDataDBType))
// 	}
// }
