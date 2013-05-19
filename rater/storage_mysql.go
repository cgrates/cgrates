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

package rater

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/cgrates/cgrates/cdrs"
	_ "github.com/go-sql-driver/mysql"
)

type MySQLStorage struct {
	Db *sql.DB
}

var (
	mysql_schema = `
CREATE TABLE ratingprofile IF NOT EXISTS (
	id SERIAL PRIMARY KEY,
	fallbackkey VARCHAR(512),
);
CREATE TABLE ratingdestinations IF NOT EXISTS (
	id SERIAL PRIMARY KEY,
	ratingprofile INTEGER REFERENCES ratingprofile(id) ON DELETE CASCADE,
	destination INTEGER REFERENCES destination(id) ON DELETE CASCADE
);
CREATE TABLE destination IF NOT EXISTS (
	id SERIAL PRIMARY KEY,
	ratingprofile INTEGER REFERENCES ratingprofile(id) ON DELETE CASCADE,
	name VARCHAR(512),
	prefixes TEXT
);
CREATE TABLE activationprofile  IF NOT EXISTS(
	id SERIAL PRIMARY KEY,
	destination INTEGER REFERENCES destination(id) ON DELETE CASCADE,
	activationtime TIMESTAMP
);
CREATE TABLE interval IF NOT EXISTS(
	id SERIAL PRIMARY KEY,
	activationprofile INTEGER REFERENCES activationprofile(id) ON DELETE CASCADE,
	years TEXT,
	months TEXT,
	monthdays TEXT,
	weekdays TEXT,
	starttime TIMESTAMP,
	endtime TIMESTAMP,
	weight FLOAT8,
	connectfee FLOAT8,
	price FLOAT8,
	pricedunits FLOAT8,
	rateincrements FLOAT8
);
CREATE TABLE minutebucket IF NOT EXISTS(
	id SERIAL PRIMARY KEY,
	destination INTEGER REFERENCES destination(id) ON DELETE CASCADE,
	seconds FLOAT8,
	weight FLOAT8,
	price FLOAT8,
	percent FLOAT8
);
CREATE TABLE unitcounter IF NOT EXISTS(
	id SERIAL PRIMARY KEY,
	direction TEXT,
	balance TEXT,
	units FLOAT8
);
CREATE TABLE unitcounterbucket IF NOT EXISTS(
	id SERIAL PRIMARY KEY,
	unitcounter INTEGER REFERENCES unitcounter(id) ON DELETE CASCADE,
	minutebucket INTEGER REFERENCES minutebucket(id) ON DELETE CASCADE
);
CREATE TABLE actiontrigger IF NOT EXISTS(
	id SERIAL PRIMARY KEY,
	destination INTEGER REFERENCES destination(id) ON DELETE CASCADE,
	actions INTEGER REFERENCES action(id) ON DELETE CASCADE,
	balance TEXT,
	direction TEXT,
	thresholdvalue FLOAT8,
	weight FLOAT8,
	executed BOOL
);
CREATE TABLE balance IF NOT EXISTS(
	id SERIAL PRIMARY KEY,
	name TEXT;
	value FLOAT8
);
CREATE TABLE userbalance IF NOT EXISTS(
	id SERIAL PRIMARY KEY,
	unitcounter INTEGER REFERENCES unitcounter(id) ON DELETE CASCADE,
	minutebucket INTEGER REFERENCES minutebucket(id) ON DELETE CASCADE
	actiontriggers INTEGER REFERENCES actiontrigger(id) ON DELETE CASCADE,
	balances INTEGER REFERENCES balance(id) ON DELETE CASCADE,
	type TEXT
);
CREATE TABLE actiontiming IF NOT EXISTS(
	id SERIAL PRIMARY KEY,
	tag TEXT,
	userbalances INTEGER REFERENCES userbalance(id) ON DELETE CASCADE,
	timing INTEGER REFERENCES interval(id) ON DELETE CASCADE,
	actions INTEGER REFERENCES action(id) ON DELETE CASCADE,
	weight FLOAT8
);
CREATE TABLE action IF NOT EXISTS(
	id SERIAL PRIMARY KEY,
	minutebucket INTEGER REFERENCES minutebucket(id) ON DELETE CASCADE,
	actiontype TEXT,
	balance TEXT,
	direction TEXT,
	units FLOAT8,
	weight FLOAT8
);
`
)

