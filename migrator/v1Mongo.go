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
	"log"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/engine"
	"gopkg.in/mgo.v2"
//	"gopkg.in/mgo.v2/bson"
)

type  v1Mongo struct{
	session			*mgo.Session
	db 				string
	v1ms			engine.Marshaler
	qryIter 		*mgo.Iter 

}

type AcKeyValue struct {
	Key   string
	Value v1Actions
}
type AtKeyValue struct {
	Key   string
	Value v1ActionPlans
}

func NewMongoStorage(host, port, db, user, pass, storageType string, cdrsIndexes []string) (v1ms *v1Mongo, err error) {
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

	func (v1ms *v1Mongo) getKeysForPrefix(prefix string) ([]string, error){
return nil,nil
}

//Account methods
//get
 func (v1ms *v1Mongo) getv1Account() (v1Acnt *v1Account, err error){
 	if v1ms.qryIter==nil{
 	v1ms.qryIter = v1ms.session.DB(v1ms.db).C(v1AccountDBPrefix).Find(nil).Iter()
	}
 	v1ms.qryIter.Next(&v1Acnt) 

	if v1Acnt==nil{
		v1ms.qryIter=nil
			return nil,utils.ErrNoMoreData

	}
	return v1Acnt,nil
 }

//set
func (v1ms *v1Mongo) setV1Account( x *v1Account) (err error) {
	if err := v1ms.session.DB(v1ms.db).C(v1AccountDBPrefix).Insert(x); err != nil {
		return err
	}
	return
}

//Action methods
//get
func (v1ms *v1Mongo) getV1ActionPlans() (v1aps *v1ActionPlans, err error){
var strct *AtKeyValue
	if v1ms.qryIter==nil{
 	v1ms.qryIter = v1ms.session.DB(v1ms.db).C("actiontimings").Find(nil).Iter()
	}
 	v1ms.qryIter.Next(&strct) 
		log.Print("Done migrating!",strct)

	if strct==nil{
		v1ms.qryIter=nil
			return nil,utils.ErrNoMoreData
	}

	v1aps=&strct.Value
	return v1aps,nil
}

//set
func (v1ms *v1Mongo) setV1Actions(x *v1ActionPlans) (err error) {
	key:=utils.ACTION_PLAN_PREFIX + (*x)[0].Id
		log.Print("Done migrating!",(*x)[0])

	if err := v1ms.session.DB(v1ms.db).C("actiontimings").Insert(&AtKeyValue{key, *x}); err != nil {
		return err
	}
	return
}

//Actions methods
//get
func (v1ms *v1Mongo) getV1ActionPlans() (v1aps *v1ActionPlans, err error){
var strct *AtKeyValue
	if v1ms.qryIter==nil{
 	v1ms.qryIter = v1ms.session.DB(v1ms.db).C("actiontimings").Find(nil).Iter()
	}
 	v1ms.qryIter.Next(&strct) 
		log.Print("Done migrating!",strct)

	if strct==nil{
		v1ms.qryIter=nil
			return nil,utils.ErrNoMoreData
	}

	v1aps=&strct.Value
	return v1aps,nil
}

func (v1ms *v1Mongo) setV1onMongoAction(key string, x *v1Actions) (err error) {
	if err := v1ms.session.DB(v1ms.db).C("actions").Insert(&AcKeyValue{key, *x}); err != nil {
		return err
	}
	return
}

// func (v1ms *v1Mongo) setV1onMongoActionTrigger(pref string, x *v1ActionTriggers) (err error) {
// 	if err := v1ms.session.DB(v1ms.db).C(pref).Insert(x); err != nil {
// 		return err
// 	}
// 	return
// }

// func (v1ms *v1Mongo) setV1onMongoSharedGroup(pref string, x *v1SharedGroup) (err error) {
// 	if err := v1ms.session.DB(v1ms.db).C(pref).Insert(x); err != nil {
// 		return err
// 	}
// 	return
// }
// func (v1ms *v1Mongo) DropV1Colection(pref string) (err error) {
// 	if err := v1ms.session.DB(v1ms.db).C(pref).DropCollection(); err != nil {
// 		return err
// 	}
// 	return
// }