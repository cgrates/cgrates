// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ers

import (
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"

	"gorm.io/gorm/logger"
)

func TestCgrCDRSetURL(t *testing.T) {
	rdr := new(CgrCDR)
	exp := &CgrCDR{
		connString: "cgrates:CGRateS.org@tcp(127.0.0.1:3306)/cgrates2?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
		connType:   "mysql",
		tableName:  "cdrs2",
	}
	inURL := "*mysql://cgrates:CGRateS.org@127.0.0.1:3306"
	if err := rdr.setURL(inURL, &config.EventReaderOpts{
		SQLDBName:    utils.StringPointer("cgrates2"),
		SQLTableName: utils.StringPointer("cdrs2"),
		PgSSLMode:    utils.StringPointer("enabled"),
	}); err != nil {
		t.Fatal(err)
	} else if exp.connString != rdr.connString {
		t.Errorf("Expected: %q ,received: %q", exp.connString, rdr.connString)
	} else if exp.connType != rdr.connType {
		t.Errorf("Expected: %q ,received: %q", exp.connType, rdr.connType)
	} else if exp.tableName != rdr.tableName {
		t.Errorf("Expected: %q ,received: %q", exp.tableName, rdr.tableName)
	}

	rdr = new(CgrCDR)
	exp = &CgrCDR{
		connString: "host=127.0.0.1 port=3306 dbname=cgrates2 user=cgrates password=CGRateS.org sslmode=enabled",
		connType:   "postgres",
		tableName:  "cdrs2",
	}
	inURL = "*postgres://cgrates:CGRateS.org@127.0.0.1:3306"
	if err := rdr.setURL(inURL, &config.EventReaderOpts{
		SQLDBName:    utils.StringPointer("cgrates2"),
		SQLTableName: utils.StringPointer("cdrs2"),
		PgSSLMode:    utils.StringPointer("enabled"),
	}); err != nil {
		t.Fatal(err)
	} else if exp.connString != rdr.connString {
		t.Errorf("Expected: %q ,received: %q", exp.connString, rdr.connString)
	} else if exp.connType != rdr.connType {
		t.Errorf("Expected: %q ,received: %q", exp.connType, rdr.connType)
	} else if exp.tableName != rdr.tableName {
		t.Errorf("Expected: %q ,received: %q", exp.tableName, rdr.tableName)
	}

	rdr = new(CgrCDR)
	exp = &CgrCDR{
		connString: "cgrates:CGRateS.org@tcp(127.0.0.1:3306)/cgrates?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
		connType:   "mysql",
		tableName:  utils.CDRsTBL,
	}
	inURL = "*mysql://cgrates:CGRateS.org@127.0.0.1:3306"
	if err := rdr.setURL(inURL, &config.EventReaderOpts{}); err != nil {
		t.Fatal(err)
	} else if exp.connString != rdr.connString {
		t.Errorf("Expected: %q ,received: %q", exp.connString, rdr.connString)
	} else if exp.connType != rdr.connType {
		t.Errorf("Expected: %q ,received: %q", exp.connType, rdr.connType)
	} else if exp.tableName != rdr.tableName {
		t.Errorf("Expected: %q ,received: %q", exp.tableName, rdr.tableName)
	}

	rdr = new(CgrCDR)
	inURL = "*postgres2://cgrates:CGRateS.org@127.0.0.1:3306"
	if err := rdr.setURL(inURL, &config.EventReaderOpts{}); err == nil || err.Error() != "unknown dbType postgres2" {
		t.Errorf("Expected error: 'unknown dbType postgres2' ,received: %v", err)
	}
}

func TestCgrCDRServePostgresErr(t *testing.T) {
	tmp := logger.Default
	logger.Default = logger.Default.LogMode(logger.Silent)
	rdr := &CgrCDR{
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

func TestCgrCDRServeBadType(t *testing.T) {
	tmp := logger.Default
	logger.Default = logger.Default.LogMode(logger.Silent)
	rdr := &CgrCDR{
		connType: "sqlite",
	}
	expected := "db type <sqlite> not supported"
	err := rdr.Serve()
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nreceived: <%+v>", expected, err)
	}
	logger.Default = tmp
}
