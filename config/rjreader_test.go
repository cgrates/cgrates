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
	"bufio"
	"os"
	"reflect"
	"strings"
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
	raw := &rawJSON{rdr: bufio.NewReader(strings.NewReader(envStr))}
	expected := []byte(`{"data_db":{"db_type":"redis","db_host":"127.0.0.1","db_port":6379,"db_name":"10","db_user":"*env:TESTVAR","db_password":",/**/","redis_sentinel":""}}`)
	rply := []byte{}
	bit, err := raw.ReadByte()
	for ; err == nil; bit, err = raw.ReadByte() {
		rply = append(rply, bit)
	}
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %+v\n, recived: %+v", string(expected), string(rply))
	}
}

func TestEnvRawJsonconsumeComent(t *testing.T) {
	raw := &rawJSON{rdr: bufio.NewReader(strings.NewReader(`//comment
a/*comment*/b`))}
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
		t.Errorf("Expected: %+v\n, recived: %+v", string(expected), string(rply))
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
		t.Errorf("Expected: %+v\n, recived: %+v", string(expected), string(rply))
	}
}

func TestEnvRawJsonReadByteWC(t *testing.T) {
	raw := &rawJSON{rdr: bufio.NewReader(strings.NewReader(`c/*first comment*///another comment    

		cgrates`))}
	expected := (byte)('c')
	if rply, err := raw.ReadByteWC(); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected: %+v\n, recived: %+v", string(expected), string(rply))
	}
	if rply, err := raw.ReadByteWC(); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected: %+v\n, recived: %+v", string(expected), string(rply))
	}
}

func TestEnvRawJsonPeekByteWC(t *testing.T) {
	raw := &rawJSON{rdr: bufio.NewReader(strings.NewReader(`c/*first comment*///another comment    

		bgrates`))}
	expected := (byte)('c')
	if rply, err := raw.PeekByteWC(); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected: %+v\n, recived: %+v", string(expected), string(rply))
	}
	if rply, err := raw.ReadByteWC(); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected: %+v\n, recived: %+v", string(expected), string(rply))
	}
	expected = (byte)('b')
	if rply, err := raw.PeekByteWC(); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected: %+v\n, recived: %+v", string(expected), string(rply))
	}
	if rply, err := raw.ReadByteWC(); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected: %+v\n, recived: %+v", string(expected), string(rply))
	}
}

func TestEnvRawJsonreadFirstNonWhiteSpace(t *testing.T) {
	raw := &rawJSON{rdr: bufio.NewReader(strings.NewReader(`    

		cgrates`))}
	expected := (byte)('c')
	if rply, err := raw.readFirstNonWhiteSpace(); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected: %+v\n, recived: %+v", string(expected), string(rply))
	}
}

func TestEnvReaderRead(t *testing.T) {
	os.Setenv("TESTVAR", "cgRates")
	envR := NewRawJSONReader(strings.NewReader(envStr))
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
		t.Errorf("Expected: %+v\n, recived: %+v", (string(expected)), (string(rply)))
	}
}

