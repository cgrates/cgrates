/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"io/ioutil"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
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

func (self *SQLStorage) CreateTablesFromScript(scriptPath string) error {
	fileContent, err := ioutil.ReadFile(scriptPath)
	if err != nil {
		return err
	}
	qries := strings.Split(string(fileContent), ";") // Script has normally multiple queries separate by ';' go driver does not understand this so we handle it here
	for _, qry := range qries {
		qry = strings.TrimSpace(qry) // Avoid empty queries
		if len(qry) == 0 {
			continue
		}
		if _, err := self.Db.Exec(qry); err != nil {
			return err
		}
	}
	return nil
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

func (self *SQLStorage) GetTPTableIds(tpid, table, distinct string, filters map[string]string) ([]string, error) {
	qry := fmt.Sprintf("SELECT DISTINCT %s FROM %s where tpid='%s'", distinct, table, tpid)
	for key, value := range filters {
		if key != "" && value != "" {
			qry += fmt.Sprintf(" AND %s='%s'", key, value)
		}
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

func (self *SQLStorage) SetTPTiming(tpid string, tm *utils.TPTiming) error {
	if _, err := self.Db.Exec(fmt.Sprintf("INSERT INTO %s (tpid, id, years, months, month_days, week_days, time) VALUES('%s','%s','%s','%s','%s','%s','%s') ON DUPLICATE KEY UPDATE years=values(years), months=values(months), month_days=values(month_days), week_days=values(week_days), time=values(time)",
		utils.TBL_TP_TIMINGS, tpid, tm.Id, tm.Years.Serialize(";"), tm.Months.Serialize(";"), tm.MonthDays.Serialize(";"),
		tm.WeekDays.Serialize(";"), tm.StartTime)); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) RemTPData(table, tpid string, args ...string) error {
	q := fmt.Sprintf("DELETE FROM %s WHERE tpid='%s' AND id='%s'", table, tpid, args[0])
	switch table {
	case utils.TBL_TP_RATE_PROFILES:
		q = fmt.Sprintf("DELETE FROM %s WHERE tpid='%s' AND loadid='%s' AND tenant='%s' AND category='%s' AND direction='%s' AND subject='%s'",
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
func (self *SQLStorage) GetTPDestination(tpid, destTag string) (*Destination, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT prefix FROM %s WHERE tpid='%s' AND id='%s'", utils.TBL_TP_DESTINATIONS, tpid, destTag))
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
		d.AddPrefix(pref)
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
	var buffer bytes.Buffer // Use bytes buffer istead of string concatenation since that becomes quite heavy on large prefixes
	buffer.WriteString(fmt.Sprintf("INSERT INTO %s (tpid, id, prefix) VALUES ", utils.TBL_TP_DESTINATIONS))
	for idx, prefix := range dest.Prefixes {
		if idx != 0 {
			buffer.WriteRune(',')
		}
		buffer.WriteString(fmt.Sprintf("('%s','%s','%s')", tpid, dest.Id, prefix))
		idx++
	}
	buffer.WriteString(" ON DUPLICATE KEY UPDATE prefix=values(prefix)")
	if _, err := self.Db.Exec(buffer.String()); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) SetTPRates(tpid string, rts map[string][]*utils.RateSlot) error {
	if len(rts) == 0 {
		return nil //Nothing to set
	}
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("INSERT INTO %s (tpid, id, connect_fee, rate, rate_unit, rate_increment, group_interval_start) VALUES ",
		utils.TBL_TP_RATES))
	i := 0
	for rtId, rtRows := range rts {
		for _, rt := range rtRows {
			if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
				buffer.WriteRune(',')
			}
			buffer.WriteString(fmt.Sprintf("('%s', '%s', %f, %f, '%s', '%s','%s')",
				tpid, rtId, rt.ConnectFee, rt.Rate, rt.RateUnit, rt.RateIncrement, rt.GroupIntervalStart))
			i++
		}
	}
	buffer.WriteString(" ON DUPLICATE KEY UPDATE connect_fee=values(connect_fee),rate=values(rate),rate_increment=values(rate_increment),group_interval_start=values(group_interval_start)")
	if _, err := self.Db.Exec(buffer.String()); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) SetTPDestinationRates(tpid string, drs map[string][]*utils.DestinationRate) error {
	if len(drs) == 0 {
		return nil //Nothing to set
	}
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("INSERT INTO %s (tpid,id,destinations_id,rates_id,rounding_method,rounding_decimals) VALUES ", utils.TBL_TP_DESTINATION_RATES))
	i := 0
	for drId, drRows := range drs {
		for _, dr := range drRows {
			if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
				buffer.WriteRune(',')
			}
			buffer.WriteString(fmt.Sprintf("('%s','%s','%s','%s','%s', %d)", tpid, drId, dr.DestinationId, dr.RateId, dr.RoundingMethod, dr.RoundingDecimals))
			i++
		}
	}
	buffer.WriteString(" ON DUPLICATE KEY UPDATE destinations_id=values(destinations_id),rates_id=values(rates_id),rounding_method=values(rounding_method),rounding_decimals=values(rounding_decimals)")
	if _, err := self.Db.Exec(buffer.String()); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) SetTPRatingPlans(tpid string, drts map[string][]*utils.TPRatingPlanBinding) error {
	if len(drts) == 0 {
		return nil //Nothing to set
	}
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("INSERT INTO %s (tpid, id, destrates_id, timing_id, weight) VALUES ", utils.TBL_TP_RATING_PLANS))
	i := 0
	for drtId, drtRows := range drts {
		for _, drt := range drtRows {
			if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
				buffer.WriteRune(',')
			}
			buffer.WriteString(fmt.Sprintf("('%s','%s','%s','%s',%f)", tpid, drtId, drt.DestinationRatesId, drt.TimingId, drt.Weight))
			i++
		}
	}
	buffer.WriteString(" ON DUPLICATE KEY UPDATE weight=values(weight)")
	if _, err := self.Db.Exec(buffer.String()); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) SetTPRatingProfiles(tpid string, rps map[string]*utils.TPRatingProfile) error {
	if len(rps) == 0 {
		return nil //Nothing to set
	}
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("INSERT INTO %s (tpid,loadid,direction,tenant,category,subject,activation_time,rating_plan_id,fallback_subjects) VALUES ",
		utils.TBL_TP_RATE_PROFILES))
	i := 0
	for _, rp := range rps {
		for _, rpa := range rp.RatingPlanActivations {
			if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
				buffer.WriteRune(',')
			}
			buffer.WriteString(fmt.Sprintf("('%s', '%s', '%s', '%s', '%s', '%s', '%s','%s','%s')", tpid, rp.LoadId, rp.Direction, rp.Tenant, rp.Category,
				rp.Subject, rpa.ActivationTime, rpa.RatingPlanId, rpa.FallbackSubjects))
			i++
		}
	}
	buffer.WriteString(" ON DUPLICATE KEY UPDATE fallback_subjects=values(fallback_subjects)")
	if _, err := self.Db.Exec(buffer.String()); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) SetTPSharedGroups(tpid string, sgs map[string]*SharedGroup) error {
	if len(sgs) == 0 {
		return nil //Nothing to set
	}
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("INSERT INTO %s (tpid,id,account,strategy,rate_subject) VALUES ", utils.TBL_TP_SHARED_GROUPS))
	i := 0
	for sgId, sg := range sgs {
		for account, params := range sg.AccountParameters {
			if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
				buffer.WriteRune(',')
			}
			buffer.WriteString(fmt.Sprintf("('%s','%s','%s','%s','%s')",
				tpid, sgId, account, params.Strategy, params.RatingSubject))
			i++
		}
	}
	buffer.WriteString(" ON DUPLICATE KEY UPDATE account=values(account),strategy=values(strategy),rate_subject=values(rate_subject)")
	if _, err := self.Db.Exec(buffer.String()); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) SetTPLCRs(tpid string, lcrs map[string]*LCR) error {
	if len(lcrs) == 0 {
		return nil //Nothing to set
	}
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("INSERT INTO %s (tpid,direction,tenant,customer,destination_id,category,strategy,suppliers,activation_time,weight) VALUES ", utils.TBL_TP_LCRS))
	i := 0
	for _, lcr := range lcrs {
		for _, act := range lcr.Activations {
			for _, entry := range act.Entries {
				if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
					buffer.WriteRune(',')
				}
				buffer.WriteString(fmt.Sprintf("('%s','%s','%s','%s','%s','%s','%s','%v','%v')",
					tpid, lcr.Direction, lcr.Tenant, lcr.Customer, entry.DestinationId, entry.Category, entry.Strategy, entry.Suppliers, act.ActivationTime, entry.Weight))
				i++
			}
		}
	}
	if _, err := self.Db.Exec(buffer.String()); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) SetTPActions(tpid string, acts map[string][]*utils.TPAction) error {
	if len(acts) == 0 {
		return nil //Nothing to set
	}
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("INSERT INTO %s (tpid,id,action,balance_type,direction,units,expiry_time,destination_id,rating_subject,shared_group,balance_weight,extra_parameters,weight) VALUES ", utils.TBL_TP_ACTIONS))
	i := 0
	for actId, actRows := range acts {
		for _, act := range actRows {
			if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
				buffer.WriteRune(',')
			}
			buffer.WriteString(fmt.Sprintf("('%s','%s','%s','%s','%s',%f,'%s','%s','%s','%s',%f,'%s',%f)",
				tpid, actId, act.Identifier, act.BalanceType, act.Direction, act.Units, act.ExpiryTime,
				act.DestinationId, act.RatingSubject, act.SharedGroup, act.BalanceWeight, act.ExtraParameters, act.Weight))
			i++
		}
	}
	buffer.WriteString(" ON DUPLICATE KEY UPDATE action=values(action),balance_type=values(balance_type),direction=values(direction),units=values(units),expiry_time=values(expiry_time),destination_id=values(destination_id),rating_subject=values(rating_subject),shared_group=values(shared_group),balance_weight=values(balance_weight),extra_parameters=values(extra_parameters),weight=values(weight)")
	if _, err := self.Db.Exec(buffer.String()); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) GetTPActions(tpid, actsId string) (*utils.TPActions, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT action,balance_type,direction,units,expiry_time,destination_id,rating_subject,shared_group,balance_weight,extra_parameters,weight FROM %s WHERE tpid='%s' AND id='%s'", utils.TBL_TP_ACTIONS, tpid, actsId))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	acts := &utils.TPActions{TPid: tpid, ActionsId: actsId}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one result
		var action, balanceId, dir, destId, rateSubject, sharedGroup, expTime, extraParameters string
		var units, balanceWeight, weight float64
		if err = rows.Scan(&action, &balanceId, &dir, &units, &expTime, &destId, &rateSubject, &sharedGroup, &balanceWeight, &extraParameters, &weight); err != nil {
			return nil, err
		}
		acts.Actions = append(acts.Actions, &utils.TPAction{
			Identifier:      action,
			BalanceType:     balanceId,
			Direction:       dir,
			Units:           units,
			ExpiryTime:      expTime,
			DestinationId:   destId,
			RatingSubject:   rateSubject,
			BalanceWeight:   balanceWeight,
			SharedGroup:     sharedGroup,
			ExtraParameters: extraParameters,
			Weight:          weight})
	}
	if i == 0 {
		return nil, nil
	}
	return acts, nil
}

