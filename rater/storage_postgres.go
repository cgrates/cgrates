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

// Extracts destinations from StorDB on specific tariffplan id
func (psl *PostgresStorage) GetTPDestination( tpid, destTag string ) (*Destination, error) {
	return nil, nil
}

func (psl *PostgresStorage) GetActions(string) (as Actions, err error) {
	return
}

func (psl *PostgresStorage) SetActions(key string, as Actions) (err error) { return }

func (psl *PostgresStorage) GetUserBalance(string) (ub *UserBalance, err error) { return }

func (psl *PostgresStorage) SetUserBalance(ub *UserBalance) (err error) { return }

func (psl *PostgresStorage) GetActionTimings(key string) (ats ActionTimings, err error) { return }

func (psl *PostgresStorage) SetActionTimings(key string, ats ActionTimings) (err error) { return }

func (psl *PostgresStorage) GetAllActionTimings(tpid string) (ats map[string]ActionTimings, err error) { return }

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

func (psl *PostgresStorage) GetTpDestinations(tpid string) ([]*Destination, error) {
	var dests []*Destination
	rows, err := psl.Db.Query("SELECT * FROM tp_destinations WHERE tpid=?", tpid)
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

func (psl *PostgresStorage) GetTpRates(tpid string) (map[string][]*Rate, error) {
	rts := make(map[string][]*Rate)
	rows, err := psl.Db.Query("SELECT * FROM tp_rates WHERE tpid=?", tpid)
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

func (psl *PostgresStorage) GetTpTimings(tpid string) (map[string]*Timing, error) {
	tms := make(map[string]*Timing)
	rows, err := psl.Db.Query("SELECT * FROM tp_timings WHERE tpid=?", tpid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		var tpid, tag, years, months, month_days, week_days, start_time string
		if err := rows.Scan(&id, &tpid, &tag, &years, &months, &month_days, &week_days, &start_time); err != nil {
			return nil, err
		}
		tms[tag] = NewTiming(years, months, month_days, week_days, start_time)
	}
	return tms, rows.Err()
}

func (psl *PostgresStorage) GetTpRateTimings(tpid string) ([]*RateTiming, error) {
	var rts []*RateTiming
	rows, err := psl.Db.Query("SELECT * FROM tp_rate_timings WHERE tpid=?", tpid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		var weight float64
		var tpid, tag, rates_tag, timings_tag string
		if err := rows.Scan(&id, &tpid, &tag, &rates_tag, &timings_tag, &weight); err != nil {
			return nil, err
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

func (psl *PostgresStorage) GetTpRatingProfiles(tpid string) (map[string]*RatingProfile, error) {
	rpfs := make(map[string]*RatingProfile)
	rows, err := psl.Db.Query("SELECT * FROM tp_rate_profiles WHERE tpid=?", tpid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		var tpid, tenant, tor, direction, subject, fallbacksubject, rates_timing_tag, activation_time string

		if err := rows.Scan(&id, &tpid, &tenant, &tor, &direction, &subject, &fallbacksubject, &rates_timing_tag, &activation_time); err != nil {
			return nil, err
		}
		key := fmt.Sprintf("%s:%s:%s:%s", direction, tenant, tor, subject)
		rp, ok := rpfs[key]
		if !ok {
			rp = &RatingProfile{Id: key}
			rpfs[key] = rp
		}
		rp.tor = tor
		rp.direction = direction
		rp.subject = subject
		rp.fallbackSubject = fallbacksubject
		rp.ratesTimingTag = rates_timing_tag
		rp.activationTime = activation_time
	}
	return rpfs, rows.Err()
}
func (psl *PostgresStorage) GetTpActions(tpid string) (map[string][]*Action, error) {
	as := make(map[string][]*Action)
	rows, err := psl.Db.Query("SELECT * FROM tp_actions WHERE tpid=?", tpid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		var units, rate, minutes_weight, weight float64
		var tpid, tag, action, balances_tag, direction, destinations_tag, rate_type string
		if err := rows.Scan(&id, &tpid, &tag, &action, &balances_tag, &direction, &units, &destinations_tag, &rate_type, &rate, &minutes_weight, &weight); err != nil {
			return nil, err
		}
		var a *Action
		if balances_tag != MINUTES {
			a = &Action{
				ActionType: action,
				BalanceId:  balances_tag,
				Direction:  direction,
				Units:      units,
			}
		} else {
			var percent, price float64
			if rate_type == PERCENT {
				percent = rate
			}
			if rate_type == ABSOLUTE {
				price = rate
			}
			a = &Action{
				Id:         utils.GenUUID(),
				ActionType: action,
				BalanceId:  balances_tag,
				Direction:  direction,
				Weight:     weight,
				MinuteBucket: &MinuteBucket{
					Seconds:       units,
					Weight:        minutes_weight,
					Price:         price,
					Percent:       percent,
					DestinationId: destinations_tag,
				},
			}
		}
		as[tag] = append(as[tag], a)
	}
	return as, rows.Err()
}

func (psl *PostgresStorage) GetTpActionTimings(tpid string) (ats map[string][]*ActionTiming, err error) {
	ats = make(map[string][]*ActionTiming)
	rows, err := psl.Db.Query("SELECT * FROM tp_action_timings WHERE tpid=?", tpid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		var weight float64
		var tpid, tag, actions_tag, timings_tag string
		if err := rows.Scan(&id, &tpid, &tag, &actions_tag, &timings_tag, &weight); err != nil {
			return nil, err
		}

		at := &ActionTiming{
			Id:        utils.GenUUID(),
			Tag:       timings_tag,
			Weight:    weight,
			ActionsId: actions_tag,
		}
		ats[tag] = append(ats[tag], at)
	}
	return ats, rows.Err()
}

func (psl *PostgresStorage) GetTpActionTriggers(tpid string) (map[string][]*ActionTrigger, error) {
	ats := make(map[string][]*ActionTrigger)
	rows, err := psl.Db.Query("SELECT * FROM tp_action_triggers WHERE tpid=?", tpid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		var threshold, weight float64
		var tpid, tag, balances_tag, direction, destinations_tag, actions_tag string
		if err := rows.Scan(&id, &tpid, &tag, &balances_tag, &direction, &threshold, &destinations_tag, &actions_tag, &weight); err != nil {
			return nil, err
		}

		at := &ActionTrigger{
			Id:             utils.GenUUID(),
			BalanceId:      balances_tag,
			Direction:      direction,
			ThresholdValue: threshold,
			DestinationId:  destinations_tag,
			ActionsId:      actions_tag,
			Weight:         weight,
		}
		ats[tag] = append(ats[tag], at)
	}
	return ats, rows.Err()
}

func (psl *PostgresStorage) GetTpAccountActions(tpid string) ([]*AccountAction, error) {
	var acs []*AccountAction
	rows, err := psl.Db.Query("SELECT * FROM tp_account_actions WHERE tpid=?", tpid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		var tpid, tenant, account, direction, action_timings_tag, action_triggers_tag string
		if err := rows.Scan(&id, &tpid, &tenant, &account, &direction, &action_timings_tag, &action_triggers_tag); err != nil {
			return nil, err
		}

		aa := &AccountAction{
			Tenant:            tenant,
			Account:           account,
			Direction:         direction,
			ActionTimingsTag:  action_timings_tag,
			ActionTriggersTag: action_triggers_tag,
		}
		acs = append(acs, aa)
	}
	return acs, rows.Err()
}
