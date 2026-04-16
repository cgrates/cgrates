//go:build integration

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
package general_tests

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// TestERSSQLFilterUnquote checks that JSON_UNQUOTE wraps only JSON_VALUE,
// not the whole comparison (which produces invalid SQL on MySQL 8).
func TestERSSQLFilterUnquote(t *testing.T) {
	var dbcfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbcfg = engine.InternalDBCfg
	case utils.MetaRedis:
		dbcfg = engine.RedisDBCfg
	case utils.MetaMySQL:
		dbcfg = engine.MySQLDBCfg
	case utils.MetaMongo:
		dbcfg = engine.MongoDBCfg
	case utils.MetaPostgres:
		dbcfg = engine.PostgresDBCfg
	default:
		t.Fatal("unsupported dbtype value")
	}

	cdb, err := gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates")),
		&gorm.Config{AllowGlobalUpdate: true})
	if err != nil {
		t.Fatal(err)
	}
	if err = cdb.Exec(`CREATE DATABASE IF NOT EXISTS cgrates2;`).Error; err != nil {
		t.Fatal(err)
	}
	sqlCDB, err := cdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlCDB.SetConnMaxLifetime(5 * time.Second)

	db, err := gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates2")),
		&gorm.Config{AllowGlobalUpdate: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = db.Migrator().DropTable("cdrs")
		_ = db.Exec(`DROP DATABASE cgrates2;`).Error
		if d, err := db.DB(); err == nil {
			_ = d.Close()
		}
		if d, err := cdb.DB(); err == nil {
			_ = d.Close()
		}
	})

	// Create CDR table schema.
	fileContent, err := os.ReadFile("/usr/share/cgrates/storage/mysql/create_cdrs_tables.sql")
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	for qry := range strings.SplitSeq(string(fileContent), ";") {
		qry = strings.TrimSpace(qry)
		if len(qry) == 0 {
			continue
		}
		if _, err := sqlDB.Exec(qry); err != nil {
			t.Fatal(err)
		}
	}

	// Insert one CDR with RateID=RateID2 (should be matched by filter).
	tx := db.Begin()
	tx = tx.Table(utils.CDRsTBL)
	saved := tx.Save(&utils.CDRSQLTable{
		Tenant:    cdr2.Tenant,
		Opts:      cdr2.Opts,
		Event:     cdr2.Event,
		CreatedAt: time.Now(),
	})
	if saved.Error != nil {
		tx.Rollback()
		t.Fatal(saved.Error)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatal(err)
	}

	tpFiles := map[string]string{
		utils.FiltersCsv: `#Tenant[0],ID[1],Type[2],Path[3],Values[4]
cgrates.org,FLTR_SQL_RatingID,*eq,~*req.opts.*rateSCost.CostIntervals[0].Increments[0].RateID,RateID2
cgrates.org,FLTR_VARS,*string,~*vars.*readerID,mysql`,
	}

	jsonCfg := `{
"general": {
	"reply_timeout": "10s",
	"default_timezone": "UTC"
},
"logger": {
	"level": 7
},
"db": {
	"db_conns": {
		"*default": {
			"db_type": "*internal",
			"opts":{
				"internalDBRewriteInterval": "0s",
				"internalDBDumpInterval": "0s"
			}
		}
	}
},
"admins": {
	"enabled": true
},
"sessions": {
	"enabled": true
},
"ers": {
	"enabled": true,
	"readers": [
		{
			"id": "mysql",
			"type": "*sql",
			"run_delay": "1m",
			"source_path": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
			"opts": {
				"sqlDBName":"cgrates2",
				"sqlTableName":"cdrs",
				"sqlBatchSize": 10
			},
			"start_delay": "500ms",
			"tenant": "cgrates.org",
			"filters": [
				"FLTR_SQL_RatingID",
				"FLTR_VARS"
			],
			"flags": ["*dryRun"],
			"fields":[
				{"tag": "ToR", "path": "*cgreq.ToR", "type": "*variable", "value": "~*req.event.ToR", "mandatory": true},
				{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.event.Account", "mandatory": true},
				{"tag": "Destination", "path": "*cgreq.Destination", "type": "*variable", "value": "~*req.event.Destination", "mandatory": true}
			]
		}
	]
}
}`

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigJSON:       jsonCfg,
		DBCfg:            dbcfg,
		Encoding:         *utils.Encoding,
		TpFiles:          tpFiles,
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	ng.Run(t)
	time.Sleep(500 * time.Millisecond)

	t.Run("VerifyFilterMatchesWithCorrectUnquote", func(t *testing.T) {
		time.Sleep(100 * time.Millisecond)
		var records int
		scanner := bufio.NewScanner(strings.NewReader(buf.String()))
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "<ERs> DRY_RUN, reader: <mysql>") {
				records++
			}
		}
		if err := scanner.Err(); err != nil {
			t.Fatalf("error reading log: %v", err)
		}
		if records != 1 {
			t.Fatalf("expected 1 DRY_RUN record, got %d\nlog output:\n%s", records, buf.String())
		}

		idx := strings.Index(buf.String(), "CGREvent: ")
		if idx == -1 {
			t.Fatalf("CGREvent not found in log output:\n%s", buf.String())
		}
		var cgrEv utils.CGREvent
		if err := json.NewDecoder(strings.NewReader(buf.String()[idx+len("CGREvent: "):])).Decode(&cgrEv); err != nil {
			t.Fatal(err)
		}
		if got := cgrEv.Event["Account"]; got != "1001" {
			t.Errorf("expected Account=1001, got %v", got)
		}
	})
}
