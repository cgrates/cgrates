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

package engine

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/cgrates/cgrates/utils"
	"time"
)

type SQLStorage struct {
	Db *sql.DB
}

func (self *SQLStorage) Close() {
	self.Db.Close()
}

func (self *SQLStorage) Flush() (err error) {
	return
}

// Return a list with all TPids defined in the system, even if incomplete, isolated in some table.
func (self *SQLStorage) GetTPIds() ([]string, error) {
	rows, err := self.Db.Query(
		fmt.Sprintf("(SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s)", utils.TBL_TP_TIMINGS, utils.TBL_TP_DESTINATIONS, utils.TBL_TP_RATES, utils.TBL_TP_DESTINATION_RATES, utils.TBL_TP_RATING_PLANS, utils.TBL_TP_RATE_PROFILES))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if i == 0 {
		return nil, nil
	}
	return ids, nil
}

func (self *SQLStorage) SetTPTiming(tpid string, tm *utils.TPTiming) error {
	if _, err := self.Db.Exec(fmt.Sprintf("INSERT INTO %s (tpid, tag, years, months, month_days, week_days, time) VALUES('%s','%s','%s','%s','%s','%s','%s') ON DUPLICATE KEY UPDATE years=values(years), months=values(months), month_days=values(month_days), week_days=values(week_days), time=values(time)",
		utils.TBL_TP_TIMINGS, tpid, tm.Id, tm.Years.Serialize(";"), tm.Months.Serialize(";"), tm.MonthDays.Serialize(";"),
		tm.WeekDays.Serialize(";"), tm.StartTime)); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) ExistsTPTiming(tpid, tmId string) (bool, error) {
	var exists bool
	err := self.Db.QueryRow(fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s WHERE tpid='%s' AND tag='%s')", utils.TBL_TP_TIMINGS, tpid, tmId)).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (self *SQLStorage) GetTPTiming(tpid, tmId string) (*utils.TPTiming, error) {
	var years, months, monthDays, weekDays, time string
	err := self.Db.QueryRow(fmt.Sprintf("SELECT years, months, month_days, week_days, time FROM %s WHERE tpid='%s' AND tag='%s' LIMIT 1",
		utils.TBL_TP_TIMINGS, tpid, tmId)).Scan(&years, &months, &monthDays, &weekDays, &time)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	}
	return NewTiming(tmId, years, months, monthDays, weekDays, time), nil
}

func (self *SQLStorage) GetTPTimingIds(tpid string) ([]string, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT DISTINCT tag FROM %s where tpid='%s'", utils.TBL_TP_TIMINGS, tpid))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if i == 0 {
		return nil, nil
	}
	return ids, nil
}



func (self *SQLStorage) RemTPData(table, tpid string, args ...string) error {
	q := fmt.Sprintf("DELETE FROM %s WHERE tpid='%s' AND tag='%s'", table, tpid, args[0])
	switch table {
	case utils.TBL_TP_RATE_PROFILES:
		q = fmt.Sprintf("DELETE FROM %s WHERE tpid='%s' AND loadid='%s' AND tenant='%s' AND tor='%s' AND direction='%s' AND subject='%s'", 
				table, tpid, args[0], args[1], args[2], args[3], args[4])
	case utils.TBL_TP_ACCOUNT_ACTIONS:
		q = fmt.Sprintf("DELETE FROM %s WHERE tpid='%s' AND loadid='%s' AND tenant='%s' AND account='%s' AND direction='%s'", 
				table, tpid, args[0], args[1], args[2], args[3])
	}
	if _, err := self.Db.Exec(q); err != nil {
		return err
	}
	return nil
}

// Extracts destinations from StorDB on specific tariffplan id
func (self *SQLStorage) GetTPDestinationIds(tpid string) ([]string, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT DISTINCT tag FROM %s where tpid='%s'", utils.TBL_TP_DESTINATIONS, tpid))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if i == 0 {
		return nil, nil
	}
	return ids, nil
}

func (self *SQLStorage) ExistsTPDestination(tpid, destTag string) (bool, error) {
	var exists bool
	err := self.Db.QueryRow(fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s WHERE tpid='%s' AND tag='%s')", utils.TBL_TP_DESTINATIONS, tpid, destTag)).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Extracts destinations from StorDB on specific tariffplan id
func (self *SQLStorage) GetTPDestination(tpid, destTag string) (*Destination, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT prefix FROM %s WHERE tpid='%s' AND tag='%s'", utils.TBL_TP_DESTINATIONS, tpid, destTag))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	d := &Destination{Id: destTag}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one prefix
		var pref string
		err = rows.Scan(&pref)
		if err != nil {
			return nil, err
		}
		d.Prefixes = append(d.Prefixes, pref)
	}
	if i == 0 {
		return nil, nil
	}
	return d, nil
}

