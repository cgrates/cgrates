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
package config

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

var (
	envStr = `{//nonprocess string
		/***********************************************/

	// Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
	// Copyright (C) ITsysCOM GmbH
	//
	// This file contains the default configuration hardcoded into CGRateS.
	// This is what you get when you load CGRateS with an empty configuration file.
	"data_db": {								// database used to store runtime data (eg: accounts, cdr stats)
		"db_type": "redis",						// data_db type: <*redis|*mongo|*internal>
		"db_host": "127.0.0.1",					/* data_db host address*/
		"db_port": 6379, 						// data_db port to reach the database
		"db_name": "10",/*/*asd*/ 						// data_db database name to connect to
		"db_user": "*env:TESTVAR", 					// username to use when connecting to data_db
		"db_password": ",/**/", 						// password to use when connecting to data_db
		"redis_sentinel":"",					// redis_sentinel is the name of sentinel
	},/*Multiline coment 
	Line1
	Line2
	Line3
	*/
/**/		}//`
)

func TestEnvRawJsonReadByte(t *testing.T) {
	raw := NewRjReaderFromBytes([]byte(envStr))
	expected := []byte(`{"data_db":{"db_type":"redis","db_host":"127.0.0.1","db_port":6379,"db_name":"10","db_user":"*env:TESTVAR","db_password":",/**/","redis_sentinel":""}}`)
	rply := []byte{}
	bit, err := raw.ReadByte()
	for ; err == nil; bit, err = raw.ReadByte() {
		rply = append(rply, bit)
	}
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %+v\n, received: %+v", string(expected), string(rply))
	}
}

func TestEnvRawJsonconsumeComent(t *testing.T) {
	raw := NewRjReaderFromBytes([]byte(`//comment
a/*comment*/b`))
	expected := (byte)('a')
	if r, err := raw.consumeComent('d'); err != nil {
		t.Error(err)
	} else if r {
		t.Errorf("Expected to not replace comment")
	}
	if r, err := raw.consumeComent('/'); err != nil {
		t.Error(err)
	} else if !r {
		t.Errorf("Expected to replace comment")
	}
	if rply, err := raw.ReadByte(); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected: %+v\n, received: %+v", string(expected), string(rply))
	}
	expected = (byte)('b')
	if r, err := raw.consumeComent('*'); err != nil {
		t.Error(err)
	} else if !r {
		t.Errorf("Expected to replace comment")
	}
	if rply, err := raw.ReadByte(); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected: %+v\n, received: %+v", string(expected), string(rply))
	}
}

func TestEnvRawJsonReadByteWC(t *testing.T) {
	raw := NewRjReaderFromBytes([]byte(`c/*first comment*///another comment    

		cgrates`))
	expected := (byte)('c')
	if rply, err := raw.ReadByteWC(); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected: %+v\n, received: %+v", string(expected), string(rply))
	}
	if rply, err := raw.ReadByteWC(); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected: %+v\n, received: %+v", string(expected), string(rply))
	}
}

func TestEnvRawJsonPeekByteWC(t *testing.T) {
	raw := NewRjReaderFromBytes([]byte(`c/*first comment*///another comment    

		bgrates`))
	expected := (byte)('c')
	if rply, err := raw.PeekByteWC(); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected: %+v\n, received: %+v", string(expected), string(rply))
	}
	if rply, err := raw.ReadByteWC(); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected: %+v\n, received: %+v", string(expected), string(rply))
	}
	expected = (byte)('b')
	if rply, err := raw.PeekByteWC(); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected: %+v\n, received: %+v", string(expected), string(rply))
	}
	if rply, err := raw.ReadByteWC(); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected: %+v\n, received: %+v", string(expected), string(rply))
	}
}

func TestEnvRawJsonreadFirstNonWhiteSpace(t *testing.T) {
	raw := NewRjReaderFromBytes([]byte(`    

		cgrates`))
	expected := (byte)('c')
	if rply, err := raw.readFirstNonWhiteSpace(); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected: %+v\n, received: %+v", string(expected), string(rply))
	}
}

