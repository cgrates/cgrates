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
package engine

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"reflect"

	"github.com/cgrates/cgrates/utils"
	"github.com/ugorji/go/codec"
	"gopkg.in/mgo.v2/bson"
)

type Storage interface {
	Close()
	Flush(string) error
	GetKeysForPrefix(string) ([]string, error)
	RebuildReverseForPrefix(string) error
	GetVersions(itm string) (vrs Versions, err error)
	SetVersions(vrs Versions, overwrite bool) (err error)
	RemoveVersions(vrs Versions) (err error)
	SelectDatabase(dbName string) (err error)
}

// OnlineStorage contains methods to use for administering online data
type DataDB interface {
	Storage
	Marshaler() Marshaler
	HasData(string, string) (bool, error)
	LoadRatingCache(dstIDs, rvDstIDs, rplIDs, rpfIDs, actIDs, aplIDs, aapIDs, atrgIDs, sgIDs, lcrIDs, dcIDs []string) error
	GetRatingPlan(string, bool, string) (*RatingPlan, error)
	SetRatingPlan(*RatingPlan, string) error
	GetRatingProfile(string, bool, string) (*RatingProfile, error)
	SetRatingProfile(*RatingProfile, string) error
	RemoveRatingProfile(string, string) error
	GetDestination(string, bool, string) (*Destination, error)
	SetDestination(*Destination, string) error
	RemoveDestination(string, string) error
	SetReverseDestination(*Destination, string) error
	GetReverseDestination(string, bool, string) ([]string, error)
	UpdateReverseDestination(*Destination, *Destination, string) error
	GetLCR(string, bool, string) (*LCR, error)
	SetLCR(*LCR, string) error
	SetCdrStats(*CdrStats) error
	GetCdrStats(string) (*CdrStats, error)
	GetAllCdrStats() ([]*CdrStats, error)
	GetDerivedChargers(string, bool, string) (*utils.DerivedChargers, error)
	SetDerivedChargers(string, *utils.DerivedChargers, string) error
	GetActions(string, bool, string) (Actions, error)
	SetActions(string, Actions, string) error
	RemoveActions(string, string) error
	GetSharedGroup(string, bool, string) (*SharedGroup, error)
	SetSharedGroup(*SharedGroup, string) error
	GetActionTriggers(string, bool, string) (ActionTriggers, error)
	SetActionTriggers(string, ActionTriggers, string) error
	RemoveActionTriggers(string, string) error
	GetActionPlan(string, bool, string) (*ActionPlan, error)
	SetActionPlan(string, *ActionPlan, bool, string) error
	GetAllActionPlans() (map[string]*ActionPlan, error)
	GetAccountActionPlans(acntID string, skipCache bool, transactionID string) (apIDs []string, err error)
	SetAccountActionPlans(acntID string, apIDs []string, overwrite bool) (err error)
	RemAccountActionPlans(acntID string, apIDs []string) (err error)
	PushTask(*Task) error
	PopTask() (*Task, error)
	LoadAccountingCache(alsIDs, rvAlsIDs, rlIDs []string) error
	GetAccount(string) (*Account, error)
	SetAccount(*Account) error
	RemoveAccount(string) error
	GetCdrStatsQueue(string) (*StatsQueue, error)
	SetCdrStatsQueue(*StatsQueue) error
	GetSubscribers() (map[string]*SubscriberData, error)
	SetSubscriber(string, *SubscriberData) error
	RemoveSubscriber(string) error
	SetUser(*UserProfile) error
	GetUser(string) (*UserProfile, error)
	GetUsers() ([]*UserProfile, error)
	RemoveUser(string) error
	SetAlias(*Alias, string) error
	GetAlias(string, bool, string) (*Alias, error)
	RemoveAlias(string, string) error
	SetReverseAlias(*Alias, string) error
	GetReverseAlias(string, bool, string) ([]string, error)
	GetResourceLimit(string, bool, string) (*ResourceLimit, error)
	SetResourceLimit(*ResourceLimit, string) error
	RemoveResourceLimit(string, string) error
	GetLoadHistory(int, bool, string) ([]*utils.LoadInstance, error)
	AddLoadHistory(*utils.LoadInstance, int, string) error
	GetStructVersion() (*StructVersion, error)
	SetStructVersion(*StructVersion) error
	GetReqFilterIndexes(dbKey string) (indexes map[string]map[string]utils.StringMap, err error)
	SetReqFilterIndexes(dbKey string, indexes map[string]map[string]utils.StringMap) (err error)
	MatchReqFilterIndex(dbKey, fieldValKey string) (itemIDs utils.StringMap, err error)
	// CacheDataFromDB loads data to cache, prefix represents the cache prefix, IDs should be nil if all available data should be loaded
	CacheDataFromDB(prefix string, IDs []string, mustBeCached bool) error // ToDo: Move this to dataManager
}

type StorDB interface {
	CdrStorage
	LoadReader
	LoadWriter
}

type CdrStorage interface {
	Storage
	SetCDR(*CDR, bool) error
	SetSMCost(smc *SMCost) error
	GetSMCosts(cgrid, runid, originHost, originIDPrfx string) ([]*SMCost, error)
	RemoveSMCost(*SMCost) error
	GetCDRs(*utils.CDRsFilter, bool) ([]*CDR, int64, error)
}

type LoadStorage interface {
	Storage
	LoadReader
	LoadWriter
}