func (self *SQLStorage) SetTPDestination(tpid string, dest *Destination) error {
	if len(dest.Prefixes) == 0 {
		return nil
	}
	vals := ""
	for idx, prefix := range dest.Prefixes {
		if idx != 0 {
			vals += ","
		}
		vals += fmt.Sprintf("('%s','%s','%s')", tpid, dest.Id, prefix)
	}
	q := fmt.Sprintf("INSERT INTO %s (tpid, tag, prefix) VALUES %s ON DUPLICATE KEY UPDATE prefix=values(prefix)", 
			utils.TBL_TP_DESTINATIONS, vals)
	if _, err := self.Db.Exec(q); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) ExistsTPRate(tpid, rtId string) (bool, error) {
	var exists bool
	err := self.Db.QueryRow(fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s WHERE tpid='%s' AND tag='%s')", utils.TBL_TP_RATES, tpid, rtId)).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (self *SQLStorage) SetTPRates(tpid string, rts map[string][]*utils.RateSlot) error {
	if len(rts) == 0 {
		return nil //Nothing to set
	}
	vals := ""
	i := 0
	for rtId, rtRows := range rts {
		for _, rt := range rtRows {
			if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
				vals += ","
			}
			vals += fmt.Sprintf("('%s', '%s', %f, %f, '%s', '%s','%s','%s', %d)",
				tpid, rtId, rt.ConnectFee, rt.Rate, rt.RateUnit, rt.RateIncrement, rt.GroupIntervalStart,
				rt.RoundingMethod, rt.RoundingDecimals)
			i++
		}
	}
	qry := fmt.Sprintf("INSERT INTO %s (tpid, tag, connect_fee, rate, rate_unit, rate_increment, group_interval_start, rounding_method, rounding_decimals) VALUES %s ON DUPLICATE KEY UPDATE connect_fee=values(connect_fee), rate=values(rate), rate_increment=values(rate_increment), group_interval_start=values(group_interval_start), rounding_method=values(rounding_method), rounding_decimals=values(rounding_decimals)", utils.TBL_TP_RATES, vals)
	if _, err := self.Db.Exec(qry); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) GetTPRate(tpid, rtId string) (*utils.TPRate, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT connect_fee, rate, rate_unit, rate_increment, group_interval_start, rounding_method, rounding_decimals FROM %s WHERE tpid='%s' AND tag='%s'", utils.TBL_TP_RATES, tpid, rtId))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rt := &utils.TPRate{TPid: tpid, RateId: rtId}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one prefix
		var connectFee, rate float64
		var roundingDecimals int
		var rateUnit, rateIncrement, groupIntervalStart, roundingMethod string
		err = rows.Scan(&connectFee, &rate, &rateUnit, &rateIncrement, &groupIntervalStart, &roundingMethod, &roundingDecimals)
		if err != nil {
			return nil, err
		}
		if rs, err := utils.NewRateSlot(connectFee, rate, rateUnit, rateIncrement, groupIntervalStart, roundingMethod, roundingDecimals); err!=nil {
			return nil, err
		} else {
			rt.RateSlots = append(rt.RateSlots, rs) 
		}
	}
	if i == 0 {
		return nil, nil
	}
	return rt, nil
}

func (self *SQLStorage) GetTPRateIds(tpid string) ([]string, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT DISTINCT tag FROM %s where tpid='%s'", utils.TBL_TP_RATES, tpid))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if i == 0 {
		return nil, nil
	}
	return ids, nil
}

func (self *SQLStorage) ExistsTPDestinationRate(tpid, drId string) (bool, error) {
	var exists bool
	err := self.Db.QueryRow(fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s WHERE tpid='%s' AND tag='%s')", utils.TBL_TP_DESTINATION_RATES, tpid, drId)).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (self *SQLStorage) SetTPDestinationRates(tpid string, drs map[string][]*utils.DestinationRate) error {
	if len(drs) == 0 {
		return nil //Nothing to set
	}
	vals := ""
	i := 0
	for drId, drRows := range drs {
		for _, dr := range drRows {
			if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
				vals += ","
			}
			vals += fmt.Sprintf("('%s','%s','%s','%s')",
				tpid, drId, dr.DestinationId, dr.RateId)
			i++
		}
	}
	qry := fmt.Sprintf("INSERT INTO %s (tpid,tag,destinations_tag,rates_tag) VALUES %s ON DUPLICATE KEY UPDATE destinations_tag=values(destinations_tag),rates_tag=values(rates_tag)", utils.TBL_TP_DESTINATION_RATES, vals)
	if _, err := self.Db.Exec(qry); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) GetTPDestinationRate(tpid, drId string) (*utils.TPDestinationRate, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT destinations_tag, rates_tag FROM %s WHERE tpid='%s' AND tag='%s'", utils.TBL_TP_DESTINATION_RATES, tpid, drId))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dr := &utils.TPDestinationRate{TPid: tpid, DestinationRateId: drId}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one prefix
		var dstTag, ratesTag string
		err = rows.Scan(&dstTag, &ratesTag)
		if err != nil {
			return nil, err
		}
		dr.DestinationRates = append(dr.DestinationRates, &utils.DestinationRate{dstTag, ratesTag, nil})
	}
	if i == 0 {
		return nil, nil
	}
	return dr, nil
}

func (self *SQLStorage) GetTPDestinationRateIds(tpid string) ([]string, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT DISTINCT tag FROM %s where tpid='%s'", utils.TBL_TP_DESTINATION_RATES, tpid))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if i == 0 {
		return nil, nil
	}
	return ids, nil
}

