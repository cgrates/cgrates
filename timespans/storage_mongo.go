/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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

package timespans

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	//"log"
)

type MongoStorage struct {
	session *mgo.Session
	db      *mgo.Database
}

func NewMongoStorage(host, port, db, user, pass string) (StorageGetter, error) {
	dial := fmt.Sprintf(host)
	if user != "" && pass != "" {
		dial = fmt.Sprintf("%s:%s@%s", user, pass, dial)
	}
	if port != "" {
		dial += ":" + port
	}
	session, err := mgo.Dial(dial)
	if err != nil {
		Logger.Err(fmt.Sprintf("Could not connect to logger database: %v", err))
		return nil, err
	}
	ndb := session.DB(db)
	session.SetMode(mgo.Monotonic, true)
	index := mgo.Index{Key: []string{"key"}, Background: true}
	err = ndb.C("activationperiods").EnsureIndex(index)
	err = ndb.C("destinations").EnsureIndex(index)
	err = ndb.C("actions").EnsureIndex(index)
	index = mgo.Index{Key: []string{"id"}, Background: true}
	err = ndb.C("userbalances").EnsureIndex(index)
	err = ndb.C("actiontimings").EnsureIndex(index)

	return &MongoStorage{db: ndb, session: session}, nil
}

func (ms *MongoStorage) Close() {
	ms.session.Close()
}

func (ms *MongoStorage) Flush() (err error) {
	err = ms.db.C("activationperiods").DropCollection()
	if err != nil {
		return
	}
	err = ms.db.C("destinations").DropCollection()
	if err != nil {
		return
	}
	err = ms.db.C("actions").DropCollection()
	if err != nil {
		return
	}
	err = ms.db.C("userbalances").DropCollection()
	if err != nil {
		return
	}
	err = ms.db.C("actiontimings").DropCollection()
	if err != nil {
		return
	}
	return nil
}

/*
Helper type for activation periods storage.
*/
type ApKeyValue struct {
	Key         string
	FallbackKey string `omitempty`
	Value       []*ActivationPeriod
}

type AcKeyValue struct {
	Key   string
	Value []*Action
}

type AtKeyValue struct {
	Key   string
	Value []*ActionTiming
}

type LogEntry struct {
	Id       string `bson:"_id,omitempty"`
	CallCost *CallCost
}

func (ms *MongoStorage) GetActivationPeriodsOrFallback(key string) ([]*ActivationPeriod, string, error) {
	result := new(ApKeyValue)
	err := ms.db.C("activationperiods").Find(bson.M{"key": key}).One(&result)
	return result.Value, result.FallbackKey, err
}

func (ms *MongoStorage) SetActivationPeriodsOrFallback(key string, aps []*ActivationPeriod, fallbackKey string) error {
	return ms.db.C("activationperiods").Insert(&ApKeyValue{Key: key, FallbackKey: fallbackKey, Value: aps})
}

func (ms *MongoStorage) GetDestination(key string) (result *Destination, err error) {
	result = new(Destination)
	err = ms.db.C("destinations").Find(bson.M{"id": key}).One(result)
	if err != nil {
		result = nil
	}
	return
}

func (ms *MongoStorage) SetDestination(dest *Destination) error {
	return ms.db.C("destinations").Insert(dest)
}

func (ms *MongoStorage) GetActions(key string) (as []*Action, err error) {
	result := AcKeyValue{}
	err = ms.db.C("actions").Find(bson.M{"key": key}).One(&result)
	return result.Value, err
}

func (ms *MongoStorage) SetActions(key string, as []*Action) error {
	return ms.db.C("actions").Insert(&AcKeyValue{Key: key, Value: as})
}

func (ms *MongoStorage) GetUserBalance(key string) (result *UserBalance, err error) {
	result = new(UserBalance)
	err = ms.db.C("userbalances").Find(bson.M{"id": key}).One(result)
	return
}

func (ms *MongoStorage) SetUserBalance(ub *UserBalance) error {
	return ms.db.C("userbalances").Insert(ub)
}

func (ms *MongoStorage) GetActionTimings(key string) (ats []*ActionTiming, err error) {
	result := AtKeyValue{}
	err = ms.db.C("actiontimings").Find(bson.M{"key": key}).One(&result)
	return result.Value, err
}

func (ms *MongoStorage) SetActionTimings(key string, ats []*ActionTiming) error {
	return ms.db.C("actiontimings").Insert(&AtKeyValue{key, ats})
}

func (ms *MongoStorage) GetAllActionTimings() (ats map[string][]*ActionTiming, err error) {
	result := AtKeyValue{}
	iter := ms.db.C("actiontimings").Find(bson.M{"key": fmt.Sprintf("/^%s/", ACTION_TIMING_PREFIX)}).Iter()
	ats = make(map[string][]*ActionTiming)
	for iter.Next(&result) {
		ats[result.Key] = result.Value
	}
	return
}

func (ms *MongoStorage) LogCallCost(uuid string, cc *CallCost) error {
	return ms.db.C("cclog").Insert(&LogEntry{uuid, cc})

}

func (ms *MongoStorage) GetCallCostLog(uuid string) (cc *CallCost, err error) {
	result := new(LogEntry)
	err = ms.db.C("cclog").Find(bson.M{"_id": uuid}).One(result)
	cc = result.CallCost
	return
}
