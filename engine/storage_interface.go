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
	"fmt"
	"reflect"

	"github.com/cgrates/cgrates/utils"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/ugorji/go/codec"
)

type Storage interface {
	Close()
	Flush(string) error
	GetKeysForPrefix(string) ([]string, error)
	RebuildReverseForPrefix(string) error
	RemoveReverseForPrefix(string) error
	GetVersions(itm string) (vrs Versions, err error)
	SetVersions(vrs Versions, overwrite bool) (err error)
	RemoveVersions(vrs Versions) (err error)
	SelectDatabase(dbName string) (err error)
	GetStorageType() string
	IsDBEmpty() (resp bool, err error)
}

// OnlineStorage contains methods to use for administering online data
type DataDB interface {
	Storage
	Marshaler() Marshaler
	HasDataDrv(string, string, string) (bool, error)
	GetRatingPlanDrv(string) (*RatingPlan, error)
	SetRatingPlanDrv(*RatingPlan) error
	RemoveRatingPlanDrv(key string) (err error)
	GetRatingProfileDrv(string) (*RatingProfile, error)
	SetRatingProfileDrv(*RatingProfile) error
	RemoveRatingProfileDrv(string) error
	GetDestination(string, bool, string) (*Destination, error)
	SetDestination(*Destination, string) error
	RemoveDestination(string, string) error
	SetReverseDestination(*Destination, string) error
	GetReverseDestination(string, bool, string) ([]string, error)
	UpdateReverseDestination(*Destination, *Destination, string) error
	GetDerivedChargersDrv(string) (*utils.DerivedChargers, error)
	SetDerivedChargers(string, *utils.DerivedChargers, string) error
	RemoveDerivedChargersDrv(id, transactionID string) (err error)
	GetActionsDrv(string) (Actions, error)
	SetActionsDrv(string, Actions) error
	RemoveActionsDrv(string) error
	GetSharedGroupDrv(string) (*SharedGroup, error)
	SetSharedGroupDrv(*SharedGroup) error
	RemoveSharedGroupDrv(id, transactionID string) (err error)
	GetActionTriggersDrv(string) (ActionTriggers, error)
	SetActionTriggersDrv(string, ActionTriggers) error
	RemoveActionTriggersDrv(string) error
	GetActionPlan(string, bool, string) (*ActionPlan, error)
	SetActionPlan(string, *ActionPlan, bool, string) error
	RemoveActionPlan(key string, transactionID string) error
	GetAllActionPlans() (map[string]*ActionPlan, error)
	GetAccountActionPlans(acntID string, skipCache bool,
		transactionID string) (apIDs []string, err error)
	SetAccountActionPlans(acntID string, apIDs []string, overwrite bool) (err error)
	RemAccountActionPlans(acntID string, apIDs []string) (err error)
	PushTask(*Task) error
	PopTask() (*Task, error)
	GetAccount(string) (*Account, error)
	SetAccount(*Account) error
	RemoveAccount(string) error
	GetSubscribersDrv() (map[string]*SubscriberData, error)
	SetSubscriberDrv(string, *SubscriberData) error
	RemoveSubscriberDrv(string) error
	SetUserDrv(*UserProfile) error
	GetUserDrv(string) (*UserProfile, error)
	GetUsersDrv() ([]*UserProfile, error)
	RemoveUserDrv(string) error
	SetAlias(*Alias, string) error
	GetAlias(string, bool, string) (*Alias, error)
	RemoveAlias(string, string) error
	SetReverseAlias(*Alias, string) error
	GetReverseAlias(string, bool, string) ([]string, error)
	GetResourceProfileDrv(string, string) (*ResourceProfile, error)
	SetResourceProfileDrv(*ResourceProfile) error
	RemoveResourceProfileDrv(string, string) error
	GetResourceDrv(string, string) (*Resource, error)
	SetResourceDrv(*Resource) error
	RemoveResourceDrv(string, string) error
	GetTimingDrv(string) (*utils.TPTiming, error)
	SetTimingDrv(*utils.TPTiming) error
	RemoveTimingDrv(string) error
	GetLoadHistory(int, bool, string) ([]*utils.LoadInstance, error)
	AddLoadHistory(*utils.LoadInstance, int, string) error
	GetFilterIndexesDrv(cacheID, itemIDPrefix, filterType string,
		fldNameVal map[string]string) (indexes map[string]utils.StringMap, err error)
	SetFilterIndexesDrv(cacheID, itemIDPrefix string,
		indexes map[string]utils.StringMap, commit bool, transactionID string) (err error)
	RemoveFilterIndexesDrv(cacheID, itemIDPrefix string) (err error)
	MatchFilterIndexDrv(cacheID, itemIDPrefix,
		filterType, fieldName, fieldVal string) (itemIDs utils.StringMap, err error)
	GetStatQueueProfileDrv(tenant string, ID string) (sq *StatQueueProfile, err error)
	SetStatQueueProfileDrv(sq *StatQueueProfile) (err error)
	RemStatQueueProfileDrv(tenant, id string) (err error)
	GetStoredStatQueueDrv(tenant, id string) (sq *StoredStatQueue, err error)
	SetStoredStatQueueDrv(sq *StoredStatQueue) (err error)
	RemStoredStatQueueDrv(tenant, id string) (err error)
	GetThresholdProfileDrv(tenant string, ID string) (tp *ThresholdProfile, err error)
	SetThresholdProfileDrv(tp *ThresholdProfile) (err error)
	RemThresholdProfileDrv(tenant, id string) (err error)
	GetThresholdDrv(string, string) (*Threshold, error)
	SetThresholdDrv(*Threshold) error
	RemoveThresholdDrv(string, string) error
	GetFilterDrv(string, string) (*Filter, error)
	SetFilterDrv(*Filter) error
	RemoveFilterDrv(string, string) error
	GetSupplierProfileDrv(string, string) (*SupplierProfile, error)
	SetSupplierProfileDrv(*SupplierProfile) error
	RemoveSupplierProfileDrv(string, string) error
	GetAttributeProfileDrv(string, string) (*AttributeProfile, error)
	SetAttributeProfileDrv(*AttributeProfile) error
	RemoveAttributeProfileDrv(string, string) error
	GetChargerProfileDrv(string, string) (*ChargerProfile, error)
	SetChargerProfileDrv(*ChargerProfile) error
	RemoveChargerProfileDrv(string, string) error
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
	GetTpIds(string) ([]string, error)
	GetTpTableIds(string, string, utils.TPDistinctIds,
		map[string]string, *utils.Paginator) ([]string, error)
	GetTPTimings(string, string) ([]*utils.ApierTPTiming, error)
	GetTPDestinations(string, string) ([]*utils.TPDestination, error)
	GetTPRates(string, string) ([]*utils.TPRate, error)
	GetTPDestinationRates(string, string, *utils.Paginator) ([]*utils.TPDestinationRate, error)
	GetTPRatingPlans(string, string, *utils.Paginator) ([]*utils.TPRatingPlan, error)
	GetTPRatingProfiles(*utils.TPRatingProfile) ([]*utils.TPRatingProfile, error)
	GetTPSharedGroups(string, string) ([]*utils.TPSharedGroups, error)
	GetTPUsers(*utils.TPUsers) ([]*utils.TPUsers, error)
	GetTPAliases(*utils.TPAliases) ([]*utils.TPAliases, error)
	GetTPDerivedChargers(*utils.TPDerivedChargers) ([]*utils.TPDerivedChargers, error)
	GetTPActions(string, string) ([]*utils.TPActions, error)
	GetTPActionPlans(string, string) ([]*utils.TPActionPlan, error)
	GetTPActionTriggers(string, string) ([]*utils.TPActionTriggers, error)
	GetTPAccountActions(*utils.TPAccountActions) ([]*utils.TPAccountActions, error)
	GetTPResources(string, string) ([]*utils.TPResource, error)
	GetTPStats(string, string) ([]*utils.TPStats, error)
	GetTPThresholds(string, string) ([]*utils.TPThreshold, error)
	GetTPFilters(string, string) ([]*utils.TPFilterProfile, error)
	GetTPSuppliers(string, string) ([]*utils.TPSupplierProfile, error)
	GetTPAttributes(string, string) ([]*utils.TPAttributeProfile, error)
	GetTPChargers(string, string) ([]*utils.TPChargerProfile, error)
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
	SetTPUsers([]*utils.TPUsers) error
	SetTPAliases([]*utils.TPAliases) error
	SetTPDerivedChargers([]*utils.TPDerivedChargers) error
	SetTPActions([]*utils.TPActions) error
	SetTPActionPlans([]*utils.TPActionPlan) error
	SetTPActionTriggers([]*utils.TPActionTriggers) error
	SetTPAccountActions([]*utils.TPAccountActions) error
	SetTPResources([]*utils.TPResource) error
	SetTPStats([]*utils.TPStats) error
	SetTPThresholds([]*utils.TPThreshold) error
	SetTPFilters([]*utils.TPFilterProfile) error
	SetTPSuppliers([]*utils.TPSupplierProfile) error
	SetTPAttributes([]*utils.TPAttributeProfile) error
	SetTPChargers([]*utils.TPChargerProfile) error
}

// NewMarshaler returns the marshaler type selected by mrshlerStr
func NewMarshaler(mrshlerStr string) (ms Marshaler, err error) {
	switch mrshlerStr {
	case utils.MSGPACK:
		ms = NewCodecMsgpackMarshaler()
	case utils.JSON:
		ms = new(JSONMarshaler)
	default:
		err = fmt.Errorf("Unsupported marshaler: %v", mrshlerStr)
	}
	return
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