// Sets actionTimings in sqlDB. Imput is expected in form map[actionTimingId][]rows, eg a full .csv file content
func (self *SQLStorage) SetTPActionTimings(tpid string, ats map[string][]*utils.TPActionTiming) error {
	if len(ats) == 0 {
		return nil //Nothing to set
	}
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("INSERT INTO %s (tpid,id,actions_id,timing_id,weight) VALUES ", utils.TBL_TP_ACTION_PLANS))
	i := 0
	for atId, atRows := range ats {
		for _, at := range atRows {
			if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
				buffer.WriteRune(',')
			}
			buffer.WriteString(fmt.Sprintf("('%s','%s','%s','%s',%f)", tpid, atId, at.ActionsId, at.TimingId, at.Weight))
			i++
		}
	}
	buffer.WriteString(" ON DUPLICATE KEY UPDATE timing_id=values(timing_id),weight=values(weight)")
	if _, err := self.Db.Exec(buffer.String()); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) GetTPActionTimings(tpid, atId string) (map[string][]*utils.TPActionTiming, error) {
	ats := make(map[string][]*utils.TPActionTiming)
	q := fmt.Sprintf("SELECT id,actions_id,timing_id,weight FROM %s WHERE tpid='%s'", utils.TBL_TP_ACTION_PLANS, tpid)
	if atId != "" {
		q += fmt.Sprintf(" AND id='%s'", atId)
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
		ats[tag] = append(ats[tag], &utils.TPActionTiming{ActionsId: actionsId, TimingId: timingId, Weight: weight})
	}
	return ats, nil
}

