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
	"github.com/cgrates/mgo"
)

type v1Mongo struct {
	session *mgo.Session
	db      string
	v1ms    engine.Marshaler
	qryIter *mgo.Iter
}

type AcKeyValue struct {
	Key   string
	Value v1Actions
}
type AtKeyValue struct {
	Key   string
	Value v1ActionPlans
}

func newv1MongoStorage(host, port, db, user, pass, storageType string, cdrsIndexes []string) (v1ms *v1Mongo, err error) {
	url := host
	if port != "" {
		url += ":" + port
	}
	if user != "" && pass != "" {
		url = fmt.Sprintf("%s:%s@%s", user, pass, url)
	}
	if db != "" {
		url += "/" + db
	}
	session, err := mgo.Dial(url)
	if err != nil {
		return nil, err
	}
	session.SetMode(mgo.Strong, true)
	v1ms = &v1Mongo{db: db, session: session, v1ms: engine.NewCodecMsgpackMarshaler()}
	return
}
func (v1ms *v1Mongo) Close() {}
func (v1ms *v1Mongo) getKeysForPrefix(prefix string) ([]string, error) {
	return nil, nil
}

//Account methods
//V1
//get
func (v1ms *v1Mongo) getv1Account() (v1Acnt *v1Account, err error) {
	if v1ms.qryIter == nil {
		v1ms.qryIter = v1ms.session.DB(v1ms.db).C(v1AccountDBPrefix).Find(nil).Iter()
	}
	v1ms.qryIter.Next(&v1Acnt)

	if v1Acnt == nil {
		v1ms.qryIter = nil
		return nil, utils.ErrNoMoreData

	}
	return v1Acnt, nil
}

//set
func (v1ms *v1Mongo) setV1Account(x *v1Account) (err error) {
	if err := v1ms.session.DB(v1ms.db).C(v1AccountDBPrefix).Insert(x); err != nil {
		return err
	}
	return
}

//V2
//get
func (v1ms *v1Mongo) getv2Account() (v2Acnt *v2Account, err error) {
	if v1ms.qryIter == nil {
		v1ms.qryIter = v1ms.session.DB(v1ms.db).C(utils.ACCOUNT_PREFIX).Find(nil).Iter()
	}
	v1ms.qryIter.Next(&v2Acnt)

	if v2Acnt == nil {
		v1ms.qryIter = nil
		return nil, utils.ErrNoMoreData

	}
	return v2Acnt, nil
}

//set
func (v1ms *v1Mongo) setV2Account(x *v2Account) (err error) {
	if err := v1ms.session.DB(v1ms.db).C(utils.ACCOUNT_PREFIX).Insert(x); err != nil {
		return err
	}
	return
}

//Action methods
//get
func (v1ms *v1Mongo) getV1ActionPlans() (v1aps *v1ActionPlans, err error) {
	var strct *AtKeyValue
	if v1ms.qryIter == nil {
		v1ms.qryIter = v1ms.session.DB(v1ms.db).C("actiontimings").Find(nil).Iter()
	}
	v1ms.qryIter.Next(&strct)
	if strct == nil {
		v1ms.qryIter = nil
		return nil, utils.ErrNoMoreData
	}
	v1aps = &strct.Value
	return v1aps, nil
}

//set
func (v1ms *v1Mongo) setV1ActionPlans(x *v1ActionPlans) (err error) {
	key := utils.ACTION_PLAN_PREFIX + (*x)[0].Id
	if err := v1ms.session.DB(v1ms.db).C("actiontimings").Insert(&AtKeyValue{key, *x}); err != nil {
		return err
	}
	return
}

//Actions methods
//get
func (v1ms *v1Mongo) getV1Actions() (v1acs *v1Actions, err error) {
	var strct *AcKeyValue
	if v1ms.qryIter == nil {
		v1ms.qryIter = v1ms.session.DB(v1ms.db).C("actions").Find(nil).Iter()
	}
	v1ms.qryIter.Next(&strct)
	if strct == nil {
		v1ms.qryIter = nil
		return nil, utils.ErrNoMoreData
	}

	v1acs = &strct.Value
	return v1acs, nil
}

