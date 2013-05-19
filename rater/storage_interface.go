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
	gmsgpack "github.com/ugorji/go-msgpack"
	"github.com/vmihailenco/msgpack"
	"strings"
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
	// sources
	SESSION_MANAGER_SOURCE = "SMR"
	MEDIATOR_SOURCE        = "MED"
	SCHED_SOURCE           = "MED"
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
	GetActions(string) ([]*Action, error)
	SetActions(string, []*Action) error
	GetUserBalance(string) (*UserBalance, error)
	SetUserBalance(*UserBalance) error
	GetActionTimings(string) ([]*ActionTiming, error)
	SetActionTimings(string, []*ActionTiming) error
	GetAllActionTimings() (map[string][]*ActionTiming, error)
	GetCdr(string) (CDR, error)
	SetCdr(CDR) error
	//GetAllActionTimingsLogs() (map[string][]*ActionTiming, error)
	LogCallCost(uuid, source string, cc *CallCost) error
	LogError(uuid, source, errstr string) error
	LogActionTrigger(ubId, source string, at *ActionTrigger, as []*Action) error
	LogActionTiming(source string, at *ActionTiming, as []*Action) error
	GetCallCostLog(uuid, source string) (*CallCost, error)
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

type JSONBufMarshaler struct {
	buf bytes.Buffer
}

func (jbm *JSONBufMarshaler) Marshal(v interface{}) (data []byte, err error) {
	jbm.buf.Reset()
	if err = json.NewEncoder(&jbm.buf).Encode(v); err == nil {
		data = jbm.buf.Bytes()
	}
	return
}

func (jbm *JSONBufMarshaler) Unmarshal(data []byte, v interface{}) error {
	jbm.buf.Reset()
	jbm.buf.Write(data)
	return json.NewDecoder(&jbm.buf).Decode(v)
}

type MsgpackMarshaler struct{}

func (jm *MsgpackMarshaler) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (jm *MsgpackMarshaler) Unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}

type GoMsgpackMarshaler struct{}

func (jm *GoMsgpackMarshaler) Marshal(v interface{}) ([]byte, error) {
	return gmsgpack.Marshal(v)
}

func (jm *GoMsgpackMarshaler) Unmarshal(data []byte, v interface{}) error {
	return gmsgpack.Unmarshal(data, v, nil)
}

type GOBMarshaler struct {
	buf bytes.Buffer
}

func (gm *GOBMarshaler) Marshal(v interface{}) (data []byte, err error) {
	gm.buf.Reset()
	if err = gob.NewEncoder(&gm.buf).Encode(v); err == nil {
		data = gm.buf.Bytes()
	}
	return
}

func (gm *GOBMarshaler) Unmarshal(data []byte, v interface{}) error {
	gm.buf.Reset()
	gm.buf.Write(data)
	return gob.NewDecoder(&gm.buf).Decode(v)
}

type storer interface {
	store() string
	restore(string)
}

type MyMarshaler struct {
	buf bytes.Buffer
}

func (mm *MyMarshaler) Marshal(v interface{}) (data []byte, err error) {
	switch v.(type) {
	case []*Action:
		result := ""
		for _, a := range v.([]*Action) {
			result += a.store() + "~"
		}
		result = strings.TrimRight(result, "~")
		return []byte(result), nil
	case []*ActionTiming:
		result := ""
		for _, at := range v.([]*ActionTiming) {
			result += at.store() + "~"
		}
		result = strings.TrimRight(result, "~")
		return []byte(result), nil
	case storer:
		s := v.(storer)
		return []byte(s.store()), nil
	}
	mm.buf.Reset()
	if err = json.NewEncoder(&mm.buf).Encode(v); err == nil {
		data = mm.buf.Bytes()
	}
	return
}

func (mm *MyMarshaler) Unmarshal(data []byte, v interface{}) (err error) {
	switch v.(type) {
	case *[]*Action:
		as := v.(*[]*Action)
		for _, a_string := range strings.Split(string(data), "~") {
			if len(a_string) > 0 {
				a := &Action{}
				a.restore(a_string)
				*as = append(*as, a)
			}
		}
		return nil
	case *[]*ActionTiming:
		ats := v.(*[]*ActionTiming)
		for _, at_string := range strings.Split(string(data), "~") {
			if len(at_string) > 0 {
				at := &ActionTiming{}
				at.restore(at_string)
				*ats = append(*ats, at)
			}
		}
		return nil
	case storer:
		s := v.(storer)
		s.restore(string(data))
		return nil

	}
	mm.buf.Reset()
	mm.buf.Write(data)
	return json.NewDecoder(&mm.buf).Decode(v)
}
