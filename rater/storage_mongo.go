/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package rater

import (
	"fmt"
	"errors"
	"github.com/cgrates/cgrates/utils"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"time"
)

type MongoStorage struct {
	session *mgo.Session
	db      *mgo.Database
}

func NewMongoStorage(host, port, db, user, pass string) (DataStorage, error) {
	dial := fmt.Sprintf(host)
	if user != "" && pass != "" {
		dial = fmt.Sprintf("%s:%s@%s", user, pass, dial)
	}
	if port != "" {
		dial += ":" + port
	}
	session, err := mgo.Dial(dial)
	if err != nil {
		log.Printf(fmt.Sprintf("Could not connect to logger database: %v", err))
		return nil, err
	}
	ndb := session.DB(db)
	session.SetMode(mgo.Monotonic, true)
	index := mgo.Index{Key: []string{"key"}, Background: true}
	err = ndb.C("actions").EnsureIndex(index)
	err = ndb.C("actiontimings").EnsureIndex(index)
	index = mgo.Index{Key: []string{"id"}, Background: true}
	err = ndb.C("ratingprofiles").EnsureIndex(index)
	err = ndb.C("destinations").EnsureIndex(index)
	err = ndb.C("userbalances").EnsureIndex(index)

	return &MongoStorage{db: ndb, session: session}, nil
}

func (ms *MongoStorage) Close() {
	ms.session.Close()
}

func (ms *MongoStorage) Flush() (err error) {
	err = ms.db.C("ratingprofiles").DropCollection()
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

type AcKeyValue struct {
	Key   string
	Value Actions
}

type AtKeyValue struct {
	Key   string
	Value ActionTimings
}

type LogCostEntry struct {
	Id       string `bson:"_id,omitempty"`
	CallCost *CallCost
	Source   string
}

type LogTimingEntry struct {
	ActionTiming *ActionTiming
	Actions      Actions
	LogTime      time.Time
	Source       string
}

type LogTriggerEntry struct {
	ubId          string
	ActionTrigger *ActionTrigger
	Actions       Actions
	LogTime       time.Time
	Source        string
}

type LogErrEntry struct {
	Id     string `bson:"_id,omitempty"`
	ErrStr string
	Source string
}

func (ms *MongoStorage) GetRatingProfile(key string) (rp *RatingProfile, err error) {
	rp = new(RatingProfile)
	err = ms.db.C("ratingprofiles").Find(bson.M{"id": key}).One(&rp)
	return
}

func (ms *MongoStorage) SetRatingProfile(rp *RatingProfile) error {
	return ms.db.C("ratingprofiles").Insert(rp)
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

// Extracts destinations from StorDB on specific tariffplan id
func (ms *MongoStorage) GetTPDestination( tpid, destTag string ) (*Destination, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) GetActions(key string) (as Actions, err error) {
	result := AcKeyValue{}
	err = ms.db.C("actions").Find(bson.M{"key": key}).One(&result)
	return result.Value, err
}

func (ms *MongoStorage) SetActions(key string, as Actions) error {
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

func (ms *MongoStorage) GetActionTimings(key string) (ats ActionTimings, err error) {
	result := AtKeyValue{}
	err = ms.db.C("actiontimings").Find(bson.M{"key": key}).One(&result)
	return result.Value, err
}

func (ms *MongoStorage) SetActionTimings(key string, ats ActionTimings) error {
	return ms.db.C("actiontimings").Insert(&AtKeyValue{key, ats})
}

func (ms *MongoStorage) GetAllActionTimings() (ats map[string]ActionTimings, err error) {
	result := AtKeyValue{}
	iter := ms.db.C("actiontimings").Find(nil).Iter()
	ats = make(map[string]ActionTimings)
	for iter.Next(&result) {
		ats[result.Key] = result.Value
	}
	return
}

func (ms *MongoStorage) LogCallCost(uuid, source string, cc *CallCost) error {
	return ms.db.C("cclog").Insert(&LogCostEntry{uuid, cc, source})
}

func (ms *MongoStorage) GetCallCostLog(uuid, source string) (cc *CallCost, err error) {
	result := new(LogCostEntry)
	err = ms.db.C("cclog").Find(bson.M{"_id": uuid, "source": source}).One(result)
	cc = result.CallCost
	return
}

func (ms *MongoStorage) LogActionTrigger(ubId, source string, at *ActionTrigger, as Actions) (err error) {
	return ms.db.C("actlog").Insert(&LogTriggerEntry{ubId, at, as, time.Now(), source})
}

func (ms *MongoStorage) LogActionTiming(source string, at *ActionTiming, as Actions) (err error) {
	return ms.db.C("actlog").Insert(&LogTimingEntry{at, as, time.Now(), source})
}

func (ms *MongoStorage) LogError(uuid, source, errstr string) (err error) {
	return ms.db.C("errlog").Insert(&LogErrEntry{uuid, errstr, source})
}

func (ms *MongoStorage) SetCdr(utils.CDR) error {
	return nil
}

func (ms *MongoStorage) SetRatedCdr(utils.CDR, *CallCost) error {
	return nil
}

func (ms *MongoStorage) GetDestinations(tpid string) ([]*Destination, error) {
	return nil, nil
}

func (ms *MongoStorage) GetTpDestinations(tpid string) ([]*Destination, error) {
	return nil, nil
}

func (ms *MongoStorage) GetTpRates(string) (map[string][]*Rate, error) {
	return nil, nil
}
func (ms *MongoStorage) GetTpTimings(string) (map[string]*Timing, error) {
	return nil, nil
}
func (ms *MongoStorage) GetTpRateTimings(string) ([]*RateTiming, error) {
	return nil, nil
}
func (ms *MongoStorage) GetTpRatingProfiles(string) (map[string]*RatingProfile, error) {
	return nil, nil
}
func (ms *MongoStorage) GetTpActions(string) (map[string][]*Action, error) {
	return nil, nil
}
func (ms *MongoStorage) GetTpActionTimings(string) (map[string][]*ActionTiming, error) {
	return nil, nil
}
func (ms *MongoStorage) GetTpActionTriggers(string) (map[string][]*ActionTrigger, error) {
	return nil, nil
}
func (ms *MongoStorage) GetTpAccountActions(string) ([]*AccountAction, error) {
	return nil, nil
}