func NewMySQLStorage(host, port, name, user, password string) (DataStorage, error) {
	db, err := sql.Open("mysql", "cgrates:testus@tcp(192.168.0.17:3306)/cgrates?charset=utf8")
	if err != nil {
		return nil, err
	}
	return &MySQLStorage{db}, nil
}

func (mys *MySQLStorage) Close() {}

func (mys *MySQLStorage) Flush() (err error) {
	return
}

func (mys *MySQLStorage) GetRatingProfile(string) (rp *RatingProfile, err error) {
	/*row := mys.Db.QueryRow(fmt.Sprintf("SELECT * FROM ratingprofiles WHERE id='%s'", id))
	err = row.Scan(&rp, &cc.Direction, &cc.Tenant, &cc.TOR, &cc.Subject, &cc.Destination, &cc.Cost, &cc.ConnectFee, &timespansJson)
	err = json.Unmarshal([]byte(timespansJson), cc.Timespans)*/
	return
}

func (mys *MySQLStorage) SetRatingProfile(rp *RatingProfile) (err error) {
	return
}

func (mys *MySQLStorage) GetDestination(string) (d *Destination, err error) {
	return
}

func (mys *MySQLStorage) SetDestination(d *Destination) (err error) {
	return
}

func (mys *MySQLStorage) GetActions(string) (as []*Action, err error) {
	return
}

func (mys *MySQLStorage) SetActions(key string, as []*Action) (err error) { return }

func (mys *MySQLStorage) GetUserBalance(string) (ub *UserBalance, err error) { return }

func (mys *MySQLStorage) SetUserBalance(ub *UserBalance) (err error) { return }

func (mys *MySQLStorage) GetActionTimings(key string) (ats []*ActionTiming, err error) { return }

func (mys *MySQLStorage) SetActionTimings(key string, ats []*ActionTiming) (err error) { return }

func (mys *MySQLStorage) GetAllActionTimings() (ats map[string][]*ActionTiming, err error) { return }

func (mys *MySQLStorage) LogCallCost(uuid, source string, cc *CallCost) (err error) {
	if mys.Db == nil {
		//timespans.Logger.Warning("Cannot write log to database.")
		return
	}
	tss, err := json.Marshal(cc.Timespans)
	if err != nil {
		Logger.Err(fmt.Sprintf("Error marshalling timespans to json: %v", err))
	}
	_, err = mys.Db.Exec(fmt.Sprintf("INSERT INTO cdr VALUES ('%s', '%s','%s', '%s', '%s', '%s', '%s', '%s', %v, %v, '%s')",
		uuid,
		source,
		cc.Direction,
		cc.Tenant,
		cc.TOR,
		cc.Subject,
		cc.Account,
		cc.Destination,
		cc.Cost,
		cc.ConnectFee,
		tss))
	if err != nil {
		Logger.Err(fmt.Sprintf("failed to execute insert statement: %v", err))
	}
	return
}

func (mys *MySQLStorage) GetCallCostLog(uuid, source string) (cc *CallCost, err error) {
	row := mys.Db.QueryRow(fmt.Sprintf("SELECT * FROM cdr WHERE uuid='%s' AND source='%s'", uuid, source))
	var uuid_found string
	var timespansJson string
	err = row.Scan(&uuid_found, &cc.Direction, &cc.Tenant, &cc.TOR, &cc.Subject, &cc.Destination, &cc.Cost, &cc.ConnectFee, &timespansJson)
	err = json.Unmarshal([]byte(timespansJson), cc.Timespans)
	return
}

func (mys *MySQLStorage) LogActionTrigger(ubId, source string, at *ActionTrigger, as []*Action) (err error) {
	return
}
func (mys *MySQLStorage) LogActionTiming(source string, at *ActionTiming, as []*Action) (err error) {
	return
}
func (mys *MySQLStorage) LogError(uuid, source, errstr string) (err error) { return }

func (mys *MySQLStorage) GetCdr(string) (cdrs.CDR, error) {
	return nil, nil
}
func (mys *MySQLStorage) SetCdr(string, cdrs.CDR) error {
	return nil
}