func (self *SQLStorage) SetTPActionTriggers(tpid string, ats map[string][]*utils.TPActionTrigger) error {
	if len(ats) == 0 {
		return nil //Nothing to set
	}
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("INSERT INTO %s (tpid,id,balance_type,direction,threshold_type,threshold_value,recurrent,destination_id,actions_id,weight) VALUES ",
		utils.TBL_TP_ACTION_TRIGGERS))
	i := 0
	for atId, atRows := range ats {
		for _, atsRow := range atRows {
			if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
				buffer.WriteRune(',')
			}
			buffer.WriteString(fmt.Sprintf("('%s','%s','%s','%s','%s', %f, %t, '%s','%s',%f)",
				tpid, atId, atsRow.BalanceType, atsRow.Direction, atsRow.ThresholdType,
				atsRow.ThresholdValue, atsRow.Recurrent, atsRow.DestinationId, atsRow.ActionsId, atsRow.Weight))
			i++
		}
	}
	buffer.WriteString(" ON DUPLICATE KEY UPDATE recurrent=values(recurrent), weight=values(weight)")
	if _, err := self.Db.Exec(buffer.String()); err != nil {
		return err
	}
	return nil
}

// Sets a group of account actions. Map key has the role of grouping within a tpid
func (self *SQLStorage) SetTPAccountActions(tpid string, aa map[string]*utils.TPAccountActions) error {
	if len(aa) == 0 {
		return nil //Nothing to set
	}
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("INSERT INTO %s (tpid, loadid, tenant, account, direction, action_plan_id, action_triggers_id) VALUES ", utils.TBL_TP_ACCOUNT_ACTIONS))
	i := 0
	for _, aActs := range aa {
		if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
			buffer.WriteRune(',')
		}
		buffer.WriteString(fmt.Sprintf("('%s','%s','%s','%s','%s','%s','%s')",
			tpid, aActs.LoadId, aActs.Tenant, aActs.Account, aActs.Direction, aActs.ActionPlanId, aActs.ActionTriggersId))
		i++
	}
	buffer.WriteString(" ON DUPLICATE KEY UPDATE action_plan_id=values(action_plan_id), action_triggers_id=values(action_triggers_id)")
	if _, err := self.Db.Exec(buffer.String()); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) LogCallCost(cgrid, source, runid string, cc *CallCost) (err error) {
	//ToDo: Add cgrid to logCallCost
	if self.Db == nil {
		//timespans.Logger.Warning("Cannot write log to database.")
		return
	}
	tss, err := json.Marshal(cc.Timespans)
	if err != nil {
		Logger.Err(fmt.Sprintf("Error marshalling timespans to json: %v", err))
	}
	_, err = self.Db.Exec(fmt.Sprintf("INSERT INTO %s (cost_time,cost_source,cgrid,runid,tor,direction,tenant,category,account,subject,destination,cost,timespans) VALUES (now(),'%s','%s','%s','%s','%s','%s','%s','%s','%s','%s',%f,'%s') ON DUPLICATE KEY UPDATE cost_time=now(),cost_source=values(cost_source),tor=values(tor),direction=values(direction),tenant=values(tenant),category=values(category),account=values(account),subject=values(subject),destination=values(destination),cost=values(cost),timespans=values(timespans)",
		utils.TBL_COST_DETAILS,
		source,
		cgrid,
		runid,
		cc.TOR,
		cc.Direction,
		cc.Tenant,
		cc.Category,
		cc.Account,
		cc.Subject,
		cc.Destination,
		cc.Cost,
		tss))
	if err != nil {
		Logger.Err(fmt.Sprintf("failed to execute insert statement: %v", err))
	}
	return
}

