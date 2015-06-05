/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

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
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/go-sql-driver/mysql"
	//"github.com/jinzhu/gorm"
	"github.com/cgrates/gorm"
)

type SQLStorage struct {
	Db *sql.DB
	db gorm.DB
}

func (self *SQLStorage) Close() {
	self.Db.Close()
	self.db.Close()
}

func (self *SQLStorage) Flush(placeholder string) (err error) {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (self *SQLStorage) GetKeysForPrefix(prefix string) ([]string, error) {
	return nil, nil
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

// ToDo: TEST
func (self *SQLStorage) GetTPTableIds(tpid, table string, distinct utils.TPDistinctIds, filters map[string]string, pagination *utils.Paginator) ([]string, error) {

	qry := fmt.Sprintf("SELECT DISTINCT %s FROM %s where tpid='%s'", distinct, table, tpid)
	for key, value := range filters {
		if key != "" && value != "" {
			qry += fmt.Sprintf(" AND %s='%s'", key, value)
		}
	}
	if pagination != nil {
		if len(pagination.SearchTerm) != 0 {
			qry += fmt.Sprintf(" AND (%s LIKE '%%%s%%'", distinct[0], pagination.SearchTerm)
			for _, d := range distinct[1:] {
				qry += fmt.Sprintf(" OR %s LIKE '%%%s%%'", d, pagination.SearchTerm)
			}
			qry += fmt.Sprintf(")")
		}
		if pagination.Limit != nil { // Keep Postgres compatibility by adding offset only when limit defined
			qry += fmt.Sprintf(" LIMIT %d", *pagination.Limit)
			if pagination.Offset != nil {
				qry += fmt.Sprintf(" OFFSET %d", *pagination.Offset)
			}
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

func (self *SQLStorage) SetTPTiming(tm *utils.ApierTPTiming) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (self *SQLStorage) RemTPData(table, tpid string, args ...string) error {
	tx := self.db.Begin()
	if len(table) == 0 { // Remove tpid out of all tables
		for _, tblName := range []string{utils.TBL_TP_TIMINGS, utils.TBL_TP_DESTINATIONS, utils.TBL_TP_RATES, utils.TBL_TP_DESTINATION_RATES, utils.TBL_TP_RATING_PLANS, utils.TBL_TP_RATE_PROFILES,
			utils.TBL_TP_SHARED_GROUPS, utils.TBL_TP_CDR_STATS, utils.TBL_TP_LCRS, utils.TBL_TP_ACTIONS, utils.TBL_TP_ACTION_PLANS, utils.TBL_TP_ACTION_TRIGGERS, utils.TBL_TP_ACCOUNT_ACTIONS, utils.TBL_TP_DERIVED_CHARGERS} {
			if err := tx.Table(tblName).Where("tpid = ?", tpid).Delete(nil).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		tx.Commit()
		return nil
	}
	// Remove from a single table
	tx = tx.Table(table).Where("tpid = ?", tpid)
	switch table {
	default:
		tx = tx.Where("tag = ?", args[0])
	case utils.TBL_TP_RATE_PROFILES:
		tx = tx.Where("loadid = ?", args[0]).Where("direction = ?", args[1]).Where("tenant = ?", args[2]).Where("category = ?", args[3]).Where("subject = ?", args[4])
	case utils.TBL_TP_ACCOUNT_ACTIONS:
		tx = tx.Where("loadid = ?", args[0]).Where("direction = ?", args[1]).Where("tenant = ?", args[2]).Where("account = ?", args[3])
	case utils.TBL_TP_DERIVED_CHARGERS:
		tx = tx.Where("loadid = ?", args[0]).Where("direction = ?", args[1]).Where("tenant = ?", args[2]).Where("category = ?", args[3]).Where("account = ?", args[4]).Where("subject = ?", args[5])
	}
	if err := tx.Delete(nil).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPDestination(tpid string, dest *Destination) error {
	if len(dest.Prefixes) == 0 {
		return nil
	}
	tx := self.db.Begin()
	if err := tx.Where(&TpDestination{Tpid: tpid, Tag: dest.Id}).Delete(TpDestination{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, prefix := range dest.Prefixes {
		save := tx.Save(&TpDestination{
			Tpid:      tpid,
			Tag:       dest.Id,
			Prefix:    prefix,
			CreatedAt: time.Now(),
		})
		if save.Error != nil {
			tx.Rollback()
			return save.Error
		}
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
		if err := tx.Where(&TpRate{Tpid: tpid, Tag: rtId}).Delete(TpRate{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, rs := range rSlots {
			save := tx.Save(&TpRate{
				Tpid:               tpid,
				Tag:                rtId,
				ConnectFee:         rs.ConnectFee,
				Rate:               rs.Rate,
				RateUnit:           rs.RateUnit,
				RateIncrement:      rs.RateIncrement,
				GroupIntervalStart: rs.GroupIntervalStart,
				CreatedAt:          time.Now(),
			})
			if save.Error != nil {
				tx.Rollback()
				return save.Error
			}

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
		if err := tx.Where(&TpDestinationRate{Tpid: tpid, Tag: drId}).Delete(TpDestinationRate{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, dr := range dRates {
			saved := tx.Save(&TpDestinationRate{
				Tpid:             tpid,
				Tag:              drId,
				DestinationsTag:  dr.DestinationId,
				RatesTag:         dr.RateId,
				RoundingMethod:   dr.RoundingMethod,
				RoundingDecimals: dr.RoundingDecimals,
				CreatedAt:        time.Now(),
			})
			if saved.Error != nil {
				tx.Rollback()
				return saved.Error
			}
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
		if err := tx.Where(&TpRatingPlan{Tpid: tpid, Tag: rpId}).Delete(TpRatingPlan{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, rp := range rPlans {
			saved := tx.Save(&TpRatingPlan{
				Tpid:         tpid,
				Tag:          rpId,
				DestratesTag: rp.DestinationRatesId,
				TimingTag:    rp.TimingId,
				Weight:       rp.Weight,
				CreatedAt:    time.Now(),
			})
			if saved.Error != nil {
				tx.Rollback()
				return saved.Error
			}
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
		if err := tx.Where(&TpRatingProfile{Tpid: tpid, Loadid: rpf.LoadId, Direction: rpf.Direction, Tenant: rpf.Tenant, Category: rpf.Category, Subject: rpf.Subject}).Delete(TpRatingProfile{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, ra := range rpf.RatingPlanActivations {
			saved := tx.Save(&TpRatingProfile{
				Tpid:             rpf.TPid,
				Loadid:           rpf.LoadId,
				Tenant:           rpf.Tenant,
				Category:         rpf.Category,
				Subject:          rpf.Subject,
				Direction:        rpf.Direction,
				ActivationTime:   ra.ActivationTime,
				RatingPlanTag:    ra.RatingPlanId,
				FallbackSubjects: ra.FallbackSubjects,
				CdrStatQueueIds:  ra.CdrStatQueueIds,
				CreatedAt:        time.Now(),
			})
			if saved.Error != nil {
				tx.Rollback()
				return saved.Error
			}
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
		if err := tx.Where(&TpSharedGroup{Tpid: tpid, Tag: sgId}).Delete(TpSharedGroup{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, sg := range sGroups {
			saved := tx.Save(&TpSharedGroup{
				Tpid:          tpid,
				Tag:           sgId,
				Account:       sg.Account,
				Strategy:      sg.Strategy,
				RatingSubject: sg.RatingSubject,
				CreatedAt:     time.Now(),
			})
			if saved.Error != nil {
				tx.Rollback()
				return saved.Error
			}
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
		if err := tx.Where(&TpCdrStat{Tpid: tpid, Tag: csId}).Delete(TpCdrStat{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, cs := range cStats {
			ql, _ := strconv.Atoi(cs.QueueLength)
			saved := tx.Save(&TpCdrStat{
				Tpid:                tpid,
				Tag:                 csId,
				QueueLength:         ql,
				TimeWindow:          cs.TimeWindow,
				Metrics:             cs.Metrics,
				SetupInterval:       cs.SetupInterval,
				Tors:                cs.TORs,
				CdrHosts:            cs.CdrHosts,
				CdrSources:          cs.CdrSources,
				ReqTypes:            cs.ReqTypes,
				Directions:          cs.Directions,
				Tenants:             cs.Tenants,
				Categories:          cs.Categories,
				Accounts:            cs.Accounts,
				Subjects:            cs.Subjects,
				DestinationPrefixes: cs.DestinationPrefixes,
				UsageInterval:       cs.UsageInterval,
				Suppliers:           cs.Suppliers,
				DisconnectCauses:    cs.DisconnectCauses,
				MediationRunids:     cs.MediationRunIds,
				RatedAccounts:       cs.RatedAccounts,
				RatedSubjects:       cs.RatedSubjects,
				CostInterval:        cs.CostInterval,
				ActionTriggers:      cs.ActionTriggers,
				CreatedAt:           time.Now(),
			})
			if saved.Error != nil {
				tx.Rollback()
				return saved.Error
			}
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
		tmpDc := &TpDerivedCharger{}
		if err := tmpDc.SetDerivedChargersId(dcId); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Where(tmpDc).Delete(TpDerivedCharger{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, dc := range dChargers {
			newDc := &TpDerivedCharger{
				Tpid:             tpid,
				Runid:            dc.RunId,
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
				SupplierField:    dc.SupplierField,
				CreatedAt:        time.Now(),
			}
			if err := newDc.SetDerivedChargersId(dcId); err != nil {
				tx.Rollback()
				return err
			}
			if err := tx.Save(newDc).Error; err != nil {
				tx.Rollback()
				return err
			}
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
	buffer.WriteString(fmt.Sprintf("INSERT INTO %s (tpid,direction,tenant,customer,destination_tag,category,strategy,suppliers,activation_time,weight) VALUES ", utils.TBL_TP_LCRS))
	i := 0
	for _, lcr := range lcrs {
		for _, act := range lcr.Activations {
			for _, entry := range act.Entries {
				if i != 0 { //Consecutive values after the first will be prefixed with "," as separator
					buffer.WriteRune(',')
				}
				buffer.WriteString(fmt.Sprintf("('%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%v', '%v')",
					tpid, lcr.Tenant, lcr.Category, lcr.Direction, lcr.Account, lcr.Subject, entry.DestinationId, entry.RPCategory, entry.Strategy, entry.RPCategory, act.ActivationTime, entry.Weight))
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
		if err := tx.Where(&TpAction{Tpid: tpid, Tag: acId}).Delete(TpAction{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, ac := range acs {
			saved := tx.Save(&TpAction{
				Tpid:            tpid,
				Tag:             acId,
				Action:          ac.Identifier,
				BalanceTag:      ac.BalanceId,
				BalanceType:     ac.BalanceType,
				Direction:       ac.Direction,
				Units:           ac.Units,
				ExpiryTime:      ac.ExpiryTime,
				TimingTags:      ac.TimingTags,
				DestinationTags: ac.DestinationIds,
				RatingSubject:   ac.RatingSubject,
				Category:        ac.Category,
				SharedGroup:     ac.SharedGroup,
				BalanceWeight:   ac.BalanceWeight,
				ExtraParameters: ac.ExtraParameters,
				Weight:          ac.Weight,
				CreatedAt:       time.Now(),
			})
			if saved.Error != nil {
				tx.Rollback()
				return saved.Error
			}
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) GetTPActions(tpid, actsId string) (*utils.TPActions, error) {
	acts := &utils.TPActions{TPid: tpid, ActionsId: actsId}
	var tpActions []*TpAction
	if err := self.db.Where(&TpAction{Tpid: tpid, Tag: actsId}).Find(&tpActions).Error; err != nil {
		return nil, err
	}
	for _, tpAct := range tpActions {
		acts.Actions = append(acts.Actions, &utils.TPAction{
			Identifier:      tpAct.Action,
			BalanceType:     tpAct.BalanceType,
			Direction:       tpAct.Direction,
			Units:           tpAct.Units,
			ExpiryTime:      tpAct.ExpiryTime,
			TimingTags:      tpAct.TimingTags,
			DestinationIds:  tpAct.DestinationTags,
			RatingSubject:   tpAct.RatingSubject,
			Category:        tpAct.Category,
			BalanceWeight:   tpAct.BalanceWeight,
			SharedGroup:     tpAct.SharedGroup,
			ExtraParameters: tpAct.ExtraParameters,
			Weight:          tpAct.Weight})
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
		if err := tx.Where(&TpActionPlan{Tpid: tpid, Tag: apId}).Delete(TpActionPlan{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, ap := range aPlans {
			saved := tx.Save(&TpActionPlan{
				Tpid:       tpid,
				Tag:        apId,
				ActionsTag: ap.ActionsId,
				TimingTag:  ap.TimingId,
				Weight:     ap.Weight,
				CreatedAt:  time.Now(),
			})
			if saved.Error != nil {
				tx.Rollback()
				return saved.Error
			}
		}
	}
	r := tx.Commit()
	return r.Error
}

func (self *SQLStorage) GetTPActionTimings(tpid, tag string) (map[string][]*utils.TPActionTiming, error) {
	ats := make(map[string][]*utils.TPActionTiming)
	var tpActionPlans []TpActionPlan
	if err := self.db.Where(&TpActionPlan{Tpid: tpid, Tag: tag}).Find(&tpActionPlans).Error; err != nil {
		return nil, err
	}
	for _, tpAp := range tpActionPlans {
		ats[tpAp.Tag] = append(ats[tpAp.Tag], &utils.TPActionTiming{ActionsId: tpAp.ActionsTag, TimingId: tpAp.TimingTag, Weight: tpAp.Weight})
	}
	return ats, nil
}

func (self *SQLStorage) SetTPActionTriggers(tpid string, ats map[string][]*utils.TPActionTrigger) error {
	if len(ats) == 0 {
		return nil //Nothing to set
	}
	tx := self.db.Begin()
	for atId, aTriggers := range ats {
		if err := tx.Where(&TpActionTrigger{Tpid: tpid, Tag: atId}).Delete(TpActionTrigger{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, at := range aTriggers {
			id := at.Id
			if id == "" {
				id = utils.GenUUID()
			}
			saved := tx.Save(&TpActionTrigger{
				Tpid:                   tpid,
				UniqueId:               id,
				Tag:                    atId,
				ThresholdType:          at.ThresholdType,
				ThresholdValue:         at.ThresholdValue,
				Recurrent:              at.Recurrent,
				MinSleep:               at.MinSleep,
				BalanceTag:             at.BalanceId,
				BalanceType:            at.BalanceType,
				BalanceDirection:       at.BalanceDirection,
				BalanceDestinationTags: at.BalanceDestinationIds,
				BalanceWeight:          at.BalanceWeight,
				BalanceExpiryTime:      at.BalanceExpirationDate,
				BalanceTimingTags:      at.BalanceTimingTags,
				BalanceRatingSubject:   at.BalanceRatingSubject,
				BalanceCategory:        at.BalanceCategory,
				BalanceSharedGroup:     at.BalanceSharedGroup,
				MinQueuedItems:         at.MinQueuedItems,
				ActionsTag:             at.ActionsId,
				Weight:                 at.Weight,
				CreatedAt:              time.Now(),
			})
			if saved.Error != nil {
				tx.Rollback()
				return saved.Error
			}
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
		if err := tx.Where(&TpAccountAction{Tpid: tpid, Loadid: aa.LoadId, Direction: aa.Direction, Tenant: aa.Tenant, Account: aa.Account}).Delete(TpAccountAction{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		saved := tx.Save(&TpAccountAction{
			Tpid:              aa.TPid,
			Loadid:            aa.LoadId,
			Tenant:            aa.Tenant,
			Account:           aa.Account,
			Direction:         aa.Direction,
			ActionPlanTag:     aa.ActionPlanId,
			ActionTriggersTag: aa.ActionTriggersId,
			CreatedAt:         time.Now(),
		})
		if saved.Error != nil {
			tx.Rollback()
			return saved.Error
		}
	}
	tx.Commit()
	return nil

}

func (self *SQLStorage) LogCallCost(cgrid, source, runid string, cc *CallCost) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (self *SQLStorage) GetCallCostLog(cgrid, source, runid string) (*CallCost, error) {
	var tpCostDetail TblCostDetail
	if err := self.db.Where(&TblCostDetail{Cgrid: cgrid, Runid: runid, CostSource: source}).First(&tpCostDetail).Error; err != nil {
		return nil, err
	}
	if len(tpCostDetail.Timespans) == 0 {
		return nil, nil // No costs returned
	}
	cc := new(CallCost)
	cc.TOR = tpCostDetail.Tor
	cc.Direction = tpCostDetail.Direction
	cc.Category = tpCostDetail.Category
	cc.Tenant = tpCostDetail.Tenant
	cc.Account = tpCostDetail.Account
	cc.Subject = tpCostDetail.Subject
	cc.Destination = tpCostDetail.Destination
	cc.Cost = tpCostDetail.Cost
	if err := json.Unmarshal([]byte(tpCostDetail.Timespans), &cc.Timespans); err != nil {
		return nil, err
	}
	return cc, nil
}

func (self *SQLStorage) LogActionTrigger(ubId, source string, at *ActionTrigger, as Actions) (err error) {
	return
}
func (self *SQLStorage) LogActionTiming(source string, at *ActionTiming, as Actions) (err error) {
	return
}
func (self *SQLStorage) LogError(uuid, source, runid, errstr string) (err error) { return }

func (self *SQLStorage) SetCdr(cdr *StoredCdr) error {
	extraFields, err := json.Marshal(cdr.ExtraFields)
	if err != nil {
		return err
	}
	tx := self.db.Begin()
	saved := tx.Save(&TblCdrsPrimary{
		Cgrid:           cdr.CgrId,
		Tor:             cdr.TOR,
		Accid:           cdr.AccId,
		Cdrhost:         cdr.CdrHost,
		Cdrsource:       cdr.CdrSource,
		Reqtype:         cdr.ReqType,
		Direction:       cdr.Direction,
		Tenant:          cdr.Tenant,
		Category:        cdr.Category,
		Account:         cdr.Account,
		Subject:         cdr.Subject,
		Destination:     cdr.Destination,
		SetupTime:       cdr.SetupTime,
		AnswerTime:      cdr.AnswerTime,
		Usage:           cdr.Usage.Seconds(),
		Pdd:             cdr.Pdd.Seconds(),
		Supplier:        cdr.Supplier,
		DisconnectCause: cdr.DisconnectCause,
		CreatedAt:       time.Now()})
	if saved.Error != nil {
		tx.Rollback()
		return saved.Error
	}
	// Save extra fields
	if err := tx.Save(&TblCdrsExtra{Cgrid: cdr.CgrId, ExtraFields: string(extraFields), CreatedAt: time.Now()}).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetRatedCdr(storedCdr *StoredCdr) error {
	return errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (self *SQLStorage) GetStoredCdrs(qryFltr *utils.CdrsFilter) ([]*StoredCdr, int64, error) {
	var cdrs []*StoredCdr
	// Select string
	var selectStr string
	if qryFltr.FilterOnRated { // We use different tables to query account data in case of derived
		selectStr = fmt.Sprintf("%s.cgrid,%s.id,%s.tor,%s.accid,%s.cdrhost,%s.cdrsource,%s.reqtype,%s.direction,%s.tenant,%s.category,%s.account,%s.subject,%s.destination,%s.setup_time,%s.answer_time,%s.usage,%s.pdd,%s.supplier,%s.disconnect_cause,%s.extra_fields,%s.runid,%s.cost,%s.tor,%s.direction,%s.tenant,%s.category,%s.account,%s.subject,%s.destination,%s.cost,%s.timespans",
			utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY, utils.TBL_RATED_CDRS, utils.TBL_RATED_CDRS, utils.TBL_RATED_CDRS,
			utils.TBL_RATED_CDRS, utils.TBL_RATED_CDRS, utils.TBL_RATED_CDRS, utils.TBL_RATED_CDRS, utils.TBL_RATED_CDRS, utils.TBL_RATED_CDRS, utils.TBL_RATED_CDRS, utils.TBL_RATED_CDRS, utils.TBL_RATED_CDRS, utils.TBL_RATED_CDRS,
			utils.TBL_CDRS_EXTRA, utils.TBL_RATED_CDRS, utils.TBL_RATED_CDRS, utils.TBL_COST_DETAILS, utils.TBL_COST_DETAILS, utils.TBL_COST_DETAILS, utils.TBL_COST_DETAILS, utils.TBL_COST_DETAILS,
			utils.TBL_COST_DETAILS, utils.TBL_COST_DETAILS, utils.TBL_COST_DETAILS, utils.TBL_COST_DETAILS)
	} else {
		selectStr = fmt.Sprintf("%s.cgrid,%s.id,%s.tor,%s.accid,%s.cdrhost,%s.cdrsource,%s.reqtype,%s.direction,%s.tenant,%s.category,%s.account,%s.subject,%s.destination,%s.setup_time,%s.answer_time,%s.usage,%s.pdd,%s.supplier,%s.disconnect_cause,%s.extra_fields,%s.runid,%s.cost,%s.tor,%s.direction,%s.tenant,%s.category,%s.account,%s.subject,%s.destination,%s.cost,%s.timespans",
			utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_EXTRA, utils.TBL_RATED_CDRS, utils.TBL_RATED_CDRS, utils.TBL_COST_DETAILS, utils.TBL_COST_DETAILS, utils.TBL_COST_DETAILS, utils.TBL_COST_DETAILS, utils.TBL_COST_DETAILS,
			utils.TBL_COST_DETAILS, utils.TBL_COST_DETAILS, utils.TBL_COST_DETAILS, utils.TBL_COST_DETAILS)

	}
	// Join string
	joinStr := fmt.Sprintf("LEFT JOIN %s ON %s.cgrid=%s.cgrid LEFT JOIN %s ON %s.cgrid=%s.cgrid LEFT JOIN %s ON %s.cgrid=%s.cgrid AND %s.runid=%s.runid", utils.TBL_CDRS_EXTRA, utils.TBL_CDRS_PRIMARY,
		utils.TBL_CDRS_EXTRA, utils.TBL_RATED_CDRS, utils.TBL_CDRS_PRIMARY, utils.TBL_RATED_CDRS, utils.TBL_COST_DETAILS, utils.TBL_RATED_CDRS, utils.TBL_COST_DETAILS, utils.TBL_RATED_CDRS, utils.TBL_COST_DETAILS)
	q := self.db.Table(utils.TBL_CDRS_PRIMARY).Select(selectStr).Joins(joinStr)
	// Query filter
	for _, tblName := range []string{utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_EXTRA, utils.TBL_COST_DETAILS, utils.TBL_RATED_CDRS} {
		q = q.Where(fmt.Sprintf("(%s.deleted_at IS NULL OR %s.deleted_at <= '0001-01-02')", tblName, tblName)) // Soft deletes
	}
	// Add filters, use in to replace the high number of ORs
	if len(qryFltr.CgrIds) != 0 {
		q = q.Where(utils.TBL_CDRS_PRIMARY+".cgrid in (?)", qryFltr.CgrIds)
	}
	if len(qryFltr.NotCgrIds) != 0 {
		q = q.Where(utils.TBL_CDRS_PRIMARY+".cgrid not in (?)", qryFltr.NotCgrIds)
	}
	if len(qryFltr.RunIds) != 0 {
		q = q.Where(utils.TBL_RATED_CDRS+".runid in (?)", qryFltr.RunIds)
	}
	if len(qryFltr.NotRunIds) != 0 {
		q = q.Where(utils.TBL_RATED_CDRS+".runid not in (?)", qryFltr.NotRunIds)
	}
	if len(qryFltr.Tors) != 0 {
		q = q.Where(utils.TBL_CDRS_PRIMARY+".tor in (?)", qryFltr.Tors)
	}
	if len(qryFltr.NotTors) != 0 {
		q = q.Where(utils.TBL_CDRS_PRIMARY+".tor not in (?)", qryFltr.NotTors)
	}
	if len(qryFltr.CdrHosts) != 0 {
		q = q.Where(utils.TBL_CDRS_PRIMARY+".cdrhost in (?)", qryFltr.CdrHosts)
	}
	if len(qryFltr.NotCdrHosts) != 0 {
		q = q.Where(utils.TBL_CDRS_PRIMARY+".cdrhost not in (?)", qryFltr.NotCdrHosts)
	}
	if len(qryFltr.CdrSources) != 0 {
		q = q.Where(utils.TBL_CDRS_PRIMARY+".cdrsource in (?)", qryFltr.CdrSources)
	}
	if len(qryFltr.NotCdrSources) != 0 {
		q = q.Where(utils.TBL_CDRS_PRIMARY+".cdrsource not in (?)", qryFltr.NotCdrSources)
	}
	if len(qryFltr.ReqTypes) != 0 {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".reqtype in (?)", qryFltr.ReqTypes)
	}
	if len(qryFltr.NotReqTypes) != 0 {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".reqtype not in (?)", qryFltr.NotReqTypes)
	}
	if len(qryFltr.Directions) != 0 {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".direction in (?)", qryFltr.Directions)
	}
	if len(qryFltr.NotDirections) != 0 {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".direction not in (?)", qryFltr.NotDirections)
	}
	if len(qryFltr.Tenants) != 0 {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".tenant in (?)", qryFltr.Tenants)
	}
	if len(qryFltr.NotTenants) != 0 {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".tenant not in (?)", qryFltr.NotTenants)
	}
	if len(qryFltr.Categories) != 0 {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".category in (?)", qryFltr.Categories)
	}
	if len(qryFltr.NotCategories) != 0 {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".category not in (?)", qryFltr.NotCategories)
	}
	if len(qryFltr.Accounts) != 0 {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".account in (?)", qryFltr.Accounts)
	}
	if len(qryFltr.NotAccounts) != 0 {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".account not in (?)", qryFltr.NotAccounts)
	}
	if len(qryFltr.Subjects) != 0 {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".subject in (?)", qryFltr.Subjects)
	}
	if len(qryFltr.NotSubjects) != 0 {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".subject not in (?)", qryFltr.NotSubjects)
	}
	if len(qryFltr.DestPrefixes) != 0 { // A bit ugly but still more readable than scopes provided by gorm
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		qIds := bytes.NewBufferString("(")
		for idx, destPrefix := range qryFltr.DestPrefixes {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.destination LIKE '%s%%'", tblName, destPrefix))
		}
		qIds.WriteString(" )")
		q = q.Where(qIds.String())
	}
	if len(qryFltr.NotDestPrefixes) != 0 { // A bit ugly but still more readable than scopes provided by gorm
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		qIds := bytes.NewBufferString("(")
		for idx, destPrefix := range qryFltr.NotDestPrefixes {
			if idx != 0 {
				qIds.WriteString(" AND")
			}
			qIds.WriteString(fmt.Sprintf(" %s.destination not LIKE '%%%s%%'", tblName, destPrefix))
		}
		qIds.WriteString(" )")
		q = q.Where(qIds.String())
	}
	if len(qryFltr.Suppliers) != 0 {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".supplier in (?)", qryFltr.Subjects)
	}
	if len(qryFltr.NotSuppliers) != 0 {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".supplier not in (?)", qryFltr.NotSubjects)
	}
	if len(qryFltr.DisconnectCauses) != 0 {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".disconnect_cause in (?)", qryFltr.DisconnectCauses)
	}
	if len(qryFltr.NotDisconnectCauses) != 0 {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".disconnect_cause not in (?)", qryFltr.NotDisconnectCauses)
	}
	if len(qryFltr.RatedAccounts) != 0 {
		q = q.Where(utils.TBL_COST_DETAILS+".account in (?)", qryFltr.RatedAccounts)
	}
	if len(qryFltr.NotRatedAccounts) != 0 {
		q = q.Where(utils.TBL_COST_DETAILS+".account not in (?)", qryFltr.NotRatedAccounts)
	}
	if len(qryFltr.RatedSubjects) != 0 {
		q = q.Where(utils.TBL_COST_DETAILS+".subject in (?)", qryFltr.RatedSubjects)
	}
	if len(qryFltr.NotRatedSubjects) != 0 {
		q = q.Where(utils.TBL_COST_DETAILS+".subject not in (?)", qryFltr.NotRatedSubjects)
	}
	if len(qryFltr.Costs) != 0 {
		q = q.Where(utils.TBL_RATED_CDRS+".cost in (?)", qryFltr.Costs)
	}
	if len(qryFltr.NotCosts) != 0 {
		q = q.Where(utils.TBL_RATED_CDRS+".cost not in (?)", qryFltr.NotCosts)
	}
	if len(qryFltr.ExtraFields) != 0 { // Extra fields searches, implemented as contains in extra field
		qIds := bytes.NewBufferString("(")
		needOr := false
		for field, value := range qryFltr.ExtraFields {
			if needOr {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(` %s.extra_fields LIKE '%%"%s":"%s"%%'`, utils.TBL_CDRS_EXTRA, field, value))
			needOr = true
		}
		qIds.WriteString(" )")
		q = q.Where(qIds.String())
	}
	if len(qryFltr.NotExtraFields) != 0 { // Extra fields searches, implemented as contains in extra field
		qIds := bytes.NewBufferString("(")
		needAnd := false
		for field, value := range qryFltr.NotExtraFields {
			if needAnd {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(` %s.extra_fields LIKE '%%"%s":"%s"%%'`, utils.TBL_CDRS_EXTRA, field, value))
			needAnd = true
		}
		qIds.WriteString(" )")
		q = q.Where(qIds.String())
	}
	if qryFltr.OrderIdStart != 0 { // Keep backwards compatible by testing 0 value
		q = q.Where(utils.TBL_CDRS_PRIMARY+".id >= ?", qryFltr.OrderIdStart)
	}
	if qryFltr.OrderIdEnd != 0 {
		q = q.Where(utils.TBL_CDRS_PRIMARY+".id < ?", qryFltr.OrderIdEnd)
	}
	if qryFltr.SetupTimeStart != nil {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".setup_time >= ?", qryFltr.SetupTimeStart)
	}
	if qryFltr.SetupTimeEnd != nil {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".setup_time < ?", qryFltr.SetupTimeEnd)
	}
	if qryFltr.AnswerTimeStart != nil && !qryFltr.AnswerTimeStart.IsZero() { // With IsZero we keep backwards compatible with ApierV1
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".answer_time >= ?", qryFltr.AnswerTimeStart)
	}
	if qryFltr.AnswerTimeEnd != nil && !qryFltr.AnswerTimeEnd.IsZero() {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".answer_time < ?", qryFltr.AnswerTimeEnd)
	}
	if qryFltr.CreatedAtStart != nil && !qryFltr.CreatedAtStart.IsZero() { // With IsZero we keep backwards compatible with ApierV1
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".created_at >= ?", qryFltr.CreatedAtStart)
	}
	if qryFltr.CreatedAtEnd != nil && !qryFltr.CreatedAtEnd.IsZero() {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".created_at < ?", qryFltr.CreatedAtEnd)
	}
	if qryFltr.UpdatedAtStart != nil && !qryFltr.UpdatedAtStart.IsZero() { // With IsZero we keep backwards compatible with ApierV1
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".updated_at >= ?", qryFltr.UpdatedAtStart)
	}
	if qryFltr.UpdatedAtEnd != nil && !qryFltr.UpdatedAtEnd.IsZero() {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".updated_at < ?", qryFltr.UpdatedAtEnd)
	}
	if qryFltr.UsageStart != nil {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".usage >= ?", qryFltr.UsageStart)
	}
	if qryFltr.UsageEnd != nil {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".usage < ?", qryFltr.UsageEnd)
	}

	if qryFltr.CostStart != nil {
		if qryFltr.CostEnd == nil {
			q = q.Where(utils.TBL_RATED_CDRS+".cost >= ?", *qryFltr.CostStart)
		} else if *qryFltr.CostStart == 0.0 && *qryFltr.CostEnd == -1.0 { // Special case when we want to skip errors
			q = q.Where(fmt.Sprintf("( %s.cost IS NULL OR %s.cost >= 0.0 )", utils.TBL_RATED_CDRS, utils.TBL_RATED_CDRS))
		} else {
			q = q.Where(utils.TBL_RATED_CDRS+".cost >= ?", *qryFltr.CostStart)
			q = q.Where(utils.TBL_RATED_CDRS+".cost < ?", *qryFltr.CostEnd)
		}
	} else if qryFltr.CostEnd != nil {
		if *qryFltr.CostEnd == -1.0 { // Non-rated CDRs
			q = q.Where(utils.TBL_RATED_CDRS + ".cost IS NULL") // Need to include it otherwise all CDRs will be returned
		} else { // Above limited CDRs, since costStart is empty, make sure we query also NULL cost
			q = q.Where(fmt.Sprintf("( %s.cost IS NULL OR %s.cost < %f )", utils.TBL_RATED_CDRS, utils.TBL_RATED_CDRS, *qryFltr.CostEnd))
		}
	}
	if qryFltr.Paginator.Limit != nil {
		q = q.Limit(*qryFltr.Paginator.Limit)
	}
	if qryFltr.Paginator.Offset != nil {
		q = q.Offset(*qryFltr.Paginator.Offset)
	}
	if qryFltr.Count {
		var cnt int64
		if err := q.Count(&cnt).Error; err != nil {
			//if err := q.Debug().Count(&cnt).Error; err != nil {
			return nil, 0, err
		}
		return nil, cnt, nil
	}

	// Execute query
	rows, err := q.Rows()
	if err != nil {
		return nil, 0, err
	}
	for rows.Next() {
		var cgrid, tor, accid, cdrhost, cdrsrc, reqtype, direction, tenant, category, account, subject, destination, runid, ccTor,
			ccDirection, ccTenant, ccCategory, ccAccount, ccSubject, ccDestination, ccSupplier, ccDisconnectCause sql.NullString
		var extraFields, ccTimespansBytes []byte
		var setupTime, answerTime mysql.NullTime
		var orderid int64
		var usage, pdd, cost, ccCost sql.NullFloat64
		var extraFieldsMp map[string]string
		var ccTimespans TimeSpans
		if err := rows.Scan(&cgrid, &orderid, &tor, &accid, &cdrhost, &cdrsrc, &reqtype, &direction, &tenant, &category, &account, &subject, &destination,
			&setupTime, &answerTime, &usage, &pdd, &ccSupplier, &ccDisconnectCause,
			&extraFields, &runid, &cost, &ccTor, &ccDirection, &ccTenant, &ccCategory, &ccAccount, &ccSubject, &ccDestination, &ccCost, &ccTimespansBytes); err != nil {
			return nil, 0, err
		}
		if len(extraFields) != 0 {
			if err := json.Unmarshal(extraFields, &extraFieldsMp); err != nil {
				return nil, 0, fmt.Errorf("JSON unmarshal error for cgrid: %s, runid: %v, error: %s", cgrid.String, runid.String, err.Error())
			}
		}
		if len(ccTimespansBytes) != 0 {
			if err := json.Unmarshal(ccTimespansBytes, &ccTimespans); err != nil {
				return nil, 0, fmt.Errorf("JSON unmarshal callcost error for cgrid: %s, runid: %v, error: %s", cgrid.String, runid.String, err.Error())
			}
		}
		usageDur, _ := time.ParseDuration(strconv.FormatFloat(usage.Float64, 'f', -1, 64) + "s")
		pddDur, _ := time.ParseDuration(strconv.FormatFloat(pdd.Float64, 'f', -1, 64) + "s")
		storCdr := &StoredCdr{
			CgrId: cgrid.String, OrderId: orderid, TOR: tor.String, AccId: accid.String, CdrHost: cdrhost.String, CdrSource: cdrsrc.String, ReqType: reqtype.String,
			Direction: direction.String, Tenant: tenant.String,
			Category: category.String, Account: account.String, Subject: subject.String, Destination: destination.String,
			SetupTime: setupTime.Time, AnswerTime: answerTime.Time, Usage: usageDur, Pdd: pddDur, Supplier: ccSupplier.String, DisconnectCause: ccDisconnectCause.String,
			ExtraFields: extraFieldsMp, MediationRunId: runid.String, RatedAccount: ccAccount.String, RatedSubject: ccSubject.String, Cost: cost.Float64,
		}
		if ccTimespans != nil {
			storCdr.CostDetails = &CallCost{Direction: ccDirection.String, Category: ccCategory.String, Tenant: ccTenant.String, Subject: ccSubject.String, Account: ccAccount.String, Destination: ccDestination.String, TOR: ccTor.String,
				Cost: ccCost.Float64, Timespans: ccTimespans}
		}
		if !cost.Valid { //There was no cost provided, will fakely insert 0 if we do not handle it and reflect on re-rating
			storCdr.Cost = -1
		}
		cdrs = append(cdrs, storCdr)
	}
	return cdrs, 0, nil
}

// Remove CDR data out of all CDR tables based on their cgrid
func (self *SQLStorage) RemStoredCdrs(cgrIds []string) error {
	if len(cgrIds) == 0 {
		return nil
	}
	tx := self.db.Begin()
	for _, tblName := range []string{utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_EXTRA, utils.TBL_COST_DETAILS, utils.TBL_RATED_CDRS} {
		txI := tx.Table(tblName)
		for idx, cgrId := range cgrIds {
			if idx == 0 {
				txI = txI.Where("cgrid = ?", cgrId)
			} else {
				txI = txI.Or("cgrid = ?", cgrId)
			}
		}
		if err := txI.Update("deleted_at", time.Now()).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) GetTpDestinations(tpid, tag string) (map[string]*Destination, error) {
	dests := make(map[string]*Destination)
	var tpDests []TpDestination
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("tag = ?", tag)
	}
	if err := q.Find(&tpDests).Error; err != nil {
		return nil, err
	}

	for _, tpDest := range tpDests {
		var dest *Destination
		var found bool
		if dest, found = dests[tpDest.Tag]; !found {
			dest = &Destination{Id: tpDest.Tag}
			dests[tpDest.Tag] = dest
		}
		dest.AddPrefix(tpDest.Prefix)
	}
	return dests, nil
}

func (self *SQLStorage) GetTpRates(tpid, tag string) (map[string]*utils.TPRate, error) {
	rts := make(map[string]*utils.TPRate)
	var tpRates []TpRate
	q := self.db.Where("tpid = ?", tpid).Order("id")
	if len(tag) != 0 {
		q = q.Where("tag = ?", tag)
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
			RateId:    tr.Tag,
			RateSlots: []*utils.RateSlot{rs},
		}

		// same tag only to create rate groups
		er, exists := rts[tr.Tag]
		if exists {
			er.RateSlots = append(er.RateSlots, r.RateSlots[0])
		} else {
			rts[tr.Tag] = r
		}
	}
	return rts, nil
}

func (self *SQLStorage) GetTpDestinationRates(tpid, tag string, pagination *utils.Paginator) (map[string]*utils.TPDestinationRate, error) {
	rts := make(map[string]*utils.TPDestinationRate)
	var tpDestinationRates []TpDestinationRate
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("tag = ?", tag)
	}
	if pagination != nil {
		if pagination.Limit != nil {
			q = q.Limit(*pagination.Limit)
		}
		if pagination.Offset != nil {
			q = q.Offset(*pagination.Offset)
		}
	}
	if err := q.Find(&tpDestinationRates).Error; err != nil {
		return nil, err
	}

	for _, tpDr := range tpDestinationRates {
		dr := &utils.TPDestinationRate{
			TPid:              tpid,
			DestinationRateId: tpDr.Tag,
			DestinationRates: []*utils.DestinationRate{
				&utils.DestinationRate{
					DestinationId:    tpDr.DestinationsTag,
					RateId:           tpDr.RatesTag,
					RoundingMethod:   tpDr.RoundingMethod,
					RoundingDecimals: tpDr.RoundingDecimals,
					MaxCost:          tpDr.MaxCost,
					MaxCostStrategy:  tpDr.MaxCostStrategy,
				},
			},
		}
		existingDR, exists := rts[tpDr.Tag]
		if exists {
			existingDR.DestinationRates = append(existingDR.DestinationRates, dr.DestinationRates[0])
		} else {
			existingDR = dr
		}
		rts[tpDr.Tag] = existingDR

	}
	return rts, nil
}

func (self *SQLStorage) GetTpTimings(tpid, tag string) (map[string]*utils.ApierTPTiming, error) {
	tms := make(map[string]*utils.ApierTPTiming)
	var tpTimings []TpTiming
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("tag = ?", tag)
	}
	if err := q.Find(&tpTimings).Error; err != nil {
		return nil, err
	}
	for _, tpTm := range tpTimings {
		tms[tpTm.Tag] = &utils.ApierTPTiming{TPid: tpTm.Tpid, TimingId: tpTm.Tag, Years: tpTm.Years, Months: tpTm.Months, MonthDays: tpTm.MonthDays, WeekDays: tpTm.WeekDays, Time: tpTm.Time}
	}
	return tms, nil
}

func (self *SQLStorage) GetTpRatingPlans(tpid, tag string, pagination *utils.Paginator) (map[string][]*utils.TPRatingPlanBinding, error) {
	rpbns := make(map[string][]*utils.TPRatingPlanBinding)

	var tpRatingPlans []TpRatingPlan
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("tag = ?", tag)
	}
	if err := q.Find(&tpRatingPlans).Error; err != nil {
		return nil, err
	}
	if pagination != nil {
		if pagination.Limit != nil {
			q = q.Limit(*pagination.Limit)
		}
		if pagination.Offset != nil {
			q = q.Offset(*pagination.Offset)
		}
	}

	for _, tpRp := range tpRatingPlans {
		rpb := &utils.TPRatingPlanBinding{
			DestinationRatesId: tpRp.DestratesTag,
			TimingId:           tpRp.TimingTag,
			Weight:             tpRp.Weight,
		}
		if _, exists := rpbns[tpRp.Tag]; exists {
			rpbns[tpRp.Tag] = append(rpbns[tpRp.Tag], rpb)
		} else { // New
			rpbns[tpRp.Tag] = []*utils.TPRatingPlanBinding{rpb}
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
			RatingPlanId:     tpRpf.RatingPlanTag,
			FallbackSubjects: tpRpf.FallbackSubjects,
			CdrStatQueueIds:  tpRpf.CdrStatQueueIds,
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
		q = q.Where("tag = ?", tag)
	}
	if err := q.Find(&tpCdrStats).Error; err != nil {
		return nil, err
	}

	for _, tpSg := range tpCdrStats {
		sgs[tpSg.Tag] = append(sgs[tpSg.Tag], &utils.TPSharedGroup{
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
		q = q.Where("tag = ?", tag)
	}
	if err := q.Find(&tpCdrStats).Error; err != nil {
		return nil, err
	}

	for _, tpCs := range tpCdrStats {
		css[tpCs.Tag] = append(css[tpCs.Tag], &utils.TPCdrStat{
			QueueLength:         strconv.Itoa(tpCs.QueueLength),
			TimeWindow:          tpCs.TimeWindow,
			Metrics:             tpCs.Metrics,
			SetupInterval:       tpCs.SetupInterval,
			TORs:                tpCs.Tors,
			CdrHosts:            tpCs.CdrHosts,
			CdrSources:          tpCs.CdrSources,
			ReqTypes:            tpCs.ReqTypes,
			Directions:          tpCs.Directions,
			Tenants:             tpCs.Tenants,
			Categories:          tpCs.Categories,
			Accounts:            tpCs.Accounts,
			Subjects:            tpCs.Subjects,
			DestinationPrefixes: tpCs.DestinationPrefixes,
			UsageInterval:       tpCs.UsageInterval,
			Suppliers:           tpCs.Suppliers,
			DisconnectCauses:    tpCs.DisconnectCauses,
			MediationRunIds:     tpCs.MediationRunids,
			RatedAccounts:       tpCs.RatedAccounts,
			RatedSubjects:       tpCs.RatedSubjects,
			CostInterval:        tpCs.CostInterval,
			ActionTriggers:      tpCs.ActionTriggers,
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
			RunId:                tpDcMdl.Runid,
			RunFilters:           tpDcMdl.RunFilters,
			ReqTypeField:         tpDcMdl.ReqTypeField,
			DirectionField:       tpDcMdl.DirectionField,
			TenantField:          tpDcMdl.TenantField,
			CategoryField:        tpDcMdl.CategoryField,
			AccountField:         tpDcMdl.AccountField,
			SubjectField:         tpDcMdl.SubjectField,
			DestinationField:     tpDcMdl.DestinationField,
			SetupTimeField:       tpDcMdl.SetupTimeField,
			AnswerTimeField:      tpDcMdl.AnswerTimeField,
			UsageField:           tpDcMdl.UsageField,
			SupplierField:        tpDcMdl.SupplierField,
			DisconnectCauseField: tpDcMdl.DisconnectCauseField,
		})
	}
	return dcs, nil
}

func (self *SQLStorage) GetTpLCRs(tpid, tag string) (map[string]*LCR, error) {
	lcrs := make(map[string]*LCR)
	q := fmt.Sprintf("SELECT * FROM %s WHERE tpid='%s'", utils.TBL_TP_LCRS, tpid)
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
		var tpid, direction, tenant, category, account, subject, destinationId, rpCategory, strategy, strategyParams, suppliers, activationTimeString string
		var weight float64
		if err := rows.Scan(&id, &tpid, &direction, &tenant, &category, &account, &subject, &destinationId, &rpCategory, &strategy, &strategyParams, &suppliers, &activationTimeString, &weight); err != nil {
			return nil, err
		}
		tag := utils.LCRKey(direction, tenant, category, account, subject)
		lcr, found := lcrs[tag]
		activationTime, _ := utils.ParseTimeDetectLayout(activationTimeString)
		if !found {
			lcr = &LCR{
				Direction: direction,
				Tenant:    tenant,
				Category:  category,
				Account:   account,
				Subject:   subject,
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
			DestinationId:  destinationId,
			RPCategory:     category,
			Strategy:       strategy,
			StrategyParams: strategyParams,
			Weight:         weight,
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
		q = q.Where("tag = ?", tag)
	}
	if err := q.Find(&tpActions).Error; err != nil {
		return nil, err
	}

	for _, tpAc := range tpActions {
		a := &utils.TPAction{
			Identifier:      tpAc.Action,
			BalanceId:       tpAc.BalanceTag,
			BalanceType:     tpAc.BalanceType,
			Direction:       tpAc.Direction,
			Units:           tpAc.Units,
			ExpiryTime:      tpAc.ExpiryTime,
			TimingTags:      tpAc.TimingTags,
			DestinationIds:  tpAc.DestinationTags,
			RatingSubject:   tpAc.RatingSubject,
			Category:        tpAc.Category,
			SharedGroup:     tpAc.SharedGroup,
			BalanceWeight:   tpAc.BalanceWeight,
			ExtraParameters: tpAc.ExtraParameters,
			Weight:          tpAc.Weight,
		}
		as[tpAc.Tag] = append(as[tpAc.Tag], a)
	}
	return as, nil
}

func (self *SQLStorage) GetTpActionTriggers(tpid, tag string) (map[string][]*utils.TPActionTrigger, error) {
	ats := make(map[string][]*utils.TPActionTrigger)
	var tpActionTriggers []TpActionTrigger
	if err := self.db.Where(&TpActionTrigger{Tpid: tpid, Tag: tag}).Find(&tpActionTriggers).Error; err != nil {
		return nil, err
	}
	for _, tpAt := range tpActionTriggers {
		at := &utils.TPActionTrigger{
			Id:                    tpAt.UniqueId,
			ThresholdType:         tpAt.ThresholdType,
			ThresholdValue:        tpAt.ThresholdValue,
			Recurrent:             tpAt.Recurrent,
			MinSleep:              tpAt.MinSleep,
			BalanceId:             tpAt.BalanceTag,
			BalanceType:           tpAt.BalanceType,
			BalanceDirection:      tpAt.BalanceDirection,
			BalanceDestinationIds: tpAt.BalanceDestinationTags,
			BalanceWeight:         tpAt.BalanceWeight,
			BalanceExpirationDate: tpAt.BalanceExpiryTime,
			BalanceTimingTags:     tpAt.BalanceTimingTags,
			BalanceRatingSubject:  tpAt.BalanceRatingSubject,
			BalanceCategory:       tpAt.BalanceCategory,
			BalanceSharedGroup:    tpAt.BalanceSharedGroup,
			Weight:                tpAt.Weight,
			ActionsId:             tpAt.ActionsTag,
			MinQueuedItems:        tpAt.MinQueuedItems,
		}
		ats[tpAt.Tag] = append(ats[tpAt.Tag], at)
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
			ActionPlanId:     tpAa.ActionPlanTag,
			ActionTriggersId: tpAa.ActionTriggersTag,
		}
		aas[aacts.KeyId()] = aacts
	}
	return aas, nil
}