func (self *SQLStorage) ExistsTPRatingPlan(tpid, drtId string) (bool, error) {
	var exists bool
	err := self.Db.QueryRow(fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s WHERE tpid='%s' AND tag='%s')", utils.TBL_TP_RATING_PLANS, tpid, drtId)).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (self *SQLStorage) SetTPRatingPlans(tpid string, drts map[string][]*utils.TPRatingPlanBinding) error {
	if len(drts) == 0 {
		return nil //Nothing to set
	}
	vals := ""
	i := 0
	for drtId, drtRows := range drts {
		for _, drt := range drtRows {
			if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
				vals += ","
			}
			vals += fmt.Sprintf("('%s','%s','%s','%s',%f)",
				tpid, drtId, drt.DestinationRatesId, drt.TimingId, drt.Weight)
			i++
		}
	}
	qry := fmt.Sprintf("INSERT INTO %s (tpid, tag, destrates_tag, timing_tag, weight) VALUES %s ON DUPLICATE KEY UPDATE weight=values(weight)", utils.TBL_TP_RATING_PLANS, vals)
	if _, err := self.Db.Exec(qry); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) GetTPRatingPlan(tpid, drtId string) (*utils.TPRatingPlan, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT destrates_tag, timing_tag, weight from %s where tpid='%s' and tag='%s'", utils.TBL_TP_RATING_PLANS, tpid, drtId))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	drt := &utils.TPRatingPlan{TPid: tpid, RatingPlanId: drtId}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one result
		var drTag, timingTag string
		var weight float64
		err = rows.Scan(&drTag, &timingTag, &weight)
		if err != nil {
			return nil, err
		}
		drt.RatingPlanBindings = append(drt.RatingPlanBindings, &utils.TPRatingPlanBinding{DestinationRatesId:drTag, TimingId:timingTag, Weight:weight})
	}
	if i == 0 {
		return nil, nil
	}
	return drt, nil
}

func (self *SQLStorage) GetTPRatingPlanIds(tpid string) ([]string, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT DISTINCT tag FROM %s where tpid='%s'", utils.TBL_TP_RATING_PLANS, tpid))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if i == 0 {
		return nil, nil
	}
	return ids, nil
}

func (self *SQLStorage) ExistsTPRatingProfile(tpid, rpId string) (bool, error) {
	var exists bool
	err := self.Db.QueryRow(fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s WHERE tpid='%s' AND tag='%s')", utils.TBL_TP_RATE_PROFILES, tpid, rpId)).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (self *SQLStorage) SetTPRatingProfiles(tpid string, rps map[string]*utils.TPRatingProfile) error {
	if len(rps) == 0 {
		return nil //Nothing to set
	}
	vals := ""
	i := 0
	for _, rp := range rps {
		for _, rpa := range rp.RatingPlanActivations {
			if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
				vals += ","
			}
			vals += fmt.Sprintf("('%s', '%s', '%s', '%s', '%s', '%s', '%s','%s','%s')", tpid, rp.LoadId, rp.Tenant, rp.TOR, rp.Direction,
				rp.Subject, rpa.ActivationTime, rpa.RatingPlanId, rpa.FallbackSubjects)
			i++
		}
	}
	qry := fmt.Sprintf("INSERT INTO %s (tpid,loadid,tenant,tor,direction,subject,activation_time,rating_plan_tag,fallback_subjects) VALUES %s ON DUPLICATE KEY UPDATE fallback_subjects=values(fallback_subjects)", utils.TBL_TP_RATE_PROFILES, vals)
	if _, err := self.Db.Exec(qry); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) GetTPRatingProfile(tpid, loadId string) (*utils.TPRatingProfile, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT tenant,tor,direction,subject,activation_time,rating_plan_tag,fallback_subjects FROM %s WHERE tpid='%s' AND loadid='%s'", utils.TBL_TP_RATE_PROFILES, tpid, loadId))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rp := &utils.TPRatingProfile{TPid: tpid, LoadId: loadId}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one result
		var tenant, tor, direction, subject, drtId, fallbackSubj, aTime string
		err = rows.Scan(&tenant, &tor, &direction, &subject, &aTime, &drtId, &fallbackSubj)
		if err != nil {
			return nil, err
		}
		if i == 1 { // Save some info on first iteration
			rp.Tenant = tenant
			rp.TOR = tor
			rp.Direction = direction
			rp.Subject = subject
		}
		rp.RatingPlanActivations = append(rp.RatingPlanActivations, &utils.TPRatingActivation{aTime, drtId, fallbackSubj})
	}
	if i == 0 {
		return nil, nil
	}
	return rp, nil
}

func (self *SQLStorage) GetTPRatingProfileIds(filters *utils.AttrTPRatingProfileIds) ([]string, error) {
	qry := fmt.Sprintf("SELECT DISTINCT loadid FROM %s where tpid='%s'", utils.TBL_TP_RATE_PROFILES, filters.TPid)
	if filters.Tenant != "" {
		qry += fmt.Sprintf(" AND tenant='%s'", filters.Tenant)
	}
	if filters.TOR != "" {
		qry += fmt.Sprintf(" AND tor='%s'", filters.TOR)
	}
	if filters.Direction != "" {
		qry += fmt.Sprintf(" AND direction='%s'", filters.Direction)
	}
	if filters.Subject != "" {
		qry += fmt.Sprintf(" AND subject='%s'", filters.Subject)
	}
	rows, err := self.Db.Query(qry)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if i == 0 {
		return nil, nil
	}
	return ids, nil
}

