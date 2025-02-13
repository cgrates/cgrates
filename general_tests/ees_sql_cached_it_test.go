//go:build integration
// +build integration

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

package general_tests

import (
	"bytes"
	"fmt"
	"math/rand"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type testModelSql struct {
	Cgrid         string
	UsageDuration int64
	Cost          int64
}

func (*testModelSql) TableName() string {
	return "cdrs"
}

func initDB(t *testing.T) {
	dbConnString := "cgrates:CGRateS.org@tcp(127.0.0.1:3306)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'"

	mainDB, err := gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates")), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to main DB: %v", err)
	}
	if err != nil {
		t.Fatalf("Failed to get database instance: %v", err)
	}

	// Create the test database if not exists
	if err := mainDB.Exec("CREATE DATABASE IF NOT EXISTS cgrates2").Error; err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Connect to the test database
	testDB, err := gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates2")), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test DB: %v", err)
	}

	sqlDB, _ := testDB.DB()

	// Drop the table if it exists and recreate
	if testDB.Migrator().HasTable("cdrs") {
		if err := testDB.Migrator().DropTable("cdrs"); err != nil {
			t.Fatalf("Failed to drop existing table: %v", err)
		}
	}
	if err := testDB.Migrator().CreateTable(&testModelSql{}); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Register cleanup to drop the table and close the connection
	t.Cleanup(func() {
		if err := mainDB.Exec("DROP DATABASE IF EXISTS cgrates2").Error; err != nil {
			t.Logf("Failed to drop database during cleanup: %v", err)
		}
		if err := sqlDB.Close(); err != nil {
			t.Logf("Failed to close database connection: %v", err)
		}
	})
}
func sendEvents(client *birpc.Client, t *testing.T) {
	n := 500
	var wg sync.WaitGroup
	wg.Add(n)

	var reply map[string]map[string]interface{}
	for range n {
		go func() {
			defer wg.Done()
			if err := client.Call(context.Background(), utils.EeSv1ProcessEvent,
				&engine.CGREventWithEeIDs{
					EeIDs: []string{"SQLExporter1"},
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     utils.GenUUID(),
						Event: map[string]interface{}{
							utils.OriginID:  utils.GenUUID(),
							"UsageDuration": 10 * time.Second,
							utils.Cost:      rand.Intn(100),
						},
					},
				}, &reply); err != nil {
				t.Error(err)
			}

		}()
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(20 * time.Second):
		t.Errorf("timed out waiting for %s replies", utils.EeSv1ProcessEvent)
	}
}
func TestSQLExporterCached(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
	initDB(t)
	content := `{
"general": {
    "log_level": 7               
},

"data_db": {
    "db_type": "*internal"
},

"stor_db": {
    "db_type": "*internal"
},

"ees": {
    "enabled": true,
    "exporters": [
      {
                "id": "SQLExporter1",
                "type": "*sql",
                "filters": [],
                "export_path": "mysql://cgrates:CGRateS.org@localhost:3306",
                "attempts": 3,
                "opts": {
                    "sqlDBName": "cgrates2",
                    "sqlTableName": "cdrs",
                },
                "fields":[
                    {"tag": "Cgrid", "path": "*exp.Cgrid", "type": "*variable", "value": "~*req.OriginID", "mandatory": true},
             	    {"tag": "Cost", "path": "*exp.Cost", "type": "*variable", "value": "~*req.Cost", "mandatory": true},
              	    {"tag": "UsageDuration", "path": "*exp.Usage_duration", "type": "*variable", "value": "~*req.UsageDuration", "mandatory": true}
                ]
            }
    ]
},
}`
	var buf bytes.Buffer
	ng := engine.TestEngine{
		ConfigJSON: content,
		LogBuffer:  &buf,
	}
	client, _ := ng.Run(t)

	t.Run("ExportSQLEvent not cached exporter", func(t *testing.T) {
		sendEvents(client, t)
		regex := regexp.MustCompile(`Error 1040: Too many connections`)
		if regex.Match(buf.Bytes()) {
			t.Error("Expected to not get 'Too many connections'")
		}

	})

}

func TestExporterNotCached(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
	initDB(t)
	content := `{
		"general": {
			"log_level": 7               
		},
		
		"data_db": {
			"db_type": "*internal"
		},
		
		"stor_db": {
			"db_type": "*internal"
		},
		
		"ees": {
			"enabled": true,
		    "cache":{
			"*sql": {"limit": 0, "ttl": "", "static_ttl": false},
			},
			"exporters": [
			  {
						"id": "SQLExporter1",
						"type": "*sql",
						"filters": [],
						"cache":{
						"*sql": {"limit": 0, "ttl": "", "static_ttl": false},
						},
						"export_path": "mysql://cgrates:CGRateS.org@localhost:3306",
						"attempts": 3,
						"opts": {
							"sqlDBName": "cgrates2",
							"sqlTableName": "cdrs",
						},
						"fields":[
							{"tag": "Cgrid", "path": "*exp.Cgrid", "type": "*variable", "value": "~*req.OriginID", "mandatory": true},
							 {"tag": "Cost", "path": "*exp.Cost", "type": "*variable", "value": "~*req.Cost", "mandatory": true},
							  {"tag": "UsageDuration", "path": "*exp.Usage_duration", "type": "*variable", "value": "~*req.UsageDuration", "mandatory": true}
						]
					}
			]
		},
		}`
	var buf bytes.Buffer
	ng := engine.TestEngine{
		ConfigJSON: content,
		LogBuffer:  &buf,
	}
	client, _ := ng.Run(t)
	t.Run("ExportSQLEvent cached exporter", func(t *testing.T) {
		sendEvents(client, t)
		regex := regexp.MustCompile(`Error 1040`)
		if !regex.Match(buf.Bytes()) {
			t.Error("Dint detected 'Too many connections' error as expected")
		}
	})
}