func TestEnvReaderRead(t *testing.T) {
	os.Setenv("TESTVAR", "cgRates")
	envR := NewRjReaderFromBytes([]byte(envStr))
	expected := []byte(`{"data_db":{"db_type":"redis","db_host":"127.0.0.1","db_port":6379,"db_name":"10","db_user":"cgRates","db_password":",/**/","redis_sentinel":""}}`)
	rply := []byte{}
	buf := make([]byte, 20)
	n, err := envR.Read(buf)
	for ; err == nil && n > 0; n, err = envR.Read(buf) {
		rply = append(rply, buf[:n]...)
		buf = make([]byte, 20)
	}
	rply = append(rply, buf[:n]...)
	for i, bit := range rply {
		if bit == 0 {
			t.Errorf("Recivied a zero bit on position %+v\n", i)
		}
	}

	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %+v\n, received: %+v", (string(expected)), (string(rply)))
	}
}

func TestEnvReaderRead2(t *testing.T) {
	os.Setenv("TESTVARNoZero", "cgr1")
	envR := NewRjReaderFromBytes([]byte(`{"origin_host": "*env:TESTVARNoZero",
        "origin_realm": "*env:TESTVARNoZero",}`))
	expected := []byte(`{"origin_host":"cgr1","origin_realm":"cgr1"}`)
	rply := []byte{}
	buf := make([]byte, 20)
	n, err := envR.Read(buf)
	for ; err == nil && n > 0; n, err = envR.Read(buf) {
		rply = append(rply, buf[:n]...)
		buf = make([]byte, 20)
	}
	rply = append(rply, buf[:n]...)
	for i, bit := range rply {
		if bit == 0 {
			t.Errorf("Recivied a zero bit on position %+v\n", i)
		}
	}

	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %+q\n, received: %+q", (string(expected)), (string(rply)))
	}
}

func TestEnvReaderreadEnvName(t *testing.T) {
	envR := NewRjReaderFromBytes([]byte(`Test_VAR1 } Var2_TEST'`))
	expected := []byte("Test_VAR1")
	if rply, endindx, err := envR.readEnvName(0); err != nil {
		t.Error(err)
	} else if endindx != 9 {
		t.Errorf("Wrong endindx returned %v", endindx)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %+v, received: %+v", (string(expected)), (string(rply)))
	}
	expected = []byte("Var2_TEST")
	if rply, endindx, err := envR.readEnvName(12); err != nil {
		t.Error(err)
	} else if endindx != 21 {
		t.Errorf("Wrong endindx returned %v", endindx)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %+v, received: %+v", (string(expected)), (string(rply)))
	}
}

func TestEnvReaderreplaceEnv(t *testing.T) {
	os.Setenv("Test_VAR1", "5")
	os.Setenv("Test_VAR2", "aVeryLongEnviormentalVariable")
	envR := NewRjReaderFromBytes([]byte(`*env:Test_VAR1,/*comment*/ }*env:Test_VAR2"`))
	// expected := []byte("5}   ")
	if err := envR.replaceEnv(0); err != nil {
		t.Error(err)
	}
	if err := envR.replaceEnv(15); err != nil {
		t.Error(err)
	}
}

func TestEnvReadercheckMeta(t *testing.T) {
	envR := NewRjReaderFromBytes([]byte("*env:Var"))
	envR.indx = 1
	if !envR.checkMeta() {
		t.Errorf("Expectiv true ")
	}
	envR = NewRjReaderFromBytes([]byte("*enva:Var"))
	envR.indx = 1
	if envR.checkMeta() {
		t.Errorf("Expectiv to get false received true")
	}
}

func TestIsNewLine(t *testing.T) {
	for char, expected := range map[byte]bool{'a': false, '\n': true, ' ': false, '\t': false, '\r': true} {
		if rply := isNewLine(char); expected != rply {
			t.Errorf("Expected: %+v, received: %+v", expected, rply)
		}
	}
}

