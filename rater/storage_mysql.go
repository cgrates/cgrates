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
	"github.com/cgrates/cgrates/utils"
	_ "github.com/go-sql-driver/mysql"
)

type MySQLStorage struct {
	Db *sql.DB
}

func NewMySQLStorage(host, port, name, user, password string) (DataStorage, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", user, password, host, port, name))
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

func (mys *MySQLStorage) GetActions(string) (as Actions, err error) {
	return
}

func (mys *MySQLStorage) SetActions(key string, as Actions) (err error) { return }

func (mys *MySQLStorage) GetUserBalance(string) (ub *UserBalance, err error) { return }

func (mys *MySQLStorage) SetUserBalance(ub *UserBalance) (err error) { return }

func (mys *MySQLStorage) GetActionTimings(key string) (ats ActionTimings, err error) { return }

func (mys *MySQLStorage) SetActionTimings(key string, ats ActionTimings) (err error) { return }

func (mys *MySQLStorage) GetAllActionTimings() (ats map[string]ActionTimings, err error) { return }

func (mys *MySQLStorage) LogCallCost(uuid, source string, cc *CallCost) (err error) {
	if mys.Db == nil {
		//timespans.Logger.Warning("Cannot write log to database.")
		return
	}
	tss, err := json.Marshal(cc.Timespans)
	if err != nil {
		Logger.Err(fmt.Sprintf("Error marshalling timespans to json: %v", err))
	}
	_, err = mys.Db.Exec(fmt.Sprintf("INSERT INTO callcosts VALUES ('NULL','%s', '%s','%s', '%s', '%s', '%s', '%s', '%s', %v, %v, '%s')",
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
	row := mys.Db.QueryRow(fmt.Sprintf("SELECT * FROM callcosts WHERE uuid='%s' AND source='%s'", uuid, source))
	var uuid_found string
	var timespansJson string
	err = row.Scan(&uuid_found, &cc.Direction, &cc.Tenant, &cc.TOR, &cc.Subject, &cc.Destination, &cc.Cost, &cc.ConnectFee, &timespansJson)
	err = json.Unmarshal([]byte(timespansJson), cc.Timespans)
	return
}

func (mys *MySQLStorage) LogActionTrigger(ubId, source string, at *ActionTrigger, as Actions) (err error) {
	return
}
func (mys *MySQLStorage) LogActionTiming(source string, at *ActionTiming, as Actions) (err error) {
	return
}
func (mys *MySQLStorage) LogError(uuid, source, errstr string) (err error) { return }

func (mys *MySQLStorage) SetCdr(cdr utils.CDR) (err error) {
	startTime, err := cdr.GetAnswerTime()
	if err != nil {
		return err
	}
	_, err = mys.Db.Exec(fmt.Sprintf("INSERT INTO cdrs_primary VALUES (NULL, '%s', '%s','%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', %d)",
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
	_, err = mys.Db.Exec(fmt.Sprintf("INSERT INTO cdrs_extra VALUES ('NULL','%s', '%s')",
		cdr.GetCgrId(),
		cdr.GetExtraFields(),
	))
	if err != nil {
		Logger.Err(fmt.Sprintf("failed to execute cdr insert statement: %v", err))
	}

	return
}

func (mys *MySQLStorage) SetRatedCdr(cdr utils.CDR, cc *CallCost) (err error) {
	_, err = mys.Db.Exec(fmt.Sprintf("INSERT INTO rated_cdrs VALUES ('%s', '%s', '%s', '%s')",
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

func (mys *MySQLStorage) GetAllDestinations(tpid string) ([]*Destination, error) {
	var dests []*Destination
	rows, err := mys.Db.Query("SELECT * FROM tp_destinations WHERE tpid=?", tpid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		var tpid, tag, prefix string
		if err := rows.Scan(id, tpid, &tag, &prefix); err != nil {
			return nil, err
		}
		var dest *Destination
		for _, d := range dests {
			if d.Id == tag {
				dest = d
				break
			}
		}
		if dest == nil {
			dest = &Destination{Id: tag}
			dests = append(dests, dest)
		}
		dest.Prefixes = append(dest.Prefixes, prefix)
	}
	return dests, rows.Err()
}

func (mys *MySQLStorage) GetAllRates(tpid string) (map[string][]*Rate, error) {
	rts := make(map[string][]*Rate)
	rows, err := mys.Db.Query("SELECT * FROM tp_rates WHERE tpid=?", tpid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		var tpid, tag, destinations_tag string
		var connect_fee, rate, priced_units, rate_increments float64
		if err := rows.Scan(&id, &tpid, &tag, &destinations_tag, &connect_fee, &rate, &priced_units, &rate_increments); err != nil {
			return nil, err
		}

		r := &Rate{
			DestinationsTag: destinations_tag,
			ConnectFee:      connect_fee,
			Price:           rate,
			PricedUnits:     priced_units,
			RateIncrements:  rate_increments,
		}

		rts[tag] = append(rts[tag], r)
	}
	return rts, rows.Err()
}

func (mys *MySQLStorage) GetAllTimings(string) (map[string][]*Timing, error) {
	tms := make(map[string][]*Timing)
	rows, err := dbr.db.Query("SELECT * FROM tp_timings WHERE tpid=?", tpid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		var tpid, tag, years, months, month_days, week_days, start_time string
		if err := rows.Scan(&id, &tpid, &tag, &years, &months, &month_days, &week_days, &start_time); err != nil {
			return nil, err
		}
		t := NewTiming(years, months, month_days, week_days, start_time)
		tms[tag] = append(tms[tag], t)
	}
	return tms, rows.Err()
}

func (mys *MySQLStorage) GetAllRateTimings(string) ([]*RateTiming, error) {
	var rts []*RateTiming
	rows, err := mys.Db.Query("SELECT * FROM tp_rate_timings WHERE tpid=?", tpid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		var weight float64
		var tpid, tag, rates_tag, timings_tag string
		if err := rows.Scan(&id, &tpid, &tag, &rates_tag, &timings_tag, &weight); err != nil {
			return err
		}
		rt := &RateTiming{
			Tag:        tag,
			RatesTag:   rates_tag,
			Weight:     weight,
			TimingsTag: timings_tag,
		}
		rts = append(rts, rt)
	}
	return rts, rows.Err()
}

func (mys *MySQLStorage) GetAllRatingProfiles(string) (map[string]*RatingProfile, error) {
	rpfs := make(map[string]*RatingProfile)
	rows, err := mys.Db.Query("SELECT * FROM tp_rate_profiles WHERE tpid=?", tpid)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int
		var tpid, tenant, tor, direction, subject, fallbacksubject, rates_timing_tag, activation_time string

		if err := rows.Scan(&id, &tpid, &tenant, &tor, &direction, &subject, &fallbacksubject, &rates_timing_tag, &activation_time); err != nil {
			return err
		}
		at, err := time.Parse(time.RFC3339, activation_time)
		if err != nil {
			return errors.New(fmt.Sprintf("Cannot parse activation time from %v", activation_time))
		}
		key := fmt.Sprintf("%s:%s:%s:%s", direction, tenant, tor, subject)
		rp, ok := rpfs[key]
		if !ok {
			rp = &RatingProfile{Id: key}
			rpfs[key] = rp
		}

	}
	return rpfs, rows.Err()
}
func (mys *MySQLStorage) GetAllActions(string) (map[string][]*Action, error) {
	return nil, nil
}
func (mys *MySQLStorage) GetAllActionTriggers(string) (map[string][]*ActionTrigger, error) {
	return nil, nil
}
