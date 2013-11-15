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
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
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
	ExistsData(string, string)(bool, error)
	GetRatingPlan(string) (*RatingPlan, error)
	SetRatingPlan(*RatingPlan) error
	GetRatingProfile(string) (*RatingProfile, error)
	SetRatingProfile(*RatingProfile) error
	GetDestination(string) (*Destination, error)
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
	GetTPIds() ([]string, error)
	SetTPTiming(string, *utils.TPTiming) error
	ExistsTPTiming(string, string) (bool, error)
	GetTPTiming(string, string) (*utils.TPTiming, error)
	GetTPTimingIds(string) ([]string, error)
	RemTPData(string, string, string) error
	SetTPDestination(string, *Destination) error
	ExistsTPDestination(string, string) (bool, error)
	GetTPDestination(string, string) (*Destination, error)
	GetTPDestinationIds(string) ([]string, error)
	ExistsTPRate(string, string) (bool, error)
	SetTPRates(string, map[string][]*utils.RateSlot) error
	GetTPRate(string, string) (*utils.TPRate, error)
	GetTPRateIds(string) ([]string, error)
	ExistsTPDestinationRate(string, string) (bool, error)
	SetTPDestinationRates(string, map[string][]*utils.DestinationRate) error
	GetTPDestinationRate(string, string) (*utils.TPDestinationRate, error)
	GetTPDestinationRateIds(string) ([]string, error)
	ExistsTPRatingPlan(string, string) (bool, error)
	SetTPRatingPlans(string, map[string][]*utils.RatingPlan) error
	GetTPRatingPlan(string, string) (*utils.TPRatingPlan, error)
	GetTPRatingPlanIds(string) ([]string, error)
	ExistsTPRatingProfile(string, string) (bool, error)
	SetTPRatingProfiles(string, map[string][]*utils.TPRatingProfile) error
	GetTPRatingProfile(string, string) (*utils.TPRatingProfile, error)
	GetTPRatingProfileIds(*utils.AttrTPRatingProfileIds) ([]string, error)
	ExistsTPActions(string, string) (bool, error)
	SetTPActions(string, map[string][]*Action) error
	GetTPActions(string, string) (*utils.TPActions, error)
	GetTPActionIds(string) ([]string, error)
	ExistsTPActionTimings(string, string) (bool, error)
	SetTPActionTimings(string, map[string][]*utils.ApiActionTiming) error
	GetTPActionTimings(string, string) (map[string][]*utils.ApiActionTiming, error)
	GetTPActionTimingIds(string) ([]string, error)
	ExistsTPActionTriggers(string, string) (bool, error)
	SetTPActionTriggers(string, map[string][]*ActionTrigger) error
	GetTPActionTriggerIds(string) ([]string, error)
	ExistsTPAccountActions(string, string) (bool, error)
	SetTPAccountActions(string, map[string]*AccountAction) error
	GetTPAccountActionIds(string) ([]string, error)
	// loader functions
	GetTpDestinations(string, string) ([]*Destination, error)
	GetTpTimings(string, string) (map[string]*utils.TPTiming, error)
	GetTpRates(string, string) (map[string]*utils.TPRate, error)
	GetTpDestinationRates(string, string) (map[string]*utils.TPDestinationRate, error)
	GetTpRatingPlans(string, string) (*utils.TPRatingPlan, error)
	GetTpRatingProfiles(string, string) (map[string]*utils.TPRatingProfile, error)
	GetTpActions(string, string) (map[string][]*Action, error)
	GetTpActionTimings(string, string) (map[string][]*utils.ApiActionTiming, error)
	GetTpActionTriggers(string, string) (map[string][]*utils.ApiActionTrigger, error)
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

type CodecMsgpackMarshaler struct {
	mh *codec.MsgpackHandle
}

/*** The following functions to be removed after go 1.2 ****/

var (
	bigen     = binary.BigEndian
	bsAll0xff = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
)

// TimeEncodeExt encodes a time.Time as a byte slice.
// Configure this to support the Time Extension, e.g. using tag 1.
func TimeEncodeExt(rv reflect.Value) (bs []byte, err error) {
	rvi := rv.Interface()
	switch iv := rvi.(type) {
	case time.Time:
		bs = encodeTime(iv)
	default:
		err = fmt.Errorf("codec/msgpack: TimeEncodeExt expects a time.Time. Received %T", rvi)
	}
	return
}

// TimeDecodeExt decodes a time.Time from the byte slice parameter, and sets it into the reflect value.
// Configure this to support the Time Extension, e.g. using tag 1.
func TimeDecodeExt(rv reflect.Value, bs []byte) (err error) {
	tt, err := decodeTime(bs)
	if err == nil {
		rv.Set(reflect.ValueOf(tt))
	}
	return
}
func pruneSignExt(v []byte) (n int) {
	l := len(v)
	if l < 2 {
		return
	}
	if v[0] == 0 {
		n2 := n + 1
		for v[n] == 0 && n2 < l && (v[n2]&(1<<7) == 0) {
			n++
			n2++
		}
		return
	}
	if v[0] == 0xff {
		n2 := n + 1
		for v[n] == 0xff && n2 < l && (v[n2]&(1<<7) != 0) {
			n++
			n2++
		}
		return
	}
	return
}

