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
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type MongoStorage struct {
	db      *mgo.Database
	session *mgo.Session
}

func NewMongoStorage(address, db string) (*MongoStorage, error) {
	session, err := mgo.Dial(address)
	if err != nil {
		Logger.Err("Could not contact mongo server")
		return nil, err
	}
	session.SetMode(mgo.Monotonic, true)

	index := mgo.Index{Key: []string{"key"}, Unique: true, DropDups: true, Background: true}
	err = session.DB(db).C("activationPeriods").EnsureIndex(index)
	index = mgo.Index{Key: []string{"id"}, Unique: true, DropDups: true, Background: true}
	err = session.DB(db).C("destinations").EnsureIndex(index)
	err = session.DB(db).C("tariffPlans").EnsureIndex(index)
	err = session.DB(db).C("userBalance").EnsureIndex(index)

	return &MongoStorage{db: session.DB(db), session: session}, nil
}

func (ms *MongoStorage) Close() {
	ms.session.Close()
}

/*
Helper type for activation periods storage.
*/
type KeyValue struct {
	Key               string
	FallbackKey       string
	ActivationPeriods []*ActivationPeriod
}

func (ms *MongoStorage) GetActivationPeriodsOrFallback(key string) (aps []*ActivationPeriod, fallbackKey string, err error) {
	ndb := ms.db.C("activationPeriods")
	result := KeyValue{}
	err = ndb.Find(bson.M{"key": key}).One(&result)
	return result.ActivationPeriods, result.FallbackKey, err
}

func (ms *MongoStorage) SetActivationPeriodsOrFallback(key string, aps []*ActivationPeriod, fallbackKey string) error {
	ndb := ms.db.C("activationPeriods")
	return ndb.Insert(&KeyValue{key, fallbackKey, aps})
}

func (ms *MongoStorage) GetDestination(key string) (result *Destination, err error) {
	ndb := ms.db.C("destinations")
	result = &Destination{}
	err = ndb.Find(bson.M{"id": key}).One(result)
	return
}

func (ms *MongoStorage) SetDestination(dest *Destination) error {
	ndb := ms.db.C("destinations")
	return ndb.Insert(&dest)
}

func (ms *MongoStorage) GetUserBalance(key string) (result *UserBalance, err error) {
	ndb := ms.db.C("userBalance")
	result = &UserBalance{}
	err = ndb.Find(bson.M{"id": key}).One(result)
	return
}

func (ms *MongoStorage) SetUserBalance(ub *UserBalance) error {
	ndb := ms.db.C("userBalance")
	return ndb.Insert(&ub)
}