func (self *SQLStorage) GetCallCostLog(cgrid, source, runid string) (cc *CallCost, err error) {
	qry := fmt.Sprintf("SELECT tor,direction,tenant,category,account,subject,destination,cost,timespans FROM %s WHERE cgrid='%s' AND runid='%s'",
		utils.TBL_COST_DETAILS, cgrid, runid)
	if len(source) != 0 {
		qry += fmt.Sprintf(" AND cost_source='%s'", source)
	}
	row := self.Db.QueryRow(qry)
	var timespansJson string
	cc = &CallCost{Cost: -1}
	err = row.Scan(&cc.TOR, &cc.Direction, &cc.Tenant, &cc.Category, &cc.Account, &cc.Subject, &cc.Destination, &cc.Cost, &timespansJson)
	if len(timespansJson) == 0 { // No costs returned
		return nil, nil
	}
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
func (self *SQLStorage) LogError(uuid, source, runid, errstr string) (err error) { return }

func (self *SQLStorage) SetCdr(cdr *utils.StoredCdr) (err error) {
	_, err = self.Db.Exec(fmt.Sprintf("INSERT INTO  %s (cgrid,tor,accid,cdrhost,cdrsource,reqtype,direction,tenant,category,account,subject,destination,setup_time,answer_time,`usage`) VALUES ('%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s', %d)",
		utils.TBL_CDRS_PRIMARY,
		cdr.CgrId,
		cdr.TOR,
		cdr.AccId,
		cdr.CdrHost,
		cdr.CdrSource,
		cdr.ReqType,
		cdr.Direction,
		cdr.Tenant,
		cdr.Category,
		cdr.Account,
		cdr.Subject,
		cdr.Destination,
		cdr.SetupTime,
		cdr.AnswerTime,
		cdr.Usage,
	))
	if err != nil {
		Logger.Err(fmt.Sprintf("failed to execute cdr insert statement: %v", err))
	}
	extraFields, err := json.Marshal(cdr.ExtraFields)
	if err != nil {
		Logger.Err(fmt.Sprintf("Error marshalling cdr extra fields to json: %v", err))
	}
	_, err = self.Db.Exec(fmt.Sprintf("INSERT INTO  %s (cgrid,extra_fields) VALUES ('%s', '%s')",
		utils.TBL_CDRS_EXTRA,
		cdr.CgrId,
		extraFields,
	))
	if err != nil {
		Logger.Err(fmt.Sprintf("failed to execute cdr insert statement: %v", err))
	}

	return
}

func (self *SQLStorage) SetRatedCdr(storedCdr *utils.StoredCdr, extraInfo string) (err error) {
	_, err = self.Db.Exec(fmt.Sprintf("INSERT INTO %s (mediation_time,cgrid,runid,reqtype,direction,tenant,category,account,subject,destination,setup_time,answer_time,`usage`,cost,extra_info) VALUES (now(),'%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s',%d,%f,'%s') ON DUPLICATE KEY UPDATE mediation_time=now(),reqtype=values(reqtype),direction=values(direction),tenant=values(tenant),category=values(category),account=values(account),subject=values(subject),destination=values(destination),setup_time=values(setup_time),answer_time=values(answer_time),`usage`=values(`usage`),cost=values(cost),extra_info=values(extra_info)",
		utils.TBL_RATED_CDRS,
		storedCdr.CgrId,
		storedCdr.MediationRunId,
		storedCdr.ReqType,
		storedCdr.Direction,
		storedCdr.Tenant,
		storedCdr.Category,
		storedCdr.Account,
		storedCdr.Subject,
		storedCdr.Destination,
		storedCdr.SetupTime,
		storedCdr.AnswerTime,
		storedCdr.Usage,
		storedCdr.Cost,
		extraInfo))
	if err != nil {
		Logger.Err(fmt.Sprintf("failed to execute cdr insert statement: %s", err.Error()))
	}
	return
}

// Return a slice of CDRs from storDb using optional filters.a
// ignoreErr - do not consider cdrs with rating errors
// ignoreRated - do not consider cdrs which were already rated, including here the ones with errors
func (self *SQLStorage) GetStoredCdrs(cgrIds, runIds, tors, cdrHosts, cdrSources, reqTypes, directions, tenants, categories, accounts, subjects, destPrefixes []string, orderIdStart, orderIdEnd int64,
	timeStart, timeEnd time.Time, ignoreErr, ignoreRated, ignoreDerived bool) ([]*utils.StoredCdr, error) {
	var cdrs []*utils.StoredCdr
	var q *bytes.Buffer // Need to query differently since in case of primary, unmediated CDRs some values will be missing
	if ignoreDerived {
		q = bytes.NewBufferString(fmt.Sprintf("SELECT %s.cgrid,%s.tbid,%s.tor,%s.accid,%s.cdrhost,%s.cdrsource,%s.reqtype,%s.direction,%s.tenant,%s.category,%s.account,%s.subject,%s.destination,%s.setup_time,%s.answer_time,%s.`usage`,%s.extra_fields,%s.runid,%s.cost FROM %s LEFT JOIN %s ON %s.cgrid=%s.cgrid LEFT JOIN %s ON %s.cgrid=%s.cgrid",
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_EXTRA,
			utils.TBL_RATED_CDRS,
			utils.TBL_RATED_CDRS,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_EXTRA,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_EXTRA,
			utils.TBL_RATED_CDRS,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_RATED_CDRS))
	} else {
		q = bytes.NewBufferString(fmt.Sprintf("SELECT %s.cgrid,%s.tbid,%s.tor,%s.accid,%s.cdrhost,%s.cdrsource,%s.reqtype,%s.direction,%s.tenant,%s.category,%s.account,%s.subject,%s.destination,%s.setup_time,%s.answer_time,%s.`usage`,%s.extra_fields,%s.runid,%s.cost FROM %s LEFT JOIN %s ON %s.cgrid=%s.cgrid LEFT JOIN %s ON %s.cgrid=%s.cgrid",
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_RATED_CDRS,
			utils.TBL_RATED_CDRS,
			utils.TBL_RATED_CDRS,
			utils.TBL_RATED_CDRS,
			utils.TBL_RATED_CDRS,
			utils.TBL_RATED_CDRS,
			utils.TBL_RATED_CDRS,
			utils.TBL_RATED_CDRS,
			utils.TBL_RATED_CDRS,
			utils.TBL_RATED_CDRS,
			utils.TBL_CDRS_EXTRA,
			utils.TBL_RATED_CDRS,
			utils.TBL_RATED_CDRS,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_EXTRA,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_EXTRA,
			utils.TBL_RATED_CDRS,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_RATED_CDRS))
	}
	fltr := new(bytes.Buffer)
	if len(cgrIds) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idxId, cgrId := range cgrIds {
			if idxId != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.cgrid='%s'", utils.TBL_CDRS_PRIMARY, cgrId))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(runIds) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, runId := range runIds {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.runid='%s'", utils.TBL_RATED_CDRS, runId))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(tors) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, host := range tors {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.tor='%s'", utils.TBL_CDRS_PRIMARY, host))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(cdrHosts) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, host := range cdrHosts {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.cdrhost='%s'", utils.TBL_CDRS_PRIMARY, host))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(cdrSources) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, src := range cdrSources {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.cdrsource='%s'", utils.TBL_CDRS_PRIMARY, src))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(reqTypes) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, reqType := range reqTypes {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.reqtype='%s'", utils.TBL_CDRS_PRIMARY, reqType))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(directions) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, direction := range directions {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.direction='%s'", utils.TBL_CDRS_PRIMARY, direction))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(tenants) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, tenant := range tenants {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf("  %s.tenant='%s'", utils.TBL_CDRS_PRIMARY, tenant))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(categories) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, category := range categories {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.category='%s'", utils.TBL_CDRS_PRIMARY, category))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(accounts) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, account := range accounts {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.account='%s'", utils.TBL_CDRS_PRIMARY, account))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(subjects) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, subject := range subjects {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.subject='%s'", utils.TBL_CDRS_PRIMARY, subject))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(destPrefixes) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, destPrefix := range destPrefixes {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.destination LIKE '%s%%'", utils.TBL_CDRS_PRIMARY, destPrefix))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if orderIdStart != 0 {
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.WriteString(fmt.Sprintf(" %s.tbid>=%d", utils.TBL_CDRS_PRIMARY, orderIdStart))
	}
	if orderIdEnd != 0 {
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.WriteString(fmt.Sprintf(" %s.tbid<%d", utils.TBL_CDRS_PRIMARY, orderIdEnd))
	}
	if !timeStart.IsZero() {
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.WriteString(fmt.Sprintf(" %s.answer_time>='%s'", utils.TBL_CDRS_PRIMARY, timeStart))
	}
	if !timeEnd.IsZero() {
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.WriteString(fmt.Sprintf(" %s.answer_time<'%s'", utils.TBL_CDRS_PRIMARY, timeEnd))
	}
	if ignoreRated {
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		if ignoreErr {
			fltr.WriteString(fmt.Sprintf(" %s.cost IS NULL", utils.TBL_RATED_CDRS))
		} else {
			fltr.WriteString(fmt.Sprintf(" (%s.cost=-1 OR %s.cost IS NULL)", utils.TBL_RATED_CDRS, utils.TBL_RATED_CDRS))
		}
	} else if ignoreErr {
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.WriteString(fmt.Sprintf(" (%s.cost!=-1 OR %s.cost IS NULL)", utils.TBL_RATED_CDRS, utils.TBL_RATED_CDRS))
	}
	if ignoreDerived {
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.WriteString(fmt.Sprintf(" (%s.runid='%s' OR %s.cost IS NULL)", utils.TBL_RATED_CDRS, utils.DEFAULT_RUNID, utils.TBL_RATED_CDRS))
	}
	if fltr.Len() != 0 {
		q.WriteString(fmt.Sprintf(" WHERE %s", fltr.String()))
	}
	rows, err := self.Db.Query(q.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var cgrid, tor, accid, cdrhost, cdrsrc, reqtype, direction, tenant, category, account, subject, destination, runid sql.NullString
		var extraFields []byte
		var setupTime, answerTime mysql.NullTime
		var orderid int64
		var usage sql.NullInt64
		var cost sql.NullFloat64 // So we can export unmediated CDRs
		var extraFieldsMp map[string]string
		if err := rows.Scan(&cgrid, &orderid, &tor, &accid, &cdrhost, &cdrsrc, &reqtype, &direction, &tenant, &category, &account, &subject, &destination, &setupTime, &answerTime, &usage,
			&extraFields, &runid, &cost); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(extraFields, &extraFieldsMp); err != nil {
			return nil, fmt.Errorf("JSON unmarshal error for cgrid: %s, runid: %v, error: %s", cgrid, runid, err.Error())
		}
		storCdr := &utils.StoredCdr{
			CgrId: cgrid.String, OrderId: orderid, TOR: tor.String, AccId: accid.String, CdrHost: cdrhost.String, CdrSource: cdrsrc.String, ReqType: reqtype.String,
			Direction: direction.String, Tenant: tenant.String,
			Category: category.String, Account: account.String, Subject: subject.String, Destination: destination.String,
			SetupTime: setupTime.Time, AnswerTime: answerTime.Time, Usage: time.Duration(usage.Int64),
			ExtraFields: extraFieldsMp, MediationRunId: runid.String, Cost: cost.Float64,
		}
		if !cost.Valid { //There was no cost provided, will fakely insert 0 if we do not handle it and reflect on re-rating
			storCdr.Cost = -1
		}
		cdrs = append(cdrs, storCdr)
	}
	return cdrs, nil
}

