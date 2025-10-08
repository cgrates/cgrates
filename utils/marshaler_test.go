/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/ugorji/go/codec"
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
	expErr := "msgpack decode error [pos 1]: invalid byte descriptor for decoding bytes, got: 0x74"
	if err := cmm.Unmarshal(data, &v); err == nil || err.Error() != expErr {
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
	if err := bm.Unmarshal(data, &v); err != io.ErrUnexpectedEOF {
		t.Errorf("Expected error <%v>, Received <%v>", io.ErrUnexpectedEOF, err)
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

func TestMarshalerDecodeIntoNilIface(t *testing.T) {
	ms := NewCodecMsgpackMarshaler()

	mp := map[string]any{
		"key1": "value1",
		"key2": 2.,
	}
	expBytes1 := []byte{130, 164, 107, 101, 121, 49, 166, 118, 97, 108, 117, 101, 49, 164, 107, 101, 121, 50, 203, 64, 0, 0, 0, 0, 0, 0, 0}
	expBytes2 := []byte{130, 164, 107, 101, 121, 50, 203, 64, 0, 0, 0, 0, 0, 0, 0, 164, 107, 101, 121, 49, 166, 118, 97, 108, 117, 101, 49}
	b, err := ms.Marshal(mp)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, expBytes1) && !bytes.Equal(b, expBytes2) {
		t.Fatalf("expected: %+v or\n%+v,\nreceived: %+v", expBytes1, expBytes2, b)
	}

	decodedMap := make(map[string]any)
	err = ms.Unmarshal(b, &decodedMap)
	if err != nil {
		t.Fatal(err)
	}
	for key, value := range decodedMap {
		if value != mp[key] {
			t.Fatalf("for key %s, expected: %+v,\nreceived: %+v",
				key, mp[key], value)
		}
	}
}

func TestMarshalerMsgPackDecode(t *testing.T) {
	type stc struct {
		Name string
	}

	type structWithTime struct {
		Date time.Time
	}

	var s stc
	mp := make(map[string]any)
	var slc []string
	var slcB []byte
	var arr [1]int
	var nm int
	var fl float64
	var str string
	var bl bool
	var td time.Duration
	var tm time.Time
	stcWithTime := structWithTime{}

	tests := []struct {
		name     string
		expBytes []byte
		val      any
		decode   any
		rng      bool
	}{
		{
			name:     "map",
			expBytes: []byte{129, 164, 107, 101, 121, 49, 166, 118, 97, 108, 117, 101, 49},
			val:      map[string]any{"key1": "value1"},
			decode:   mp,
			rng:      true,
		},
		{
			name:     "int",
			expBytes: []byte{1},
			val:      1,
			decode:   nm,
			rng:      false,
		},
		{
			name:     "string",
			expBytes: []byte{164, 116, 101, 115, 116},
			val:      "test",
			decode:   str,
			rng:      false,
		},
		{
			name:     "float64",
			expBytes: []byte{203, 63, 248, 0, 0, 0, 0, 0, 0},
			val:      1.5,
			decode:   fl,
			rng:      false,
		},
		{
			name:     "boolean",
			expBytes: []byte{195},
			val:      true,
			decode:   bl,
			rng:      false,
		},
		{
			name:     "slice",
			expBytes: []byte{145, 164, 118, 97, 108, 49},
			val:      []string{"val1"},
			decode:   slc,
			rng:      true,
		},
		{
			name:     "array",
			expBytes: []byte{145, 1},
			val:      [1]int{1},
			decode:   arr,
			rng:      true,
		},
		{
			name:     "struct",
			expBytes: []byte{129, 164, 78, 97, 109, 101, 164, 116, 101, 115, 116},
			val:      stc{"test"},
			decode:   s,
			rng:      true,
		},
		{
			name:     "time duration",
			expBytes: []byte{210, 59, 154, 202, 0},
			val:      1 * time.Second,
			decode:   td,
			rng:      false,
		},
		{
			name:     "slice of bytes",
			expBytes: []byte{162, 5, 8},
			val:      []byte{5, 8},
			decode:   slcB,
			rng:      true,
		},
		{
			name:     "time.Time",
			expBytes: []byte{168, 41, 96, 67, 92, 100, 165, 51, 220},
			val:      time.Date(2023, 7, 5, 9, 11, 56, 173543639, time.UTC), // 2023-07-05 09:11:56.173543639 UTC
			decode:   tm,
			rng:      true,
		},
		{
			name:     "struct with time.Time",
			expBytes: []byte{129, 164, 68, 97, 116, 101, 168, 41, 96, 67, 92, 100, 165, 51, 220},
			val: structWithTime{
				Date: time.Date(2023, 7, 5, 9, 11, 56, 173543639, time.UTC), // 2023-07-05 09:11:56.173543639 UTC
			},
			decode: stcWithTime,
			rng:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := NewCodecMsgpackMarshaler()

			b, err := ms.Marshal(tt.val)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(b, tt.expBytes) {
				t.Fatalf("expected: %+v,\nreceived: %+v", tt.expBytes, b)
			}

			err = ms.Unmarshal(b, &tt.decode)
			if err != nil {
				t.Fatal(err)
			}

			if tt.rng {
				if !reflect.DeepEqual(tt.decode, tt.val) {
					t.Errorf("expected %v, received %v", tt.val, tt.decode)
				}
			} else {
				if tt.decode != tt.val {
					t.Errorf("expected %v, received %v", tt.val, tt.decode)
				}
			}
		})
	}
}
