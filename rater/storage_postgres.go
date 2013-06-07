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
	_ "github.com/bmizerany/pq"
	"github.com/cgrates/cgrates/utils"
)

type PostgresStorage struct {
	Db *sql.DB
}

func NewPostgresStorage(host, port, name, user, password string) (DataStorage, error) {
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", host, port, name, user, password))
	if err != nil {
		return nil, err
	}
	return &PostgresStorage{db}, nil
}

func (psl *PostgresStorage) Close() {}

func (psl *PostgresStorage) Flush() (err error) {
	return
}

func (psl *PostgresStorage) GetRatingProfile(string) (rp *RatingProfile, err error) {
	/*row := psl.Db.QueryRow(fmt.Sprintf("SELECT * FROM ratingprofiles WHERE id='%s'", id))
	err = row.Scan(&rp, &cc.Direction, &cc.Tenant, &cc.TOR, &cc.Subject, &cc.Destination, &cc.Cost, &cc.ConnectFee, &timespansJson)
	err = json.Unmarshal([]byte(timespansJson), cc.Timespans)*/
	return
}

func (psl *PostgresStorage) SetRatingProfile(rp *RatingProfile) (err error) {
	return
}

func (psl *PostgresStorage) GetDestination(string) (d *Destination, err error) {
	return
}

func (psl *PostgresStorage) SetDestination(d *Destination) (err error) {
	return
}

func (psl *PostgresStorage) GetActions(string) (as Actions, err error) {
	return
}

func (psl *PostgresStorage) SetActions(key string, as Actions) (err error) { return }

func (psl *PostgresStorage) GetUserBalance(string) (ub *UserBalance, err error) { return }

func (psl *PostgresStorage) SetUserBalance(ub *UserBalance) (err error) { return }

func (psl *PostgresStorage) GetActionTimings(key string) (ats ActionTimings, err error) { return }

func (psl *PostgresStorage) SetActionTimings(key string, ats ActionTimings) (err error) { return }

func (psl *PostgresStorage) GetAllActionTimings() (ats map[string]ActionTimings, err error) { return }

func (psl *PostgresStorage) LogCallCost(uuid, source string, cc *CallCost) (err error) {
	if psl.Db == nil {
		//timespans.Logger.Warning("Cannot write log to database.")
		return
	}
	tss, err := json.Marshal(cc.Timespans)
	if err != nil {
		Logger.Err(fmt.Sprintf("Error marshalling timespans to json: %v", err))
	}
	_, err = psl.Db.Exec(fmt.Sprintf("INSERT INTO cdr VALUES ('%s', '%s','%s', '%s', '%s', '%s', '%s', '%s', %v, %v, '%s')",
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

func (psl *PostgresStorage) GetCallCostLog(uuid, source string) (cc *CallCost, err error) {
	row := psl.Db.QueryRow(fmt.Sprintf("SELECT * FROM cdr WHERE uuid='%s' AND source='%s'", uuid, source))
	var uuid_found string
	var timespansJson string
	err = row.Scan(&uuid_found, &cc.Direction, &cc.Tenant, &cc.TOR, &cc.Subject, &cc.Destination, &cc.Cost, &cc.ConnectFee, &timespansJson)
	err = json.Unmarshal([]byte(timespansJson), cc.Timespans)
	return
}

func (psl *PostgresStorage) LogActionTrigger(ubId, source string, at *ActionTrigger, as Actions) (err error) {
	return
}
func (psl *PostgresStorage) LogActionTiming(source string, at *ActionTiming, as Actions) (err error) {
	return
}
func (psl *PostgresStorage) LogError(uuid, source, errstr string) (err error) { return }

func (psl *PostgresStorage) SetCdr(cdr utils.CDR) (err error) {
	startTime, err := cdr.GetAnswerTime()
	if err != nil {
		return err
	}
	_, err = psl.Db.Exec(fmt.Sprintf("INSERT INTO cdrs_primary VALUES ('%s', '%s','%s', '%s', '%s', '%s', '%s', '%s', %v, %v, '%s')",
		cdr.GetCgrId(),
		cdr.GetAccId(),
		cdr.GetCdrHost(),
		cdr.GetReqType(),
		cdr.GetDirection(),
		cdr.GetTenant(),
		cdr.GetTOR(),
		cdr.GetAccount(),
		cdr.GetSubject(),
		cdr.GetDestination(),
		startTime,
		cdr.GetDuration(),
	))
	if err != nil {
		Logger.Err(fmt.Sprintf("failed to execute cdr insert statement: %v", err))
	}
	_, err = psl.Db.Exec(fmt.Sprintf("INSERT INTO cdrs_extra VALUES ('%s', '%s')",
		cdr.GetCgrId(),
		cdr.GetExtraFields(),
	))
	if err != nil {
		Logger.Err(fmt.Sprintf("failed to execute cdr insert statement: %v", err))
	}

	return
}

func (psl *PostgresStorage) SetRatedCdr(cdr utils.CDR, cc *CallCost) (err error) {
	if err != nil {
		return err
	}
	_, err = psl.Db.Exec(fmt.Sprintf("INSERT INTO cdrs_extra VALUES ('%s', '%s', '%s', '%s')",
		cdr.GetCgrId(),
		cc.Cost,
		"cgrcostid",
		"cdrsrc",
	))
	if err != nil {
		Logger.Err(fmt.Sprintf("failed to execute cdr insert statement: %v", err))
	}

	return
}

func (psl *PostgresStorage) GetDestinations(tpid string) ([]*Destination, error) {
	return nil, nil
}
