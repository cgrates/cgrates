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
)

type SQLStorage struct {
	Db *sql.DB
}

func (self *SQLStorage) Close() {}

func (self *SQLStorage) Flush() (err error) {
	return
}

func (self *SQLStorage) GetRatingProfile(string) (rp *RatingProfile, err error) {
	/*row := self.Db.QueryRow(fmt.Sprintf("SELECT * FROM ratingprofiles WHERE id='%s'", id))
	err = row.Scan(&rp, &cc.Direction, &cc.Tenant, &cc.TOR, &cc.Subject, &cc.Destination, &cc.Cost, &cc.ConnectFee, &timespansJson)
	err = json.Unmarshal([]byte(timespansJson), cc.Timespans)*/
	return
}

func (self *SQLStorage) SetRatingProfile(rp *RatingProfile) (err error) {
	return
}

func (self *SQLStorage) GetDestination(string) (d *Destination, err error) {
	return
}

func (self *SQLStorage) SetDestination(d *Destination) (err error) {
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

func (self *SQLStorage) SetTPRate(rt *utils.TPRate) error {
	for _, rtSlot := range rt.RateSlots {
		if _, err := self.Db.Exec(fmt.Sprintf("INSERT INTO %s (tpid, tag, connect_fee, rate, rated_units, rate_increments, weight) VALUES ('%s', '%s', %f, %f, %d, %d, %f)",
			utils.TBL_TP_RATES, rt.TPid, rt.RateId, rtSlot.ConnectFee, rtSlot.Rate, rtSlot.RatedUnits, rtSlot.RateIncrements,
			rtSlot.Weight)); err != nil {
			return err
		}
	}
	return nil
}

func (self *SQLStorage) GetTPRate(tpid, rtId string) (*utils.TPRate, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT connect_fee, rate, rated_units, rate_increments, weight FROM %s WHERE tpid='%s' AND tag='%s'", utils.TBL_TP_RATES, tpid, rtId))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rt := &utils.TPRate{TPid: tpid, RateId: rtId}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one prefix
		var connectFee, rate, weight float64
		var ratedUnits, rateIncrements int
		err = rows.Scan(&connectFee, &rate, &ratedUnits, &rateIncrements, &weight)
		if err != nil {
			return nil, err
		}
		rt.RateSlots = append(rt.RateSlots, utils.RateSlot{connectFee, rate, ratedUnits, rateIncrements, weight})
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

func (self *SQLStorage) SetTPDestinationRate(dr *utils.TPDestinationRate) error {
	if len(dr.DestinationRates) == 0 {
		return nil //Nothing to set
	}
	// Using multiple values in query to spare some network processing time
	qry := fmt.Sprintf("INSERT INTO %s (tpid, tag, destinations_tag, rates_tag) VALUES ", utils.TBL_TP_DESTINATION_RATES)
	for idx, drPair := range dr.DestinationRates {
		if idx != 0 { //Consecutive values after the first will be prefixed with "," as separator
			qry += ","
		}
		qry += fmt.Sprintf("('%s','%s','%s','%s')", dr.TPid, dr.DestinationRateId, drPair.DestinationId, drPair.RateId)
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

func (self *SQLStorage) SetTPDestRateTiming(drt *utils.TPDestRateTiming) error {
	if len(drt.DestRateTimings) == 0 {
		return nil //Nothing to set
	}
	// Using multiple values in query to spare some network processing time
	qry := fmt.Sprintf("INSERT INTO %s (tpid, tag, destrates_tag, timing_tag, weight) VALUES ", utils.TBL_TP_DESTRATE_TIMINGS)
	for idx, drtPair := range drt.DestRateTimings {
		if idx != 0 { //Consecutive values after the first will be prefixed with "," as separator
			qry += ","
		}
		qry += fmt.Sprintf("('%s','%s','%s','%s',%f)", drt.TPid, drt.DestRateTimingId, drtPair.DestRatesId, drtPair.TimingId, drtPair.Weight)
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

func (self *SQLStorage) GetActions(string) (as Actions, err error) {
	return
}

func (self *SQLStorage) SetActions(key string, as Actions) (err error) { return }

func (self *SQLStorage) GetUserBalance(string) (ub *UserBalance, err error) { return }

func (self *SQLStorage) SetUserBalance(ub *UserBalance) (err error) { return }

func (self *SQLStorage) GetActionTimings(key string) (ats ActionTimings, err error) { return }

func (self *SQLStorage) SetActionTimings(key string, ats ActionTimings) (err error) { return }

func (self *SQLStorage) GetAllActionTimings() (ats map[string]ActionTimings, err error) {
	return
}

func (self *SQLStorage) LogCallCost(uuid, source string, cc *CallCost) (err error) {
	if self.Db == nil {
		//timespans.Logger.Warning("Cannot write log to database.")
		return
	}
	tss, err := json.Marshal(cc.Timespans)
	if err != nil {
		Logger.Err(fmt.Sprintf("Error marshalling timespans to json: %v", err))
	}
	_, err = self.Db.Exec(fmt.Sprintf("INSERT INTO callcosts VALUES ('NULL','%s', '%s','%s', '%s', '%s', '%s', '%s', '%s', %v, %v, '%s')",
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

func (self *SQLStorage) GetCallCostLog(uuid, source string) (cc *CallCost, err error) {
	row := self.Db.QueryRow(fmt.Sprintf("SELECT * FROM callcosts WHERE uuid='%s' AND source='%s'", uuid, source))
	var uuid_found string
	var timespansJson string
	err = row.Scan(&uuid_found, &cc.Direction, &cc.Tenant, &cc.TOR, &cc.Subject, &cc.Destination, &cc.Cost, &cc.ConnectFee, &timespansJson)
	err = json.Unmarshal([]byte(timespansJson), cc.Timespans)
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
	_, err = self.Db.Exec(fmt.Sprintf("INSERT INTO cdrs_primary VALUES (NULL, '%s', '%s','%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', %d)",
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

func (self *SQLStorage) SetRatedCdr(cdr utils.CDR, cc *CallCost) (err error) {
	_, err = self.Db.Exec(fmt.Sprintf("INSERT INTO rated_cdrs VALUES ('%s', '%s', '%s', '%s')",
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

func (self *SQLStorage) GetAllRatedCdr() ([]utils.CDR, error) {
	return nil, nil
}

func (self *SQLStorage) GetTpDestinations(tpid, tag string) ([]*Destination, error) {
	var dests []*Destination
	q := fmt.Sprintf("SELECT * FROM %s WHERE tpid=%s", utils.TBL_TP_DESTINATIONS, tpid)
	if tag != "" {
		q += "AND tag=" + tag
	}
	rows, err := self.Db.Query(q, tpid)
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

func (self *SQLStorage) GetTpRates(tpid, tag string) (map[string]*Rate, error) {
	rts := make(map[string]*Rate)
	q := fmt.Sprintf("SELECT * FROM %s WHERE tpid=%s", utils.TBL_TP_RATES, tpid)
	if tag != "" {
		q += "AND tag=" + tag
	}
	rows, err := self.Db.Query(q, tpid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		var tpid, tag string
		var connect_fee, rate, priced_units, rate_increments float64
		if err := rows.Scan(&id, &tpid, &tag, &connect_fee, &rate, &priced_units, &rate_increments); err != nil {
			return nil, err
		}

		r := &Rate{
			Tag:            tag,
			ConnectFee:     connect_fee,
			Price:          rate,
			PricedUnits:    priced_units,
			RateIncrements: rate_increments,
		}

		rts[tag] = r
	}
	return rts, rows.Err()
}

func (self *SQLStorage) GetTpDestinationRates(tpid, tag string) (map[string][]*DestinationRate, error) {
	rts := make(map[string][]*DestinationRate)
	q := fmt.Sprintf("SELECT * FROM %s WHERE tpid=%s", utils.TBL_TP_DESTINATION_RATES, tpid)
	if tag != "" {
		q += "AND tag=" + tag
	}
	rows, err := self.Db.Query(q, tpid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		var tpid, tag, destinations_tag, rate_tag string
		if err := rows.Scan(&id, &tpid, &tag, destinations_tag, rate_tag); err != nil {
			return nil, err
		}

		dr := &DestinationRate{
			Tag:             tag,
			DestinationsTag: destinations_tag,
			RateTag:         rate_tag,
		}

		rts[tag] = append(rts[tag], dr)
	}
	return rts, rows.Err()
}

func (self *SQLStorage) GetTpTimings(tpid, tag string) (map[string]*Timing, error) {
	tms := make(map[string]*Timing)
	q := fmt.Sprintf("SELECT * FROM %s WHERE tpid=%s", utils.TBL_TP_TIMINGS, tpid)
	if tag != "" {
		q += "AND tag=" + tag
	}
	rows, err := self.Db.Query(q, tpid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		var tpid, tag, years, months, month_days, week_days, start_time string
		if err := rows.Scan(&id, &tpid, &tag, &years, &months, &month_days, &week_days, &start_time); err != nil {
			return nil, err
		}
		tms[tag] = NewTiming(tag, years, months, month_days, week_days, start_time)
	}
	return tms, rows.Err()
}

func (self *SQLStorage) GetTpDestinationRateTimings(tpid, tag string) ([]*DestinationRateTiming, error) {
	var rts []*DestinationRateTiming
	q := fmt.Sprintf("SELECT * FROM %s WHERE tpid=%s", utils.TBL_TP_DESTRATE_TIMINGS, tpid)
	if tag != "" {
		q += "AND tag=" + tag
	}
	rows, err := self.Db.Query(q, tpid)
	if err != nil {
		return nil, err
	}
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
			TimingsTag:          timings_tag,
		}
		rts = append(rts, rt)
	}
	return rts, rows.Err()
}

func (self *SQLStorage) GetTpRatingProfiles(tpid, tag string) (map[string]*RatingProfile, error) {
	rpfs := make(map[string]*RatingProfile)
	q := fmt.Sprintf("SELECT * FROM %s WHERE tpid=%s", utils.TBL_TP_RATE_PROFILES, tpid)
	if tag != "" {
		q += "AND tag=" + tag
	}
	rows, err := self.Db.Query(q, tpid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		var tpid, tag, tenant, tor, direction, subject, fallback_subject, rates_timing_tag, activation_time string

		if err := rows.Scan(&id, &tag, &tpid, &tenant, &tor, &direction, &subject, &fallback_subject, &rates_timing_tag, &activation_time); err != nil {
			return nil, err
		}
		key := fmt.Sprintf("%s:%s:%s:%s", direction, tenant, tor, subject)
		rp, ok := rpfs[key]
		if !ok || rp.tag != tag {
			rp = &RatingProfile{Id: key, tag: tag}
			rpfs[key] = rp
		}
		rp.ratesTimingTag = rates_timing_tag
		rp.activationTime = activation_time
		if fallback_subject != "" {
			rp.FallbackKey = fmt.Sprintf("%s:%s:%s:%s", direction, tenant, tor, fallback_subject)
		}
	}
	return rpfs, rows.Err()
}
func (self *SQLStorage) GetTpActions(tpid, tag string) (map[string][]*Action, error) {
	as := make(map[string][]*Action)
	q := fmt.Sprintf("SELECT * FROM %s WHERE tpid=%s", utils.TBL_TP_ACTIONS, tpid)
	if tag != "" {
		q += "AND tag=" + tag
	}
	rows, err := self.Db.Query(q, tpid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		var units, rate, minutes_weight, weight float64
		var tpid, tag, action, balance_tag, direction, destinations_tag, rate_type string
		if err := rows.Scan(&id, &tpid, &tag, &action, &balance_tag, &direction, &units, &destinations_tag, &rate_type, &rate, &minutes_weight, &weight); err != nil {
			return nil, err
		}
		var a *Action
		if balance_tag != MINUTES {
			a = &Action{
				ActionType: action,
				BalanceId:  balance_tag,
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
				BalanceId:  balance_tag,
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

func (self *SQLStorage) GetTpActionTimings(tpid, tag string) (ats map[string][]*ActionTiming, err error) {
	ats = make(map[string][]*ActionTiming)
	q := fmt.Sprintf("SELECT * FROM %s WHERE tpid=%s", utils.TBL_TP_ACTION_TIMINGS, tpid)
	if tag != "" {
		q += "AND tag=" + tag
	}
	rows, err := self.Db.Query(q, tpid)
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

func (self *SQLStorage) GetTpActionTriggers(tpid, tag string) (map[string][]*ActionTrigger, error) {
	ats := make(map[string][]*ActionTrigger)
	q := fmt.Sprintf("SELECT * FROM %s WHERE tpid=%s", utils.TBL_TP_ACTION_TRIGGERS, tpid)
	if tag != "" {
		q += "AND tag=" + tag
	}
	rows, err := self.Db.Query(q, tpid)
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

func (self *SQLStorage) GetTpAccountActions(tpid, tag string) ([]*AccountAction, error) {
	var acs []*AccountAction
	q := fmt.Sprintf("SELECT * FROM %s WHERE tpid=%s", utils.TBL_TP_ACCOUNT_ACTIONS, tpid)
	if tag != "" {
		q += "AND tag=" + tag
	}
	rows, err := self.Db.Query(q, tpid)
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
