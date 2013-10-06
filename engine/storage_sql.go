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
	self.Close()
}

func (self *SQLStorage) Flush() (err error) {
	return
}

// Return a list with all TPids defined in the system, even if incomplete, isolated in some table.
func (self *SQLStorage) GetTPIds() ([]string, error) {
	rows, err := self.Db.Query(
		fmt.Sprintf("(SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s)", utils.TBL_TP_TIMINGS, utils.TBL_TP_DESTINATIONS, utils.TBL_TP_RATES, utils.TBL_TP_DESTINATION_RATES, utils.TBL_TP_DESTRATE_TIMINGS, utils.TBL_TP_RATE_PROFILES))
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

func (self *SQLStorage) SetTPTiming(tpid string, tm *Timing) error {
	if _, err := self.Db.Exec(fmt.Sprintf("INSERT INTO %s (tpid, tag, years, months, month_days, week_days, time) VALUES('%s','%s','%s','%s','%s','%s','%s')",
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

func (self *SQLStorage) GetTPTiming(tpid, tmId string) (*Timing, error) {
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
	for _, prefix := range dest.Prefixes {
		if _, err := self.Db.Exec(fmt.Sprintf("INSERT INTO %s (tpid, tag, prefix) VALUES( '%s','%s','%s')", utils.TBL_TP_DESTINATIONS, tpid, dest.Id, prefix)); err != nil {
			return err
		}
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

func (self *SQLStorage) SetTPRates(tpid string, rts map[string][]*LoadRate) error {
	if len(rts) == 0 {
		return nil //Nothing to set
	}
	qry := fmt.Sprintf("INSERT INTO %s (tpid, tag, connect_fee, rate, rate_unit, rate_increment, group_interval_start, rounding_method, rounding_decimals, weight) VALUES ", utils.TBL_TP_RATES)
	i := 0
	for rtId, rtRows := range rts {
		for _, rt := range rtRows {
			if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
				qry += ","
			}
			qry += fmt.Sprintf("('%s', '%s', %f, %f, %d, %d,%d,'%s', %d, %f)",
				tpid, rtId, rt.ConnectFee, rt.Price, rt.RateUnit, rt.RateIncrement, rt.GroupIntervalStart,
				rt.RoundingMethod, rt.RoundingDecimals, rt.Weight)
			i++
		}
	}
	if _, err := self.Db.Exec(qry); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) GetTPRate(tpid, rtId string) (*utils.TPRate, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT connect_fee, rate, rate_unit, rate_increment, group_interval_start, rounding_method, rounding_decimals, weight FROM %s WHERE tpid='%s' AND tag='%s'", utils.TBL_TP_RATES, tpid, rtId))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rt := &utils.TPRate{TPid: tpid, RateId: rtId}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one prefix
		var connectFee, rate, weight float64
		var roundingDecimals int
		var rateUnit, rateIncrement, groupIntervalStart time.Duration
		var roundingMethod string
		err = rows.Scan(&connectFee, &rate, &rateUnit, &rateIncrement, &groupIntervalStart, &roundingMethod, &roundingDecimals, &weight)
		if err != nil {
			return nil, err
		}
		rt.RateSlots = append(rt.RateSlots, utils.RateSlot{connectFee, rate, rateUnit.String(), rateIncrement.String(),
			groupIntervalStart.String(), roundingMethod, roundingDecimals, weight})
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

func (self *SQLStorage) SetTPDestinationRates(tpid string, drs map[string][]*DestinationRate) error {
	if len(drs) == 0 {
		return nil //Nothing to set
	}
	qry := fmt.Sprintf("INSERT INTO %s (tpid, tag, destinations_tag, rates_tag) VALUES ", utils.TBL_TP_DESTINATION_RATES)
	i := 0
	for drId, drRows := range drs {
		for _, dr := range drRows {
			if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
				qry += ","
			}
			qry += fmt.Sprintf("('%s','%s','%s','%s')",
				tpid, drId, dr.DestinationsTag, dr.RateTag)
			i++
		}
	}
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
		dr.DestinationRates = append(dr.DestinationRates, utils.DestinationRate{dstTag, ratesTag})
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

func (self *SQLStorage) ExistsTPDestRateTiming(tpid, drtId string) (bool, error) {
	var exists bool
	err := self.Db.QueryRow(fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s WHERE tpid='%s' AND tag='%s')", utils.TBL_TP_DESTRATE_TIMINGS, tpid, drtId)).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (self *SQLStorage) SetTPDestRateTimings(tpid string, drts map[string][]*DestinationRateTiming) error {
	if len(drts) == 0 {
		return nil //Nothing to set
	}
	qry := fmt.Sprintf("INSERT INTO %s (tpid, tag, destrates_tag, timing_tag, weight) VALUES ", utils.TBL_TP_DESTRATE_TIMINGS)
	i := 0
	for drtId, drtRows := range drts {
		for _, drt := range drtRows {
			if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
				qry += ","
			}
			qry += fmt.Sprintf("('%s','%s','%s','%s',%f)",
				tpid, drtId, drt.DestinationRatesTag, drt.TimingTag, drt.Weight)
			i++
		}
	}
	if _, err := self.Db.Exec(qry); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) GetTPDestRateTiming(tpid, drtId string) (*utils.TPDestRateTiming, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT destrates_tag, timing_tag, weight from %s where tpid='%s' and tag='%s'", utils.TBL_TP_DESTRATE_TIMINGS, tpid, drtId))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	drt := &utils.TPDestRateTiming{TPid: tpid, DestRateTimingId: drtId}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one result
		var drTag, timingTag string
		var weight float64
		err = rows.Scan(&drTag, &timingTag, &weight)
		if err != nil {
			return nil, err
		}
		drt.DestRateTimings = append(drt.DestRateTimings, utils.DestRateTiming{drTag, timingTag, weight})
	}
	if i == 0 {
		return nil, nil
	}
	return drt, nil
}

func (self *SQLStorage) GetTPDestRateTimingIds(tpid string) ([]string, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT DISTINCT tag FROM %s where tpid='%s'", utils.TBL_TP_DESTRATE_TIMINGS, tpid))
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

func (self *SQLStorage) SetTPRatingProfiles(tpid string, rps map[string][]*RatingProfile) error {
	if len(rps) == 0 {
		return nil //Nothing to set
	}
	qry := fmt.Sprintf("INSERT INTO %s (tpid,tag,tenant,tor,direction,subject,activation_time,destrates_timing_tag,rates_fallback_subject) VALUES ",
		utils.TBL_TP_RATE_PROFILES)
	i := 0
	for rpId, rp := range rps {
		for _, rpa := range rp {
			if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
				qry += ","
			}
			qry += fmt.Sprintf("('%s', '%s', '%s', '%s', '%s', '%s', '%s','%s','%s')", tpid, rpId, rpa.Tenant, rpa.TOR, rpa.Direction,
				rpa.Subject, rpa.ActivationTime, rpa.DestRatesTimingTag, rpa.RatesFallbackSubject)
			i++
		}

	}
	if _, err := self.Db.Exec(qry); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) GetTPRatingProfile(tpid, rpId string) (*utils.TPRatingProfile, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT tenant,tor,direction,subject,activation_time,destrates_timing_tag,rates_fallback_subject FROM %s WHERE tpid='%s' AND tag='%s'", utils.TBL_TP_RATE_PROFILES, tpid, rpId))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rp := &utils.TPRatingProfile{TPid: tpid, RatingProfileId: rpId}
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
			rp.RatesFallbackSubject = fallbackSubj
		}
		rp.RatingActivations = append(rp.RatingActivations, utils.RatingActivation{aTime, drtId})
	}
	if i == 0 {
		return nil, nil
	}
	return rp, nil
}

func (self *SQLStorage) GetTPRatingProfileIds(filters *utils.AttrTPRatingProfileIds) ([]string, error) {
	qry := fmt.Sprintf("SELECT DISTINCT tag FROM %s where tpid='%s'", utils.TBL_TP_RATE_PROFILES, filters.TPid)
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

func (self *SQLStorage) SetTPActions(tpid string, acts map[string][]*Action) error {
	if len(acts) == 0 {
		return nil //Nothing to set
	}
	qry := fmt.Sprintf("INSERT INTO %s (tpid,tag,action,balance_type,direction,units,expiry_time,destination_tag,rating_subject,balance_weight,extra_parameters,weight) VALUES ", utils.TBL_TP_ACTIONS)
	i := 0
	for actId, actRows := range acts {
		for _, act := range actRows {
			if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
				qry += ","
			}
			qry += fmt.Sprintf("('%s','%s','%s','%s','%s',%f,'%s','%s','%s',%f,'%s',%f)",
				tpid, actId, act.ActionType, act.BalanceId, act.Direction, act.Balance.Value, act.ExpirationString,
				act.Balance.DestinationId, act.Balance.RateSubject, act.Balance.Weight, act.ExtraParameters, act.Weight)
			i++
		}
	}
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
		acts.Actions = append(acts.Actions, utils.Action{action, balanceId, dir, units, expTime, destId, rateSubject, balanceWeight, extraParameters, weight})
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
func (self *SQLStorage) SetTPActionTimings(tpid string, ats map[string][]*ActionTiming) error {
	if len(ats) == 0 {
		return nil //Nothing to set
	}
	qry := fmt.Sprintf("INSERT INTO %s (tpid,tag,actions_tag,timing_tag,weight) VALUES ", utils.TBL_TP_ACTION_TIMINGS)
	i := 0
	for atId, atRows := range ats {
		for _, at := range atRows {
			if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
				qry += ","
			}
			qry += fmt.Sprintf("('%s','%s','%s','%s',%f)",
				tpid, atId, at.ActionsTag, at.TimingsTag, at.Weight)
			i++
		}
	}
	if _, err := self.Db.Exec(qry); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) GetTPActionTimings(tpid, atId string) (map[string][]*utils.TPActionTimingsRow, error) {
	ats := make(map[string][]*utils.TPActionTimingsRow)
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
		ats[tag] = append(ats[tag], &utils.TPActionTimingsRow{actionsId, timingId, weight})
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

func (self *SQLStorage) SetTPActionTriggers(tpid string, ats map[string][]*ActionTrigger) error {
	if len(ats) == 0 {
		return nil //Nothing to set
	}
	qry := fmt.Sprintf("INSERT INTO %s (tpid,tag,balance_type,direction,threshold_type,threshold_value,destination_tag,actions_tag,weight) VALUES ",
		utils.TBL_TP_ACTION_TRIGGERS)
	i := 0
	for atId, atRows := range ats {
		for _, atsRow := range atRows {
			if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
				qry += ","
			}
			qry += fmt.Sprintf("('%s','%s','%s','%s','%s', %f, '%s','%s',%f)",
				tpid, atId, atsRow.BalanceId, atsRow.Direction, atsRow.ThresholdType,
				atsRow.ThresholdValue, atsRow.DestinationId, atsRow.ActionsId, atsRow.Weight)
			i++
		}
	}
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

func (self *SQLStorage) SetTPAccountActions(tpid string, aa map[string]*AccountAction) error {
	if len(aa) == 0 {
		return nil //Nothing to set
	}
	qry := fmt.Sprintf("INSERT INTO %s (tpid, tag, tenant, account, direction, action_timings_tag, action_triggers_tag) VALUES ",
		utils.TBL_TP_ACCOUNT_ACTIONS)
	i := 0
	for aaId, aActs := range aa {
		if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
			qry += ","
		}
		qry += fmt.Sprintf("('%s','%s','%s','%s','%s','%s','%s')",
			tpid, aaId, aActs.Tenant, aActs.Account, aActs.Direction, aActs.ActionTimingsTag, aActs.ActionTriggersTag)
		i++
	}
	if _, err := self.Db.Exec(qry); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) GetTPAccountActionIds(tpid string) ([]string, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT DISTINCT tag FROM %s where tpid='%s'", utils.TBL_TP_ACCOUNT_ACTIONS, tpid))
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
	_, err = self.Db.Exec(fmt.Sprintf("INSERT INTO cdrs_primary VALUES (NULL, '%s', '%s','%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', FROM_UNIXTIME(%d), %d)",
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

func (self *SQLStorage) GetAllRatedCdr() ([]utils.CDR, error) {
	return nil, nil
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

func (self *SQLStorage) GetTpRates(tpid, tag string) (map[string][]*LoadRate, error) {
	rts := make(map[string][]*LoadRate)
	q := fmt.Sprintf("SELECT tag, connect_fee, rate, rate_unit, rate_increment, group_interval_start, rounding_method, rounding_decimals, weight FROM %s WHERE tpid='%s' ", utils.TBL_TP_RATES, tpid)
	if tag != "" {
		q += fmt.Sprintf(" AND tag='%s'", tag)
	}
	rows, err := self.Db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var tag, roundingMethod string
		var connect_fee, rate, weight float64
		var rate_unit, rate_increment, group_interval_start time.Duration
		var roundingDecimals int
		if err := rows.Scan(&tag, &connect_fee, &rate, &rate_unit, &rate_increment, &group_interval_start, &roundingMethod, &roundingDecimals, &weight); err != nil {
			return nil, err
		}
		r := &LoadRate{
			Tag:                tag,
			ConnectFee:         connect_fee,
			Price:              rate,
			RateUnit:           rate_unit,
			RateIncrement:      rate_increment,
			GroupIntervalStart: group_interval_start,
			RoundingMethod:     roundingMethod,
			RoundingDecimals:   roundingDecimals,
			Weight:             weight,
		}
		// same tag only to create rate groups
		existingRates, exists := rts[tag]
		if exists {
			if err := existingRates[len(existingRates)-1].ValidNextGroup(r); err != nil {
				return nil, err
			}
		}
		rts[tag] = append(rts[tag], r)
	}
	return rts, nil
}

func (self *SQLStorage) GetTpDestinationRates(tpid, tag string) (map[string][]*DestinationRate, error) {
	rts := make(map[string][]*DestinationRate)
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
		dr := &DestinationRate{
			Tag:             tag,
			DestinationsTag: destinations_tag,
			RateTag:         rate_tag,
		}
		rts[tag] = append(rts[tag], dr)
	}
	return rts, nil
}

func (self *SQLStorage) GetTpTimings(tpid, tag string) (map[string]*Timing, error) {
	tms := make(map[string]*Timing)
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

func (self *SQLStorage) GetTpDestinationRateTimings(tpid, tag string) ([]*DestinationRateTiming, error) {
	var rts []*DestinationRateTiming
	q := fmt.Sprintf("SELECT * FROM %s WHERE tpid='%s'", utils.TBL_TP_DESTRATE_TIMINGS, tpid)
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
		rt := &DestinationRateTiming{
			Tag:                 tag,
			DestinationRatesTag: destination_rates_tag,
			Weight:              weight,
			TimingTag:           timings_tag,
		}
		rts = append(rts, rt)
	}
	return rts, nil
}

func (self *SQLStorage) GetTpRatingProfiles(tpid, tag string) (map[string]*RatingProfile, error) {
	rpfs := make(map[string]*RatingProfile)
	q := fmt.Sprintf("SELECT tag,tenant,tor,direction,subject,activation_time,destrates_timing_tag,rates_fallback_subject FROM %s WHERE tpid='%s'",
		utils.TBL_TP_RATE_PROFILES, tpid)
	if tag != "" {
		q += fmt.Sprintf(" AND tag='%s'", tag)
	}
	rows, err := self.Db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var tag, tenant, tor, direction, subject, fallback_subject, destrates_timing_tag, activation_time string
		if err := rows.Scan(&tag, &tenant, &tor, &direction, &subject, &activation_time, &destrates_timing_tag, &fallback_subject); err != nil {
			return nil, err
		}
		key := fmt.Sprintf("%s:%s:%s:%s", direction, tenant, tor, subject)
		rp, ok := rpfs[key]
		if !ok || rp.Tag != tag {
			rp = &RatingProfile{Id: key, Tag: tag}
			rpfs[key] = rp
		}
		rp.DestRatesTimingTag = destrates_timing_tag
		rp.ActivationTime = activation_time
		if fallback_subject != "" {
			rp.FallbackKey = fmt.Sprintf("%s:%s:%s:%s", direction, tenant, tor, fallback_subject)
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

func (self *SQLStorage) GetTpActionTimings(tpid, tag string) (ats map[string][]*ActionTiming, err error) {
	ats = make(map[string][]*ActionTiming)
	q := fmt.Sprintf("SELECT * FROM %s WHERE tpid='%s'", utils.TBL_TP_ACTION_TIMINGS, tpid)
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
	return ats, nil
}

func (self *SQLStorage) GetTpActionTriggers(tpid, tag string) (map[string][]*ActionTrigger, error) {
	ats := make(map[string][]*ActionTrigger)
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
		var tpid, tag, balances_tag, direction, destinations_tag, actions_tag, thresholdType string
		if err := rows.Scan(&tpid, &tag, &balances_tag, &direction, &thresholdType, &threshold, &destinations_tag, &actions_tag, &weight); err != nil {
			return nil, err
		}

		at := &ActionTrigger{
			Id:             utils.GenUUID(),
			BalanceId:      balances_tag,
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

func (self *SQLStorage) GetTpAccountActions(tpid, tag string) (map[string]*AccountAction, error) {
	q := fmt.Sprintf("SELECT tag, tenant, account, direction, action_timings_tag, action_triggers_tag FROM %s WHERE tpid='%s'", utils.TBL_TP_ACCOUNT_ACTIONS, tpid)
	if tag != "" {
		q += fmt.Sprintf(" AND tag='%s'", tag)
	}
	rows, err := self.Db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	aa := make(map[string]*AccountAction)
	for rows.Next() {
		var aaId, tenant, account, direction, action_timings_tag, action_triggers_tag string
		if err := rows.Scan(&aaId, &tenant, &account, &direction, &action_timings_tag, &action_triggers_tag); err != nil {
			return nil, err
		}
		aa[aaId] = &AccountAction{
			Tenant:            tenant,
			Account:           account,
			Direction:         direction,
			ActionTimingsTag:  action_timings_tag,
			ActionTriggersTag: action_triggers_tag,
		}
	}
	return aa, nil
}
