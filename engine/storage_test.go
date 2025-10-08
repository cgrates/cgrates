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
package engine

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestMsgpackStructsAdded(t *testing.T) {
	var a = struct{ First string }{"test"}
	var b = struct {
		First  string
		Second string
	}{}
	m := utils.NewCodecMsgpackMarshaler()
	buf, err := m.Marshal(&a)
	if err != nil {
		t.Error("error marshaling structure: ", err)
	}
	err = m.Unmarshal(buf, &b)
	if err != nil || b.First != "test" || b.Second != "" {
		t.Error("error unmarshalling structure: ", b, err)
	}
}

func TestMsgpackStructsMissing(t *testing.T) {
	var a = struct {
		First  string
		Second string
	}{"test1", "test2"}
	var b = struct{ First string }{}
	m := utils.NewCodecMsgpackMarshaler()
	buf, err := m.Marshal(&a)
	if err != nil {
		t.Error("error marshaling structure: ", err)
	}
	err = m.Unmarshal(buf, &b)
	if err != nil || b.First != "test1" {
		t.Error("error unmarshalling structure: ", b, err)
	}
}

func TestMsgpackTime(t *testing.T) {
	t1 := time.Date(2013, 8, 28, 22, 27, 30, 11, time.UTC)
	m := utils.NewCodecMsgpackMarshaler()
	buf, err := m.Marshal(&t1)
	if err != nil {
		t.Error("error marshaling structure: ", err)
	}
	var t2 time.Time
	err = m.Unmarshal(buf, &t2)
	if err != nil || t1 != t2 || !t1.Equal(t2) {
		t.Errorf("error unmarshalling structure: %#v %#v %v", t1, t2, err)
	}
}

// Install fails to detect them and starting server will panic, these tests will fix this
func TestStoreInterfaces(t *testing.T) {
	rds := new(RedisStorage)
	var _ DataDB = rds
}

func TestStorageDecodeCodecMsgpackMarshaler(t *testing.T) {
	type stc struct {
		Name string
	}

	var s stc
	mp := make(map[string]any)
	var slc []string
	var slcB []byte
	var arr *[1]int
	var nm int
	var fl float64
	var str string
	var bl bool
	var td time.Duration

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
			val:      &[1]int{1},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := utils.NewCodecMsgpackMarshaler()

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

func TestComposeURI(t *testing.T) {
	tests := []struct {
		name     string
		scheme   string
		host     string
		port     string
		db       string
		user     string
		pass     string
		expected string
		parseErr bool
	}{
		{
			name:     "multiple nodes",
			scheme:   "mongodb",
			host:     "clusternode1:1230,clusternode2:1231,clusternode3",
			port:     "1232",
			db:       "cgrates",
			user:     "user",
			pass:     "pass",
			expected: "mongodb://user:pass@clusternode1:1230,clusternode2:1231,clusternode3:1232/cgrates",
		},
		{
			name:     "no port",
			scheme:   "mongodb",
			host:     "localhost:1234",
			port:     "0",
			db:       "cgrates",
			user:     "user",
			pass:     "pass",
			expected: "mongodb://user:pass@localhost:1234/cgrates",
		},
		{
			name:     "with port",
			scheme:   "mongodb",
			host:     "localhost",
			port:     "1234",
			db:       "cgrates",
			user:     "user",
			pass:     "pass",
			expected: "mongodb://user:pass@localhost:1234/cgrates",
		},
		{
			name:     "no password",
			scheme:   "mongodb",
			host:     "localhost",
			port:     "1234",
			db:       "cgrates",
			user:     "user",
			pass:     "",
			expected: "mongodb://localhost:1234/cgrates",
		},
		{
			name:     "no db",
			scheme:   "mongodb",
			host:     "localhost",
			port:     "1234",
			db:       "",
			user:     "user",
			pass:     "pass",
			expected: "mongodb://user:pass@localhost:1234",
		},
		{
			name:     "different scheme",
			scheme:   "mongodb+srv",
			host:     "cgr.abcdef.mongodb.net",
			port:     "0",
			db:       "?retryWrites=true&w=majority",
			user:     "user",
			pass:     "pass",
			expected: "mongodb+srv://user:pass@cgr.abcdef.mongodb.net/?retryWrites=true&w=majority",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := composeMongoURI(tt.scheme, tt.host, tt.port, tt.db, tt.user, tt.pass)
			if url != tt.expected {
				t.Errorf("expected %v,\nreceived %v", tt.expected, url)
			}
		})
	}
}