func (self *SQLStorage) ExistsTPActions(tpid, actsId string) (bool, error) {
	var exists bool
	err := self.Db.QueryRow(fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s WHERE tpid='%s' AND tag='%s')", utils.TBL_TP_ACTIONS, tpid, actsId)).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (self *SQLStorage) SetTPActions(tpid string, acts map[string][]*utils.TPAction) error {
	if len(acts) == 0 {
		return nil //Nothing to set
	}
	vals := ""
	i := 0
	for actId, actRows := range acts {
		for _, act := range actRows {
			if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
				vals += ","
			}
			vals += fmt.Sprintf("('%s','%s','%s','%s','%s',%f,'%s','%s','%s',%f,'%s',%f)",
				tpid, actId, act.Identifier, act.BalanceType, act.Direction, act.Units, act.ExpiryTime,
				act.DestinationId, act.RatingSubject, act.BalanceWeight, act.ExtraParameters, act.Weight)
			i++
		}
	}
	qry := fmt.Sprintf("INSERT INTO %s (tpid,tag,action,balance_type,direction,units,expiry_time,destination_tag,rating_subject,balance_weight,extra_parameters,weight) VALUES %s ON DUPLICATE KEY UPDATE action=values(action),balance_type=values(balance_type),direction=values(direction),units=values(units),expiry_time=values(expiry_time),destination_tag=values(destination_tag),rating_subject=values(rating_subject),balance_weight=values(balance_weight),extra_parameters=values(extra_parameters),weight=values(weight)", utils.TBL_TP_ACTIONS, vals)
	if _, err := self.Db.Exec(qry); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) GetTPActions(tpid, actsId string) (*utils.TPActions, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT action,balance_type,direction,units,expiry_time,destination_tag,rating_subject,balance_weight,extra_parameters,weight FROM %s WHERE tpid='%s' AND tag='%s'", utils.TBL_TP_ACTIONS, tpid, actsId))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	acts := &utils.TPActions{TPid: tpid, ActionsId: actsId}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one result
		var action, balanceId, dir, destId, rateSubject, expTime, extraParameters string
		var units, balanceWeight, weight float64
		if err = rows.Scan(&action, &balanceId, &dir, &units, &expTime, &destId, &rateSubject, &balanceWeight, &extraParameters, &weight); err != nil {
			return nil, err
		}
		acts.Actions = append(acts.Actions, &utils.TPAction{action, balanceId, dir, units, expTime, destId, rateSubject, balanceWeight, extraParameters, weight})
	}
	if i == 0 {
		return nil, nil
	}
	return acts, nil
}

func (self *SQLStorage) GetTPActionIds(tpid string) ([]string, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT DISTINCT tag FROM %s where tpid='%s'", utils.TBL_TP_ACTIONS, tpid))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if i == 0 {
		return nil, nil
	}
	return ids, nil
}

func (self *SQLStorage) ExistsTPActionTimings(tpid, atId string) (bool, error) {
	var exists bool
	err := self.Db.QueryRow(fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s WHERE tpid='%s' AND tag='%s')", utils.TBL_TP_ACTION_TIMINGS, tpid, atId)).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Sets actionTimings in sqlDB. Imput is expected in form map[actionTimingId][]rows, eg a full .csv file content
func (self *SQLStorage) SetTPActionTimings(tpid string, ats map[string][]*utils.TPActionTiming) error {
	if len(ats) == 0 {
		return nil //Nothing to set
	}
	vals := ""
	i := 0
	for atId, atRows := range ats {
		for _, at := range atRows {
			if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
				vals += ","
			}
			vals += fmt.Sprintf("('%s','%s','%s','%s',%f)",
				tpid, atId, at.ActionsId, at.TimingId, at.Weight)
			i++
		}
	}
	qry := fmt.Sprintf("INSERT INTO %s (tpid,tag,actions_tag,timing_tag,weight) VALUES %s ON DUPLICATE KEY UPDATE timing_tag=values(timing_tag),weight=values(weight)", utils.TBL_TP_ACTION_TIMINGS, vals)
	if _, err := self.Db.Exec(qry); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) GetTPActionTimings(tpid, atId string) (map[string][]*utils.TPActionTiming, error) {
	ats := make(map[string][]*utils.TPActionTiming)
	q := fmt.Sprintf("SELECT tag,actions_tag,timing_tag,weight FROM %s WHERE tpid='%s'", utils.TBL_TP_ACTION_TIMINGS, tpid)
	if atId != "" {
		q += fmt.Sprintf(" AND tag='%s'", atId)
	}
	rows, err := self.Db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one result
		var tag, actionsId, timingId string
		var weight float64
		if err = rows.Scan(&tag, &actionsId, &timingId, &weight); err != nil {
			return nil, err
		}
		ats[tag] = append(ats[tag], &utils.TPActionTiming{actionsId, timingId, weight})
	}
	return ats, nil
}

func (self *SQLStorage) GetTPActionTimingIds(tpid string) ([]string, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT DISTINCT tag FROM %s where tpid='%s'", utils.TBL_TP_ACTION_TIMINGS, tpid))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if i == 0 {
		return nil, nil
	}
	return ids, nil
}

func (self *SQLStorage) ExistsTPActionTriggers(tpid, atId string) (bool, error) {
	var exists bool
	err := self.Db.QueryRow(fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s WHERE tpid='%s' AND tag='%s')", utils.TBL_TP_ACTION_TRIGGERS, tpid, atId)).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (self *SQLStorage) SetTPActionTriggers(tpid string, ats map[string][]*utils.TPActionTrigger) error {
	if len(ats) == 0 {
		return nil //Nothing to set
	}
	vals := ""
	i := 0
	for atId, atRows := range ats {
		for _, atsRow := range atRows {
			if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
				vals += ","
			}
			vals += fmt.Sprintf("('%s','%s','%s','%s','%s', %f, '%s','%s',%f)",
				tpid, atId, atsRow.BalanceType, atsRow.Direction, atsRow.ThresholdType,
				atsRow.ThresholdValue, atsRow.DestinationId, atsRow.ActionsId, atsRow.Weight)
			i++
		}
	}
	qry := fmt.Sprintf("INSERT INTO %s (tpid,tag,balance_type,direction,threshold_type,threshold_value,destination_tag,actions_tag,weight) VALUES %s ON DUPLICATE KEY UPDATE weight=values(weight)", utils.TBL_TP_ACTION_TRIGGERS, vals)
	if _, err := self.Db.Exec(qry); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) GetTPActionTriggerIds(tpid string) ([]string, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT DISTINCT tag FROM %s where tpid='%s'", utils.TBL_TP_ACTION_TRIGGERS, tpid))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if i == 0 {
		return nil, nil
	}
	return ids, nil
}

