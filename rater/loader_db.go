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
	"errors"
	"fmt"
	"log"
	"time"
)

type DbReader struct {
	tpid              string
	db                *sql.DB
	actions           map[string][]*Action
	actionsTimings    map[string][]*ActionTiming
	actionsTriggers   map[string][]*ActionTrigger
	accountActions    []*UserBalance
	destinations      []*Destination
	rates             map[string][]*Rate
	timings           map[string][]*Timing
	activationPeriods map[string]*ActivationPeriod
	ratingProfiles    map[string]*RatingProfile
}

func NewDbReader(db *sql.DB, tpid string) *DbReader {
	c := new(DbReader)
	c.db = db
	c.tpid = tpid
	c.actions = make(map[string][]*Action)
	c.actionsTimings = make(map[string][]*ActionTiming)
	c.actionsTriggers = make(map[string][]*ActionTrigger)
	c.rates = make(map[string][]*Rate)
	c.timings = make(map[string][]*Timing)
	c.activationPeriods = make(map[string]*ActivationPeriod)
	c.ratingProfiles = make(map[string]*RatingProfile)
	return c
}

func (dbr *DbReader) WriteToDatabase(storage DataStorage, flush, verbose bool) (err error) {
	if flush {
		storage.Flush()
	}
	if verbose {
		log.Print("Destinations")
	}
	for _, d := range dbr.destinations {
		err = storage.SetDestination(d)
		if err != nil {
			return err
		}
		if verbose {
			log.Print(d.Id, " : ", d.Prefixes)
		}
	}
	if verbose {
		log.Print("Rating profiles")
	}
	for _, rp := range dbr.ratingProfiles {
		err = storage.SetRatingProfile(rp)
		if err != nil {
			return err
		}
		if verbose {
			log.Print(rp.Id)
		}
	}
	if verbose {
		log.Print("Action timings")
	}
	for k, ats := range dbr.actionsTimings {
		err = storage.SetActionTimings(k, ats)
		if err != nil {
			return err
		}
		if verbose {
			log.Println(k)
		}
	}
	if verbose {
		log.Print("Actions")
	}
	for k, as := range dbr.actions {
		err = storage.SetActions(k, as)
		if err != nil {
			return err
		}
		if verbose {
			log.Println(k)
		}
	}
	if verbose {
		log.Print("Account actions")
	}
	for _, ub := range dbr.accountActions {
		err = storage.SetUserBalance(ub)
		if err != nil {
			return err
		}
		if verbose {
			log.Println(ub.Id)
		}
	}
	return
}

func (dbr *DbReader) LoadDestinations() error {
	rows, err := dbr.db.Query("SELECT * FROM tp_destinations WHERE tpid=?", dbr.tpid)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int
		var tpid, tag, prefix string
		if err := rows.Scan(id, tpid, &tag, &prefix); err != nil {
			return err
		}
		var dest *Destination
		for _, d := range dbr.destinations {
			if d.Id == tag {
				dest = d
				break
			}
		}
		if dest == nil {
			dest = &Destination{Id: tag}
			dbr.destinations = append(dbr.destinations, dest)
		}
		dest.Prefixes = append(dest.Prefixes, prefix)
	}
	return rows.Err()
}

func (dbr *DbReader) LoadRates() error {
	rows, err := dbr.db.Query("SELECT * FROM tp_rates WHERE tpid=?", dbr.tpid)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int
		var tpid, tag, destinations_tag string
		var connect_fee, rate, priced_units, rate_increments float64
		if err := rows.Scan(&id, &tpid, &tag, &destinations_tag, &connect_fee, &rate, &priced_units, &rate_increments); err != nil {
			return err
		}

		r := &Rate{
			DestinationsTag: destinations_tag,
			ConnectFee:      connect_fee,
			Price:           rate,
			PricedUnits:     priced_units,
			RateIncrements:  rate_increments,
		}

		dbr.rates[tag] = append(dbr.rates[tag], r)
	}
	return rows.Err()
}

func (dbr *DbReader) LoadTimings() error {
	rows, err := dbr.db.Query("SELECT * FROM tp_timings WHERE tpid=?", dbr.tpid)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int
		var tpid, tag, years, months, month_days, week_days, start_time string
		if err := rows.Scan(&id, &tpid, &tag, &years, &months, &month_days, &week_days, &start_time); err != nil {
			return err
		}
		t := NewTiming(years, months, month_days, week_days, start_time)
		dbr.timings[tag] = append(dbr.timings[tag], t)
	}
	return rows.Err()
}

func (dbr *DbReader) LoadRateTimings() error {
	rows, err := dbr.db.Query("SELECT * FROM tp_rate_timings WHERE tpid=?", dbr.tpid)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int
		var weight float64
		var tpid, tag, rates_tag, timings_tag string
		if err := rows.Scan(&id, &tpid, &tag, &rates_tag, &timings_tag, &weight); err != nil {
			return err
		}
		ts, exists := dbr.timings[timings_tag]
		if !exists {
			return errors.New(fmt.Sprintf("Could not get timing for tag %v", timings_tag))
		}
		for _, t := range ts {
			rt := &RateTiming{
				RatesTag: rates_tag,
				Weight:   weight,
				timing:   t,
			}
			rs, exists := dbr.rates[rates_tag]
			if !exists {
				return errors.New(fmt.Sprintf("Could not find rate for tag %v", rates_tag))
			}
			for _, r := range rs {
				_, exists := dbr.activationPeriods[tag]
				if !exists {
					dbr.activationPeriods[tag] = &ActivationPeriod{}
				}
				dbr.activationPeriods[tag].AddIntervalIfNotPresent(rt.GetInterval(r))
			}
		}
	}
	return rows.Err()
}

