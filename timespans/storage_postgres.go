/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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

package timespans

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type PostgresStorage struct {
	Db *sql.DB
}

func NewPostgresStorage(host, port, name, user, password string) (StorageGetter, error) {
	db, err := sql.Open("postgress", fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", host, port, name, user, password))
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

func (psl *PostgresStorage) GetActions(string) (as []*Action, err error) {
	return
}

func (psl *PostgresStorage) SetActions(key string, as []*Action) (err error) { return }

func (psl *PostgresStorage) GetUserBalance(string) (ub *UserBalance, err error) { return }

func (psl *PostgresStorage) SetUserBalance(ub *UserBalance) (err error) { return }

func (psl *PostgresStorage) GetActionTimings(key string) (ats []*ActionTiming, err error) { return }

func (psl *PostgresStorage) SetActionTimings(key string, ats []*ActionTiming) (err error) { return }

func (psl *PostgresStorage) GetAllActionTimings() (ats map[string][]*ActionTiming, err error) { return }

func (psl *PostgresStorage) LogCallCost(uuid string, cc *CallCost) (err error) {
	if psl.Db == nil {
		//timespans.Logger.Warning("Cannot write log to database.")
		return
	}
	tss, err := json.Marshal(cc.Timespans)
	if err != nil {
		Logger.Err(fmt.Sprintf("Error marshalling timespans to json: %v", err))
	}
	_, err = psl.Db.Exec(fmt.Sprintf("INSERT INTO cdr VALUES ('%s','%s', '%s', '%s', '%s', '%s', '%s', %v, %v, '%s')",
		uuid,
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

func (psl *PostgresStorage) GetCallCostLog(uuid string) (cc *CallCost, err error) {
	row := psl.Db.QueryRow(fmt.Sprintf("SELECT * FROM cdr WHERE uuid='%s'", uuid))
	var uuid_found string
	var timespansJson string
	err = row.Scan(&uuid_found, &cc.Direction, &cc.Tenant, &cc.TOR, &cc.Subject, &cc.Destination, &cc.Cost, &cc.ConnectFee, &timespansJson)
	err = json.Unmarshal([]byte(timespansJson), cc.Timespans)
	return
}

func (psl *PostgresStorage) LogActionTrigger(ubId string, at *ActionTrigger, as []*Action) (err error) {
	return
}
func (psl *PostgresStorage) LogActionTiming(at *ActionTiming, as []*Action) (err error) { return }