//set
func (v1ms *v1Mongo) setV1Actions(x *v1Actions) (err error) {
	key := utils.ACTION_PREFIX + (*x)[0].Id
	if err := v1ms.session.DB(v1ms.db).C("actions").Insert(&AcKeyValue{key, *x}); err != nil {
		return err
	}
	return
}

//ActionTriggers methods
//get
func (v1ms *v1Mongo) getV1ActionTriggers() (v1acts *v1ActionTriggers, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (v1ms *v1Mongo) setV1ActionTriggers(x *v1ActionTriggers) (err error) {
	return utils.ErrNotImplemented
}

//Actions methods
//get
func (v1ms *v1Mongo) getV1SharedGroup() (v1sg *v1SharedGroup, err error) {
	if v1ms.qryIter == nil {
		v1ms.qryIter = v1ms.session.DB(v1ms.db).C(utils.SHARED_GROUP_PREFIX).Find(nil).Iter()
	}
	v1ms.qryIter.Next(&v1sg)
	if v1sg == nil {
		v1ms.qryIter = nil
		return nil, utils.ErrNoMoreData

	}
	return v1sg, nil
}

//set
func (v1ms *v1Mongo) setV1SharedGroup(x *v1SharedGroup) (err error) {
	if err := v1ms.session.DB(v1ms.db).C(utils.SHARED_GROUP_PREFIX).Insert(x); err != nil {
		return err
	}
	return
}

//Stats methods
//get
func (v1ms *v1Mongo) getV1Stats() (v1st *v1Stat, err error) {
	if v1ms.qryIter == nil {
		v1ms.qryIter = v1ms.session.DB(v1ms.db).C(utils.CDR_STATS_PREFIX).Find(nil).Iter()
	}
	v1ms.qryIter.Next(&v1st)
	if v1st == nil {
		v1ms.qryIter = nil
		return nil, utils.ErrNoMoreData

	}
	return v1st, nil
}

//set
func (v1ms *v1Mongo) setV1Stats(x *v1Stat) (err error) {
	if err := v1ms.session.DB(v1ms.db).C(utils.CDR_STATS_PREFIX).Insert(x); err != nil {
		return err
	}
	return
}

//Stats methods
//get
func (v1ms *v1Mongo) getV2ActionTrigger() (v2at *v2ActionTrigger, err error) {
	if v1ms.qryIter == nil {
		v1ms.qryIter = v1ms.session.DB(v1ms.db).C(utils.ACTION_TRIGGER_PREFIX).Find(nil).Iter()
	}
	v1ms.qryIter.Next(&v2at)
	if v2at == nil {
		v1ms.qryIter = nil
		return nil, utils.ErrNoMoreData

	}
	return v2at, nil
}

//set
func (v1ms *v1Mongo) setV2ActionTrigger(x *v2ActionTrigger) (err error) {
	if err := v1ms.session.DB(v1ms.db).C(utils.ACTION_TRIGGER_PREFIX).Insert(x); err != nil {
		return err
	}
	return
}

//AttributeProfile methods
//get
func (v1ms *v1Mongo) getV1AttributeProfile() (v1attrPrf *v1AttributeProfile, err error) {
	if v1ms.qryIter == nil {
		v1ms.qryIter = v1ms.session.DB(v1ms.db).C(utils.AttributeProfilePrefix).Find(nil).Iter()
	}
	v1ms.qryIter.Next(&v1attrPrf)
	if v1attrPrf == nil {
		v1ms.qryIter = nil
		return nil, utils.ErrNoMoreData

	}
	return v1attrPrf, nil
}

//set
func (v1ms *v1Mongo) setV1AttributeProfile(x *v1AttributeProfile) (err error) {
	if err := v1ms.session.DB(v1ms.db).C(utils.AttributeProfilePrefix).Insert(x); err != nil {
		return err
	}
	return
}
