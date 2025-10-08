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

package ers

import (
	"testing"

	"github.com/cgrates/cgrates/config"

	"github.com/cgrates/cgrates/utils"
	"gorm.io/gorm/logger"
)

func TestSQLSetURL(t *testing.T) {
	sql := new(SQLEventReader)
	expsql := &SQLEventReader{
		connString: "cgrates:CGRateS.org@tcp(127.0.0.1:3306)/cgrates2?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
		connType:   "mysql",
		tableName:  "cdrs2",
	}
	inURL := "*mysql://cgrates:CGRateS.org@127.0.0.1:3306"
	if err := sql.setURL(inURL, &config.EventReaderOpts{
		SQLDBName:    utils.StringPointer("cgrates2"),
		SQLTableName: utils.StringPointer("cdrs2"),
		PgSSLMode:    utils.StringPointer("enabled"),
	}); err != nil {
		t.Fatal(err)
	} else if expsql.connString != sql.connString {
		t.Errorf("Expected: %q ,received: %q", expsql.connString, sql.connString)
	} else if expsql.connType != sql.connType {
		t.Errorf("Expected: %q ,received: %q", expsql.connType, sql.connType)
	} else if expsql.tableName != sql.tableName {
		t.Errorf("Expected: %q ,received: %q", expsql.tableName, sql.tableName)
	}
	sql = new(SQLEventReader)
	expsql = &SQLEventReader{
		connString: "host=127.0.0.1 port=3306 dbname=cgrates2 user=cgrates password=CGRateS.org sslmode=enabled",
		connType:   "postgres",
		tableName:  "cdrs2",
	}
	inURL = "*postgres://cgrates:CGRateS.org@127.0.0.1:3306"
	if err := sql.setURL(inURL, &config.EventReaderOpts{
		SQLDBName:    utils.StringPointer("cgrates2"),
		SQLTableName: utils.StringPointer("cdrs2"),
		PgSSLMode:    utils.StringPointer("enabled"),
	}); err != nil {
		t.Fatal(err)
	} else if expsql.connString != sql.connString {
		t.Errorf("Expected: %q ,received: %q", expsql.connString, sql.connString)
	} else if expsql.connType != sql.connType {
		t.Errorf("Expected: %q ,received: %q", expsql.connType, sql.connType)
	} else if expsql.tableName != sql.tableName {
		t.Errorf("Expected: %q ,received: %q", expsql.tableName, sql.tableName)
	}
	sql = new(SQLEventReader)
	expsql = &SQLEventReader{
		connString: "host=127.0.0.1 port=3306 dbname=cgrates2 user=cgrates password=CGRateS.org sslmode=enabled",
		connType:   "postgres",
		tableName:  "cdrs2",
	}
	inURL = "*postgres://cgrates:CGRateS.org@127.0.0.1:3306"
	if err := sql.setURL(inURL, &config.EventReaderOpts{
		SQLDBName:    utils.StringPointer("cgrates2"),
		SQLTableName: utils.StringPointer("cdrs2"),
		PgSSLMode:    utils.StringPointer("enabled"),
	}); err != nil {
		t.Fatal(err)
	} else if expsql.connString != sql.connString {
		t.Errorf("Expected: %q ,received: %q", expsql.connString, sql.connString)
	} else if expsql.connType != sql.connType {
		t.Errorf("Expected: %q ,received: %q", expsql.connType, sql.connType)
	} else if expsql.tableName != sql.tableName {
		t.Errorf("Expected: %q ,received: %q", expsql.tableName, sql.tableName)
	}
	inURL = "*postgres2://cgrates:CGRateS.org@127.0.0.1:3306?dbName=cgrates2&tableName=cdrs2&sslmode=enabled"
	if err := sql.setURL(inURL, &config.EventReaderOpts{}); err == nil || err.Error() != "unknown db_type postgres2" {
		t.Errorf("Expected error: 'unknown db_type postgres2' ,received: %v", err)
	}
}

func TestSQLReaderServePostgresErr(t *testing.T) {
	tmp := logger.Default
	logger.Default = logger.Default.LogMode(logger.Silent)
	rdr := &SQLEventReader{
		connType:   utils.Postgres,
		connString: "host=127.0.0.1 port=9999 dbname=cdrs user=cgrates password=CGRateS.org sslmode=disabled",
	}
	expected := "cannot parse `host=127.0.0.1 port=9999 dbname=cdrs user=cgrates password=xxxxx sslmode=disabled`: failed to configure TLS (sslmode is invalid)"
	err := rdr.Serve()
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nreceived: <%+v>", expected, err)
	}
	logger.Default = tmp
}

func TestSQLReaderServeBadType(t *testing.T) {
	tmp := logger.Default
	logger.Default = logger.Default.LogMode(logger.Silent)
	rdr := &SQLEventReader{
		connType:   utils.Postgres,
		connString: "host=127.0.0.1 port=9999 dbname=cdrs user=cgrates password=CGRateS.org sslmode=disabled",
	}
	expected := "cannot parse `host=127.0.0.1 port=9999 dbname=cdrs user=cgrates password=xxxxx sslmode=disabled`: failed to configure TLS (sslmode is invalid)"
	err := rdr.Serve()
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nreceived: <%+v>", expected, err)
	}
	logger.Default = tmp
}
