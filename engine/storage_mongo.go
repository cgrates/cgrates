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

package engine

import (
	"errors"
	"fmt"
	"github.com/cgrates/cgrates/history"
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
	if historyScribe != nil {
		response := 0
		historyScribe.Record(&history.Record{RATING_PROFILE_PREFIX + rp.Id, rp}, &response)
	}
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
	if historyScribe != nil {
		response := 0
		historyScribe.Record(&history.Record{DESTINATION_PREFIX + dest.Id, dest}, &response)
	}
	return ms.db.C("destinations").Insert(dest)
}

func (ms *MongoStorage) GetTPIds() ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) SetTPTiming(tpid string, tm *Timing) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) ExistsTPTiming(tpid, tmId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) GetTPTiming(tpid, tmId string) (*Timing, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) GetTPTimingIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) GetTPDestinationIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) ExistsTPDestination(tpid, destTag string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

// Extracts destinations from StorDB on specific tariffplan id
func (ms *MongoStorage) GetTPDestination(tpid, destTag string) (*Destination, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) SetTPDestination(tpid string, dest *Destination) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) ExistsTPRate(tpid, rtId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) SetTPRates(tpid string, rts map[string][]*Rate) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) GetTPRate(tpid, rtId string) (*utils.TPRate, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) GetTPRateIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) ExistsTPDestinationRate(tpid, drId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) SetTPDestinationRates(tpid string, drs map[string][]*DestinationRate) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) GetTPDestinationRate(tpid, drId string) (*utils.TPDestinationRate, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) GetTPDestinationRateIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) ExistsTPDestRateTiming(tpid, drtId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) SetTPDestRateTimings(tpid string, drts map[string][]*DestinationRateTiming) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) GetTPDestRateTiming(tpid, drtId string) (*utils.TPDestRateTiming, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) GetTPDestRateTimingIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) ExistsTPRatingProfile(tpid, rpId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) SetTPRatingProfiles(tpid string, rps map[string][]*RatingProfile) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) GetTPRatingProfile(tpid, rpId string) (*utils.TPRatingProfile, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) GetTPRatingProfileIds(filters *utils.AttrTPRatingProfileIds) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) ExistsTPActions(tpid, aId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) SetTPActions(tpid string, acts map[string][]*Action) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) GetTPActions(tpid, aId string) (*utils.TPActions, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) GetTPActionIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) ExistsTPActionTimings(tpid, atId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) SetTPActionTimings(tpid string, ats map[string][]*ActionTiming) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) GetTPActionTimings(tpid, atId string) (map[string][]*utils.TPActionTimingsRow, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) GetTPActionTimingIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) ExistsTPActionTriggers(tpid, atId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) SetTPActionTriggers(tpid string, ats map[string][]*ActionTrigger) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) GetTPActionTriggerIds(tpid string) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) ExistsTPAccountActions(tpid, aaId string) (bool, error) {
	return false, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) SetTPAccountActions(tpid string, aa map[string]*AccountAction) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) GetTPAccountActionIds(tpid string) ([]string, error) {
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

func (ms *MongoStorage) SetRatedCdr(cdr utils.CDR, cc *CallCost, extraInfo string) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (ms *MongoStorage) GetAllRatedCdr() ([]utils.CDR, error) {
	return nil, nil
}

func (ms *MongoStorage) GetDestinations(tpid string) ([]*Destination, error) {
	return nil, nil
}

func (ms *MongoStorage) GetTpDestinations(tpid, tag string) ([]*Destination, error) {
	return nil, nil
}

func (ms *MongoStorage) GetTpRates(tpid, tag string) (map[string]*Rate, error) {
	return nil, nil
}
func (ms *MongoStorage) GetTpDestinationRates(tpid, tag string) (map[string][]*DestinationRate, error) {
	return nil, nil
}
func (ms *MongoStorage) GetTpTimings(tpid, tag string) (map[string]*Timing, error) {
	return nil, nil
}
func (ms *MongoStorage) GetTpDestinationRateTimings(tpid, tag string) ([]*DestinationRateTiming, error) {
	return nil, nil
}
func (ms *MongoStorage) GetTpRatingProfiles(tpid, tag string) (map[string]*RatingProfile, error) {
	return nil, nil
}
func (ms *MongoStorage) GetTpActions(tpid, tag string) (map[string][]*Action, error) {
	return nil, nil
}
func (ms *MongoStorage) GetTpActionTimings(tpid, tag string) (map[string][]*ActionTiming, error) {
	return nil, nil
}
func (ms *MongoStorage) GetTpActionTriggers(tpid, tag string) (map[string][]*ActionTrigger, error) {
	return nil, nil
}
func (ms *MongoStorage) GetTpAccountActions(tpid, tag string) (map[string]*AccountAction, error) {
	return nil, nil
}
