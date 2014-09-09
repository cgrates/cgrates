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
	"io/ioutil"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type SQLStorage struct {
	Db *sql.DB
	db gorm.DB
}

func (self *SQLStorage) Close() {
	self.Db.Close()
	self.db.Close()
}

func (self *SQLStorage) Flush() (err error) {
	cfg := config.CgrConfig()
	for _, scriptName := range []string{CREATE_CDRS_TABLES_SQL, CREATE_TARIFFPLAN_TABLES_SQL} {
		if err := self.CreateTablesFromScript(path.Join(cfg.DataFolderPath, "storage", "mysql", scriptName)); err != nil {
			return err
		}
	}
	for _, tbl := range []string{utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_EXTRA} {
		if _, err := self.Db.Query(fmt.Sprintf("SELECT 1 from %s", tbl)); err != nil {
			return err
		}
	}
	return nil
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
		fmt.Sprintf("(SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s)",
			utils.TBL_TP_TIMINGS,
			utils.TBL_TP_DESTINATIONS,
			utils.TBL_TP_RATES,
			utils.TBL_TP_DESTINATION_RATES,
			utils.TBL_TP_RATING_PLANS,
			utils.TBL_TP_RATE_PROFILES))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make([]string, 0)
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

func (self *SQLStorage) GetTPTableIds(tpid, table string, distinct utils.TPDistinctIds, filters map[string]string, pagination *utils.TPPagination) ([]string, error) {

	qry := fmt.Sprintf("SELECT DISTINCT %s FROM %s where tpid='%s'", distinct, table, tpid)
	for key, value := range filters {
		if key != "" && value != "" {
			qry += fmt.Sprintf(" AND %s='%s'", key, value)
		}
	}
	if pagination.SearchTerm != "" {
		qry += fmt.Sprintf(" AND (%s LIKE '%%%s%%'", distinct[0], pagination.SearchTerm)
		for _, d := range distinct[1:] {
			qry += fmt.Sprintf(" OR %s LIKE '%%%s%%'", d, pagination.SearchTerm)
		}
		qry += fmt.Sprintf(")")
	}
	if pagination != nil {
		limLow, limHigh := pagination.GetLimit()
		qry += fmt.Sprintf(" LIMIT %d,%d", limLow, limHigh)
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

		cols, err := rows.Columns()            // Get the column names; remember to check err
		vals := make([]string, len(cols))      // Allocate enough values
		ints := make([]interface{}, len(cols)) // Make a slice of []interface{}
		for i := range ints {
			ints[i] = &vals[i] // Copy references into the slice
		}

		err = rows.Scan(ints...)
		if err != nil {
			return nil, err
		}
		finalId := vals[0]
		if len(vals) > 1 {
			finalId = strings.Join(vals, utils.CONCATENATED_KEY_SEP)
		}
		ids = append(ids, finalId)
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
		q = fmt.Sprintf("DELETE FROM %s WHERE tpid='%s' AND loadid='%s' AND direction='%s' AND tenant='%s' AND category='%s' AND subject='%s'",
			table, tpid, args[0], args[1], args[2], args[3], args[4])
	case utils.TBL_TP_ACCOUNT_ACTIONS:
		q = fmt.Sprintf("DELETE FROM %s WHERE tpid='%s' AND loadid='%s' AND direction='%s' AND tenant='%s' AND account='%s'",
			table, tpid, args[0], args[1], args[2], args[3])
	case utils.TBL_TP_DERIVED_CHARGERS:
		q = fmt.Sprintf("DELETE FROM %s WHERE tpid='%s' AND loadid='%s' AND direction='%s' AND tenant='%s' AND category='%s' AND account='%s' AND subject='%s'",
			table, tpid, args[0], args[1], args[2], args[3], args[4], args[5])
	}
	if _, err := self.Db.Exec(q); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) SetTPDestination(tpid string, dest *Destination) error {
	if len(dest.Prefixes) == 0 {
		return nil
	}
	tx := self.db.Begin()
	tx.Where("tpid = ?", tpid).Where("id = ?", dest.Id).Delete(TpDestination{})
	for _, prefix := range dest.Prefixes {
		tx.Save(TpDestination{
			Tpid:   tpid,
			Id:     dest.Id,
			Prefix: prefix,
		})
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPRates(tpid string, rts map[string][]*utils.RateSlot) error {
	if len(rts) == 0 {
		return nil //Nothing to set
	}
	tx := self.db.Begin()
	for rtId, rSlots := range rts {
		tx.Where("tpid = ?", tpid).Where("id = ?", rtId).Delete(TpRate{})
		for _, rs := range rSlots {
			tx.Save(TpRate{
				Tpid:               tpid,
				Id:                 rtId,
				ConnectFee:         rs.ConnectFee,
				Rate:               rs.Rate,
				RateUnit:           rs.RateUnit,
				RateIncrement:      rs.RateIncrement,
				GroupIntervalStart: rs.GroupIntervalStart,
			})
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPDestinationRates(tpid string, drs map[string][]*utils.DestinationRate) error {
	if len(drs) == 0 {
		return nil //Nothing to set
	}

	tx := self.db.Begin()
	for drId, dRates := range drs {
		tx.Where("tpid = ?", tpid).Where("id = ?", drId).Delete(TpDestinationRate{})
		for _, dr := range dRates {
			tx.Save(TpDestinationRate{
				Tpid:             tpid,
				Id:               drId,
				DestinationsId:   dr.DestinationId,
				RatesId:          dr.RateId,
				RoundingMethod:   dr.RoundingMethod,
				RoundingDecimals: dr.RoundingDecimals,
			})
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPRatingPlans(tpid string, drts map[string][]*utils.TPRatingPlanBinding) error {
	if len(drts) == 0 {
		return nil //Nothing to set
	}
	tx := self.db.Begin()
	for rpId, rPlans := range drts {
		tx.Where("tpid = ?", tpid).Where("id = ?", rpId).Delete(TpRatingPlan{})
		for _, rp := range rPlans {
			tx.Save(TpRatingPlan{
				Tpid:        tpid,
				Id:          rpId,
				DestratesId: rp.DestinationRatesId,
				TimingId:    rp.TimingId,
				Weight:      rp.Weight,
			})
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPRatingProfiles(tpid string, rpfs map[string]*utils.TPRatingProfile) error {
	if len(rpfs) == 0 {
		return nil //Nothing to set
	}
	tx := self.db.Begin()
	for _, rpf := range rpfs {
		// parse identifiers
		tx.Where("tpid = ?", tpid).
			Where("direction = ?", rpf.Direction).
			Where("tenant = ?", rpf.Tenant).
			Where("subject = ?", rpf.Subject).
			Where("category = ?", rpf.Category).
			Where("loadid = ?", rpf.LoadId).
			Delete(TpRatingProfile{})
		for _, ra := range rpf.RatingPlanActivations {
			tx.Save(TpRatingProfile{
				Tpid:             rpf.TPid,
				Loadid:           rpf.LoadId,
				Tenant:           rpf.Tenant,
				Category:         rpf.Category,
				Subject:          rpf.Subject,
				Direction:        rpf.Direction,
				ActivationTime:   ra.ActivationTime,
				RatingPlanId:     ra.RatingPlanId,
				FallbackSubjects: ra.FallbackSubjects,
			})
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPSharedGroups(tpid string, sgs map[string][]*utils.TPSharedGroup) error {
	if len(sgs) == 0 {
		return nil //Nothing to set
	}
	tx := self.db.Begin()
	for sgId, sGroups := range sgs {
		tx.Where("tpid = ?", tpid).Where("id = ?", sgId).Delete(TpSharedGroup{})
		for _, sg := range sGroups {
			tx.Save(TpSharedGroup{
				Tpid:          tpid,
				Id:            sgId,
				Account:       sg.Account,
				Strategy:      sg.Strategy,
				RatingSubject: sg.RatingSubject,
			})
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPCdrStats(tpid string, css map[string][]*utils.TPCdrStat) error {
	if len(css) == 0 {
		return nil //Nothing to set
	}
	tx := self.db.Begin()
	for csId, cStats := range css {
		tx.Where("tpid = ?", tpid).Where("id = ?", csId).Delete(TpCdrStat{})
		for _, cs := range cStats {
			tx.Save(TpCdrStat{
				Tpid:              tpid,
				Id:                csId,
				QueueLength:       cs.QueueLength,
				TimeWindow:        cs.TimeWindow,
				Metrics:           cs.Metrics,
				SetupInterval:     cs.SetupInterval,
				Tor:               cs.TOR,
				CdrHost:           cs.CdrHost,
				CdrSource:         cs.CdrSource,
				ReqType:           cs.ReqType,
				Direction:         cs.Direction,
				Tenant:            cs.Tenant,
				Category:          cs.Category,
				Account:           cs.Account,
				Subject:           cs.Subject,
				DestinationPrefix: cs.DestinationPrefix,
				UsageInterval:     cs.UsageInterval,
				MediationRunIds:   cs.MediationRunIds,
				RatedAccount:      cs.RatedAccount,
				RatedSubject:      cs.RatedSubject,
				CostInterval:      cs.CostInterval,
				ActionTriggers:    cs.ActionTriggers,
			})
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPDerivedChargers(tpid string, sgs map[string][]*utils.TPDerivedCharger) error {
	if len(sgs) == 0 {
		return nil //Nothing to set
	}
	tx := self.db.Begin()
	for dcId, dChargers := range sgs {
		// parse identifiers
		tmpDc := TpDerivedCharger{}
		if err := tmpDc.SetDerivedChargersId(dcId); err != nil {
			tx.Rollback()
			return err
		}
		tx.Where("tpid = ?", tpid).
			Where("direction = ?", tmpDc.Direction).
			Where("tenant = ?", tmpDc.Tenant).
			Where("account = ?", tmpDc.Account).
			Where("category = ?", tmpDc.Category).
			Where("subject = ?", tmpDc.Subject).
			Where("loadid = ?", tmpDc.Loadid).
			Delete(TpDerivedCharger{})
		for _, dc := range dChargers {
			newDc := TpDerivedCharger{
				Tpid:             tpid,
				RunId:            dc.RunId,
				RunFilters:       dc.RunFilters,
				ReqTypeField:     dc.ReqTypeField,
				DirectionField:   dc.DirectionField,
				TenantField:      dc.TenantField,
				CategoryField:    dc.CategoryField,
				AccountField:     dc.AccountField,
				SubjectField:     dc.SubjectField,
				DestinationField: dc.DestinationField,
				SetupTimeField:   dc.SetupTimeField,
				AnswerTimeField:  dc.AnswerTimeField,
				UsageField:       dc.UsageField,
			}
			if err := newDc.SetDerivedChargersId(dcId); err != nil {
				tx.Rollback()
				return err
			}
			tx.Save(newDc)
		}
	}
	tx.Commit()
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

	tx := self.db.Begin()
	for acId, acs := range acts {
		tx.Where("tpid = ?", tpid).Where("id = ?", acId).Delete(TpAction{})
		for _, ac := range acs {
			tx.Save(TpAction{
				Tpid:            tpid,
				Id:              acId,
				Action:          ac.Identifier,
				BalanceType:     ac.BalanceType,
				Direction:       ac.Direction,
				Units:           ac.Units,
				ExpiryTime:      ac.ExpiryTime,
				DestinationId:   ac.DestinationId,
				RatingSubject:   ac.RatingSubject,
				Category:        ac.Category,
				SharedGroup:     ac.SharedGroup,
				BalanceWeight:   ac.BalanceWeight,
				ExtraParameters: ac.ExtraParameters,
				Weight:          ac.Weight,
			})
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) GetTPActions(tpid, actsId string) (*utils.TPActions, error) {
	rows, err := self.Db.Query(fmt.Sprintf("SELECT action,balance_type,direction,units,expiry_time,destination_id,rating_subject,category,shared_group,balance_weight,extra_parameters,weight FROM %s WHERE tpid='%s' AND id='%s'", utils.TBL_TP_ACTIONS, tpid, actsId))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	acts := &utils.TPActions{TPid: tpid, ActionsId: actsId}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one result
		var action, balanceId, dir, destId, rateSubject, category, sharedGroup, expTime, extraParameters string
		var units, balanceWeight, weight float64
		if err = rows.Scan(&action, &balanceId, &dir, &units, &expTime, &destId, &rateSubject, &category, &sharedGroup, &balanceWeight, &extraParameters, &weight); err != nil {
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
			Category:        category,
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
	tx := self.db.Begin()
	for apId, aPlans := range ats {
		tx.Where("tpid = ?", tpid).Where("id = ?", apId).Delete(TpActionPlan{})
		for _, ap := range aPlans {
			tx.Save(TpActionPlan{
				Tpid:      tpid,
				Id:        apId,
				ActionsId: ap.ActionsId,
				TimingId:  ap.TimingId,
				Weight:    ap.Weight,
			})
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) GetTPActionTimings(tpid, tag string) (map[string][]*utils.TPActionTiming, error) {
	ats := make(map[string][]*utils.TPActionTiming)

	var tpActionPlans []TpActionPlan
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("id = ?", tag)
	}
	if err := q.Find(&tpActionPlans).Error; err != nil {
		return nil, err
	}

	for _, tpAp := range tpActionPlans {
		ats[tpAp.Id] = append(ats[tpAp.Id], &utils.TPActionTiming{ActionsId: tpAp.ActionsId, TimingId: tpAp.TimingId, Weight: tpAp.Weight})
	}
	return ats, nil
}

func (self *SQLStorage) SetTPActionTriggers(tpid string, ats map[string][]*utils.TPActionTrigger) error {
	if len(ats) == 0 {
		return nil //Nothing to set
	}
	tx := self.db.Begin()
	for atId, aTriggers := range ats {
		tx.Where("tpid = ?", tpid).Where("id = ?", atId).Delete(TpActionTrigger{})
		for _, at := range aTriggers {
			recurrent := 0
			if at.Recurrent {
				recurrent = 1
			}
			tx.Save(TpActionTrigger{
				Tpid:                 tpid,
				Id:                   atId,
				BalanceType:          at.BalanceType,
				Direction:            at.Direction,
				ThresholdType:        at.ThresholdType,
				ThresholdValue:       at.ThresholdValue,
				Recurrent:            recurrent,
				MinSleep:             int64(at.MinSleep),
				DestinationId:        at.DestinationId,
				BalanceWeight:        at.BalanceWeight,
				BalanceExpiryTime:    at.BalanceExpirationDate,
				BalanceRatingSubject: at.BalanceRatingSubject,
				BalanceCategory:      at.BalanceCategory,
				BalanceSharedGroup:   at.BalanceSharedGroup,
				MinQueuedItems:       at.MinQueuedItems,
				ActionsId:            at.ActionsId,
				Weight:               at.Weight,
			})
		}
	}
	tx.Commit()
	return nil
}

// Sets a group of account actions. Map key has the role of grouping within a tpid
func (self *SQLStorage) SetTPAccountActions(tpid string, aas map[string]*utils.TPAccountActions) error {
	if len(aas) == 0 {
		return nil //Nothing to set
	}
	tx := self.db.Begin()
	for _, aa := range aas {
		// parse identifiers
		tx.Where("tpid = ?", tpid).
			Where("direction = ?", aa.Direction).
			Where("tenant = ?", aa.Tenant).
			Where("account = ?", aa.Account).
			Where("loadid = ?", aa.LoadId).
			Delete(TpAccountAction{})

		tx.Save(TpAccountAction{
			Tpid:             aa.TPid,
			Loadid:           aa.LoadId,
			Tenant:           aa.Tenant,
			Account:          aa.Account,
			Direction:        aa.Direction,
			ActionPlanId:     aa.ActionPlanId,
			ActionTriggersId: aa.ActionTriggersId,
		})
	}
	tx.Commit()
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
	_, err = self.Db.Exec(fmt.Sprintf("INSERT INTO  %s (cgrid,tor,accid,cdrhost,cdrsource,reqtype,direction,tenant,category,account,subject,destination,setup_time,answer_time,`usage`) VALUES ('%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s', %v)",
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
		cdr.Usage.Seconds(),
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
	_, err = self.Db.Exec(fmt.Sprintf("INSERT INTO %s (mediation_time,cgrid,runid,reqtype,direction,tenant,category,account,subject,destination,setup_time,answer_time,`usage`,cost,extra_info) VALUES (now(),'%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s',%v,%f,'%s') ON DUPLICATE KEY UPDATE mediation_time=now(),reqtype=values(reqtype),direction=values(direction),tenant=values(tenant),category=values(category),account=values(account),subject=values(subject),destination=values(destination),setup_time=values(setup_time),answer_time=values(answer_time),`usage`=values(`usage`),cost=values(cost),extra_info=values(extra_info)",
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
		storedCdr.Usage.Seconds(),
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
func (self *SQLStorage) GetStoredCdrs(cgrIds, runIds, tors, cdrHosts, cdrSources, reqTypes, directions, tenants, categories, accounts, subjects, destPrefixes, ratedAccounts, ratedSubjects []string,
	orderIdStart, orderIdEnd int64, timeStart, timeEnd time.Time, ignoreErr, ignoreRated, ignoreDerived bool) ([]*utils.StoredCdr, error) {
	var cdrs []*utils.StoredCdr
	var q *bytes.Buffer // Need to query differently since in case of primary, unmediated CDRs some values will be missing
	if ignoreDerived {
		q = bytes.NewBufferString(fmt.Sprintf("SELECT %s.cgrid,%s.tbid,%s.tor,%s.accid,%s.cdrhost,%s.cdrsource,%s.reqtype,%s.direction,%s.tenant,%s.category,%s.account,%s.subject,%s.destination,%s.setup_time,%s.answer_time,%s.`usage`,%s.extra_fields,%s.runid,%s.account,%s.subject,%s.cost FROM %s LEFT JOIN %s ON %s.cgrid=%s.cgrid LEFT JOIN %s ON %s.cgrid=%s.cgrid LEFT JOIN %s ON %s.cgrid=%s.cgrid AND %s.runid=%s.runid",
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
			utils.TBL_COST_DETAILS,
			utils.TBL_COST_DETAILS,
			utils.TBL_RATED_CDRS,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_EXTRA,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_EXTRA,
			utils.TBL_RATED_CDRS,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_RATED_CDRS,
			utils.TBL_COST_DETAILS,
			utils.TBL_RATED_CDRS,
			utils.TBL_COST_DETAILS,
			utils.TBL_RATED_CDRS,
			utils.TBL_COST_DETAILS))
	} else {
		q = bytes.NewBufferString(fmt.Sprintf("SELECT %s.cgrid,%s.tbid,%s.tor,%s.accid,%s.cdrhost,%s.cdrsource,%s.reqtype,%s.direction,%s.tenant,%s.category,%s.account,%s.subject,%s.destination,%s.setup_time,%s.answer_time,%s.`usage`,%s.extra_fields,%s.runid,%s.account,%s.subject,%s.cost FROM %s LEFT JOIN %s ON %s.cgrid=%s.cgrid LEFT JOIN %s ON %s.cgrid=%s.cgrid LEFT JOIN %s ON %s.cgrid=%s.cgrid AND %s.runid=%s.runid",
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
			utils.TBL_COST_DETAILS,
			utils.TBL_COST_DETAILS,
			utils.TBL_RATED_CDRS,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_EXTRA,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_EXTRA,
			utils.TBL_RATED_CDRS,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_RATED_CDRS,
			utils.TBL_COST_DETAILS,
			utils.TBL_RATED_CDRS,
			utils.TBL_COST_DETAILS,
			utils.TBL_RATED_CDRS,
			utils.TBL_COST_DETAILS))
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
	if len(ratedAccounts) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, ratedAccount := range ratedAccounts {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.account='%s'", utils.TBL_COST_DETAILS, ratedAccount))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(ratedSubjects) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, ratedSubject := range ratedSubjects {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.subject='%s'", utils.TBL_COST_DETAILS, ratedSubject))
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
		var cgrid, tor, accid, cdrhost, cdrsrc, reqtype, direction, tenant, category, account, subject, destination, runid, ratedAccount, ratedSubject sql.NullString
		var extraFields []byte
		var setupTime, answerTime mysql.NullTime
		var orderid int64
		var usage, cost sql.NullFloat64
		var extraFieldsMp map[string]string
		if err := rows.Scan(&cgrid, &orderid, &tor, &accid, &cdrhost, &cdrsrc, &reqtype, &direction, &tenant, &category, &account, &subject, &destination, &setupTime, &answerTime, &usage,
			&extraFields, &runid, &ratedAccount, &ratedSubject, &cost); err != nil {
			return nil, err
		}
		if len(extraFields) != 0 {
			if err := json.Unmarshal(extraFields, &extraFieldsMp); err != nil {
				return nil, fmt.Errorf("JSON unmarshal error for cgrid: %s, runid: %v, error: %s", cgrid.String, runid.String, err.Error())
			}
		}
		usageDur, _ := time.ParseDuration(strconv.FormatFloat(usage.Float64, 'f', -1, 64) + "s")
		storCdr := &utils.StoredCdr{
			CgrId: cgrid.String, OrderId: orderid, TOR: tor.String, AccId: accid.String, CdrHost: cdrhost.String, CdrSource: cdrsrc.String, ReqType: reqtype.String,
			Direction: direction.String, Tenant: tenant.String,
			Category: category.String, Account: account.String, Subject: subject.String, Destination: destination.String,
			SetupTime: setupTime.Time, AnswerTime: answerTime.Time, Usage: usageDur,
			ExtraFields: extraFieldsMp, MediationRunId: runid.String, RatedAccount: ratedAccount.String, RatedSubject: ratedSubject.String, Cost: cost.Float64,
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

func (self *SQLStorage) GetTpDestinations(tpid, tag string) (map[string]*Destination, error) {
	dests := make(map[string]*Destination)
	var tpDests []TpDestination
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("id = ?", tag)
	}
	if err := q.Find(&tpDests).Error; err != nil {
		return nil, err
	}

	for _, tpDest := range tpDests {
		var dest *Destination
		var found bool
		if dest, found = dests[tpDest.Id]; !found {
			dest = &Destination{Id: tpDest.Id}
			dests[tpDest.Id] = dest
		}
		dest.AddPrefix(tpDest.Prefix)
	}
	return dests, nil
}

func (self *SQLStorage) GetTpRates(tpid, tag string) (map[string]*utils.TPRate, error) {
	rts := make(map[string]*utils.TPRate)
	var tpRates []TpRate
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("id = ?", tag)
	}
	if err := q.Find(&tpRates).Error; err != nil {
		return nil, err
	}

	for _, tr := range tpRates {
		rs, err := utils.NewRateSlot(tr.ConnectFee, tr.Rate, tr.RateUnit, tr.RateIncrement, tr.GroupIntervalStart)
		if err != nil {
			return nil, err
		}
		r := &utils.TPRate{
			TPid:      tpid,
			RateId:    tr.Id,
			RateSlots: []*utils.RateSlot{rs},
		}

		// same tag only to create rate groups
		er, exists := rts[tr.Id]
		if exists {
			if err := ValidNextGroup(er.RateSlots[len(er.RateSlots)-1], r.RateSlots[0]); err != nil {
				return nil, err
			}
			er.RateSlots = append(er.RateSlots, r.RateSlots[0])
		} else {
			rts[tr.Id] = r
		}
	}
	return rts, nil
}

func (self *SQLStorage) GetTpDestinationRates(tpid, tag string) (map[string]*utils.TPDestinationRate, error) {
	rts := make(map[string]*utils.TPDestinationRate)
	var tpDestinationRates []TpDestinationRate
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("id = ?", tag)
	}
	if err := q.Find(&tpDestinationRates).Error; err != nil {
		return nil, err
	}

	for _, tpDr := range tpDestinationRates {
		dr := &utils.TPDestinationRate{
			TPid:              tpid,
			DestinationRateId: tpDr.Id,
			DestinationRates: []*utils.DestinationRate{
				&utils.DestinationRate{
					DestinationId:    tpDr.DestinationsId,
					RateId:           tpDr.RatesId,
					RoundingMethod:   tpDr.RoundingMethod,
					RoundingDecimals: tpDr.RoundingDecimals,
				},
			},
		}
		existingDR, exists := rts[tpDr.Id]
		if exists {
			existingDR.DestinationRates = append(existingDR.DestinationRates, dr.DestinationRates[0])
		} else {
			existingDR = dr
		}
		rts[tpDr.Id] = existingDR

	}
	return rts, nil
}

func (self *SQLStorage) GetTpTimings(tpid, tag string) (map[string]*utils.TPTiming, error) {
	tms := make(map[string]*utils.TPTiming)
	var tpTimings []TpTiming
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("id = ?", tag)
	}
	if err := q.Find(&tpTimings).Error; err != nil {
		return nil, err
	}

	for _, tpTm := range tpTimings {
		tms[tpTm.Id] = NewTiming(tpTm.Id, tpTm.Years, tpTm.Months, tpTm.MonthDays, tpTm.WeekDays, tpTm.Time)
	}

	return tms, nil
}

func (self *SQLStorage) GetTpRatingPlans(tpid, tag string) (map[string][]*utils.TPRatingPlanBinding, error) {
	rpbns := make(map[string][]*utils.TPRatingPlanBinding)

	var tpRatingPlans []TpRatingPlan
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("id = ?", tag)
	}
	if err := q.Find(&tpRatingPlans).Error; err != nil {
		return nil, err
	}

	for _, tpRp := range tpRatingPlans {
		rpb := &utils.TPRatingPlanBinding{
			DestinationRatesId: tpRp.DestratesId,
			TimingId:           tpRp.TimingId,
			Weight:             tpRp.Weight,
		}
		if _, exists := rpbns[tpRp.Id]; exists {
			rpbns[tpRp.Id] = append(rpbns[tpRp.Id], rpb)
		} else { // New
			rpbns[tpRp.Id] = []*utils.TPRatingPlanBinding{rpb}
		}
	}
	return rpbns, nil
}

func (self *SQLStorage) GetTpRatingProfiles(qryRpf *utils.TPRatingProfile) (map[string]*utils.TPRatingProfile, error) {

	rpfs := make(map[string]*utils.TPRatingProfile)
	var tpRpfs []TpRatingProfile
	q := self.db.Where("tpid = ?", qryRpf.TPid)
	if len(qryRpf.Direction) != 0 {
		q = q.Where("direction = ?", qryRpf.Direction)
	}
	if len(qryRpf.Tenant) != 0 {
		q = q.Where("tenant = ?", qryRpf.Tenant)
	}
	if len(qryRpf.Category) != 0 {
		q = q.Where("category = ?", qryRpf.Category)
	}
	if len(qryRpf.Subject) != 0 {
		q = q.Where("subject = ?", qryRpf.Subject)
	}
	if len(qryRpf.LoadId) != 0 {
		q = q.Where("loadid = ?", qryRpf.LoadId)
	}
	if err := q.Find(&tpRpfs).Error; err != nil {
		return nil, err
	}
	for _, tpRpf := range tpRpfs {

		rp := &utils.TPRatingProfile{
			TPid:      tpRpf.Tpid,
			LoadId:    tpRpf.Loadid,
			Direction: tpRpf.Direction,
			Tenant:    tpRpf.Tenant,
			Category:  tpRpf.Category,
			Subject:   tpRpf.Subject,
		}
		ra := &utils.TPRatingActivation{
			ActivationTime:   tpRpf.ActivationTime,
			RatingPlanId:     tpRpf.RatingPlanId,
			FallbackSubjects: tpRpf.FallbackSubjects,
		}
		if existingRpf, exists := rpfs[rp.KeyId()]; !exists {
			rp.RatingPlanActivations = []*utils.TPRatingActivation{ra}
			rpfs[rp.KeyId()] = rp
		} else { // Exists, update
			existingRpf.RatingPlanActivations = append(existingRpf.RatingPlanActivations, ra)
		}

	}
	return rpfs, nil
}

func (self *SQLStorage) GetTpSharedGroups(tpid, tag string) (map[string][]*utils.TPSharedGroup, error) {
	sgs := make(map[string][]*utils.TPSharedGroup)

	var tpCdrStats []TpSharedGroup
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("id = ?", tag)
	}
	if err := q.Find(&tpCdrStats).Error; err != nil {
		return nil, err
	}

	for _, tpSg := range tpCdrStats {
		sgs[tag] = append(sgs[tpSg.Id], &utils.TPSharedGroup{
			Account:       tpSg.Account,
			Strategy:      tpSg.Strategy,
			RatingSubject: tpSg.RatingSubject,
		})
	}
	return sgs, nil
}

func (self *SQLStorage) GetTpCdrStats(tpid, tag string) (map[string][]*utils.TPCdrStat, error) {
	css := make(map[string][]*utils.TPCdrStat)

	var tpCdrStats []TpCdrStat
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("id = ?", tag)
	}
	if err := q.Find(&tpCdrStats).Error; err != nil {
		return nil, err
	}

	for _, tpCs := range tpCdrStats {
		css[tag] = append(css[tpCs.Id], &utils.TPCdrStat{
			QueueLength:       tpCs.QueueLength,
			TimeWindow:        tpCs.TimeWindow,
			Metrics:           tpCs.Metrics,
			SetupInterval:     tpCs.SetupInterval,
			TOR:               tpCs.Tor,
			CdrHost:           tpCs.CdrHost,
			CdrSource:         tpCs.CdrSource,
			ReqType:           tpCs.ReqType,
			Direction:         tpCs.Direction,
			Tenant:            tpCs.Tenant,
			Category:          tpCs.Category,
			Account:           tpCs.Account,
			Subject:           tpCs.Subject,
			DestinationPrefix: tpCs.DestinationPrefix,
			UsageInterval:     tpCs.UsageInterval,
			MediationRunIds:   tpCs.MediationRunIds,
			RatedAccount:      tpCs.RatedAccount,
			RatedSubject:      tpCs.RatedSubject,
			CostInterval:      tpCs.CostInterval,
			ActionTriggers:    tpCs.ActionTriggers,
		})
	}
	return css, nil
}

func (self *SQLStorage) GetTpDerivedChargers(dc *utils.TPDerivedChargers) (map[string]*utils.TPDerivedChargers, error) {
	dcs := make(map[string]*utils.TPDerivedChargers)
	var tpDerivedChargers []TpDerivedCharger
	q := self.db.Where("tpid = ?", dc.TPid)
	if len(dc.Direction) != 0 {
		q = q.Where("direction = ?", dc.Direction)
	}
	if len(dc.Tenant) != 0 {
		q = q.Where("tenant = ?", dc.Tenant)
	}
	if len(dc.Account) != 0 {
		q = q.Where("account = ?", dc.Account)
	}
	if len(dc.Category) != 0 {
		q = q.Where("category = ?", dc.Category)
	}
	if len(dc.Subject) != 0 {
		q = q.Where("subject = ?", dc.Subject)
	}
	if len(dc.Loadid) != 0 {
		q = q.Where("loadid = ?", dc.Loadid)
	}
	if err := q.Find(&tpDerivedChargers).Error; err != nil {
		return nil, err
	}
	for _, tpDcMdl := range tpDerivedChargers {
		tpDc := &utils.TPDerivedChargers{TPid: tpDcMdl.Tpid, Loadid: tpDcMdl.Loadid, Direction: tpDcMdl.Direction, Tenant: tpDcMdl.Tenant, Category: tpDcMdl.Category,
			Account: tpDcMdl.Account, Subject: tpDcMdl.Subject}
		tag := tpDc.GetDerivedChargesId()
		if _, hasIt := dcs[tag]; !hasIt {
			dcs[tag] = tpDc
		}
		dcs[tag].DerivedChargers = append(dcs[tag].DerivedChargers, &utils.TPDerivedCharger{
			RunId:            tpDcMdl.RunId,
			RunFilters:       tpDcMdl.RunFilters,
			ReqTypeField:     tpDcMdl.ReqTypeField,
			DirectionField:   tpDcMdl.DirectionField,
			TenantField:      tpDcMdl.TenantField,
			CategoryField:    tpDcMdl.CategoryField,
			AccountField:     tpDcMdl.AccountField,
			SubjectField:     tpDcMdl.SubjectField,
			DestinationField: tpDcMdl.DestinationField,
			SetupTimeField:   tpDcMdl.SetupTimeField,
			AnswerTimeField:  tpDcMdl.AnswerTimeField,
			UsageField:       tpDcMdl.UsageField,
		})
	}
	return dcs, nil
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

	var tpActions []TpAction
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("id = ?", tag)
	}
	if err := q.Find(&tpActions).Error; err != nil {
		return nil, err
	}

	for _, tpAc := range tpActions {
		a := &utils.TPAction{
			Identifier:      tpAc.Action,
			BalanceType:     tpAc.BalanceType,
			Direction:       tpAc.Direction,
			Units:           tpAc.Units,
			ExpiryTime:      tpAc.ExpiryTime,
			DestinationId:   tpAc.DestinationId,
			RatingSubject:   tpAc.RatingSubject,
			Category:        tpAc.Category,
			SharedGroup:     tpAc.SharedGroup,
			BalanceWeight:   tpAc.BalanceWeight,
			ExtraParameters: tpAc.ExtraParameters,
			Weight:          tpAc.Weight,
		}
		as[tpAc.Id] = append(as[tpAc.Id], a)
	}
	return as, nil
}

func (self *SQLStorage) GetTpActionTriggers(tpid, tag string) (map[string][]*utils.TPActionTrigger, error) {
	ats := make(map[string][]*utils.TPActionTrigger)
	var tpActionTriggers []TpActionTrigger
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("id = ?", tag)
	}
	if err := q.Find(&tpActionTriggers).Error; err != nil {
		return nil, err
	}

	for _, tpAt := range tpActionTriggers {
		recurrent := tpAt.Recurrent == 1
		at := &utils.TPActionTrigger{
			BalanceType:           tpAt.BalanceType,
			Direction:             tpAt.Direction,
			ThresholdType:         tpAt.ThresholdType,
			ThresholdValue:        tpAt.ThresholdValue,
			Recurrent:             recurrent,
			MinSleep:              time.Duration(tpAt.MinSleep),
			DestinationId:         tpAt.DestinationId,
			BalanceWeight:         tpAt.BalanceWeight,
			BalanceExpirationDate: tpAt.BalanceExpiryTime,
			BalanceRatingSubject:  tpAt.BalanceRatingSubject,
			BalanceCategory:       tpAt.BalanceCategory,
			BalanceSharedGroup:    tpAt.BalanceSharedGroup,
			Weight:                tpAt.Weight,
			ActionsId:             tpAt.ActionsId,
			MinQueuedItems:        tpAt.MinQueuedItems,
		}
		ats[tpAt.Id] = append(ats[tpAt.Id], at)
	}
	return ats, nil
}

func (self *SQLStorage) GetTpAccountActions(aaFltr *utils.TPAccountActions) (map[string]*utils.TPAccountActions, error) {
	aas := make(map[string]*utils.TPAccountActions)
	var tpAccActs []TpAccountAction
	q := self.db.Where("tpid = ?", aaFltr.TPid)
	if len(aaFltr.Direction) != 0 {
		q = q.Where("direction = ?", aaFltr.Direction)
	}
	if len(aaFltr.Tenant) != 0 {
		q = q.Where("tenant = ?", aaFltr.Tenant)
	}
	if len(aaFltr.Account) != 0 {
		q = q.Where("account = ?", aaFltr.Account)
	}
	if len(aaFltr.LoadId) != 0 {
		q = q.Where("loadid = ?", aaFltr.LoadId)
	}
	if err := q.Find(&tpAccActs).Error; err != nil {
		return nil, err
	}
	for _, tpAa := range tpAccActs {
		aacts := &utils.TPAccountActions{
			TPid:             tpAa.Tpid,
			LoadId:           tpAa.Loadid,
			Tenant:           tpAa.Tenant,
			Account:          tpAa.Account,
			Direction:        tpAa.Direction,
			ActionPlanId:     tpAa.ActionPlanId,
			ActionTriggersId: tpAa.ActionTriggersId,
		}
		aas[aacts.KeyId()] = aacts
	}
	return aas, nil
}
