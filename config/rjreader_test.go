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
package config

import (
	"encoding/json"
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
		"db_name": "10",/*/*asd*/ 				// data_db database name to connect to
		"db_user": "*env:TESTVAR", 				// username to use when connecting to data_db
		"db_password": ",/**/", 				// password to use when connecting to data_db
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

func TestNewRjReaderError(t *testing.T) {
	expectedErrFile := "open randomfile.go: no such file or directory"
	file, err := os.Open("randomfile.go")
	if err == nil || err.Error() != expectedErrFile {
		t.Errorf("Expected %+v, receivewd %+v", expectedErrFile, err)
	}
	expectedErrReader := "invalid argument"
	if _, err := NewRjReader(file); err == nil || err.Error() != expectedErrReader {
		t.Errorf("Expected %+v, received %+v", expectedErrReader, err)
	}
}

func TestUnreadByte(t *testing.T) {
	reader := rjReader{
		indx: -1,
	}
	expected := "bufio: invalid use of UnreadByte"
	if err := reader.UnreadByte(); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
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

func TestConsumeComent(t *testing.T) {
	rjreader := new(rjReader)
	var pkbit byte = '*'
	expectedErr := "JSON_INCOMPLETE_COMMENT"
	if _, err := rjreader.consumeComent(pkbit); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
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

func TestEnvRawJsonReadByteWCError(t *testing.T) {
	raw := NewRjReaderFromBytes([]byte(`/*`))
	expected := "JSON_INCOMPLETE_COMMENT"
	if _, err := raw.ReadByteWC(); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestEnvRawJsonPeekByteWCError(t *testing.T) {
	raw := NewRjReaderFromBytes([]byte(`/*`))
	expected := "JSON_INCOMPLETE_COMMENT"
	if _, err := raw.PeekByteWC(); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	raw = NewRjReaderFromBytes([]byte(`/rand val`))
	expectedByte := (byte)('/')
	if rply, err := raw.PeekByteWC(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedByte, rply) {
		t.Errorf("Expected %+v, received %+v", expectedByte, rply)
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
	rply, endindx := envR.readEnvName(0)
	if endindx != 9 {
		t.Errorf("Wrong endindx returned %v", endindx)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %+v, received: %+v", (string(expected)), (string(rply)))
	}

	expected = []byte("Var2_TEST")
	rply, endindx = envR.readEnvName(12)
	if endindx != 21 {
		t.Errorf("Wrong endindx returned %v", endindx)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %+v, received: %+v", (string(expected)), (string(rply)))
	}
}

func TestEnvReaderreadEnvNameError(t *testing.T) {
	envR := NewRjReaderFromBytes([]byte(``))
	if name, index := envR.readEnvName(1); name != nil || index != 1 {
		t.Errorf("Expected nil, received %+v", name)
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

func TestHandleJSONErrorNil(t *testing.T) {
	os.Setenv("Test_VAR1", "5")
	os.Setenv("Test_VAR2", "aVeryLongEnviormentalVariable")
	envR := NewRjReaderFromBytes([]byte(`*env:Test_VAR1,/*comment*/ }*env:Test_VAR2"`))
	var expected error = nil
	if err := envR.replaceEnv(0); err != nil {
		t.Error(err)
	} else if newErr := envR.HandleJSONError(err); newErr != expected {
		t.Errorf("Expected %+v, received %+v", expected, newErr)
	}
}

func TestHandleJSONErrorInvalidUTF8(t *testing.T) {
	rjr := NewRjReaderFromBytes([]byte("{}"))
	expectedErr := new(json.InvalidUTF8Error)
	if err := rjr.HandleJSONError(expectedErr); err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestHandleJSONErrorInvalidUnmarshalError(t *testing.T) {
	rjr := NewRjReaderFromBytes([]byte("{}"))
	err := json.NewDecoder(rjr).Decode(nil)
	if err == nil {
		t.Fatal(err)
	}
	err = rjr.HandleJSONError(err)
	expectedErr := &json.InvalidUnmarshalError{Type: reflect.TypeOf(nil)}
	if err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestHandleJSONErrorDefaultError(t *testing.T) {
	rjr := NewRjReaderFromBytes([]byte("{}"))
	rjr.indx = 10
	if _, err := rjr.ReadByteWC(); err == nil || err != io.EOF {
		t.Errorf("Expected %+v, received %+v", io.EOF, err)
	} else if newErr := rjr.HandleJSONError(err); newErr == nil || newErr != io.EOF {
		t.Errorf("Expected %+v, received %+v", io.EOF, err)
	}
}

func TestHandleJSONErrorUnmarshalTypeError(t *testing.T) {
	rjr := NewRjReaderFromBytes([]byte("{}"))
	err := &json.UnmarshalTypeError{
		Offset: 0,
		Value:  "2",
		Type:   reflect.TypeOf(""),
		Struct: "configs",
		Field:  "field",
	}
	expMessage := fmt.Sprintf("%s at line 0 around position 0", err.Error())
	if err := rjr.HandleJSONError(err); err == nil || err.Error() != expMessage {
		t.Errorf("Expected %+v, received %+v", expMessage, err)
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

func TestReadRJReader(t *testing.T) {
	os.Setenv("TESTVARNoZero", utils.EmptyString)
	rjr := NewRjReaderFromBytes([]byte(`{"origin_host": "*env:TESTVARNoZero",
        "origin_realm": "*env:TESTVARNoZero",}`))
	expected := "NOT_FOUND:ENV_VAR:TESTVARNoZero"
	buf := make([]byte, 20)
	rjr.indx = 1
	if _, err := rjr.Read(buf); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestReadEnv(t *testing.T) {
	key := "TESTVAR2"
	if _, err := ReadEnv(key); err == nil || err.Error() != utils.ErrEnvNotFound(key).Error() {
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
	if line, character := r.getJSONOffsetLine(offset); expLine != line {
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
	if line, character := r.getJSONOffsetLine(offset); expLine != line {
		t.Errorf("Expected line %v received:%v", expLine, line)
	} else if expChar != character {
		t.Errorf("Expected line %v received:%v", expChar, character)
	}
}

func TestGetJSONOffsetLineFuncError(t *testing.T) {
	rjr := NewRjReaderFromBytes([]byte("{}"))
	rjr.indx = 7
	var eLine, eCharachter int64 = 1, 0
	if line, character := rjr.getJSONOffsetLine(int64(3)); line != eLine && character != eCharachter {
		fmt.Printf("Expected %+v and %+v, received %+v and %+v", eLine, eCharachter, line, character)
	}
}

func TestGetJSONOffsetLineReadStringEOF(t *testing.T) {
	rjr := NewRjReaderFromBytes([]byte(`{","}`))
	rjr.indx = 3
	var eLine, eCharachter int64 = 1, 0
	if line, character := rjr.getJSONOffsetLine(int64(5)); line != eLine && character != eCharachter {
		fmt.Printf("Expected %+v and %+v, received %+v and %+v", eLine, eCharachter, line, character)
	}
}

func TestGetJSONOffsetLineReadString1(t *testing.T) {
	rjr := NewRjReaderFromBytes([]byte(`{,}, {val1, val2}`))
	rjr.indx = 0
	var eLine, eCharachter int64 = 1, 0
	if line, character := rjr.getJSONOffsetLine(int64(5)); line != eLine && character != eCharachter {
		fmt.Printf("Expected %+v and %+v, received %+v and %+v", eLine, eCharachter, line, character)
	}
}

func TestGetJSONOffsetLineReadStringNilError1(t *testing.T) {
	rjr := NewRjReaderFromBytes([]byte(`"

", {,}, {val1, val2}`))
	rjr.indx = 0
	var eLine, eCharachter int64 = 1, 0
	if line, character := rjr.getJSONOffsetLine(int64(3)); line != eLine && character != eCharachter {
		fmt.Printf("Expected %+v and %+v, received %+v and %+v", eLine, eCharachter, line, character)
	}
}

func TestGetJSONOffsetLineReadLineCommentEOF(t *testing.T) {
	rjr := NewRjReaderFromBytes([]byte(`random/`))
	rjr.indx = 6
	var eLine, eCharachter int64 = 1, 0
	if line, character := rjr.getJSONOffsetLine(int64(5)); line != eLine && character != eCharachter {
		fmt.Printf("Expected %+v and %+v, received %+v and %+v", eLine, eCharachter, line, character)
	}
}

func TestGetJSONOffsetLineReadLineCommentEOF1(t *testing.T) {
	rjr := NewRjReaderFromBytes([]byte(`random//`))
	rjr.indx = 6
	var eLine, eCharachter int64 = 1, 0
	if line, character := rjr.getJSONOffsetLine(int64(5)); line != eLine && character != eCharachter {
		fmt.Printf("Expected %+v and %+v, received %+v and %+v", eLine, eCharachter, line, character)
	}
}

func TestGetJSONOffsetLineReadCommentEOF(t *testing.T) {
	rjr := NewRjReaderFromBytes([]byte(`random/*`))
	rjr.indx = 5
	var eLine, eCharachter int64 = 1, 0
	if line, character := rjr.getJSONOffsetLine(int64(5)); line != eLine && character != eCharachter {
		fmt.Printf("Expected %+v and %+v, received %+v and %+v", eLine, eCharachter, line, character)
	}
}

func TestGetJSONOffsetLineReadCommentInvalidEnding(t *testing.T) {
	rjr := NewRjReaderFromBytes([]byte(`random/*

**`))
	rjr.indx = 5
	var eLine, eCharachter int64 = 3, 2
	if line, character := rjr.getJSONOffsetLine(int64(5)); line != eLine && character != eCharachter {
		fmt.Printf("Expected %+v and %+v, received %+v and %+v", eLine, eCharachter, line, character)
	}
}

func TestGetJSONOffsetLineReadComment1(t *testing.T) {
	rjr := NewRjReaderFromBytes([]byte(`/*
**`))
	rjr.indx = 0
	var eLine, eCharachter int64 = 3, 2
	if line, character := rjr.getJSONOffsetLine(int64(5)); line != eLine && character != eCharachter {
		fmt.Printf("Expected %+v and %+v, received %+v and %+v", eLine, eCharachter, line, character)
	}
}

func TestGetJSONOffsetLineInvalidComment2(t *testing.T) {
	rjr := NewRjReaderFromBytes([]byte(`/noComm`))
	rjr.indx = 0
	var eLine, eCharachter int64 = 1, 5
	if line, character := rjr.getJSONOffsetLine(int64(5)); line != eLine && character != eCharachter {
		fmt.Printf("Expected %+v and %+v, received %+v and %+v", eLine, eCharachter, line, character)
	}
}