func TestIsWhiteSpace(t *testing.T) {
	for char, expected := range map[byte]bool{'a': false, '\n': true, ' ': true, '\t': true, '\r': true, 0: true, '1': false} {
		if rply := isWhiteSpace(char); expected != rply {
			t.Errorf("Expected: %+v, received: %+v", expected, rply)
		}
	}
}

func TestReadEnv(t *testing.T) {
	key := "TESTVAR2"
	if _, err := ReadEnv(key); !reflect.DeepEqual(err, utils.ErrEnvNotFound(key)) {
		t.Errorf("Expected: %+v, received: %+v", utils.ErrEnvNotFound(key), err)
	}
	expected := "cgRates"
	os.Setenv(key, expected)
	if rply, err := ReadEnv(key); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected: %+v, received: %+v", expected, rply)
	}
}

func TestIsAlfanum(t *testing.T) {
	for char, expected := range map[byte]bool{'a': true, '\n': false, ' ': false, '\t': false, '\r': false, 0: false, '1': true, 'Q': true, '9': true} {
		if rply := isAlfanum(char); expected != rply {
			t.Errorf("Expected: %+v, received: %+v", expected, rply)
		}
	}
}

func TestGetErrorLine(t *testing.T) {
	jsonstr := `{//nonprocess string
		/***********************************************/

	// Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
	// Copyright (C) ITsysCOM GmbH
	//
	// This file contains the default configuration hardcoded into CGRateS.
	// This is what you get when you load CGRateS with an empty configuration file.
	"data_db": {								// database used to store runtime data (eg: accounts, cdr stats)
		"db_type": "redis",	1					// data_db type: <*redis|*mongo|*internal>
		"db_host": "127.0.0.1",					/* data_db host address*/
		"db_port": 6379, 						// data_db port to reach the database
		"db_name": "10",/*/*asd*/ 						// data_db database name to connect to
		"db_user": "user", 					// username to use when connecting to data_db
		"db_password": ",/**/", 						// password to use when connecting to data_db
		"redis_sentinel":"",					// redis_sentinel is the name of sentinel
	},/*Multiline coment 
	Line1
	Line2
	Line3
	*/
/**/		}//`
	r := NewRjReaderFromBytes([]byte(jsonstr))
	var offset int64 = 31
	var expLine, expChar int64 = 10, 23
	if line, character := r.getJsonOffsetLine(offset); expLine != line {
		t.Errorf("Expected line %v received:%v", expLine, line)
	} else if expChar != character {
		t.Errorf("Expected line %v received:%v", expChar, character)
	}

}

func TestGetErrorLine2(t *testing.T) {
	jsonstr := `{//nonprocess string
		/***********************************************/

	// Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
	// Copyright (C) ITsysCOM GmbH
	//
	// This file contains the default configuration hardcoded into CGRateS.
	// This is what you get when you load CGRateS with an empty configuration file.
	"data_db": {								// database used to store runtime data (eg: accounts, cdr stats)
		"db_type": "redis",	/*some comment before*/1					// data_db type: <*redis|*mongo|*internal>
		"db_host": "127.0.0.1",					/* data_db host address*/
		"db_port": 6379, 						// data_db port to reach the database
		"db_name": "10",/*/*asd*/ 						// data_db database name to connect to
		"db_user": "user", 					// username to use when connecting to data_db
		"db_password": ",/**/", 						// password to use when connecting to data_db
		"redis_sentinel":"",					// redis_sentinel is the name of sentinel
	},/*Multiline coment 
	Line1
	Line2
	Line3
	*/
/**/		}//`
	r := NewRjReaderFromBytes([]byte(jsonstr))
	var offset int64 = 31
	var expLine, expChar int64 = 10, 46
	if line, character := r.getJsonOffsetLine(offset); expLine != line {
		t.Errorf("Expected line %v received:%v", expLine, line)
	} else if expChar != character {
		t.Errorf("Expected line %v received:%v", expChar, character)
	}

}

type reader struct {
}

func (r *reader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("test error: %v", n)
}