// Remove CDR data out of all CDR tables based on their cgrid
func (self *SQLStorage) RemStoredCdrs(cgrIds []string) error {
	if len(cgrIds) == 0 {
		return nil
	}
	buffRated := bytes.NewBufferString(fmt.Sprintf("DELETE FROM %s WHERE", utils.TBL_RATED_CDRS))
	buffCosts := bytes.NewBufferString(fmt.Sprintf("DELETE FROM %s WHERE", utils.TBL_COST_DETAILS))
	buffCdrExtra := bytes.NewBufferString(fmt.Sprintf("DELETE FROM %s WHERE", utils.TBL_CDRS_EXTRA))
	buffCdrPrimary := bytes.NewBufferString(fmt.Sprintf("DELETE FROM %s WHERE", utils.TBL_CDRS_PRIMARY))
	qryBuffers := []*bytes.Buffer{buffRated, buffCosts, buffCdrExtra, buffCdrPrimary}
	for idx, cgrId := range cgrIds {
		for _, buffer := range qryBuffers {
			if idx != 0 {
				buffer.WriteString(" OR")
			}
			buffer.WriteString(fmt.Sprintf(" cgrid='%s'", cgrId))
		}
	}
	for _, buffer := range qryBuffers {
		if _, err := self.Db.Exec(buffer.String()); err != nil {
			return err
		}
	}
	return nil
}

