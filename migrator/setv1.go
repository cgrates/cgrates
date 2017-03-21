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

import "github.com/cgrates/cgrates/engine"

func (m *Migrator) SetV1onRedis(key string, bl []byte) (err error) {
	dataDB := m.dataDB.(*engine.RedisStorage)
	if err = dataDB.Cmd("SET", key, bl).Err; err != nil {
		return err
	}
	return
}

func (m *Migrator) SetV1onMongoAccount(pref string, x *v1Account) (err error) {
	dataDB := m.dataDB.(*engine.MongoStorage)
	mgoDB := dataDB.DB()
	defer mgoDB.Session.Close()
	if err := mgoDB.C(pref).Insert(x); err != nil {
		return err
	}
	return
}

func (m *Migrator) SetV1onMongoAction(pref string, x *v1Action) (err error) {
	dataDB := m.dataDB.(*engine.MongoStorage)
	mgoDB := dataDB.DB()
	defer mgoDB.Session.Close()
	if err := mgoDB.C(pref).Insert(x); err != nil {
		return err
	}
	return
}

func (m *Migrator) SetV1onMongoActionPlan(pref string, x *v1ActionPlan) (err error) {
	dataDB := m.dataDB.(*engine.MongoStorage)
	mgoDB := dataDB.DB()
	defer mgoDB.Session.Close()
	if err := mgoDB.C(pref).Insert(x); err != nil {
		return err
	}
	return
}

func (m *Migrator) SetV1onMongoActionTrigger(pref string, x *v1ActionTrigger) (err error) {
	dataDB := m.dataDB.(*engine.MongoStorage)
	mgoDB := dataDB.DB()
	defer mgoDB.Session.Close()
	if err := mgoDB.C(pref).Insert(x); err != nil {
		return err
	}
	return
}

func (m *Migrator) SetV1onMongoSharedGroup(pref string, x *v1SharedGroup) (err error) {
	dataDB := m.dataDB.(*engine.MongoStorage)
	mgoDB := dataDB.DB()
	defer mgoDB.Session.Close()
	if err := mgoDB.C(pref).Insert(x); err != nil {
		return err
	}
	return
}
func (m *Migrator) DropV1Colection(pref string) (err error) {
	dataDB := m.dataDB.(*engine.MongoStorage)
	mgoDB := dataDB.DB()
	defer mgoDB.Session.Close()
	if err := mgoDB.C(pref).DropCollection(); err != nil {
		return err
	}
	return
}