func (self *SQLStorage) ExistsTPAccountActions(tpid, aaId string) (bool, error) {
	var exists bool
	err := self.Db.QueryRow(fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s WHERE tpid='%s' AND tag='%s')", utils.TBL_TP_ACCOUNT_ACTIONS, tpid, aaId)).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Sets a group of account actions. Map key has the role of grouping within a tpid
func (self *SQLStorage) SetTPAccountActions(tpid string, aa map[string]*utils.TPAccountActions) error {
	if len(aa) == 0 {
		return nil //Nothing to set
	}
	vals := ""
	i := 0
	for _, aActs := range aa {
		if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
			vals += ","
		}
		vals += fmt.Sprintf("('%s','%s','%s','%s','%s','%s','%s')",
			tpid, aActs.LoadId, aActs.Tenant, aActs.Account, aActs.Direction, aActs.ActionTimingsId, aActs.ActionTriggersId)
		i++
	}
	qry := fmt.Sprintf("INSERT INTO %s (tpid, loadid, tenant, account, direction, action_timings_tag, action_triggers_tag) VALUES %s ON DUPLICATE KEY UPDATE action_timings_tag=values(action_timings_tag), action_triggers_tag=values(action_triggers_tag)", utils.TBL_TP_ACCOUNT_ACTIONS, vals)
	if _, err := self.Db.Exec(qry); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) GetTPAccountActionIds(tpid string) ([]string, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT DISTINCT loadid FROM %s where tpid='%s'", utils.TBL_TP_ACCOUNT_ACTIONS, tpid))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if i == 0 {
		return nil, nil
	}
	return ids, nil
}

func (self *SQLStorage) LogCallCost(uuid, source string, cc *CallCost) (err error) {
	//ToDo: Add cgrid to logCallCost
	if self.Db == nil {
		//timespans.Logger.Warning("Cannot write log to database.")
		return
	}
	tss, err := json.Marshal(cc.Timespans)
	if err != nil {
		Logger.Err(fmt.Sprintf("Error marshalling timespans to json: %v", err))
	}
	_, err = self.Db.Exec(fmt.Sprintf("INSERT INTO %s (cgrid, accid, direction, tenant, tor, account, subject, destination, cost, connect_fee, timespans, source )VALUES ('%s', '%s','%s', '%s', '%s', '%s', '%s', '%s', %f, %f, '%s','%s')",
		utils.TBL_COST_DETAILS,
		utils.FSCgrId(uuid),
		uuid,
		cc.Direction,
		cc.Tenant,
		cc.TOR,
		cc.Account,
		cc.Subject,
		cc.Destination,
		cc.Cost,
		cc.ConnectFee,
		tss,
		source))
	if err != nil {
		Logger.Err(fmt.Sprintf("failed to execute insert statement: %v", err))
	}
	return
}

func (self *SQLStorage) GetCallCostLog(cgrid, source string) (cc *CallCost, err error) {
	row := self.Db.QueryRow(fmt.Sprintf("SELECT cgrid, accid, direction, tenant, tor, account, subject, destination, cost, connect_fee, timespans, source  FROM %s WHERE cgrid='%s' AND source='%s'", utils.TBL_COST_DETAILS, cgrid, source))
	var accid, src string
	var timespansJson string
	cc = &CallCost{Cost: -1}
	err = row.Scan(&cgrid, &accid, &cc.Direction, &cc.Tenant, &cc.TOR, &cc.Account, &cc.Subject,
		&cc.Destination, &cc.Cost, &cc.ConnectFee, &timespansJson, &src)
	if err = json.Unmarshal([]byte(timespansJson), &cc.Timespans); err != nil {
		return nil, err
	}
	return
}

func (self *SQLStorage) LogActionTrigger(ubId, source string, at *ActionTrigger, as Actions) (err error) {
	return
}
func (self *SQLStorage) LogActionTiming(source string, at *ActionTiming, as Actions) (err error) {
	return
}
func (self *SQLStorage) LogError(uuid, source, errstr string) (err error) { return }

func (self *SQLStorage) SetCdr(cdr utils.CDR) (err error) {
	startTime, err := cdr.GetAnswerTime()
	if err != nil {
		return err
	}
	_, err = self.Db.Exec(fmt.Sprintf("INSERT INTO cdrs_primary VALUES (NULL, '%s', '%s','%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', %d, %d)",
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
		startTime.Unix(),
		cdr.GetDuration(),
	))
	if err != nil {
		Logger.Err(fmt.Sprintf("failed to execute cdr insert statement: %v", err))
	}
	extraFields, err := json.Marshal(cdr.GetExtraFields())
	if err != nil {
		Logger.Err(fmt.Sprintf("Error marshalling cdr extra fields to json: %v", err))
	}
	_, err = self.Db.Exec(fmt.Sprintf("INSERT INTO cdrs_extra VALUES ('NULL','%s', '%s')",
		cdr.GetCgrId(),
		extraFields,
	))
	if err != nil {
		Logger.Err(fmt.Sprintf("failed to execute cdr insert statement: %v", err))
	}

	return
}