func encodeTime(t time.Time) []byte {
	//t := rv.Interface().(time.Time)
	tsecs, tnsecs := t.Unix(), t.Nanosecond()
	var (
		bd   byte
		btmp [8]byte
		bs   [16]byte
		i    int = 1
	)
	l := t.Location()
	if l == time.UTC {
		l = nil
	}
	if tsecs != 0 {
		bd = bd | 0x80
		bigen.PutUint64(btmp[:], uint64(tsecs))
		f := pruneSignExt(btmp[:])
		bd = bd | (byte(7-f) << 2)
		copy(bs[i:], btmp[f:])
		i = i + (8 - f)
	}
	if tnsecs != 0 {
		bd = bd | 0x40
		bigen.PutUint32(btmp[:4], uint32(tnsecs))
		f := pruneSignExt(btmp[:4])
		bd = bd | byte(3-f)
		copy(bs[i:], btmp[f:4])
		i = i + (4 - f)
	}
	if l != nil {
		bd = bd | 0x20
		// Note that Go Libs do not give access to dst flag.
		_, zoneOffset := t.Zone()
		//zoneName, zoneOffset := t.Zone()
		zoneOffset /= 60
		z := uint16(zoneOffset)
		bigen.PutUint16(btmp[:2], z)
		// clear dst flags
		bs[i] = btmp[0] & 0x3f
		bs[i+1] = btmp[1]
		i = i + 2
	}
	bs[0] = bd
	return bs[0:i]
}

// DecodeTime decodes a []byte into a time.Time.
func decodeTime(bs []byte) (tt time.Time, err error) {
	bd := bs[0]
	var (
		tsec  int64
		tnsec uint32
		tz    uint16
		i     byte = 1
		i2    byte
		n     byte
	)
	if bd&(1<<7) != 0 {
		var btmp [8]byte
		n = ((bd >> 2) & 0x7) + 1
		i2 = i + n
		copy(btmp[8-n:], bs[i:i2])
		//if first bit of bs[i] is set, then fill btmp[0..8-n] with 0xff (ie sign extend it)
		if bs[i]&(1<<7) != 0 {
			copy(btmp[0:8-n], bsAll0xff)
			//for j,k := byte(0), 8-n; j < k; j++ {	btmp[j] = 0xff }
		}
		i = i2
		tsec = int64(bigen.Uint64(btmp[:]))
	}
	if bd&(1<<6) != 0 {
		var btmp [4]byte
		n = (bd & 0x3) + 1
		i2 = i + n
		copy(btmp[4-n:], bs[i:i2])
		i = i2
		tnsec = bigen.Uint32(btmp[:])
	}
	if bd&(1<<5) == 0 {
		tt = time.Unix(tsec, int64(tnsec)).UTC()
		return
	}
	// In stdlib time.Parse, when a date is parsed without a zone name, it uses "" as zone name.
	// However, we need name here, so it can be shown when time is printed.
	// Zone name is in form: UTC-08:00.
	// Note that Go Libs do not give access to dst flag, so we ignore dst bits

	i2 = i + 2
	tz = bigen.Uint16(bs[i:i2])
	i = i2
	// sign extend sign bit into top 2 MSB (which were dst bits):
	if tz&(1<<13) == 0 { // positive
		tz = tz & 0x3fff //clear 2 MSBs: dst bits
	} else { // negative
		tz = tz | 0xc000 //set 2 MSBs: dst bits
		//tzname[3] = '-' (TODO: verify. this works here)
	}
	tzint := int16(tz)
	if tzint == 0 {
		tt = time.Unix(tsec, int64(tnsec)).UTC()
	} else {
		// For Go Time, do not use a descriptive timezone.
		// It's unnecessary, and makes it harder to do a reflect.DeepEqual.
		// The Offset already tells what the offset should be, if not on UTC and unknown zone name.
		// var zoneName = timeLocUTCName(tzint)
		tt = time.Unix(tsec, int64(tnsec)).In(time.FixedZone("", int(tzint)*60))
	}
	return
}

/*** end remove here ***/

func NewCodecMsgpackMarshaler() *CodecMsgpackMarshaler {
	cmm := &CodecMsgpackMarshaler{new(codec.MsgpackHandle)}
	mh := cmm.mh

	// configure extensions for msgpack, to enable Binary and Time support for tags 0 and 1
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	//mh.AddExt( reflect.TypeOf([]byte(nil)), 0, mh.BinaryEncodeExt, mh.BinaryDecodeExt)
	//	mh.AddExt(reflect.TypeOf(time.Time{}), 1, mh.TimeEncodeExt, mh.TimeDecodeExt)
	mh.AddExt(reflect.TypeOf(time.Time{}), 1, TimeEncodeExt, TimeDecodeExt)
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
