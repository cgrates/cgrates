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
	"github.com/ugorji/go/codec"
	"go.mongodb.org/mongo-driver/bson"
)

type Storage interface {
	Close()
	Flush(string) error
	GetKeysForPrefix(string) ([]string, error)
	RemoveKeysForPrefix(string) error
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
	HasDataDrv(string, string, string) (bool, error)
	GetRatingPlanDrv(string) (*RatingPlan, error)
	SetRatingPlanDrv(*RatingPlan) error
	RemoveRatingPlanDrv(key string) (err error)
	GetRatingProfileDrv(string) (*RatingProfile, error)
	SetRatingProfileDrv(*RatingProfile) error
	RemoveRatingProfileDrv(string) error
	GetDestinationDrv(string, string) (*Destination, error)
	SetDestinationDrv(*Destination, string) error
	RemoveDestinationDrv(string, string) error
	RemoveReverseDestinationDrv(string, string, string) error
	SetReverseDestinationDrv(string, []string, string) error
	GetReverseDestinationDrv(string, string) ([]string, error)
	GetActionsDrv(string) (Actions, error)
	SetActionsDrv(string, Actions) error
	RemoveActionsDrv(string) error
	GetSharedGroupDrv(string) (*SharedGroup, error)
	SetSharedGroupDrv(*SharedGroup) error
	RemoveSharedGroupDrv(id string) (err error)
	GetActionTriggersDrv(string) (ActionTriggers, error)
	SetActionTriggersDrv(string, ActionTriggers) error
	RemoveActionTriggersDrv(string) error
	GetActionPlanDrv(string) (*ActionPlan, error)
	SetActionPlanDrv(string, *ActionPlan) error
	RemoveActionPlanDrv(key string) error
	GetAllActionPlansDrv() (map[string]*ActionPlan, error)
	GetAccountActionPlansDrv(acntID string) (apIDs []string, err error)
	SetAccountActionPlansDrv(acntID string, apIDs []string) (err error)
	RemAccountActionPlansDrv(acntID string) (err error)
	PushTask(*Task) error
	PopTask() (*Task, error)
	GetAccountDrv(string) (*Account, error)
	SetAccountDrv(*Account) error
	RemoveAccountDrv(string) error
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
	GetIndexesDrv(idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error)
	SetIndexesDrv(idxItmType, tntCtx string,
		indexes map[string]utils.StringSet, commit bool, transactionID string) (err error)
	RemoveIndexesDrv(idxItmType, tntCtx, idxKey string) (err error)
	GetStatQueueProfileDrv(tenant string, ID string) (sq *StatQueueProfile, err error)
	SetStatQueueProfileDrv(sq *StatQueueProfile) (err error)
	RemStatQueueProfileDrv(tenant, id string) (err error)
	GetStatQueueDrv(tenant, id string) (sq *StatQueue, err error)
	SetStatQueueDrv(ssq *StoredStatQueue, sq *StatQueue) (err error)
	RemStatQueueDrv(tenant, id string) (err error)
	SetRankingProfileDrv(sq *RankingProfile) (err error)
	GetRankingProfileDrv(tenant string, id string) (sq *RankingProfile, err error)
	RemRankingProfileDrv(tenant string, id string) (err error)
	GetRankingDrv(string, string) (*Ranking, error)
	SetRankingDrv(*Ranking) error
	RemoveRankingDrv(string, string) error
	SetTrendProfileDrv(tr *TrendProfile) (err error)
	GetTrendProfileDrv(tenant string, id string) (sq *TrendProfile, err error)
	RemTrendProfileDrv(tenant string, id string) (err error)
	GetTrendDrv(string, string) (*Trend, error)
	SetTrendDrv(*Trend) error
	RemoveTrendDrv(string, string) error
	GetThresholdProfileDrv(tenant string, ID string) (tp *ThresholdProfile, err error)
	SetThresholdProfileDrv(tp *ThresholdProfile) (err error)
	RemThresholdProfileDrv(tenant, id string) (err error)
	GetThresholdDrv(string, string) (*Threshold, error)
	SetThresholdDrv(*Threshold) error
	RemoveThresholdDrv(string, string) error
	GetFilterDrv(string, string) (*Filter, error)
	SetFilterDrv(*Filter) error
	RemoveFilterDrv(string, string) error
	GetRouteProfileDrv(string, string) (*RouteProfile, error)
	SetRouteProfileDrv(*RouteProfile) error
	RemoveRouteProfileDrv(string, string) error
	GetAttributeProfileDrv(string, string) (*AttributeProfile, error)
	SetAttributeProfileDrv(*AttributeProfile) error
	RemoveAttributeProfileDrv(string, string) error
	GetChargerProfileDrv(string, string) (*ChargerProfile, error)
	SetChargerProfileDrv(*ChargerProfile) error
	RemoveChargerProfileDrv(string, string) error
	GetDispatcherProfileDrv(string, string) (*DispatcherProfile, error)
	SetDispatcherProfileDrv(*DispatcherProfile) error
	RemoveDispatcherProfileDrv(string, string) error
	GetItemLoadIDsDrv(itemIDPrefix string) (loadIDs map[string]int64, err error)
	SetLoadIDsDrv(loadIDs map[string]int64) error
	RemoveLoadIDsDrv() error
	GetDispatcherHostDrv(string, string) (*DispatcherHost, error)
	SetDispatcherHostDrv(*DispatcherHost) error
	RemoveDispatcherHostDrv(string, string) error
	SetBackupSessionsDrv(nodeID string, tenant string, sessions []*StoredSession) error
	GetSessionsBackupDrv(nodeID string, tenant string) ([]*StoredSession, error)
	RemoveSessionsBackupDrv(nodeID, tenant, cgrid string) error
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
	RemoveSMCosts(qryFltr *utils.SMCostFilter) error
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
		map[string]string, *utils.PaginatorWithSearch) ([]string, error)
	GetTPTimings(string, string) ([]*utils.ApierTPTiming, error)
	GetTPDestinations(string, string) ([]*utils.TPDestination, error)
	GetTPRates(string, string) ([]*utils.TPRateRALs, error)
	GetTPDestinationRates(string, string, *utils.Paginator) ([]*utils.TPDestinationRate, error)
	GetTPRatingPlans(string, string, *utils.Paginator) ([]*utils.TPRatingPlan, error)
	GetTPRatingProfiles(*utils.TPRatingProfile) ([]*utils.TPRatingProfile, error)
	GetTPSharedGroups(string, string) ([]*utils.TPSharedGroups, error)
	GetTPActions(string, string) ([]*utils.TPActions, error)
	GetTPActionPlans(string, string) ([]*utils.TPActionPlan, error)
	GetTPActionTriggers(string, string) ([]*utils.TPActionTriggers, error)
	GetTPAccountActions(*utils.TPAccountActions) ([]*utils.TPAccountActions, error)
	GetTPResources(string, string, string) ([]*utils.TPResourceProfile, error)
	GetTPStats(string, string, string) ([]*utils.TPStatProfile, error)
	GetTPTrends(string, string, string) ([]*utils.TPTrendsProfile, error)
	GetTPRankings(string, string, string) ([]*utils.TPRankingProfile, error)
	GetTPThresholds(string, string, string) ([]*utils.TPThresholdProfile, error)
	GetTPFilters(string, string, string) ([]*utils.TPFilterProfile, error)
	GetTPRoutes(string, string, string) ([]*utils.TPRouteProfile, error)
	GetTPAttributes(string, string, string) ([]*utils.TPAttributeProfile, error)
	GetTPChargers(string, string, string) ([]*utils.TPChargerProfile, error)
	GetTPDispatcherProfiles(string, string, string) ([]*utils.TPDispatcherProfile, error)
	GetTPDispatcherHosts(string, string, string) ([]*utils.TPDispatcherHost, error)
}

