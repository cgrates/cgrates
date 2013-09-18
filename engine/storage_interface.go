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
	"bytes"
	"encoding/gob"
	"encoding/json"
	"github.com/cgrates/cgrates/utils"
	"github.com/ugorji/go/codec"
	"github.com/vmihailenco/msgpack"
	"labix.org/v2/mgo/bson"
	"reflect"
	"time"
)

const (
	ACTION_TIMING_PREFIX      = "atm_"
	RATING_PROFILE_PREFIX     = "rpf_"
	ACTION_PREFIX             = "act_"
	USER_BALANCE_PREFIX       = "ubl_"
	DESTINATION_PREFIX        = "dst_"
	TEMP_DESTINATION_PREFIX   = "tmp_"
	LOG_CALL_COST_PREFIX      = "cco_"
	LOG_ACTION_TIMMING_PREFIX = "ltm_"
	LOG_ACTION_TRIGGER_PREFIX = "ltr_"
	LOG_ERR                   = "ler_"
	LOG_CDR                   = "cdr_"
	LOG_MEDIATED_CDR          = "mcd_"
	// sources
	SESSION_MANAGER_SOURCE = "SMR"
	MEDIATOR_SOURCE        = "MED"
	SCHED_SOURCE           = "SCH"
	RATER_SOURCE           = "RAT"
)

var (
	// for codec msgpack
	mapStrIntfTyp = reflect.TypeOf(map[string]interface{}(nil))
	sliceByteTyp  = reflect.TypeOf([]byte(nil))
	timeTyp       = reflect.TypeOf(time.Time{})
)

type Storage interface {
	Close()
	Flush() error
}

/*
Interface for storage providers.
*/
type DataStorage interface {
	Storage
	GetRatingProfile(string) (*RatingProfile, error)
	SetRatingProfile(*RatingProfile) error
	GetDestination(string) (*Destination, error)
	DestinationContainsPrefix(string, string) (int, error)
	SetDestination(*Destination) error
	GetActions(string) (Actions, error)
	SetActions(string, Actions) error
	GetUserBalance(string) (*UserBalance, error)
	SetUserBalance(*UserBalance) error
	GetActionTimings(string) (ActionTimings, error)
	SetActionTimings(string, ActionTimings) error
	GetAllActionTimings() (map[string]ActionTimings, error)
}

type CdrStorage interface {
	Storage
	SetCdr(utils.CDR) error
	SetRatedCdr(utils.CDR, *CallCost, string) error
	GetAllRatedCdr() ([]utils.CDR, error)
}

type LogStorage interface {
	Storage
	//GetAllActionTimingsLogs() (map[string]ActionsTimings, error)
	LogCallCost(uuid, source string, cc *CallCost) error
	LogError(uuid, source, errstr string) error
	LogActionTrigger(ubId, source string, at *ActionTrigger, as Actions) error
	LogActionTiming(source string, at *ActionTiming, as Actions) error
	GetCallCostLog(uuid, source string) (*CallCost, error)
}

type LoadStorage interface {
	Storage
	// Apier functions
	GetTPIds() ([]string, error)
	SetTPTiming(string, *Timing) error
	ExistsTPTiming(string, string) (bool, error)
	GetTPTiming(string, string) (*Timing, error)
	GetTPTimingIds(string) ([]string, error)
	SetTPDestination(string, *Destination) error
	ExistsTPDestination(string, string) (bool, error)
	GetTPDestination(string, string) (*Destination, error)
	GetTPDestinationIds(string) ([]string, error)
	ExistsTPRate(string, string) (bool, error)
	SetTPRates(string, map[string][]*LoadRate) error
	GetTPRate(string, string) (*utils.TPRate, error)
	GetTPRateIds(string) ([]string, error)
	ExistsTPDestinationRate(string, string) (bool, error)
	SetTPDestinationRates(string, map[string][]*DestinationRate) error
	GetTPDestinationRate(string, string) (*utils.TPDestinationRate, error)
	GetTPDestinationRateIds(string) ([]string, error)
	ExistsTPDestRateTiming(string, string) (bool, error)
	SetTPDestRateTimings(string, map[string][]*DestinationRateTiming) error
	GetTPDestRateTiming(string, string) (*utils.TPDestRateTiming, error)
	GetTPDestRateTimingIds(string) ([]string, error)
	ExistsTPRatingProfile(string, string) (bool, error)
	SetTPRatingProfiles(string, map[string][]*RatingProfile) error
	GetTPRatingProfile(string, string) (*utils.TPRatingProfile, error)
	GetTPRatingProfileIds(*utils.AttrTPRatingProfileIds) ([]string, error)
	ExistsTPActions(string, string) (bool, error)
	SetTPActions(string, map[string][]*Action) error
	GetTPActions(string, string) (*utils.TPActions, error)
	GetTPActionIds(string) ([]string, error)
	ExistsTPActionTimings(string, string) (bool, error)
	SetTPActionTimings(string, map[string][]*ActionTiming) error
	GetTPActionTimings(string, string) (map[string][]*utils.TPActionTimingsRow, error)
	GetTPActionTimingIds(string) ([]string, error)
	ExistsTPActionTriggers(string, string) (bool, error)
	SetTPActionTriggers(string, map[string][]*ActionTrigger) error
	GetTPActionTriggerIds(string) ([]string, error)
	ExistsTPAccountActions(string, string) (bool, error)
	SetTPAccountActions(string, map[string]*AccountAction) error
	GetTPAccountActionIds(string) ([]string, error)
	// loader functions
	GetTpDestinations(string, string) ([]*Destination, error)
	GetTpTimings(string, string) (map[string]*Timing, error)
	GetTpRates(string, string) (map[string]*LoadRate, error)
	GetTpDestinationRates(string, string) (map[string][]*DestinationRate, error)
	GetTpDestinationRateTimings(string, string) ([]*DestinationRateTiming, error)
	GetTpRatingProfiles(string, string) (map[string]*RatingProfile, error)
	GetTpActions(string, string) (map[string][]*Action, error)
	GetTpActionTimings(string, string) (map[string][]*ActionTiming, error)
	GetTpActionTriggers(string, string) (map[string][]*ActionTrigger, error)
	GetTpAccountActions(string, string) (map[string]*AccountAction, error)
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

type MsgpackMarshaler struct{}

func (jm *MsgpackMarshaler) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (jm *MsgpackMarshaler) Unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}

type CodecMsgpackMarshaler struct {
	mh *codec.MsgpackHandle
}

func NewCodecMsgpackMarshaler() *CodecMsgpackMarshaler {
	cmm := &CodecMsgpackMarshaler{new(codec.MsgpackHandle)}
	mh := cmm.mh
	mh.MapType = mapStrIntfTyp

	// configure extensions for msgpack, to enable Binary and Time support for tags 0 and 1
	mh.AddExt(sliceByteTyp, 0, mh.BinaryEncodeExt, mh.BinaryDecodeExt)
	mh.AddExt(timeTyp, 1, mh.TimeEncodeExt, mh.TimeDecodeExt)
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
