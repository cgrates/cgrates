/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
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

var cfgPath string
var cfg *config.CGRConfig
var cdrcCfgs map[string]*config.CdrcConfig
var cdrcCfg *config.CdrcConfig

var testLocal = flag.Bool("local", false, "Perform the tests only on local test environment, not by default.") // This flag will be passed here via "go test -local" args
var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
var storDbType = flag.String("stordb_type", "mysql", "The type of the storDb database <mysql>")
var waitRater = flag.Int("wait_rater", 300, "Number of miliseconds to wait for rater to start and cache")

var fileContent1 = `accid11,prepaid,out,cgrates.org,call,1001,1001,+4986517174963,2013-02-03 19:54:00,62,supplier1,172.16.1.1
accid12,prepaid,out,cgrates.org,call,1001,1001,+4986517174963,2013-02-03 19:54:00,62,supplier1,172.16.1.1
dummy_data
accid13,prepaid,out,cgrates.org,call,1001,1001,+4986517174963,2013-02-03 19:54:00,62,supplier1,172.16.1.1
`

var fileContent2 = `accid21,prepaid,out,cgrates.org,call,1001,1001,+4986517174963,2013-02-03 19:54:00,62,supplier1,172.16.1.1
accid22,prepaid,out,cgrates.org,call,1001,1001,+4986517174963,2013-02-03 19:54:00,62,supplier1,172.16.1.1
#accid1,prepaid,out,cgrates.org,call,1001,1001,+4986517174963,2013-02-03 19:54:00,62,supplier1,172.16.1.1
accid23,prepaid,out,cgrates.org,call,1001,1001,+4986517174963,2013-02-03 19:54:00,62,supplier1,172.16.1.1`

var fileContent3 = `accid31;prepaid;out;cgrates.org;call;1001;1001;+4986517174963;2013-02-03 19:54:00;62;supplier1;172.16.1.1
accid32;prepaid;out;cgrates.org;call;1001;1001;+4986517174963;2013-02-03 19:54:00;62;supplier1;172.16.1.1
#accid1;prepaid;out;cgrates.org;call;1001;1001;+4986517174963;2013-02-03 19:54:00;62;supplier1;172.16.1.1
accid33;prepaid;out;cgrates.org;call;1001;1001;+4986517174963;2013-02-03 19:54:00;62;supplier1;172.16.1.1`

func startEngine() error {
	enginePath, err := exec.LookPath("cgr-engine")
	if err != nil {
		return errors.New("Cannot find cgr-engine executable")
	}
	stopEngine()
	engine := exec.Command(enginePath, "-config", cfgPath)
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

// Need it here and not in init since Travis has no possibility to load local file
func TestCsvLclLoadConfigt(*testing.T) {
	if !*testLocal {
		return
	}
	cfgPath = path.Join(*dataDir, "conf", "samples", "apier")
	cfg, _ = config.NewCGRConfigFromFolder(cfgPath)
	if len(cfg.CdrcProfiles) > 0 {
		cdrcCfgs = cfg.CdrcProfiles["/var/log/cgrates/cdrc/in"]
	}
}

func TestCsvLclEmptyTables(t *testing.T) {
	if !*testLocal {
		return
	}
	if *storDbType != utils.MYSQL {
		t.Fatal("Unsupported storDbType")
	}
	mysql, err := engine.NewMySQLStorage(cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass, cfg.StorDBMaxOpenConns, cfg.StorDBMaxIdleConns)
	if err != nil {
		t.Fatal("Error on opening database connection: ", err)
	}
	for _, scriptName := range []string{utils.CREATE_CDRS_TABLES_SQL, utils.CREATE_TARIFFPLAN_TABLES_SQL} {
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
func TestCsvLclCreateCdrFiles(t *testing.T) {
	if !*testLocal {
		return
	}
	if cdrcCfgs == nil {
		t.Fatal("Empty default cdrc configuration")
	}
	for _, cdrcCfg = range cdrcCfgs { // Take the first config out, does not matter which one
		break
	}
	if err := os.RemoveAll(cdrcCfg.CdrInDir); err != nil {
		t.Fatal("Error removing folder: ", cdrcCfg.CdrInDir, err)
	}
	if err := os.MkdirAll(cdrcCfg.CdrInDir, 0755); err != nil {
		t.Fatal("Error creating folder: ", cdrcCfg.CdrInDir, err)
	}
	if err := os.RemoveAll(cdrcCfg.CdrOutDir); err != nil {
		t.Fatal("Error removing folder: ", cdrcCfg.CdrOutDir, err)
	}
	if err := os.MkdirAll(cdrcCfg.CdrOutDir, 0755); err != nil {
		t.Fatal("Error creating folder: ", cdrcCfg.CdrOutDir, err)
	}
	if err := ioutil.WriteFile(path.Join(cdrcCfg.CdrInDir, "file1.csv"), []byte(fileContent1), 0644); err != nil {
		t.Fatal(err.Error)
	}
	if err := ioutil.WriteFile(path.Join(cdrcCfg.CdrInDir, "file2.csv"), []byte(fileContent2), 0644); err != nil {
		t.Fatal(err.Error)
	}

}

func TestCsvLclProcessCdrDir(t *testing.T) {
	if !*testLocal {
		return
	}
	var cdrcCfg *config.CdrcConfig
	for _, cdrcCfg = range cdrcCfgs { // Take the first config out, does not matter which one
		break
	}
	if cdrcCfg.Cdrs == utils.INTERNAL { // For now we only test over network
		cdrcCfg.Cdrs = "127.0.0.1:2013"
	}
	if err := startEngine(); err != nil {
		t.Fatal(err.Error())
	}
	cdrc, err := NewCdrc(cdrcCfgs, true, nil, make(chan struct{}), "")
	if err != nil {
		t.Fatal(err.Error())
	}
	if err := cdrc.processCdrDir(); err != nil {
		t.Error(err)
	}
	stopEngine()
}

// Creates cdr files and starts the engine
func TestCsvLclCreateCdr3File(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := os.RemoveAll(cdrcCfg.CdrInDir); err != nil {
		t.Fatal("Error removing folder: ", cdrcCfg.CdrInDir, err)
	}
	if err := os.MkdirAll(cdrcCfg.CdrInDir, 0755); err != nil {
		t.Fatal("Error creating folder: ", cdrcCfg.CdrInDir, err)
	}
	if err := ioutil.WriteFile(path.Join(cdrcCfg.CdrInDir, "file3.csv"), []byte(fileContent3), 0644); err != nil {
		t.Fatal(err.Error)
	}
}

func TestCsvLclProcessCdr3Dir(t *testing.T) {
	if !*testLocal {
		return
	}
	if cdrcCfg.Cdrs == utils.INTERNAL { // For now we only test over network
		cdrcCfg.Cdrs = "127.0.0.1:2013"
	}
	if err := startEngine(); err != nil {
		t.Fatal(err.Error())
	}
	cdrc, err := NewCdrc(cdrcCfgs, true, nil, make(chan struct{}), "")
	if err != nil {
		t.Fatal(err.Error())
	}
	if err := cdrc.processCdrDir(); err != nil {
		t.Error(err)
	}
	stopEngine()
}