func TestEnvReaderRead2(t *testing.T) {
	os.Setenv("TESTVARNoZero", "cgr1")
	envR := NewRawJSONReader(strings.NewReader(`{"origin_host": "*env:TESTVARNoZero",
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
		t.Errorf("Expected: %+q\n, recived: %+q", (string(expected)), (string(rply)))
	}
}

func TestEnvReaderreadEnvName(t *testing.T) {
	envR := EnvReader{rd: &rawJSON{rdr: bufio.NewReader(strings.NewReader(`Test_VAR1 } Var2_TEST'`))}}
	expected := []byte("Test_VAR1")
	if rply, bit, err := envR.readEnvName(); err != nil {
		t.Error(err)
	} else if bit != '}' {
		t.Errorf("Wrong bit returned %q", bit)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %+v, recived: %+v", (string(expected)), (string(rply)))
	}
	expected = []byte("Var2_TEST")
	if rply, bit, err := envR.readEnvName(); err != nil {
		t.Error(err)
	} else if bit != '\'' {
		t.Errorf("Wrong bit returned %q", bit)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %+v, recived: %+v", (string(expected)), (string(rply)))
	}
}

func TestEnvReaderreplaceEnv(t *testing.T) {
	os.Setenv("Test_VAR1", "5")
	os.Setenv("Test_VAR2", "aVeryLongEnviormentalVariable")
	envR := EnvReader{rd: &rawJSON{rdr: bufio.NewReader(strings.NewReader(`Test_VAR1,/*comment*/ }Test_VAR2"`))}}
	expected := []byte("5}   ")
	expectedn := 1
	rply := make([]byte, 5)
	if n, err := envR.replaceEnv(rply, 0, 5); err != nil {
		t.Error(err)
	} else if expectedn != n {
		t.Errorf("Expected: %+v, recived: %+v", expectedn, n)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %q, recived: %q", (string(expected)), (string(rply)))
	}
	expected = []byte("aVery")
	expectedn = 5
	rply = make([]byte, 5)
	if n, err := envR.replaceEnv(rply, 0, 5); err != nil {
		t.Error(err)
	} else if expectedn != n {
		t.Errorf("Expected: %+v, recived: %+v", expectedn, n)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %q, recived: %q", (string(expected)), (string(rply)))
	} else if bufexp := []byte("LongEnviormentalVariable\""); !reflect.DeepEqual(bufexp, envR.buf) {
		t.Errorf("Expected: %q, recived: %q", (string(expected)), (string(rply)))
	}
}

func TestEnvReadercheckMeta(t *testing.T) {
	envR := EnvReader{rd: &rawJSON{rdr: bufio.NewReader(strings.NewReader(""))}}
	envR.m = 2
	if envR.checkMeta('n') {
		t.Errorf("Expectiv to get false recived true")
	} else if envR.m != 3 {
		t.Errorf("Expectiv the meta offset to incrase")
	}
	envR.m = 4
	if !envR.checkMeta(':') {
		t.Errorf("Expectiv true ")
	} else if envR.m != 0 {
		t.Errorf("Expectiv the meta offset to reset")
	}
	envR.m = 1
	if envR.checkMeta('v') {
		t.Errorf("Expectiv to get false recived true")
	} else if envR.m != 0 {
		t.Errorf("Expectiv the meta offset to reset")
	}
}

func TestisNewLine(t *testing.T) {
	for char, expected := range map[byte]bool{'a': false, '\n': true, ' ': false, '\t': false, '\r': true} {
		if rply := isNewLine(char); expected != rply {
			t.Errorf("Expected: %+v, recived: %+v", expected, rply)
		}
	}
}

func TestisWhiteSpace(t *testing.T) {
	for char, expected := range map[byte]bool{'a': false, '\n': true, ' ': true, '\t': true, '\r': true, 0: true, '1': false} {
		if rply := isWhiteSpace(char); expected != rply {
			t.Errorf("Expected: %+v, recived: %+v", expected, rply)
		}
	}
}

func TestReadEnv(t *testing.T) {
	key := "TESTVAR2"
	if _, err := ReadEnv(key); !reflect.DeepEqual(err, utils.ErrEnvNotFound(key)) {
		t.Errorf("Expected: %+v, recived: %+v", utils.ErrEnvNotFound(key), err)
	}
	expected := "cgRates"
	os.Setenv(key, expected)
	if rply, err := ReadEnv(key); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected: %+v, recived: %+v", expected, rply)
	}
}

func TestisAlfanum(t *testing.T) {
	for char, expected := range map[byte]bool{'a': true, '\n': false, ' ': false, '\t': false, '\r': false, 0: false, '1': true, 'Q': true, '9': true} {
		if rply := isAlfanum(char); expected != rply {
			t.Errorf("Expected: %+v, recived: %+v", expected, rply)
		}
	}
}
