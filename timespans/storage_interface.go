/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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

package timespans

import (
	"encoding/json"
	"encoding/gob"
	"bytes"
	//"log"
)

/*
Interface for storage providers.
*/
type StorageGetter interface {
	Close()
	Flush() error
	GetActivationPeriodsOrFallback(string) ([]*ActivationPeriod, string, error)
	SetActivationPeriodsOrFallback(string, []*ActivationPeriod, string) error
	GetDestination(string) (*Destination, error)
	SetDestination(*Destination) error
	GetActions(string) ([]*Action, error)
	SetActions(string, []*Action) error
	GetUserBalance(string) (*UserBalance, error)
	SetUserBalance(*UserBalance) error
	GetActionTimings(string) ([]*ActionTiming, error)
	SetActionTimings(string, []*ActionTiming) error
	GetAllActionTimings() ([]*ActionTiming, error)
}

type Marshaler interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

type MarshalStrategy struct {
	marshaler Marshaler
}

func (ms *MarshalStrategy) SetMarshaler(m Marshaler) {
	ms.marshaler = m
}

func (ms *MarshalStrategy) Marshal(v interface{}) ([]byte, error) {
	return ms.marshaler.Marshal(v)
}

func (ms *MarshalStrategy) Unmarshal(data []byte, v interface{}) error {
	return ms.marshaler.Unmarshal(data, v)
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
