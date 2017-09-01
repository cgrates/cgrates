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
	"gopkg.in/mgo.v2"
)

type  v1Mongo struct{
	session			*mgo.Session
	db 				string
	v1ms			engine.Marshaler
	qryIter 		*mgo.Iter 

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

 func (v1ms *v1Mongo) getv1Account() (v1Acnt *v1Account, err error){
 	if v1ms.qryIter==nil{
 	v1ms.qryIter = v1ms.session.DB(v1ms.db).C(v1AccountDBPrefix).Find(nil).Iter()
	}
 	v1ms.qryIter.Next(&v1Acnt) 

	if v1Acnt==nil{
		v1ms.qryIter=nil
	}
	return v1Acnt,nil
 }

func (v1ms *v1Mongo) getKeysForPrefix(prefix string) ([]string, error){
return nil,nil
}