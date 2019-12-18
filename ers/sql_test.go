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

package ers

import (
	"testing"
)

func TestSQLSetURL(t *testing.T) {
	sql := new(SQLEventReader)
	expsql := &SQLEventReader{
		connString:    "cgrates:CGRateS.org@tcp(127.0.0.1:3306)/cgrates2?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
		connType:      "mysql",
		tableName:     "cdrs2",
		expConnString: "cgrates:CGRateS.org@tcp(127.0.0.1:3306)/cgrates2?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
		expConnType:   "mysql",
		expTableName:  "cdrs2",
	}
	inURL := "*mysql://cgrates:CGRateS.org@127.0.0.1:3306?db_name=cgrates2&table_name=cdrs2&sslmode=enabled"
	outURL := "*mysql://cgrates:CGRateS.org@127.0.0.1:3306?db_name=cgrates2&table_name=cdrs2&sslmode=enabled"
	if err := sql.setURL(inURL, outURL); err != nil {
		t.Fatal(err)
	} else if expsql.connString != sql.connString {
		t.Errorf("Expected: %q ,received: %q", expsql.connString, sql.connString)
	} else if expsql.connType != sql.connType {
		t.Errorf("Expected: %q ,received: %q", expsql.connType, sql.connType)
	} else if expsql.tableName != sql.tableName {
		t.Errorf("Expected: %q ,received: %q", expsql.tableName, sql.tableName)
	} else if expsql.expConnString != sql.expConnString {
		t.Errorf("Expected: %q ,received: %q", expsql.expConnString, sql.expConnString)
	} else if expsql.expConnType != sql.expConnType {
		t.Errorf("Expected: %q ,received: %q", expsql.expConnType, sql.expConnType)
	} else if expsql.expTableName != sql.expTableName {
		t.Errorf("Expected: %q ,received: %q", expsql.expTableName, sql.expTableName)
	}

	sql = new(SQLEventReader)
	expsql = &SQLEventReader{
		connString:    "host=127.0.0.1 port=3306 dbname=cgrates2 user=cgrates password=CGRateS.org sslmode=enabled",
		connType:      "postgres",
		tableName:     "cdrs2",
		expConnString: "host=127.0.0.1 port=3306 dbname=cgrates2 user=cgrates password=CGRateS.org sslmode=enabled",
		expConnType:   "postgres",
		expTableName:  "cdrs2",
	}
	inURL = "*postgres://cgrates:CGRateS.org@127.0.0.1:3306?db_name=cgrates2&table_name=cdrs2&sslmode=enabled"
	outURL = "*postgres://cgrates:CGRateS.org@127.0.0.1:3306?db_name=cgrates2&table_name=cdrs2&sslmode=enabled"
	if err := sql.setURL(inURL, outURL); err != nil {
		t.Fatal(err)
	} else if expsql.connString != sql.connString {
		t.Errorf("Expected: %q ,received: %q", expsql.connString, sql.connString)
	} else if expsql.connType != sql.connType {
		t.Errorf("Expected: %q ,received: %q", expsql.connType, sql.connType)
	} else if expsql.tableName != sql.tableName {
		t.Errorf("Expected: %q ,received: %q", expsql.tableName, sql.tableName)
	} else if expsql.expConnString != sql.expConnString {
		t.Errorf("Expected: %q ,received: %q", expsql.expConnString, sql.expConnString)
	} else if expsql.expConnType != sql.expConnType {
		t.Errorf("Expected: %q ,received: %q", expsql.expConnType, sql.expConnType)
	} else if expsql.expTableName != sql.expTableName {
		t.Errorf("Expected: %q ,received: %q", expsql.expTableName, sql.expTableName)
	}

	sql = new(SQLEventReader)
	expsql = &SQLEventReader{
		connString:    "host=127.0.0.1 port=3306 dbname=cgrates2 user=cgrates password=CGRateS.org sslmode=enabled",
		connType:      "postgres",
		tableName:     "cdrs2",
		expConnString: "host=127.0.0.1 port=3306 dbname=cgrates2 user=cgrates password=CGRateS.org sslmode=enabled",
		expConnType:   "postgres",
		expTableName:  "cdrs2",
	}
	inURL = "*postgres://cgrates:CGRateS.org@127.0.0.1:3306?db_name=cgrates2&table_name=cdrs2&sslmode=enabled"
	outURL = "db_name=cgrates2&table_name=cdrs2&sslmode=enabled"
	if err := sql.setURL(inURL, outURL); err != nil {
		t.Fatal(err)
	} else if expsql.connString != sql.connString {
		t.Errorf("Expected: %q ,received: %q", expsql.connString, sql.connString)
	} else if expsql.connType != sql.connType {
		t.Errorf("Expected: %q ,received: %q", expsql.connType, sql.connType)
	} else if expsql.tableName != sql.tableName {
		t.Errorf("Expected: %q ,received: %q", expsql.tableName, sql.tableName)
	} else if expsql.expConnString != sql.expConnString {
		t.Errorf("Expected: %q ,received: %q", expsql.expConnString, sql.expConnString)
	} else if expsql.expConnType != sql.expConnType {
		t.Errorf("Expected: %q ,received: %q", expsql.expConnType, sql.expConnType)
	} else if expsql.expTableName != sql.expTableName {
		t.Errorf("Expected: %q ,received: %q", expsql.expTableName, sql.expTableName)
	}

	inURL = "*postgres://cgrates:CGRateS.org@127.0.0.1:3306?db_name=cgrates2&table_name=cdrs2&sslmode=enabled"
	outURL = "*postgres2://cgrates:CGRateS.org@127.0.0.1:3306?db_name=cgrates2&table_name=cdrs2&sslmode=enabled"
	if err := sql.setURL(inURL, outURL); err == nil || err.Error() != "unknown db_type postgres2" {
		t.Errorf("Expected error: 'unknown db_type postgres2' ,received: %v", err)
	}
	inURL = "*postgres2://cgrates:CGRateS.org@127.0.0.1:3306?db_name=cgrates2&table_name=cdrs2&sslmode=enabled"
	outURL = "*postgres://cgrates:CGRateS.org@127.0.0.1:3306?db_name=cgrates2&table_name=cdrs2&sslmode=enabled"
	if err := sql.setURL(inURL, outURL); err == nil || err.Error() != "unknown db_type postgres2" {
		t.Errorf("Expected error: 'unknown db_type postgres2' ,received: %v", err)
	}
}