// LoadReader reads from .csv or TP tables and provides the data ready for the tp_db or data_db.
type LoadReader interface {
	GetTpIds() ([]string, error)
	GetTpTableIds(string, string, utils.TPDistinctIds, map[string]string, *utils.Paginator) ([]string, error)
	GetTPTimings(string, string) ([]*utils.ApierTPTiming, error)
	GetTPDestinations(string, string) ([]*utils.TPDestination, error)
	GetTPRates(string, string) ([]*utils.TPRate, error)
	GetTPDestinationRates(string, string, *utils.Paginator) ([]*utils.TPDestinationRate, error)
	GetTPRatingPlans(string, string, *utils.Paginator) ([]*utils.TPRatingPlan, error)
	GetTPRatingProfiles(*utils.TPRatingProfile) ([]*utils.TPRatingProfile, error)
	GetTPSharedGroups(string, string) ([]*utils.TPSharedGroups, error)
	GetTPCdrStats(string, string) ([]*utils.TPCdrStats, error)
	GetTPLCRs(*utils.TPLcrRules) ([]*utils.TPLcrRules, error)
	GetTPUsers(*utils.TPUsers) ([]*utils.TPUsers, error)
	GetTPAliases(*utils.TPAliases) ([]*utils.TPAliases, error)
	GetTPDerivedChargers(*utils.TPDerivedChargers) ([]*utils.TPDerivedChargers, error)
	GetTPActions(string, string) ([]*utils.TPActions, error)
	GetTPActionPlans(string, string) ([]*utils.TPActionPlan, error)
	GetTPActionTriggers(string, string) ([]*utils.TPActionTriggers, error)
	GetTPAccountActions(*utils.TPAccountActions) ([]*utils.TPAccountActions, error)
	GetTPResourceLimits(string, string) ([]*utils.TPResourceLimit, error)
}

type LoadWriter interface {
	RemTpData(string, string, map[string]string) error
	SetTPTimings([]*utils.ApierTPTiming) error
	SetTPDestinations([]*utils.TPDestination) error
	SetTPRates([]*utils.TPRate) error
	SetTPDestinationRates([]*utils.TPDestinationRate) error
	SetTPRatingPlans([]*utils.TPRatingPlan) error
	SetTPRatingProfiles([]*utils.TPRatingProfile) error
	SetTPSharedGroups([]*utils.TPSharedGroups) error
	SetTPCdrStats([]*utils.TPCdrStats) error
	SetTPUsers([]*utils.TPUsers) error
	SetTPAliases([]*utils.TPAliases) error
	SetTPDerivedChargers([]*utils.TPDerivedChargers) error
	SetTPLCRs([]*utils.TPLcrRules) error
	SetTPActions([]*utils.TPActions) error
	SetTPActionPlans([]*utils.TPActionPlan) error
	SetTPActionTriggers([]*utils.TPActionTriggers) error
	SetTPAccountActions([]*utils.TPAccountActions) error
	SetTPResourceLimits([]*utils.TPResourceLimit) error
}

type Marshaler interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

type JSONMarshaler struct{}

func (jm *JSONMarshaler) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (jm *JSONMarshaler) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

type BSONMarshaler struct{}

func (jm *BSONMarshaler) Marshal(v interface{}) ([]byte, error) {
	return bson.Marshal(v)
}

func (jm *BSONMarshaler) Unmarshal(data []byte, v interface{}) error {
	return bson.Unmarshal(data, v)
}

type JSONBufMarshaler struct{}

func (jbm *JSONBufMarshaler) Marshal(v interface{}) (data []byte, err error) {
	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(v)
	data = buf.Bytes()
	return
}

func (jbm *JSONBufMarshaler) Unmarshal(data []byte, v interface{}) error {
	return json.NewDecoder(bytes.NewBuffer(data)).Decode(v)
}

type CodecMsgpackMarshaler struct {
	mh *codec.MsgpackHandle
}

func NewCodecMsgpackMarshaler() *CodecMsgpackMarshaler {
	cmm := &CodecMsgpackMarshaler{new(codec.MsgpackHandle)}
	mh := cmm.mh
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	mh.RawToString = true
	return cmm
}

func (cmm *CodecMsgpackMarshaler) Marshal(v interface{}) (b []byte, err error) {
	enc := codec.NewEncoderBytes(&b, cmm.mh)
	err = enc.Encode(v)
	return
}

func (cmm *CodecMsgpackMarshaler) Unmarshal(data []byte, v interface{}) error {
	dec := codec.NewDecoderBytes(data, cmm.mh)
	return dec.Decode(&v)
}

type BincMarshaler struct {
	bh *codec.BincHandle
}

func NewBincMarshaler() *BincMarshaler {
	return &BincMarshaler{new(codec.BincHandle)}
}

func (bm *BincMarshaler) Marshal(v interface{}) (b []byte, err error) {
	enc := codec.NewEncoderBytes(&b, bm.bh)
	err = enc.Encode(v)
	return
}

func (bm *BincMarshaler) Unmarshal(data []byte, v interface{}) error {
	dec := codec.NewDecoderBytes(data, bm.bh)
	return dec.Decode(&v)
}

type GOBMarshaler struct{}

func (gm *GOBMarshaler) Marshal(v interface{}) (data []byte, err error) {
	buf := new(bytes.Buffer)
	err = gob.NewEncoder(buf).Encode(v)
	data = buf.Bytes()
	return
}

func (gm *GOBMarshaler) Unmarshal(data []byte, v interface{}) error {
	return gob.NewDecoder(bytes.NewBuffer(data)).Decode(v)
}

// Decide the value of cacheCommit parameter based on transactionID
func cacheCommit(transactionID string) bool {
	if transactionID == utils.NonTransactional {
		return true
	}
	return false
}
