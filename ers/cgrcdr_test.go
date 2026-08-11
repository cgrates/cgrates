// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ers

import (
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
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

func TestCgrCDRProcessMessageFields(t *testing.T) {
	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
"ers": {
	"readers": [
		{
			"id": "fields",
			"fields": [
				{"tag": "refund", "path": "*opts.*refund", "type": "*constant", "value": "true"},
				{"tag": "account", "path": "*cgreq.Account", "type": "*remove"}
			]
		},
		{
			"id": "fields",
			"type": "*cgrcdr",
			"runDelay": "0",
			"sourcePath": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
			"eesSuccessIDs": ["exporter"]
		},
		{
			"id": "no_fields",
			"type": "*cgrcdr"
		}
	]
}
}`)
	if err != nil {
		t.Fatalf("config load failed: %v", err)
	}
	if fields := cfg.ERsCfg().ReaderCfg("no_fields").Fields; len(fields) != 0 {
		t.Fatalf("expected no default Fields, got %s", utils.ToJSON(fields))
	}

	events := make(chan *erEvent, 1)
	rdr, err := NewCgrCdr(cfg, 1, events, make(chan *erEvent, 1), make(chan error, 1),
		engine.NewFilterS(cfg, nil, nil), make(chan struct{}, 1), nil)
	if err != nil {
		t.Fatalf("reader creation failed: %v", err)
	}
	cdrSQL := &utils.CDRSQLTable{
		Tenant: "cgrates.org",
		Opts: utils.JSONB{
			utils.OriginID:   "usage-record",
			utils.MetaRefund: false,
		},
		Event: utils.JSONB{
			utils.AccountField: "1001",
			utils.Destination:  "1002",
		},
	}
	if err := rdr.(*CgrCDR).processMessage(cdrSQL, nil); err != nil {
		t.Fatalf("message processing failed: %v", err)
	}
	event := <-events
	if got := event.cgrEvent.APIOpts[utils.OriginID]; got != "usage-record" {
		t.Errorf("expected OriginID %q, got %#v", "usage-record", got)
	}
	if got := event.cgrEvent.APIOpts[utils.MetaRefund]; got != "true" {
		t.Errorf("expected *refund %q, got %#v", "true", got)
	}
	if got := event.cgrEvent.Event[utils.Destination]; got != "1002" {
		t.Errorf("expected Destination %q, got %#v", "1002", got)
	}
	if got, has := event.cgrEvent.Event[utils.AccountField]; has {
		t.Errorf("expected no Account field, got %#v", got)
	}
	if got := event.rawEvent[utils.OptsCfg].(utils.JSONB)[utils.MetaRefund]; got != false {
		t.Errorf("expected original *refund false, got %#v", got)
	}
	if got := event.rawEvent[utils.EventLowCase].(utils.JSONB)[utils.AccountField]; got != "1001" {
		t.Errorf("expected original Account %q, got %#v", "1001", got)
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