func (self *SQLStorage) SetRatedCdr(cdr utils.CDR, cc *CallCost, extraInfo string) (err error) {
	// ToDo: Add here source and subject
	_, err = self.Db.Exec(fmt.Sprintf("INSERT INTO %s (cgrid, subject, cost, extra_info) VALUES ('%s', '%s', %f, '%s')",
		utils.TBL_RATED_CDRS,
		cdr.GetCgrId(),
		cdr.GetSubject(),
		cc.Cost+cc.ConnectFee,
		extraInfo))
	if err != nil {
		Logger.Err(fmt.Sprintf("failed to execute cdr insert statement: %v", err))
	}

	return
}

// Return a slice of rated CDRs from storDb using optional timeStart and timeEnd as filters.
func (self *SQLStorage) GetRatedCdrs(timeStart, timeEnd time.Time) ([]utils.CDR, error) {
	var cdrs []utils.CDR
	q := fmt.Sprintf("SELECT %s.cgrid,accid,cdrhost,reqtype,direction,tenant,tor,account,%s.subject,destination,answer_timestamp,duration,extra_fields,cost FROM %s LEFT JOIN %s ON %s.cgrid=%s.cgrid LEFT JOIN %s ON %s.cgrid=%s.cgrid", utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_EXTRA, utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_EXTRA, utils.TBL_RATED_CDRS, utils.TBL_CDRS_PRIMARY, utils.TBL_RATED_CDRS)
	if !timeStart.IsZero() && !timeEnd.IsZero() {
		q += fmt.Sprintf(" WHERE answer_timestamp>=%d AND answer_timestamp<%d", timeStart.Unix(), timeEnd.Unix())
	} else if !timeStart.IsZero() {
		q += fmt.Sprintf(" WHERE answer_timestamp>=%d", timeStart.Unix())
	} else if !timeEnd.IsZero() {
		q += fmt.Sprintf(" WHERE answer_timestamp<%d", timeEnd.Unix())
	}
	rows, err := self.Db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var cgrid, accid, cdrhost, reqtype, direction, tenant, tor, account, subject, destination string
		var extraFields []byte
		var answerTimestamp, duration int64
		var cost float64
		var extraFieldsMp map[string]string
		if err := rows.Scan(&cgrid, &accid, &cdrhost, &reqtype, &direction, &tenant, &tor, &account, &subject, &destination, &answerTimestamp, &duration, &extraFields, &cost); err != nil {
			return nil, err
		}
		answerTime := time.Unix(answerTimestamp, 0)
		if err := json.Unmarshal(extraFields, &extraFieldsMp); err != nil {
			return nil, err
		}
		storCdr := &utils.RatedCDR{
			CgrId: cgrid, AccId: accid, CdrHost: cdrhost, ReqType: reqtype, Direction: direction, Tenant: tenant,
			TOR: tor, Account: account, Subject: subject, Destination: destination, AnswerTime: answerTime, Duration: duration,
			ExtraFields: extraFieldsMp, Cost: cost,
		}
		cdrs = append(cdrs, utils.CDR(storCdr))
	}
	return cdrs, nil
}

func (self *SQLStorage) GetTpDestinations(tpid, tag string) ([]*Destination, error) {
	var dests []*Destination
	q := fmt.Sprintf("SELECT * FROM %s WHERE tpid='%s'", utils.TBL_TP_DESTINATIONS, tpid)
	if tag != "" {
		q += fmt.Sprintf(" AND tag='%s'", tag)
	}
	rows, err := self.Db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var tpid, tag, prefix string
		if err := rows.Scan(&id, &tpid, &tag, &prefix); err != nil {
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
	return dests, nil
}

func (self *SQLStorage) GetTpRates(tpid, tag string) (map[string]*utils.TPRate, error) {
	rts := make(map[string]*utils.TPRate)
	q := fmt.Sprintf("SELECT tag, connect_fee, rate, rate_unit, rate_increment, group_interval_start, rounding_method, rounding_decimals FROM %s WHERE tpid='%s' ", utils.TBL_TP_RATES, tpid)
	if tag != "" {
		q += fmt.Sprintf(" AND tag='%s'", tag)
	}
	rows, err := self.Db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var tag, rate_unit, rate_increment, group_interval_start, roundingMethod string
		var connect_fee, rate float64
		var roundingDecimals int
		if err := rows.Scan(&tag, &connect_fee, &rate, &rate_unit, &rate_increment, &group_interval_start, &roundingMethod, &roundingDecimals); err != nil {
			return nil, err
		}
		rs, err := utils.NewRateSlot(connect_fee, rate, rate_unit, rate_increment, group_interval_start, roundingMethod, roundingDecimals)
		if err!=nil {
			return nil, err
		}
		r := &utils.TPRate{
			RateId: tag,
			RateSlots: []*utils.RateSlot{rs},
			}
		
		// same tag only to create rate groups
		existingRates, exists := rts[tag]
		if exists {
			rss := existingRates.RateSlots
			if err := ValidNextGroup(rss[len(rss)-1], r.RateSlots[0]); err != nil {
				return nil, err
			}
			rts[tag].RateSlots = append(rts[tag].RateSlots, r.RateSlots[0])
		} else {
			rts[tag] = r

		}
	}
	return rts, nil
}

