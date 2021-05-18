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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ugocodec/codec"
	"go.mongodb.org/mongo-driver/bson"
)

type Storage interface {
	Close()
	Flush(string) error
	GetKeysForPrefix(ctx *context.Context, prefix string) ([]string, error)
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
	HasDataDrv(*context.Context, string, string, string) (bool, error)
	GetResourceProfileDrv(string, string) (*ResourceProfile, error)
	SetResourceProfileDrv(*ResourceProfile) error
	RemoveResourceProfileDrv(string, string) error
	GetResourceDrv(string, string) (*Resource, error)
	SetResourceDrv(*Resource) error
	RemoveResourceDrv(string, string) error
	GetLoadHistory(int, bool, string) ([]*utils.LoadInstance, error)
	AddLoadHistory(*utils.LoadInstance, int, string) error
	GetIndexesDrv(ctx *context.Context, idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error)
	SetIndexesDrv(ctx *context.Context, idxItmType, tntCtx string,
		indexes map[string]utils.StringSet, commit bool, transactionID string) (err error)
	RemoveIndexesDrv(idxItmType, tntCtx, idxKey string) (err error)
	GetStatQueueProfileDrv(tenant string, ID string) (sq *StatQueueProfile, err error)
	SetStatQueueProfileDrv(sq *StatQueueProfile) (err error)
	RemStatQueueProfileDrv(tenant, id string) (err error)
	GetStatQueueDrv(tenant, id string) (sq *StatQueue, err error)
	SetStatQueueDrv(ssq *StoredStatQueue, sq *StatQueue) (err error)
	RemStatQueueDrv(tenant, id string) (err error)
	GetThresholdProfileDrv(tenant string, ID string) (tp *ThresholdProfile, err error)
	SetThresholdProfileDrv(tp *ThresholdProfile) (err error)
	RemThresholdProfileDrv(tenant, id string) (err error)
	GetThresholdDrv(string, string) (*Threshold, error)
	SetThresholdDrv(*Threshold) error
	RemoveThresholdDrv(string, string) error
	GetFilterDrv(ctx *context.Context, tnt string, id string) (*Filter, error)
	SetFilterDrv(ctx *context.Context, f *Filter) error
	RemoveFilterDrv(string, string) error
	GetRouteProfileDrv(string, string) (*RouteProfile, error)
	SetRouteProfileDrv(*RouteProfile) error
	RemoveRouteProfileDrv(string, string) error
	GetAttributeProfileDrv(ctx *context.Context, tnt string, id string) (*AttributeProfile, error)
	SetAttributeProfileDrv(ctx *context.Context, attr *AttributeProfile) error
	RemoveAttributeProfileDrv(*context.Context, string, string) error
	GetChargerProfileDrv(string, string) (*ChargerProfile, error)
	SetChargerProfileDrv(*ChargerProfile) error
	RemoveChargerProfileDrv(string, string) error
	GetDispatcherProfileDrv(*context.Context, string, string) (*DispatcherProfile, error)
	SetDispatcherProfileDrv(*context.Context, *DispatcherProfile) error
	RemoveDispatcherProfileDrv(*context.Context, string, string) error
	GetItemLoadIDsDrv(itemIDPrefix string) (loadIDs map[string]int64, err error)
	SetLoadIDsDrv(ctx *context.Context, loadIDs map[string]int64) error
	RemoveLoadIDsDrv() error
	GetDispatcherHostDrv(string, string) (*DispatcherHost, error)
	SetDispatcherHostDrv(*DispatcherHost) error
	RemoveDispatcherHostDrv(string, string) error
	GetRateProfileDrv(*context.Context, string, string) (*utils.RateProfile, error)
	SetRateProfileDrv(*context.Context, *utils.RateProfile) error
	RemoveRateProfileDrv(*context.Context, string, string) error
	GetActionProfileDrv(*context.Context, string, string) (*ActionProfile, error)
	SetActionProfileDrv(*context.Context, *ActionProfile) error
	RemoveActionProfileDrv(*context.Context, string, string) error
	GetAccountDrv(*context.Context, string, string) (*utils.Account, error)
	SetAccountDrv(ctx *context.Context, profile *utils.Account) error
	RemoveAccountDrv(*context.Context, string, string) error
}

// DataDBDriver used as a DataDB but also as a ConfigProvider
type DataDBDriver interface {
	DataDB
	config.ConfigDB
}

type StorDB interface {
	CdrStorage
	LoadReader
	LoadWriter
}

type CdrStorage interface {
	Storage
	SetCDR(*CDR, bool) error
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
	GetTpTableIds(string, string, []string,
		map[string]string, *utils.PaginatorWithSearch) ([]string, error)
	GetTPResources(string, string, string) ([]*utils.TPResourceProfile, error)
	GetTPStats(string, string, string) ([]*utils.TPStatProfile, error)
	GetTPThresholds(string, string, string) ([]*utils.TPThresholdProfile, error)
	GetTPFilters(string, string, string) ([]*utils.TPFilterProfile, error)
	GetTPRoutes(string, string, string) ([]*utils.TPRouteProfile, error)
	GetTPAttributes(string, string, string) ([]*utils.TPAttributeProfile, error)
	GetTPChargers(string, string, string) ([]*utils.TPChargerProfile, error)
	GetTPDispatcherProfiles(string, string, string) ([]*utils.TPDispatcherProfile, error)
	GetTPDispatcherHosts(string, string, string) ([]*utils.TPDispatcherHost, error)
	GetTPRateProfiles(string, string, string) ([]*utils.TPRateProfile, error)
	GetTPActionProfiles(string, string, string) ([]*utils.TPActionProfile, error)
	GetTPAccounts(string, string, string) ([]*utils.TPAccount, error)
}

type LoadWriter interface {
	RemTpData(string, string, map[string]string) error
	SetTPResources([]*utils.TPResourceProfile) error
	SetTPStats([]*utils.TPStatProfile) error
	SetTPThresholds([]*utils.TPThresholdProfile) error
	SetTPFilters([]*utils.TPFilterProfile) error
	SetTPRoutes([]*utils.TPRouteProfile) error
	SetTPAttributes([]*utils.TPAttributeProfile) error
	SetTPChargers([]*utils.TPChargerProfile) error
	SetTPDispatcherProfiles([]*utils.TPDispatcherProfile) error
	SetTPDispatcherHosts([]*utils.TPDispatcherHost) error
	SetTPRateProfiles([]*utils.TPRateProfile) error
	SetTPActionProfiles([]*utils.TPActionProfile) error
	SetTPAccounts([]*utils.TPAccount) error
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
	return transactionID == utils.NonTransactional
}