func (self *SQLStorage) GetTpDestinations(tpid, tag string) ([]*Destination, error) {
	var dests []*Destination
	q := fmt.Sprintf("SELECT * FROM %s WHERE tpid='%s'", utils.TBL_TP_DESTINATIONS, tpid)
	if len(tag) != 0 {
		q += fmt.Sprintf(" AND id='%s'", tag)
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
		dest.AddPrefix(prefix)
	}
	return dests, nil
}

func (self *SQLStorage) GetTpRates(tpid, tag string) (map[string]*utils.TPRate, error) {
	rts := make(map[string]*utils.TPRate)
	q := fmt.Sprintf("SELECT id, connect_fee, rate, rate_unit, rate_increment, group_interval_start FROM %s WHERE tpid='%s' ", utils.TBL_TP_RATES, tpid)
	if tag != "" {
		q += fmt.Sprintf(" AND id='%s'", tag)
	}
	rows, err := self.Db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var tag, rate_unit, rate_increment, group_interval_start string
		var connect_fee, rate float64
		if err := rows.Scan(&tag, &connect_fee, &rate, &rate_unit, &rate_increment, &group_interval_start); err != nil {
			return nil, err
		}
		rs, err := utils.NewRateSlot(connect_fee, rate, rate_unit, rate_increment, group_interval_start)
		if err != nil {
			return nil, err
		}
		r := &utils.TPRate{
			TPid:      tpid,
			RateId:    tag,
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
	q := fmt.Sprintf("SELECT tpid,id,destinations_id,rates_id,rounding_method,rounding_decimals FROM %s WHERE tpid='%s'", utils.TBL_TP_DESTINATION_RATES, tpid)
	if tag != "" {
		q += fmt.Sprintf(" AND id='%s'", tag)
	}
	rows, err := self.Db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var tpid, tag, destinations_tag, rate_tag, rounding_method string
		var rounding_decimals int
		if err := rows.Scan(&tpid, &tag, &destinations_tag, &rate_tag, &rounding_method, &rounding_decimals); err != nil {
			return nil, err
		}

		dr := &utils.TPDestinationRate{
			TPid:              tpid,
			DestinationRateId: tag,
			DestinationRates: []*utils.DestinationRate{
				&utils.DestinationRate{
					DestinationId:    destinations_tag,
					RateId:           rate_tag,
					RoundingMethod:   rounding_method,
					RoundingDecimals: rounding_decimals,
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
		q += fmt.Sprintf(" AND id='%s'", tag)
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
	q := fmt.Sprintf("SELECT tpid, id, destrates_id, timing_id, weight FROM %s WHERE tpid='%s'", utils.TBL_TP_RATING_PLANS, tpid)
	if tag != "" {
		q += fmt.Sprintf(" AND id='%s'", tag)
	}
	rows, err := self.Db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var weight float64
		var tpid, id, destination_rates_tag, timings_tag string
		if err := rows.Scan(&tpid, &id, &destination_rates_tag, &timings_tag, &weight); err != nil {
			return nil, err
		}
		rpb := &utils.TPRatingPlanBinding{
			DestinationRatesId: destination_rates_tag,
			TimingId:           timings_tag,
			Weight:             weight,
		}
		// Logger.Debug(fmt.Sprintf("For RatingPlan id: %s, loading RatingPlanBinding: %v", tag, rpb))
		if _, exists := rpbns[id]; exists {
			rpbns[id] = append(rpbns[id], rpb)
		} else { // New
			rpbns[id] = []*utils.TPRatingPlanBinding{rpb}
		}
	}
	return rpbns, nil
}

func (self *SQLStorage) GetTpRatingProfiles(qryRpf *utils.TPRatingProfile) (map[string]*utils.TPRatingProfile, error) {
	q := fmt.Sprintf("SELECT loadid,direction,tenant,category,subject,activation_time,rating_plan_id,fallback_subjects FROM %s WHERE tpid='%s'",
		utils.TBL_TP_RATE_PROFILES, qryRpf.TPid)
	if len(qryRpf.LoadId) != 0 {
		q += fmt.Sprintf(" AND loadid='%s'", qryRpf.LoadId)
	}
	if len(qryRpf.Direction) != 0 {
		q += fmt.Sprintf(" AND direction='%s'", qryRpf.Direction)
	}
	if len(qryRpf.Tenant) != 0 {
		q += fmt.Sprintf(" AND tenant='%s'", qryRpf.Tenant)
	}
	if len(qryRpf.Category) != 0 {
		q += fmt.Sprintf(" AND category='%s'", qryRpf.Category)
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
		var rcvLoadId, tenant, category, direction, subject, fallback_subjects, rating_plan_tag, activation_time string
		if err := rows.Scan(&rcvLoadId, &direction, &tenant, &category, &subject, &activation_time, &rating_plan_tag, &fallback_subjects); err != nil {
			return nil, err
		}
		rp := &utils.TPRatingProfile{TPid: qryRpf.TPid, LoadId: rcvLoadId, Direction: direction, Tenant: tenant, Category: category, Subject: subject}
		if existingRp, has := rpfs[rp.KeyId()]; !has {
			rp.RatingPlanActivations = []*utils.TPRatingActivation{
				&utils.TPRatingActivation{ActivationTime: activation_time, RatingPlanId: rating_plan_tag, FallbackSubjects: fallback_subjects}}
			rpfs[rp.KeyId()] = rp
		} else { // Exists, update
			existingRp.RatingPlanActivations = append(existingRp.RatingPlanActivations,
				&utils.TPRatingActivation{ActivationTime: activation_time, RatingPlanId: rating_plan_tag, FallbackSubjects: fallback_subjects})
		}
	}
	return rpfs, nil
}

func (self *SQLStorage) GetTpSharedGroups(tpid, tag string) (map[string]*SharedGroup, error) {
	sgs := make(map[string]*SharedGroup)
	q := fmt.Sprintf("SELECT * FROM %s WHERE tpid='%s'", utils.TBL_TP_SHARED_GROUPS, tpid)
	if tag != "" {
		q += fmt.Sprintf(" AND id='%s'", tag)
	}
	rows, err := self.Db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var tpid, tag, account, strategy, rateSubject string
		if err := rows.Scan(&id, &tpid, &tag, &account, &strategy, &rateSubject); err != nil {
			return nil, err
		}

		sg, found := sgs[tag]
		if found {
			sg.AccountParameters[account] = &SharingParameters{
				Strategy:      strategy,
				RatingSubject: rateSubject,
			}
		} else {
			sg = &SharedGroup{
				Id: tag,
				AccountParameters: map[string]*SharingParameters{
					account: &SharingParameters{
						Strategy:      strategy,
						RatingSubject: rateSubject,
					},
				},
			}
		}
		sgs[tag] = sg

	}
	return sgs, nil
}

func (self *SQLStorage) GetTpLCRs(tpid, tag string) (map[string]*LCR, error) {
	lcrs := make(map[string]*LCR)
	q := fmt.Sprintf("SELECT * FROM %s WHERE tpid='%s'", utils.TBL_TP_LCRS, tpid)
	if tag != "" {
		q += fmt.Sprintf(" AND id='%s'", tag)
	}
	rows, err := self.Db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var tpid, direction, tenant, customer, destinationId, category, strategy, suppliers, activationTimeString string
		var weight float64
		if err := rows.Scan(&id, &tpid, &direction, &tenant, &customer, &destinationId, &category, &strategy, &suppliers, &activationTimeString, &weight); err != nil {
			return nil, err
		}
		tag := fmt.Sprintf("%s:%s:%s", direction, tenant, customer)
		lcr, found := lcrs[tag]
		activationTime, _ := utils.ParseTimeDetectLayout(activationTimeString)
		if !found {
			lcr = &LCR{
				Direction: direction,
				Tenant:    tenant,
				Customer:  customer,
			}
		}
		var act *LCRActivation
		for _, existingAct := range lcr.Activations {
			if existingAct.ActivationTime.Equal(activationTime) {
				act = existingAct
				break
			}
		}
		if act == nil {
			act = &LCRActivation{
				ActivationTime: activationTime,
			}
			lcr.Activations = append(lcr.Activations, act)
		}
		act.Entries = append(act.Entries, &LCREntry{
			DestinationId: destinationId,
			Category:      category,
			Strategy:      strategy,
			Suppliers:     suppliers,
			Weight:        weight,
		})
		lcrs[tag] = lcr
	}
	return lcrs, nil
}

func (self *SQLStorage) GetTpActions(tpid, tag string) (map[string][]*utils.TPAction, error) {
	as := make(map[string][]*utils.TPAction)
	q := fmt.Sprintf("SELECT * FROM %s WHERE tpid='%s'", utils.TBL_TP_ACTIONS, tpid)
	if tag != "" {
		q += fmt.Sprintf(" AND id='%s'", tag)
	}
	rows, err := self.Db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var units, balance_weight, weight float64
		var tpid, tag, action, balance_type, direction, destinations_tag, rating_subject, shared_group, extra_parameters, expirationDate string
		if err := rows.Scan(&id, &tpid, &tag, &action, &balance_type, &direction, &units, &expirationDate, &destinations_tag, &rating_subject, &shared_group, &balance_weight, &extra_parameters, &weight); err != nil {
			return nil, err
		}
		a := &utils.TPAction{
			Identifier:      action,
			BalanceType:     balance_type,
			Direction:       direction,
			Units:           units,
			ExpiryTime:      expirationDate,
			DestinationId:   destinations_tag,
			RatingSubject:   rating_subject,
			SharedGroup:     shared_group,
			BalanceWeight:   balance_weight,
			ExtraParameters: extra_parameters,
			Weight:          weight,
		}
		as[tag] = append(as[tag], a)
	}
	return as, nil
}

