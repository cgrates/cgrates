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
	"bytes"
	"encoding/gob"
	"encoding/json"
	"github.com/cgrates/cgrates/utils"
	"github.com/ugorji/go/codec"
	"github.com/vmihailenco/msgpack"
	"labix.org/v2/mgo/bson"
)

const (
	ACTION_TIMING_PREFIX      = "atm_"
	RATING_PROFILE_PREFIX     = "rpf_"
	ACTION_PREFIX             = "act_"
	USER_BALANCE_PREFIX       = "ubl_"
	DESTINATION_PREFIX        = "dst_"
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

/*
Interface for storage providers.
*/
type DataStorage interface {
	Close()
	Flush() error
	GetRatingProfile(string) (*RatingProfile, error)
	SetRatingProfile(*RatingProfile) error
	GetDestination(string) (*Destination, error)
	SetDestination(*Destination) error
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
	SetTPRate(*utils.TPRate) error
	GetTPRate(string, string) (*utils.TPRate, error)
	GetTPRateIds(string) ([]string, error)
	// End Apier functions
	GetActions(string) (Actions, error)
	SetActions(string, Actions) error
	GetUserBalance(string) (*UserBalance, error)
	SetUserBalance(*UserBalance) error
	GetActionTimings(string) (ActionTimings, error)
	SetActionTimings(string, ActionTimings) error
	GetAllActionTimings() (map[string]ActionTimings, error)
	SetCdr(utils.CDR) error
	SetRatedCdr(utils.CDR, *CallCost) error
	GetAllRatedCdr() ([]utils.CDR, error)
	//GetAllActionTimingsLogs() (map[string]ActionsTimings, error)
	LogCallCost(uuid, source string, cc *CallCost) error
	LogError(uuid, source, errstr string) error
	LogActionTrigger(ubId, source string, at *ActionTrigger, as Actions) error
	LogActionTiming(source string, at *ActionTiming, as Actions) error
	GetCallCostLog(uuid, source string) (*CallCost, error)
	// loader functions
	GetTpDestinations(string, string) ([]*Destination, error)
	GetTpTimings(string, string) (map[string]*Timing, error)
	GetTpRates(string, string) (map[string]*Rate, error)
	GetTpDestinationRates(string, string) (map[string][]*DestinationRate, error)
	GetTpDestinationRateTimings(string, string) ([]*DestinationRateTiming, error)
	GetTpRatingProfiles(string, string) (map[string]*RatingProfile, error)
	GetTpActions(string, string) (map[string][]*Action, error)
	GetTpActionTimings(string, string) (map[string][]*ActionTiming, error)
	GetTpActionTriggers(string, string) (map[string][]*ActionTrigger, error)
	GetTpAccountActions(string, string) ([]*AccountAction, error)
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

type GoMsgpackMarshaler struct {
	mh *codec.MsgpackHandle
}

type BincMarshaler struct {
	bh *codec.BincHandle
}

func NewBincMarshaler() *BincMarshaler {
	return &BincMarshaler{&codec.BincHandle{}}
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

type MyMarshaler struct{}

func (mm *MyMarshaler) Marshal(v interface{}) ([]byte, error) {
	ser := v.(Serializer)
	res, err := ser.Store()
	return []byte(res), err
}

func (mm *MyMarshaler) Unmarshal(data []byte, v interface{}) error {
	ser := v.(Serializer)
	return ser.Restore(string(data))
}