func (self *SQLStorage) GetTpDestinationRates(tpid, tag string) (map[string]*utils.TPDestinationRate, error) {
	rts := make(map[string]*utils.TPDestinationRate)
	q := fmt.Sprintf("SELECT * FROM %s WHERE tpid='%s'", utils.TBL_TP_DESTINATION_RATES, tpid)
	if tag != "" {
		q += fmt.Sprintf(" AND tag='%s'", tag)
	}
	rows, err := self.Db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var tpid, tag, destinations_tag, rate_tag string
		if err := rows.Scan(&id, &tpid, &tag, &destinations_tag, &rate_tag); err != nil {
			return nil, err
		}

		dr := &utils.TPDestinationRate{
			DestinationRateId: tag,
			DestinationRates: []*utils.DestinationRate{
				&utils.DestinationRate{
					DestinationId: destinations_tag,
					RateId:        rate_tag,
				},
			},
		}
		existingDR, exists := rts[tag]
		if exists {
			existingDR.DestinationRates = append(existingDR.DestinationRates, dr.DestinationRates[0])
		} else {
			existingDR = dr
		}
		rts[tag] = existingDR
	}
	return rts, nil
}

func (self *SQLStorage) GetTpTimings(tpid, tag string) (map[string]*utils.TPTiming, error) {
	tms := make(map[string]*utils.TPTiming)
	q := fmt.Sprintf("SELECT * FROM %s WHERE tpid='%s'", utils.TBL_TP_TIMINGS, tpid)
	if tag != "" {
		q += fmt.Sprintf(" AND tag='%s'", tag)
	}
	rows, err := self.Db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var tpid, tag, years, months, month_days, week_days, start_time string
		if err := rows.Scan(&id, &tpid, &tag, &years, &months, &month_days, &week_days, &start_time); err != nil {
			return nil, err
		}
		tms[tag] = NewTiming(tag, years, months, month_days, week_days, start_time)
	}
	return tms, nil
}

func (self *SQLStorage) GetTpRatingPlans(tpid, tag string) (map[string][]*utils.TPRatingPlanBinding, error) {
	rpbns := make(map[string][]*utils.TPRatingPlanBinding)
	q := fmt.Sprintf("SELECT * FROM %s WHERE tpid='%s'", utils.TBL_TP_RATING_PLANS, tpid)
	if tag != "" {
		q += fmt.Sprintf(" AND tag='%s'", tag)
	}
	rows, err := self.Db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var weight float64
		var tpid, tag, destination_rates_tag, timings_tag string
		if err := rows.Scan(&id, &tpid, &tag, &destination_rates_tag, &timings_tag, &weight); err != nil {
			return nil, err
		}
		rpb := &utils.TPRatingPlanBinding{
			DestinationRatesId: destination_rates_tag,
			TimingId:    timings_tag,
			Weight:      weight,
		}
		if rpBnLst,exists := rpbns[tag]; exists {
			rpBnLst = append(rpBnLst, rpb)
		} else { // New
			rpbns[tag] = []*utils.TPRatingPlanBinding{rpb}
		}
	}
	return rpbns, nil
}

func (self *SQLStorage) GetTpRatingProfiles(qryRpf *utils.TPRatingProfile) (map[string]*utils.TPRatingProfile, error) {
	q := fmt.Sprintf("SELECT loadid,tenant,tor,direction,subject,activation_time,rating_plan_tag,fallback_subjects FROM %s WHERE tpid='%s'",
		utils.TBL_TP_RATE_PROFILES, qryRpf.TPid)
	if len(qryRpf.LoadId) != 0 {
		q += fmt.Sprintf(" AND loadid='%s'", qryRpf.LoadId)
	}
	if len(qryRpf.Tenant) != 0 {
		q += fmt.Sprintf(" AND tenant='%s'", qryRpf.Tenant)
	}
	if len(qryRpf.TOR) != 0 {
		q += fmt.Sprintf(" AND tor='%s'", qryRpf.TOR)
	}
	if len(qryRpf.Direction) != 0 {
		q += fmt.Sprintf(" AND direction='%s'", qryRpf.Direction)
	}
	if len(qryRpf.Subject) != 0 {
		q += fmt.Sprintf(" AND subject='%s'", qryRpf.Subject)
	}	
	rows, err := self.Db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rpfs := make(map[string]*utils.TPRatingProfile)
	for rows.Next() {
		var rcvLoadId, tenant, tor, direction, subject, fallback_subjects, rating_plan_tag, activation_time string
		if err := rows.Scan(&rcvLoadId, &tenant, &tor, &direction, &subject, &activation_time, &rating_plan_tag, &fallback_subjects); err != nil {
			return nil, err
		}
		rp := &utils.TPRatingProfile{TPid: qryRpf.TPid, LoadId: rcvLoadId, Tenant: tenant, TOR: tor, Direction:direction, Subject:subject}
		if existingRp,has := rpfs[rp.KeyId()]; !has {
			rp.RatingPlanActivations = []*utils.TPRatingActivation{
						&utils.TPRatingActivation{ActivationTime:activation_time, RatingPlanId:rating_plan_tag, FallbackSubjects:fallback_subjects}}
			rpfs[rp.KeyId()] = rp
		} else { // Exists, update
			existingRp.RatingPlanActivations = append( existingRp.RatingPlanActivations, 
					&utils.TPRatingActivation{ActivationTime:activation_time, RatingPlanId:rating_plan_tag, FallbackSubjects:fallback_subjects} )
		}
	}
	return rpfs, nil
}

