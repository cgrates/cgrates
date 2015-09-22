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
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

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

func (self *SQLStorage) Flush(placeholder string) (err error) {
	return utils.ErrNotImplemented
}

func (self *SQLStorage) GetKeysForPrefix(prefix string) ([]string, error) {
	return nil, utils.ErrNotImplemented
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
func (self *SQLStorage) GetTpIds() ([]string, error) {
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
func (self *SQLStorage) GetTpTableIds(tpid, table string, distinct utils.TPDistinctIds, filters map[string]string, pagination *utils.Paginator) ([]string, error) {

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

func (self *SQLStorage) RemTpData(table, tpid string, args ...string) error {
	tx := self.db.Begin()
	if len(table) == 0 { // Remove tpid out of all tables
		for _, tblName := range []string{utils.TBL_TP_TIMINGS, utils.TBL_TP_DESTINATIONS, utils.TBL_TP_RATES, utils.TBL_TP_DESTINATION_RATES, utils.TBL_TP_RATING_PLANS, utils.TBL_TP_RATE_PROFILES,
			utils.TBL_TP_SHARED_GROUPS, utils.TBL_TP_CDR_STATS, utils.TBL_TP_LCRS, utils.TBL_TP_ACTIONS, utils.TBL_TP_ACTION_PLANS, utils.TBL_TP_ACTION_TRIGGERS, utils.TBL_TP_ACCOUNT_ACTIONS, utils.TBL_TP_DERIVED_CHARGERS, utils.TBL_TP_ALIASES} {
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

func (self *SQLStorage) SetTpTimings(timings []TpTiming) error {
	if len(timings) == 0 {
		return nil
	}
	m := make(map[string]bool)

	tx := self.db.Begin()
	for _, timing := range timings {
		if found, _ := m[timing.Tag]; !found {
			m[timing.Tag] = true
			if err := tx.Where(&TpTiming{Tpid: timing.Tpid, Tag: timing.Tag}).Delete(TpTiming{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		save := tx.Save(&timing)
		if save.Error != nil {
			tx.Rollback()
			return save.Error
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTpDestinations(dests []TpDestination) error {
	if len(dests) == 0 {
		return nil
	}
	m := make(map[string]bool)

	tx := self.db.Begin()
	for _, dest := range dests {
		if found, _ := m[dest.Tag]; !found {
			m[dest.Tag] = true
			if err := tx.Where(&TpDestination{Tpid: dest.Tpid, Tag: dest.Tag}).Delete(TpDestination{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		save := tx.Save(&dest)
		if save.Error != nil {
			tx.Rollback()
			return save.Error
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTpRates(rs []TpRate) error {
	if len(rs) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)

	tx := self.db.Begin()
	for _, rate := range rs {
		if found, _ := m[rate.Tag]; !found {
			m[rate.Tag] = true
			if err := tx.Where(&TpRate{Tpid: rate.Tpid, Tag: rate.Tag}).Delete(TpRate{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		save := tx.Save(&rate)
		if save.Error != nil {
			tx.Rollback()
			return save.Error
		}

	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTpDestinationRates(drs []TpDestinationRate) error {
	if len(drs) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)

	tx := self.db.Begin()
	for _, dRate := range drs {
		if found, _ := m[dRate.Tag]; !found {
			m[dRate.Tag] = true
			if err := tx.Where(&TpDestinationRate{Tpid: dRate.Tpid, Tag: dRate.Tag}).Delete(TpDestinationRate{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}

		saved := tx.Save(&dRate)
		if saved.Error != nil {
			tx.Rollback()
			return saved.Error
		}

	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTpRatingPlans(drts []TpRatingPlan) error {
	if len(drts) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)

	tx := self.db.Begin()
	for _, rPlan := range drts {
		if found, _ := m[rPlan.Tag]; !found {
			m[rPlan.Tag] = true
			if err := tx.Where(&TpRatingPlan{Tpid: rPlan.Tpid, Tag: rPlan.Tag}).Delete(TpRatingPlan{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		saved := tx.Save(&rPlan)
		if saved.Error != nil {
			tx.Rollback()
			return saved.Error
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTpRatingProfiles(rpfs []TpRatingProfile) error {
	if len(rpfs) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)

	tx := self.db.Begin()
	for _, rpf := range rpfs {
		if found, _ := m[rpf.GetRatingProfileId()]; !found {
			m[rpf.GetRatingProfileId()] = true
			if err := tx.Where(&TpRatingProfile{Tpid: rpf.Tpid, Loadid: rpf.Loadid, Direction: rpf.Direction, Tenant: rpf.Tenant, Category: rpf.Category, Subject: rpf.Subject}).Delete(TpRatingProfile{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		saved := tx.Save(&rpf)
		if saved.Error != nil {
			tx.Rollback()
			return saved.Error
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTpSharedGroups(sgs []TpSharedGroup) error {
	if len(sgs) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)

	tx := self.db.Begin()
	for _, sGroup := range sgs {
		if found, _ := m[sGroup.Tag]; !found {
			m[sGroup.Tag] = true
			if err := tx.Where(&TpSharedGroup{Tpid: sGroup.Tpid, Tag: sGroup.Tag}).Delete(TpSharedGroup{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		saved := tx.Save(&sGroup)
		if saved.Error != nil {
			tx.Rollback()
			return saved.Error
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTpCdrStats(css []TpCdrstat) error {
	if len(css) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)

	tx := self.db.Begin()
	for _, cStat := range css {
		if found, _ := m[cStat.Tag]; !found {
			m[cStat.Tag] = true
			if err := tx.Where(&TpCdrstat{Tpid: cStat.Tpid, Tag: cStat.Tag}).Delete(TpCdrstat{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		saved := tx.Save(&cStat)
		if saved.Error != nil {
			tx.Rollback()
			return saved.Error
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTpDerivedChargers(sgs []TpDerivedCharger) error {
	if len(sgs) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)

	tx := self.db.Begin()
	for _, dCharger := range sgs {
		if found, _ := m[dCharger.GetDerivedChargersId()]; !found {
			m[dCharger.GetDerivedChargersId()] = true
			tmpDc := &TpDerivedCharger{}
			if err := tmpDc.SetDerivedChargersId(dCharger.GetDerivedChargersId()); err != nil {
				tx.Rollback()
				return err
			}

			if err := tx.Where(tmpDc).Delete(TpDerivedCharger{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		if err := tx.Save(&dCharger).Error; err != nil {
			tx.Rollback()
			return err
		}

	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTpLCRs(lcrs []TpLcrRule) error {
	if len(lcrs) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)

	tx := self.db.Begin()
	for _, lcr := range lcrs {
		if found, _ := m[lcr.GetLcrRuleId()]; !found {
			m[lcr.GetLcrRuleId()] = true

			if err := tx.Where(&TpLcrRule{
				Tpid:      lcr.Tpid,
				Direction: lcr.Direction,
				Tenant:    lcr.Tenant,
				Category:  lcr.Category,
				Account:   lcr.Account,
				Subject:   lcr.Subject,
			}).Delete(TpLcrRule{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		if err := tx.Save(&lcr).Error; err != nil {
			tx.Rollback()
			return err
		}

	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTpActions(acts []TpAction) error {
	if len(acts) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)

	tx := self.db.Begin()
	for _, a := range acts {
		if found, _ := m[a.Tag]; !found {
			m[a.Tag] = true
			if err := tx.Where(&TpAction{Tpid: a.Tpid, Tag: a.Tag}).Delete(TpAction{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		saved := tx.Save(&a)
		if saved.Error != nil {
			tx.Rollback()
			return saved.Error
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTpActionPlans(ats []TpActionPlan) error {
	if len(ats) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)

	tx := self.db.Begin()
	for _, aPlan := range ats {
		if found, _ := m[aPlan.Tag]; !found {
			m[aPlan.Tag] = true
			if err := tx.Where(&TpActionPlan{Tpid: aPlan.Tpid, Tag: aPlan.Tag}).Delete(TpActionPlan{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		saved := tx.Save(&aPlan)
		if saved.Error != nil {
			tx.Rollback()
			return saved.Error
		}
	}
	r := tx.Commit()
	return r.Error
}

func (self *SQLStorage) SetTpActionTriggers(ats []TpActionTrigger) error {
	if len(ats) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)

	tx := self.db.Begin()
	for _, aTrigger := range ats {
		if found, _ := m[aTrigger.Tag]; !found {
			m[aTrigger.Tag] = true
			if err := tx.Where(&TpActionTrigger{Tpid: aTrigger.Tpid, Tag: aTrigger.Tag}).Delete(TpActionTrigger{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		saved := tx.Save(&aTrigger)
		if saved.Error != nil {
			tx.Rollback()
			return saved.Error
		}

	}
	tx.Commit()
	return nil
}

// Sets a group of account actions. Map key has the role of grouping within a tpid
func (self *SQLStorage) SetTpAccountActions(aas []TpAccountAction) error {
	if len(aas) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)

	tx := self.db.Begin()
	for _, aa := range aas {
		if found, _ := m[aa.GetAccountActionId()]; !found {
			m[aa.GetAccountActionId()] = true
			if err := tx.Where(&TpAccountAction{Tpid: aa.Tpid, Loadid: aa.Loadid, Direction: aa.Direction, Tenant: aa.Tenant, Account: aa.Account}).Delete(TpAccountAction{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		saved := tx.Save(&aa)
		if saved.Error != nil {
			tx.Rollback()
			return saved.Error
		}
	}
	tx.Commit()
	return nil

}

func (self *SQLStorage) LogCallCost(cgrid, source, runid string, cc *CallCost) error {
	return utils.ErrNotImplemented
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
func (self *SQLStorage) LogActionPlan(source string, at *ActionPlan, as Actions) (err error) {
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
	return utils.ErrNotImplemented
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
	if qryFltr.Unscoped {
		q = q.Unscoped()
	} else {
		// Query filter
		for _, tblName := range []string{utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_EXTRA, utils.TBL_COST_DETAILS, utils.TBL_RATED_CDRS} {
			q = q.Where(fmt.Sprintf("(%s.deleted_at IS NULL OR %s.deleted_at <= '0001-01-02')", tblName, tblName)) // Soft deletes
		}
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
	if qryFltr.MinUsage != nil {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".usage >= ?", qryFltr.MinUsage)
	}
	if qryFltr.MaxUsage != nil {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".usage < ?", qryFltr.MaxUsage)
	}
	if qryFltr.MinPdd != nil {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".pdd >= ?", qryFltr.MinPdd)
	}
	if qryFltr.MaxPdd != nil {
		tblName := utils.TBL_CDRS_PRIMARY
		if qryFltr.FilterOnRated {
			tblName = utils.TBL_RATED_CDRS
		}
		q = q.Where(tblName+".pdd < ?", qryFltr.MaxPdd)
	}

	if qryFltr.MinCost != nil {
		if qryFltr.MaxCost == nil {
			q = q.Where(utils.TBL_RATED_CDRS+".cost >= ?", *qryFltr.MinCost)
		} else if *qryFltr.MinCost == 0.0 && *qryFltr.MaxCost == -1.0 { // Special case when we want to skip errors
			q = q.Where(fmt.Sprintf("( %s.cost IS NULL OR %s.cost >= 0.0 )", utils.TBL_RATED_CDRS, utils.TBL_RATED_CDRS))
		} else {
			q = q.Where(utils.TBL_RATED_CDRS+".cost >= ?", *qryFltr.MinCost)
			q = q.Where(utils.TBL_RATED_CDRS+".cost < ?", *qryFltr.MaxCost)
		}
	} else if qryFltr.MaxCost != nil {
		if *qryFltr.MaxCost == -1.0 { // Non-rated CDRs
			q = q.Where(utils.TBL_RATED_CDRS + ".cost IS NULL") // Need to include it otherwise all CDRs will be returned
		} else { // Above limited CDRs, since MinCost is empty, make sure we query also NULL cost
			q = q.Where(fmt.Sprintf("( %s.cost IS NULL OR %s.cost < %f )", utils.TBL_RATED_CDRS, utils.TBL_RATED_CDRS, *qryFltr.MaxCost))
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

func (self *SQLStorage) GetTpDestinations(tpid, tag string) ([]TpDestination, error) {
	var tpDests []TpDestination
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("tag = ?", tag)
	}
	if err := q.Find(&tpDests).Error; err != nil {
		return nil, err
	}

	return tpDests, nil
}

func (self *SQLStorage) GetTpRates(tpid, tag string) ([]TpRate, error) {
	var tpRates []TpRate
	q := self.db.Where("tpid = ?", tpid).Order("id")
	if len(tag) != 0 {
		q = q.Where("tag = ?", tag)
	}
	if err := q.Find(&tpRates).Error; err != nil {
		return nil, err
	}
	return tpRates, nil
}

func (self *SQLStorage) GetTpDestinationRates(tpid, tag string, pagination *utils.Paginator) ([]TpDestinationRate, error) {
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

	return tpDestinationRates, nil
}

func (self *SQLStorage) GetTpTimings(tpid, tag string) ([]TpTiming, error) {
	var tpTimings []TpTiming
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("tag = ?", tag)
	}
	if err := q.Find(&tpTimings).Error; err != nil {
		return nil, err
	}
	return tpTimings, nil
}

func (self *SQLStorage) GetTpRatingPlans(tpid, tag string, pagination *utils.Paginator) ([]TpRatingPlan, error) {
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

	return tpRatingPlans, nil
}

func (self *SQLStorage) GetTpRatingProfiles(filter *TpRatingProfile) ([]TpRatingProfile, error) {
	var tpRpfs []TpRatingProfile
	q := self.db.Where("tpid = ?", filter.Tpid)
	if len(filter.Direction) != 0 {
		q = q.Where("direction = ?", filter.Direction)
	}
	if len(filter.Tenant) != 0 {
		q = q.Where("tenant = ?", filter.Tenant)
	}
	if len(filter.Category) != 0 {
		q = q.Where("category = ?", filter.Category)
	}
	if len(filter.Subject) != 0 {
		q = q.Where("subject = ?", filter.Subject)
	}
	if len(filter.Loadid) != 0 {
		q = q.Where("loadid = ?", filter.Loadid)
	}
	if err := q.Find(&tpRpfs).Error; err != nil {
		return nil, err
	}

	return tpRpfs, nil
}

func (self *SQLStorage) GetTpSharedGroups(tpid, tag string) ([]TpSharedGroup, error) {
	var tpShareGroups []TpSharedGroup
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("tag = ?", tag)
	}
	if err := q.Find(&tpShareGroups).Error; err != nil {
		return nil, err
	}
	return tpShareGroups, nil

}

func (self *SQLStorage) GetTpLCRs(filter *TpLcrRule) ([]TpLcrRule, error) {
	var tpLcrRule []TpLcrRule
	q := self.db.Where("tpid = ?", filter.Tpid)
	if len(filter.Direction) != 0 {
		q = q.Where("direction = ?", filter.Direction)
	}
	if len(filter.Tenant) != 0 {
		q = q.Where("tenant = ?", filter.Tenant)
	}
	if len(filter.Category) != 0 {
		q = q.Where("category = ?", filter.Category)
	}
	if len(filter.Account) != 0 {
		q = q.Where("account = ?", filter.Account)
	}
	if len(filter.Subject) != 0 {
		q = q.Where("subject = ?", filter.Subject)
	}

	if err := q.Find(&tpLcrRule).Error; err != nil {
		return nil, err
	}

	return tpLcrRule, nil
}

func (self *SQLStorage) GetTpActions(tpid, tag string) ([]TpAction, error) {
	var tpActions []TpAction
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("tag = ?", tag)
	}
	if err := q.Find(&tpActions).Error; err != nil {
		return nil, err
	}

	return tpActions, nil
}

func (self *SQLStorage) GetTpActionTriggers(tpid, tag string) ([]TpActionTrigger, error) {
	var tpActionTriggers []TpActionTrigger
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("tag = ?", tag)
	}
	if err := q.Find(&tpActionTriggers).Error; err != nil {
		return nil, err
	}

	return tpActionTriggers, nil
}

func (self *SQLStorage) GetTpActionPlans(tpid, tag string) ([]TpActionPlan, error) {
	var tpActionPlans []TpActionPlan
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("tag = ?", tag)
	}
	if err := q.Find(&tpActionPlans).Error; err != nil {
		return nil, err
	}

	return tpActionPlans, nil
}

func (self *SQLStorage) GetTpAccountActions(filter *TpAccountAction) ([]TpAccountAction, error) {

	var tpAccActs []TpAccountAction
	q := self.db.Where("tpid = ?", filter.Tpid)
	if len(filter.Direction) != 0 {
		q = q.Where("direction = ?", filter.Direction)
	}
	if len(filter.Tenant) != 0 {
		q = q.Where("tenant = ?", filter.Tenant)
	}
	if len(filter.Account) != 0 {
		q = q.Where("account = ?", filter.Account)
	}
	if len(filter.Loadid) != 0 {
		q = q.Where("loadid = ?", filter.Loadid)
	}
	if err := q.Find(&tpAccActs).Error; err != nil {
		return nil, err
	}
	return tpAccActs, nil
}

func (self *SQLStorage) GetTpDerivedChargers(filter *TpDerivedCharger) ([]TpDerivedCharger, error) {
	var tpDerivedChargers []TpDerivedCharger
	q := self.db.Where("tpid = ?", filter.Tpid)
	if len(filter.Direction) != 0 {
		q = q.Where("direction = ?", filter.Direction)
	}
	if len(filter.Tenant) != 0 {
		q = q.Where("tenant = ?", filter.Tenant)
	}
	if len(filter.Account) != 0 {
		q = q.Where("account = ?", filter.Account)
	}
	if len(filter.Category) != 0 {
		q = q.Where("category = ?", filter.Category)
	}
	if len(filter.Subject) != 0 {
		q = q.Where("subject = ?", filter.Subject)
	}
	if len(filter.Loadid) != 0 {
		q = q.Where("loadid = ?", filter.Loadid)
	}
	if err := q.Find(&tpDerivedChargers).Error; err != nil {
		return nil, err
	}
	return tpDerivedChargers, nil
}

func (self *SQLStorage) GetTpCdrStats(tpid, tag string) ([]TpCdrstat, error) {
	var tpCdrStats []TpCdrstat
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("tag = ?", tag)
	}
	if err := q.Find(&tpCdrStats).Error; err != nil {
		return nil, err
	}

	return tpCdrStats, nil
}

func (self *SQLStorage) SetTpUsers(users []TpUser) error {
	if len(users) == 0 {
		return nil
	}
	m := make(map[string]bool)

	tx := self.db.Begin()
	for _, user := range users {
		if found, _ := m[user.GetId()]; !found {
			m[user.GetId()] = true
			if err := tx.Where(&TpUser{Tpid: user.Tpid, Tenant: user.Tenant, UserName: user.UserName}).Delete(TpUser{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		save := tx.Save(&user)
		if save.Error != nil {
			tx.Rollback()
			return save.Error
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) GetTpUsers(filter *TpUser) ([]TpUser, error) {
	var tpUsers []TpUser
	q := self.db.Where("tpid = ?", filter.Tpid)
	if len(filter.Tenant) != 0 {
		q = q.Where("tenant = ?", filter.Tenant)
	}
	if len(filter.UserName) != 0 {
		q = q.Where("user_name = ?", filter.UserName)
	}
	if err := q.Find(&tpUsers).Error; err != nil {
		return nil, err
	}

	return tpUsers, nil
}

func (self *SQLStorage) SetTpAliases(aliases []TpAlias) error {
	if len(aliases) == 0 {
		return nil
	}
	m := make(map[string]bool)

	tx := self.db.Begin()
	for _, alias := range aliases {
		if found, _ := m[alias.GetId()]; !found {
			m[alias.GetId()] = true
			if err := tx.Where(&TpAlias{
				Tpid:      alias.Tpid,
				Direction: alias.Direction,
				Tenant:    alias.Tenant,
				Category:  alias.Category,
				Account:   alias.Account,
				Subject:   alias.Subject,
				Context:   alias.Context,
			}).Delete(TpAlias{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		save := tx.Save(&alias)
		if save.Error != nil {
			tx.Rollback()
			return save.Error
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) GetTpAliases(filter *TpAlias) ([]TpAlias, error) {
	var tpAliases []TpAlias
	q := self.db.Where("tpid = ?", filter.Tpid)
	if len(filter.Direction) != 0 {
		q = q.Where("direction = ?", filter.Direction)
	}
	if len(filter.Tenant) != 0 {
		q = q.Where("tenant = ?", filter.Tenant)
	}
	if len(filter.Category) != 0 {
		q = q.Where("category = ?", filter.Category)
	}
	if len(filter.Account) != 0 {
		q = q.Where("account = ?", filter.Account)
	}
	if len(filter.Subject) != 0 {
		q = q.Where("subject = ?", filter.Subject)
	}
	if len(filter.Context) != 0 {
		q = q.Where("context = ?", filter.Context)
	}

	if err := q.Find(&tpAliases).Error; err != nil {
		return nil, err
	}

	return tpAliases, nil
}
