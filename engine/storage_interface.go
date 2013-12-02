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
	"labix.org/v2/mgo/bson"

	"reflect"
	"time"
)

const (
	ACTION_TIMING_PREFIX      = "atm_"
	RATING_PLAN_PREFIX        = "rpl_"
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
	// Some consts used in tests
	CREATE_CDRS_TABLES_SQL        = "create_cdrs_tables.sql"
	CREATE_COSTDETAILS_TABLES_SQL = "create_costdetails_tables.sql"
	CREATE_MEDIATOR_TABLES_SQL    = "create_mediator_tables.sql"
	CREATE_TARIFFPLAN_TABLES_SQL  = "create_tariffplan_tables.sql"
	TEST_SQL                      = "TEST_SQL"
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
	PreCache([]string, []string) error
	ExistsData(string, string) (bool, error)
	GetRatingPlan(string, bool) (*RatingPlan, error)
	SetRatingPlan(*RatingPlan) error
	GetRatingProfile(string) (*RatingProfile, error)
	SetRatingProfile(*RatingProfile) error
	GetDestination(string, bool) (*Destination, error)
	//	DestinationContainsPrefix(string, string) (int, error)
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
	GetRatedCdrs(time.Time, time.Time) ([]utils.CDR, error)
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
	RemTPData(string, string, ...string) error
	GetTPIds() ([]string, error)

	SetTPTiming(string, *utils.TPTiming) error
	GetTpTimings(string, string) (map[string]*utils.TPTiming, error)
	GetTPTimingIds(string) ([]string, error)

	SetTPDestination(string, *Destination) error
	GetTpDestinations(string, string) ([]*Destination, error)
	GetTPDestinationIds(string) ([]string, error)

	SetTPRates(string, map[string][]*utils.RateSlot) error
	GetTpRates(string, string) (map[string]*utils.TPRate, error)
	GetTPRateIds(string) ([]string, error)

	SetTPDestinationRates(string, map[string][]*utils.DestinationRate) error
	GetTpDestinationRates(string, string) (map[string]*utils.TPDestinationRate, error)
	GetTPDestinationRateIds(string) ([]string, error)

	SetTPRatingPlans(string, map[string][]*utils.TPRatingPlanBinding) error
	GetTpRatingPlans(string, string) (map[string][]*utils.TPRatingPlanBinding, error)
	GetTPRatingPlanIds(string) ([]string, error)

	SetTPRatingProfiles(string, map[string]*utils.TPRatingProfile) error
	GetTpRatingProfiles(*utils.TPRatingProfile) (map[string]*utils.TPRatingProfile, error)
	GetTPRatingProfileIds(*utils.AttrTPRatingProfileIds) ([]string, error)

	SetTPActions(string, map[string][]*utils.TPAction) error
	GetTpActions(string, string) (map[string][]*utils.TPAction, error)
	GetTPActionIds(string) ([]string, error)

	SetTPActionTimings(string, map[string][]*utils.TPActionTiming) error
	GetTPActionTimings(string, string) (map[string][]*utils.TPActionTiming, error)
	GetTPActionTimingIds(string) ([]string, error)

	SetTPActionTriggers(string, map[string][]*utils.TPActionTrigger) error
	GetTpActionTriggers(string, string) (map[string][]*utils.TPActionTrigger, error)
	GetTPActionTriggerIds(string) ([]string, error)

	SetTPAccountActions(string, map[string]*utils.TPAccountActions) error
	GetTpAccountActions(*utils.TPAccountActions) (map[string]*utils.TPAccountActions, error)
	GetTPAccountActionIds(string) ([]string, error)
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
