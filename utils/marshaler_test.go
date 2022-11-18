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

package utils

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/cgrates/ugocodec/codec"
	"go.mongodb.org/mongo-driver/bson"
)

func TestNewMarshaler(t *testing.T) {
	_, err := NewMarshaler(MsgPack)
	if err != nil {
		t.Error(err)
	}

	_, err = NewMarshaler(JSON)
	if err != nil {
		t.Error(err)
	}

	_, err = NewMarshaler("not_valid")
	errExp := "Unsupported marshaler: not_valid"
	if err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestJSONMarshaler(t *testing.T) {
	v := "text"
	jsnM := &JSONMarshaler{}
	rcv, err := jsnM.Marshal(v)
	if err != nil {
		t.Error(err)
	}
	var exp string
	json.Unmarshal([]byte(string(rcv)), &exp)
	if exp != v {
		t.Errorf("Expected %v\n but received %v", v, exp)
	}
}

func TestJSONMarshalerError(t *testing.T) {
	v := make(chan string)
	jsnM := &JSONMarshaler{}
	_, err := jsnM.Marshal(v)
	errExp := "json: unsupported type: chan string"
	if err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

type dummyStruct struct {
	Field string
}

func TestJSONUnmarshaler(t *testing.T) {
	data := []byte(`{"Field": "some_string"}`)
	jsnM := &JSONMarshaler{}
	dS := dummyStruct{
		Field: "some_string",
	}

	var ndS dummyStruct
	err := jsnM.Unmarshal(data, &ndS)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(dS, ndS) {
		t.Errorf("Expected: %s , received: %s", ToJSON(dS), ToJSON(ndS))
	}
}

func TestBSONMarshaler(t *testing.T) {
	v := bson.M{"ts": "test"}
	bsnM := &BSONMarshaler{}
	rcv, err := bsnM.Marshal(v)
	dS := dummyStruct{
		Field: "some_string",
	}
	if err != nil {
		t.Error(err)
	}
	if reflect.DeepEqual(dS, rcv) {
		t.Errorf("Expected: %s , received: %s", ToJSON(dS), ToJSON(rcv))
	}
}

func TestBSONMarshalerError(t *testing.T) {
	v := make(chan string)
	bsnM := &BSONMarshaler{}
	_, err := bsnM.Marshal(v)
	errExp := "no encoder found for chan string"
	if err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestJSONBufMarshaler(t *testing.T) {
	v := `text`
	jsnM := &JSONBufMarshaler{}
	rcv, err := jsnM.Marshal(v)
	if err != nil {
		t.Error(err)
	}
	var exp string
	json.Unmarshal([]byte(string(rcv)), &exp)
	if exp != v {
		t.Errorf("Expected %v\n but received %v", exp, v)
	}
}

func TestJSONBufUnmarshaler(t *testing.T) {
	v := []byte(`{"Field":"some_string"}`)
	jsnM := &JSONBufMarshaler{}
	s := dummyStruct{
		Field: "some_string",
	}
	var rcv dummyStruct
	err := jsnM.Unmarshal(v, &rcv)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, s) {
		t.Errorf("Expected %v\n but received %v", s, rcv)
	}
}
func TestBsonUnmarshal(t *testing.T) {
	v := bson.M{"ts": "test"}
	bsnM := &BSONMarshaler{}
	rcvM, _ := bsnM.Marshal(v)
	dS := dummyStruct{}

	var ndS dummyStruct
	err := bsnM.Unmarshal(rcvM, &ndS)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(dS, ndS) {
		t.Errorf("Expected: %s , received: %s", ToJSON(dS), ToJSON(ndS))
	}
}

func TestNewBincMarshler(t *testing.T) {
	exp := &BincMarshaler{new(codec.BincHandle)}
	if rcv := NewBincMarshaler(); !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected <%+v>, Received <%+v>", ToJSON(exp), ToJSON(rcv))
	}
}

func TestCodecMsgpackMarshalerMarshal(t *testing.T) {
	cmm := &CodecMsgpackMarshaler{&codec.MsgpackHandle{}}
	v := "texted"
	exp := []byte{166, 116, 101, 120, 116, 101, 100}
	if rcv, err := cmm.Marshal(v); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected <%v>, Received <%v>", exp, rcv)
	}

}

func TestCodecMsgpackMarshalerUnmarshal(t *testing.T) {
	cmm := &CodecMsgpackMarshaler{&codec.MsgpackHandle{}}
	data := []byte{116, 101, 100}
	v := "testv"
	expErr := "[pos 1]: invalid container type: expecting bin|str|array"
	if err := cmm.Unmarshal(data, v); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestBincMarshalerMarshal(t *testing.T) {
	bm := &BincMarshaler{&codec.BincHandle{}}
	v := "testinterface"
	exp := []byte{64, 13, 116, 101, 115, 116, 105, 110, 116, 101, 114, 102, 97, 99, 101}
	if rcv, err := bm.Marshal(v); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected <%v>, Received <%v>", exp, rcv)
	}
}
func TestBincMarshalerUnmarshal(t *testing.T) {
	bm := &BincMarshaler{&codec.BincHandle{}}
	v := "testinterce"
	data := []byte{64, 13, 11, 115, 116, 105, 110, 116, 101, 114, 102, 97, 99, 101}
	expErr := "unexpected EOF"
	if err := bm.Unmarshal(data, v); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestGOBMarshalerMarshal(t *testing.T) {
	v := "test"
	gobm := GOBMarshaler{}
	exp := []byte{7, 12, 0, 4, 116, 101, 115, 116}
	if rcv, err := gobm.Marshal(v); err != nil {
		t.Error(rcv, err)
	} else if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected <%v>, Received <%v>", exp, rcv)
	}
}

func TestGOBMarshalerUnmarshal(t *testing.T) {
	gobm := GOBMarshaler{}
	v := "testinterce"
	data := []byte{64, 13, 11, 115, 116, 105, 110, 116, 101, 114, 102, 97, 99, 101}
	expErr := "gob: attempt to decode into a non-pointer"
	if err := gobm.Unmarshal(data, v); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}