type LoadWriter interface {
	RemTpData(string, string, map[string]string) error
	SetTPTimings([]*utils.ApierTPTiming) error
	SetTPDestinations([]*utils.TPDestination) error
	SetTPRates([]*utils.TPRateRALs) error
	SetTPDestinationRates([]*utils.TPDestinationRate) error
	SetTPRatingPlans([]*utils.TPRatingPlan) error
	SetTPRatingProfiles([]*utils.TPRatingProfile) error
	SetTPSharedGroups([]*utils.TPSharedGroups) error
	SetTPActions([]*utils.TPActions) error
	SetTPActionPlans([]*utils.TPActionPlan) error
	SetTPActionTriggers([]*utils.TPActionTriggers) error
	SetTPAccountActions([]*utils.TPAccountActions) error
	SetTPResources([]*utils.TPResourceProfile) error
	SetTPStats([]*utils.TPStatProfile) error
	SetTPTrends([]*utils.TPTrendsProfile) error
	SetTPRankings([]*utils.TPRankingProfile) error
	SetTPThresholds([]*utils.TPThresholdProfile) error
	SetTPFilters([]*utils.TPFilterProfile) error
	SetTPRoutes([]*utils.TPRouteProfile) error
	SetTPAttributes([]*utils.TPAttributeProfile) error
	SetTPChargers([]*utils.TPChargerProfile) error
	SetTPDispatcherProfiles([]*utils.TPDispatcherProfile) error
	SetTPDispatcherHosts([]*utils.TPDispatcherHost) error
}

