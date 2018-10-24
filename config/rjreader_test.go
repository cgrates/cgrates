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

	if !reflect.DeepEqual(expected, rply) {
		// for i := 0; i < len(expected); i++ {
		// 	if expected[i] != rply[i] {
		// 		t.Errorf("Expected: %q\n, recived: %q pe pozitia %+v", (string(expected[i-2 : i+2])), (string(rply[i-2 : i+2])), i)
		// 		break
		// 	}
		// }
		t.Errorf("Expected: %+v\n, recived: %+v", (string(expected)), (string(rply)))
	}
}