func (self *SQLStorage) GetTpActions(tpid, tag string) (map[string][]*Action, error) {
	as := make(map[string][]*Action)
	q := fmt.Sprintf("SELECT * FROM %s WHERE tpid='%s'", utils.TBL_TP_ACTIONS, tpid)
	if tag != "" {
		q += fmt.Sprintf(" AND tag='%s'", tag)
	}
	rows, err := self.Db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var units, balance_weight, weight float64
		var tpid, tag, action, balance_type, direction, destinations_tag, rating_subject, extra_parameters, expirationDate string
		if err := rows.Scan(&id, &tpid, &tag, &action, &balance_type, &direction, &units, &expirationDate, &destinations_tag, &rating_subject, &balance_weight, &extra_parameters, &weight); err != nil {
			return nil, err
		}
		a := &Action{
			Id:               utils.GenUUID(),
			ActionType:       action,
			BalanceId:        balance_type,
			Direction:        direction,
			Weight:           weight,
			ExtraParameters:  extra_parameters,
			ExpirationString: expirationDate,
			Balance: &Balance{
				Uuid:          utils.GenUUID(),
				Value:         units,
				Weight:        balance_weight,
				RateSubject:   rating_subject,
				DestinationId: destinations_tag,
			},
		}
		as[tag] = append(as[tag], a)
	}
	return as, nil
}

func (self *SQLStorage) GetTpActionTimings(tpid, tag string) (map[string][]*utils.TPActionTiming, error) {
	q := fmt.Sprintf("SELECT tag,actions_tag,timing_tag,weight FROM %s WHERE tpid='%s'", utils.TBL_TP_ACTION_TIMINGS, tpid)
	if tag != "" {
		q += fmt.Sprintf(" AND tag='%s'", tag)
	}
	rows, err := self.Db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ats := make(map[string][]*utils.TPActionTiming)
	for rows.Next() {
		var weight float64
		var tag, actions_tag, timing_tag string
		if err := rows.Scan(&tag, &actions_tag, &timing_tag, &weight); err != nil {
			return nil, err
		}
		at := &utils.TPActionTiming {
			ActionsId: tag,
			TimingId: timing_tag,
			Weight: weight,
		}
		ats[tag] = append(ats[tag], at)
	}
	return ats, nil
}

func (self *SQLStorage) GetTpActionTriggers(tpid, tag string) (map[string][]*utils.TPActionTrigger, error) {
	ats := make(map[string][]*utils.TPActionTrigger)
	q := fmt.Sprintf("SELECT tpid,tag,balance_type,direction,threshold_type,threshold_value,destination_tag,actions_tag,weight FROM %s WHERE tpid='%s'",
		utils.TBL_TP_ACTION_TRIGGERS, tpid)
	if tag != "" {
		q += fmt.Sprintf(" AND tag='%s'", tag)
	}
	rows, err := self.Db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var threshold, weight float64
		var tpid, tag, balances_type, direction, destinations_tag, actions_tag, thresholdType string
		if err := rows.Scan(&tpid, &tag, &balances_type, &direction, &thresholdType, &threshold, &destinations_tag, &actions_tag, &weight); err != nil {
			return nil, err
		}

		at := &utils.TPActionTrigger{
			BalanceType:    balances_type,
			Direction:      direction,
			ThresholdType:  thresholdType,
			ThresholdValue: threshold,
			DestinationId:  destinations_tag,
			ActionsId:      actions_tag,
			Weight:         weight,
		}
		ats[tag] = append(ats[tag], at)
	}
	return ats, nil
}

func (self *SQLStorage) GetTpAccountActions(aaFltr *utils.TPAccountActions) (map[string]*utils.TPAccountActions, error) {
	q := fmt.Sprintf("SELECT loadid, tenant, account, direction, action_timings_tag, action_triggers_tag FROM %s WHERE tpid='%s'", utils.TBL_TP_ACCOUNT_ACTIONS, aaFltr.TPid)
	if len(aaFltr.LoadId) != 0 {
		q += fmt.Sprintf(" AND loadid='%s'", aaFltr.LoadId)
	}
	if len(aaFltr.Tenant) != 0 {
		q += fmt.Sprintf(" AND tenant='%s'", aaFltr.Tenant)
	}
	if len(aaFltr.Account) != 0 {
		q += fmt.Sprintf(" AND account='%s'", aaFltr.Account)
	}
	if len(aaFltr.Direction) != 0 {
		q += fmt.Sprintf(" AND direction='%s'", aaFltr.Direction)
	}
	rows, err := self.Db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	aa := make(map[string]*utils.TPAccountActions)
	for rows.Next() {
		var aaLoadId, tenant, account, direction, action_timings_tag, action_triggers_tag string
		if err := rows.Scan(&aaLoadId, &tenant, &account, &direction, &action_timings_tag, &action_triggers_tag); err != nil {
			return nil, err
		}
		aacts := &utils.TPAccountActions{
			TPid:             aaFltr.TPid,
			LoadId:           aaLoadId,
			Tenant:           tenant,
			Account:          account,
			Direction:        direction,
			ActionTimingsId:  action_timings_tag,
			ActionTriggersId: action_triggers_tag,
		}
		aa[aacts.KeyId()] = aacts
	}
	return aa, nil
}