func TestRjReaderNewRjReader(t *testing.T) {
	r := &reader{}

	rcv, err := NewRjReader(r)

	if err != nil {
		if err.Error() != "test error: 0" {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestRjReaderUnreadByte(t *testing.T) {
	val := "val1"
	mrsh, err := json.Marshal(val)
	if err != nil {
		t.Error(err)
	}
	rjr := &rjReader{
		buf:        mrsh,
		isInString: true,
		indx:       0,
	}

	err = rjr.UnreadByte()

	if err != nil {
		if err.Error() != bufio.ErrInvalidUnreadByte.Error() {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}
}

func TestRjReaderconsumeComent(t *testing.T) {
	val := "1"
	mrsh, err := json.Marshal(val)
	if err != nil {
		t.Error(err)
	}
	rjr := &rjReader{
		buf:        mrsh,
		isInString: true,
		indx:       20,
	}

	rcv, err := rjr.consumeComent('*')

	if err != nil {
		if err.Error() != utils.ErrJsonIncompleteComment.Error() {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	if rcv != true {
		t.Error(rcv)
	}
}

func TestRjReaderPeekByteWC(t *testing.T) {
	val := "1"
	mrsh, err := json.Marshal(val)
	if err != nil {
		t.Error(err)
	}
	rjr := &rjReader{
		buf:        mrsh,
		isInString: true,
		indx:       20,
	}

	rcv, err := rjr.PeekByteWC()

	if err != nil {
		if err.Error() != io.EOF.Error() {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	if rcv != 0 {
		t.Error(rcv)
	}
}

func TestRjReaderreadEnvName(t *testing.T) {
	val := "val1"
	mrsh, err := json.Marshal(val)
	if err != nil {
		t.Error(err)
	}
	rjr := &rjReader{
		buf:        mrsh,
		isInString: true,
		indx:       0,
	}

	rcv1, rcv2, err := rjr.readEnvName(20)

	if err != nil {
		if err.Error() != io.EOF.Error() {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	if rcv1 != nil {
		t.Error(rcv1)
	}

	if rcv2 != 20 {
		t.Error(rcv2)
	}
}

func TestRjReaderReplaceEnv(t *testing.T) {
	val := "val1"
	mrsh, err := json.Marshal(val)
	if err != nil {
		t.Error(err)
	}
	rjr := &rjReader{
		buf:        mrsh,
		isInString: true,
		indx:       0,
	}

	err = rjr.replaceEnv(20)

	if err != nil {
		if err.Error() != "NOT_FOUND:ENV_VAR:" {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}
}

func TestRjReaderHandleJSONError(t *testing.T) {
	err1 := &json.InvalidUTF8Error{
		S: "test err1",
	}
	err2 := &json.InvalidUnmarshalError{
		Type: reflect.TypeOf(fmt.Errorf("error %s", "test")),
	}
	err3 := &json.UnmarshalTypeError{
		Type: reflect.TypeOf(fmt.Errorf("error %s", "test")),
	}
	val := "val1"
	mrsh, err := json.Marshal(val)
	if err != nil {
		t.Error(err)
	}
	rjr := &rjReader{
		buf:        mrsh,
		isInString: true,
		indx:       0,
	}

	tests := []struct {
		name string
		arg  error
		exp  string
	}{
		{
			name: "case nil",
			arg:  nil,
			exp:  "",
		},
		{
			name: "case *json.InvalidUTF8Error, *json.UnmarshalFieldError",
			arg:  err1,
			exp:  `json: invalid UTF-8 in string: "test err1"`,
		},
		{
			name: "case *json.InvalidUnmarshalError",
			arg:  err2,
			exp:  `json: Unmarshal(nil *errors.errorString)`,
		},
		{
			name: "case *json.UnmarshalTypeError",
			arg:  err3,
			exp:  `json: cannot unmarshal  into Go value of type *errors.errorString at line 0 around position 0`,
		},
		{
			name: "case default",
			arg:  errors.New("test error"),
			exp:  `test error`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rjr.HandleJSONError(tt.arg)

			if tt.exp != "" {
				if err != nil {
					if err.Error() != tt.exp {
						t.Error(err)
					}
				} else {
					t.Error("was expecting an error")
				}
			}
		})
	}
}
