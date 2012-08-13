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
	ap, d, a, ub, at *mgo.Collection
	session          *mgo.Session
}

func NewMongoStorage(address, db string) (StorageGetter, error) {
	session, err := mgo.Dial(address)
	if err != nil {
		Logger.Err("Could not contact mongo server")
		return nil, err
	}
	ndb := session.DB(db)
	ap := ndb.C("activationPeriods")
	d := ndb.C("destinations")
	a := ndb.C("actions")
	ub := ndb.C("userBalance")
	at := ndb.C("actionTimings")
	session.SetMode(mgo.Monotonic, true)
	index := mgo.Index{Key: []string{"key"}, Background: true}
	err = ap.EnsureIndex(index)
	err = at.EnsureIndex(index)
	err = a.EnsureIndex(index)
	index = mgo.Index{Key: []string{"id"}, Background: true}
	err = d.EnsureIndex(index)
	err = ub.EnsureIndex(index)

	return &MongoStorage{
		ap:      ap,
		d:       d,
		a:       a,
		ub:      ub,
		at:      at,
		session: session,
	}, nil
}

func (ms *MongoStorage) Close() {
	ms.session.Close()
}

func (ms *MongoStorage) Flush() (err error) {
	err = ms.ap.DropCollection()
	if err != nil {
		return
	}
	err = ms.d.DropCollection()
	if err != nil {
		return
	}
	err = ms.a.DropCollection()
	if err != nil {
		return
	}
	err = ms.ub.DropCollection()
	if err != nil {
		return
	}
	err = ms.at.DropCollection()
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

func (ms *MongoStorage) GetActivationPeriodsOrFallback(key string) (aps []*ActivationPeriod, fallbackKey string, err error) {
	result := ApKeyValue{}
	err = ms.ap.Find(bson.M{"key": key}).One(&result)
	return result.Value, result.FallbackKey, err
}

func (ms *MongoStorage) SetActivationPeriodsOrFallback(key string, aps []*ActivationPeriod, fallbackKey string) error {
	return ms.ap.Insert(&ApKeyValue{Key: key, FallbackKey: fallbackKey, Value: aps})
}

func (ms *MongoStorage) GetDestination(key string) (result *Destination, err error) {
	result = new(Destination)
	err = ms.d.Find(bson.M{"id": key}).One(result)
	if err != nil {
		result = nil
	}
	return
}

func (ms *MongoStorage) SetDestination(dest *Destination) error {
	return ms.d.Insert(dest)
}

func (ms *MongoStorage) GetActions(key string) (as []*Action, err error) {
	result := AcKeyValue{}
	err = ms.a.Find(bson.M{"key": key}).One(&result)
	return result.Value, err
}

func (ms *MongoStorage) SetActions(key string, as []*Action) error {
	return ms.a.Insert(&AcKeyValue{Key: key, Value: as})
}

func (ms *MongoStorage) GetUserBalance(key string) (result *UserBalance, err error) {
	result = new(UserBalance)
	err = ms.ub.Find(bson.M{"id": key}).One(result)
	return
}

func (ms *MongoStorage) SetUserBalance(ub *UserBalance) error {
	return ms.ub.Insert(ub)
}

func (ms *MongoStorage) GetActionTimings(key string) (ats []*ActionTiming, err error) {
	result := AtKeyValue{}
	err = ms.at.Find(bson.M{"key": key}).One(&result)
	return result.Value, err
}

func (ms *MongoStorage) SetActionTimings(key string, ats []*ActionTiming) error {
	return ms.at.Insert(&AtKeyValue{key, ats})
}

func (ms *MongoStorage) GetAllActionTimings() (ats map[string][]*ActionTiming, err error) {
	result := AtKeyValue{}
	iter := ms.at.Find(bson.M{"key": fmt.Sprintf("/^%s/", ACTION_TIMING_PREFIX)}).Iter()
	ats = make(map[string][]*ActionTiming)
	for iter.Next(&result) {
		ats[result.Key] = result.Value
	}
	return
}