func (dbr *DbReader) LoadRatingProfiles() error {
	rows, err := dbr.db.Query("SELECT * FROM tp_rate_profiles WHERE tpid=?", dbr.tpid)
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
		rp, ok := dbr.ratingProfiles[key]
		if !ok {
			rp = &RatingProfile{Id: key}
			dbr.ratingProfiles[key] = rp
		}
		for _, d := range dbr.destinations {
			ap, exists := dbr.activationPeriods[rates_timing_tag]
			if !exists {
				return errors.New(fmt.Sprintf("Could not load rating timing for tag: %v", rates_timing_tag))
			}
			newAP := &ActivationPeriod{ActivationTime: at}
			//copy(newAP.Intervals, ap.Intervals)
			newAP.Intervals = append(newAP.Intervals, ap.Intervals...)
			rp.AddActivationPeriodIfNotPresent(d.Id, newAP)
			if fallbacksubject != "" {
				rp.FallbackKey = fmt.Sprintf("%s:%s:%s:%s", direction, tenant, tor, fallbacksubject)
			}
		}
	}
	return rows.Err()
}

func (dbr *DbReader) LoadActions() error {
	rows, err := dbr.db.Query("SELECT * FROM tp_actions WHERE tpid=?", dbr.tpid)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int
		var units, rate, minutes_weight, weight float64
		var tpid, tag, action, balances_tag, direction, destinations_tag, rate_type string
		if err := rows.Scan(&id, &tpid, &tag, &action, &balances_tag, &direction, &units, &destinations_tag, &rate_type, &rate, &minutes_weight, &weight); err != nil {
			return err
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
				Id:         GenUUID(),
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
		dbr.actions[tag] = append(dbr.actions[tag], a)
	}
	return rows.Err()
}

func (dbr *DbReader) LoadActionTimings() error {
	rows, err := dbr.db.Query("SELECT * FROM tp_action_timings WHERE tpid=?", dbr.tpid)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int
		var weight float64
		var tpid, tag, actions_tag, timings_tag string
		if err := rows.Scan(&id, &tpid, &tag, &actions_tag, &timings_tag, &weight); err != nil {
			return err
		}
		_, exists := dbr.actions[actions_tag]
		if !exists {
			return errors.New(fmt.Sprintf("ActionTiming: Could not load the action for tag: %v", actions_tag))
		}
		ts, exists := dbr.timings[timings_tag]
		if !exists {
			return errors.New(fmt.Sprintf("ActionTiming: Could not load the timing for tag: %v", timings_tag))
		}
		for _, t := range ts {
			at := &ActionTiming{
				Id:     GenUUID(),
				Tag:    timings_tag,
				Weight: weight,
				Timing: &Interval{
					Months:    t.Months,
					MonthDays: t.MonthDays,
					WeekDays:  t.WeekDays,
					StartTime: t.StartTime,
				},
				ActionsId: actions_tag,
			}
			dbr.actionsTimings[tag] = append(dbr.actionsTimings[tag], at)
		}
	}
	return rows.Err()
}

func (dbr *DbReader) LoadActionTriggers() error {
	rows, err := dbr.db.Query("SELECT * FROM tp_action_triggers WHERE tpid=?", dbr.tpid)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int
		var threshold, weight float64
		var tpid, tag, balances_tag, direction, destinations_tag, actions_tag string
		if err := rows.Scan(&id, &tpid, &tag, &balances_tag, &direction, &threshold, &destinations_tag, &actions_tag, &weight); err != nil {
			return err
		}

		at := &ActionTrigger{
			Id:             GenUUID(),
			BalanceId:      balances_tag,
			Direction:      direction,
			ThresholdValue: threshold,
			DestinationId:  destinations_tag,
			ActionsId:      actions_tag,
			Weight:         weight,
		}
		dbr.actionsTriggers[tag] = append(dbr.actionsTriggers[tag], at)
	}
	return rows.Err()
}

func (dbr *DbReader) LoadAccountActions() error {
	rows, err := dbr.db.Query("SELECT * FROM tp_account_actions WHERE tpid=?", dbr.tpid)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int
		var tpid, tenant, account, direction, action_timings_tag, action_triggers_tag string
		if err := rows.Scan(&id, &tpid, &tenant, &account, &direction, &action_timings_tag, &action_triggers_tag); err != nil {
			return err
		}

		tag := fmt.Sprintf("%s:%s:%s", direction, tenant, account)
		aTriggers, exists := dbr.actionsTriggers[action_triggers_tag]
		if action_triggers_tag != "" && !exists {
			// only return error if there was something ther for the tag
			return errors.New(fmt.Sprintf("Could not get action triggers for tag %v", action_triggers_tag))
		}
		ub := &UserBalance{
			Type:           UB_TYPE_PREPAID,
			Id:             tag,
			ActionTriggers: aTriggers,
		}
		dbr.accountActions = append(dbr.accountActions, ub)

		aTimings, exists := dbr.actionsTimings[action_timings_tag]
		if !exists {
			log.Printf("Could not get action timing for tag %v", action_timings_tag)
			// must not continue here
		}
		for _, at := range aTimings {
			at.UserBalanceIds = append(at.UserBalanceIds, tag)
		}
	}
	return rows.Err()
}