func (self *SQLStorage) GetTpActionTriggers(tpid, tag string) (map[string][]*utils.TPActionTrigger, error) {
	ats := make(map[string][]*utils.TPActionTrigger)
	q := fmt.Sprintf("SELECT tpid,id,balance_type,direction,threshold_type,threshold_value,recurrent,destination_id,actions_id,weight FROM %s WHERE tpid='%s'",
		utils.TBL_TP_ACTION_TRIGGERS, tpid)
	if tag != "" {
		q += fmt.Sprintf(" AND id='%s'", tag)
	}
	rows, err := self.Db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var threshold, weight float64
		var tpid, tag, balances_type, direction, destinations_tag, actions_tag, thresholdType string
		var recurrent bool
		if err := rows.Scan(&tpid, &tag, &balances_type, &direction, &thresholdType, &threshold, &recurrent, &destinations_tag, &actions_tag, &weight); err != nil {
			return nil, err
		}

		at := &utils.TPActionTrigger{
			BalanceType:    balances_type,
			Direction:      direction,
			ThresholdType:  thresholdType,
			ThresholdValue: threshold,
			Recurrent:      recurrent,
			DestinationId:  destinations_tag,
			ActionsId:      actions_tag,
			Weight:         weight,
		}
		ats[tag] = append(ats[tag], at)
	}
	return ats, nil
}

func (self *SQLStorage) GetTpAccountActions(aaFltr *utils.TPAccountActions) (map[string]*utils.TPAccountActions, error) {
	q := fmt.Sprintf("SELECT loadid, tenant, account, direction, action_plan_id, action_triggers_id FROM %s WHERE tpid='%s'", utils.TBL_TP_ACCOUNT_ACTIONS, aaFltr.TPid)
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
		var aaLoadId, tenant, account, direction, action_plan_tag, action_triggers_tag string
		if err := rows.Scan(&aaLoadId, &tenant, &account, &direction, &action_plan_tag, &action_triggers_tag); err != nil {
			return nil, err
		}
		aacts := &utils.TPAccountActions{
			TPid:             aaFltr.TPid,
			LoadId:           aaLoadId,
			Tenant:           tenant,
			Account:          account,
			Direction:        direction,
			ActionPlanId:     action_plan_tag,
			ActionTriggersId: action_triggers_tag,
		}
		aa[aacts.KeyId()] = aacts
	}
	return aa, nil
}