// NewMarshaler returns the marshaler type selected by mrshlerStr
func NewMarshaler(mrshlerStr string) (ms Marshaler, err error) {
	switch mrshlerStr {
	case utils.MsgPack:
		ms = NewCodecMsgpackMarshaler()
	case utils.JSON:
		ms = new(JSONMarshaler)
	default:
		err = fmt.Errorf("Unsupported marshaler: %v", mrshlerStr)
	}
	return
}

type Marshaler interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

type JSONMarshaler struct{}

func (jm *JSONMarshaler) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (jm *JSONMarshaler) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

type BSONMarshaler struct{}

func (jm *BSONMarshaler) Marshal(v any) ([]byte, error) {
	return bson.Marshal(v)
}

func (jm *BSONMarshaler) Unmarshal(data []byte, v any) error {
	return bson.Unmarshal(data, v)
}

type JSONBufMarshaler struct{}

func (jbm *JSONBufMarshaler) Marshal(v any) (data []byte, err error) {
	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(v)
	data = buf.Bytes()
	return
}

func (jbm *JSONBufMarshaler) Unmarshal(data []byte, v any) error {
	return json.NewDecoder(bytes.NewBuffer(data)).Decode(v)
}

type CodecMsgpackMarshaler struct {
	mh *codec.MsgpackHandle
}

func NewCodecMsgpackMarshaler() *CodecMsgpackMarshaler {
	mh := new(codec.MsgpackHandle)
	mh.MapType = reflect.TypeOf(map[string]any(nil))
	mh.RawToString = true
	mh.TimeNotBuiltin = true
	return &CodecMsgpackMarshaler{mh}
}

func (cmm *CodecMsgpackMarshaler) Marshal(v any) (b []byte, err error) {
	enc := codec.NewEncoderBytes(&b, cmm.mh)
	err = enc.Encode(v)
	return
}

func (cmm *CodecMsgpackMarshaler) Unmarshal(data []byte, v any) error {
	dec := codec.NewDecoderBytes(data, cmm.mh)
	return dec.Decode(v)
}

type BincMarshaler struct {
	bh *codec.BincHandle
}

func NewBincMarshaler() *BincMarshaler {
	return &BincMarshaler{new(codec.BincHandle)}
}

func (bm *BincMarshaler) Marshal(v any) (b []byte, err error) {
	enc := codec.NewEncoderBytes(&b, bm.bh)
	err = enc.Encode(v)
	return
}

func (bm *BincMarshaler) Unmarshal(data []byte, v any) error {
	dec := codec.NewDecoderBytes(data, bm.bh)
	return dec.Decode(v)
}

type GOBMarshaler struct{}

func (gm *GOBMarshaler) Marshal(v any) (data []byte, err error) {
	buf := new(bytes.Buffer)
	err = gob.NewEncoder(buf).Encode(v)
	data = buf.Bytes()
	return
}

func (gm *GOBMarshaler) Unmarshal(data []byte, v any) error {
	return gob.NewDecoder(bytes.NewBuffer(data)).Decode(v)
}

// Decide the value of cacheCommit parameter based on transactionID
func cacheCommit(transactionID string) bool {
	return transactionID == utils.NonTransactional
}
