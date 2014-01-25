/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package cdrc

import (
	"errors"
	"flag"
	"fmt"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"io/ioutil"
	"os/exec"
	"path"
	"testing"
	"time"
)

/*
README:

 Enable local tests by passing '-local' to the go test command
 It is expected that the data folder of CGRateS exists at path /usr/share/cgrates/data or passed via command arguments.
 Prior running the tests, create database and users by running:
  mysql -pyourrootpwd < /usr/share/cgrates/data/storage/mysql/create_db_with_users.sql
 What these tests do:
  * Flush tables in storDb.
  * Start engine with default configuration and give it some time to listen (here caching can slow down).
  *
*/

var cfg *config.CGRConfig

var testLocal = flag.Bool("local", false, "Perform the tests only on local test environment, not by default.") // This flag will be passed here via "go test -local" args
var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
var storDbType = flag.String("stordb_type", "mysql", "The type of the storDb database <mysql>")
var waitRater = flag.Int("wait_rater", 300, "Number of miliseconds to wait for rater to start and cache")

func init() {
	cfgPath := path.Join(*dataDir, "conf", "cgrates.cfg")
	cfg, _ = config.NewCGRConfig(&cfgPath)
}

var fileContent1 = `accid11,prepaid,out,cgrates.org,call,1001,1001,+4986517174963,2013-02-03 19:54:00,62,supplier1,172.16.1.1
accid12,prepaid,out,cgrates.org,call,1001,1001,+4986517174963,2013-02-03 19:54:00,62,supplier1,172.16.1.1
dummy_data
accid13,prepaid,out,cgrates.org,call,1001,1001,+4986517174963,2013-02-03 19:54:00,62,supplier1,172.16.1.1
`

var fileContent2 = `accid21,prepaid,out,cgrates.org,call,1001,1001,+4986517174963,2013-02-03 19:54:00,62,supplier1,172.16.1.1
accid22,prepaid,out,cgrates.org,call,1001,1001,+4986517174963,2013-02-03 19:54:00,62,supplier1,172.16.1.1
#accid1,prepaid,out,cgrates.org,call,1001,1001,+4986517174963,2013-02-03 19:54:00,62,supplier1,172.16.1.1
accid23,prepaid,out,cgrates.org,call,1001,1001,+4986517174963,2013-02-03 19:54:00,62,supplier1,172.16.1.1`

func startEngine() error {
	enginePath, err := exec.LookPath("cgr-engine")
	if err != nil {
		return errors.New("Cannot find cgr-engine executable")
	}
	stopEngine()
	engine := exec.Command(enginePath, "-cdrs", "-config", path.Join(*dataDir, "conf", "cgrates.cfg"))
	if err := engine.Start(); err != nil {
		return fmt.Errorf("Cannot start cgr-engine: %s", err.Error())
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time to rater to fire up
	return nil
}

func stopEngine() error {
	exec.Command("pkill", "cgr-engine").Run() // Just to make sure another one is not running, bit brutal maybe we can fine tune it
	return nil
}

func TestEmptyTables(t *testing.T) {
	if !*testLocal {
		return
	}
	if *storDbType != utils.MYSQL {
		t.Fatal("Unsupported storDbType")
	}
	var mysql *engine.MySQLStorage
	if d, err := engine.NewMySQLStorage(cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass); err != nil {
		t.Fatal("Error on opening database connection: ", err)
	} else {
		mysql = d.(*engine.MySQLStorage)
	}
	for _, scriptName := range []string{engine.CREATE_CDRS_TABLES_SQL, engine.CREATE_COSTDETAILS_TABLES_SQL, engine.CREATE_MEDIATOR_TABLES_SQL, engine.CREATE_TARIFFPLAN_TABLES_SQL} {
		if err := mysql.CreateTablesFromScript(path.Join(*dataDir, "storage", *storDbType, scriptName)); err != nil {
			t.Fatal("Error on mysql creation: ", err.Error())
			return // No point in going further
		}
	}
	for _, tbl := range []string{utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_EXTRA} {
		if _, err := mysql.Db.Query(fmt.Sprintf("SELECT 1 from %s", tbl)); err != nil {
			t.Fatal(err.Error())
		}
	}
}

// Creates cdr files and starts the engine
func TestCreateCdrFiles(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := ioutil.WriteFile(path.Join(cfg.CdrcCdrInDir, "file1.csv"), []byte(fileContent1), 0644); err != nil {
		t.Fatal(err.Error)
	}
	if err := ioutil.WriteFile(path.Join(cfg.CdrcCdrInDir, "file2.csv"), []byte(fileContent2), 0644); err != nil {
		t.Fatal(err.Error)
	}
}

func TestProcessCdrDir(t *testing.T) {
	if !*testLocal {
		return
	}
	if cfg.CdrcCdrs == utils.INTERNAL { // For now we only test over network
		return
	}
	if err := startEngine(); err != nil {
		t.Fatal(err.Error())
	}
	cdrc, err := NewCdrc(cfg, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	if err := cdrc.processCdrDir(); err != nil {
		t.Error(err)
	}
	stopEngine()
}
