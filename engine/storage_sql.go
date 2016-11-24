/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/jinzhu/gorm"
)

type SQLStorage struct {
	Db *sql.DB
	db *gorm.DB
}

func (self *SQLStorage) Close() {
	self.Db.Close()
	self.db.Close()
}

func (self *SQLStorage) Flush(scriptsPath string) (err error) {
	for _, scriptName := range []string{utils.CREATE_CDRS_TABLES_SQL, utils.CREATE_TARIFFPLAN_TABLES_SQL} {
		if err := self.CreateTablesFromScript(path.Join(scriptsPath, scriptName)); err != nil {
			return err
		}
	}
	if _, err := self.Db.Query(fmt.Sprintf("SELECT 1 FROM %s", utils.TBL_CDRS)); err != nil {
		return err
	}
	return nil
}

func (self *SQLStorage) GetKeysForPrefix(prefix string) ([]string, error) {
	return nil, utils.ErrNotImplemented
}

func (ms *SQLStorage) RebuildReverseForPrefix(prefix string) error {
	return utils.ErrNotImplemented
}

func (self *SQLStorage) PreloadCacheForPrefix(prefix string) error {
	return utils.ErrNotImplemented
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

func (self *SQLStorage) RemTpData(table, tpid string, args map[string]string) error {
	tx := self.db.Begin()
	if len(table) == 0 { // Remove tpid out of all tables
		for _, tblName := range []string{utils.TBL_TP_TIMINGS, utils.TBL_TP_DESTINATIONS, utils.TBL_TP_RATES, utils.TBL_TP_DESTINATION_RATES, utils.TBL_TP_RATING_PLANS, utils.TBL_TP_RATE_PROFILES,
			utils.TBL_TP_SHARED_GROUPS, utils.TBL_TP_CDR_STATS, utils.TBL_TP_LCRS, utils.TBL_TP_ACTIONS, utils.TBL_TP_ACTION_PLANS, utils.TBL_TP_ACTION_TRIGGERS, utils.TBL_TP_ACCOUNT_ACTIONS,
			utils.TBL_TP_DERIVED_CHARGERS, utils.TBL_TP_ALIASES, utils.TBLTPResourceLimits} {
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
	// Compose filters
	for key, value := range args {
		tx = tx.Where(key+" = ?", value)
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

func (self *SQLStorage) SetTPDestinations(dests []*utils.TPDestination) error {
	if len(dests) == 0 {
		return nil
	}
	tx := self.db.Begin()
	for _, dst := range dests {
		// Remove previous
		if err := tx.Where(&TpDestination{Tpid: dst.TPid, Tag: dst.Tag}).Delete(TpDestination{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, dstPrfx := range dst.Prefixes {
			if err := tx.Save(&TpDestination{Tpid: dst.TPid, Tag: dst.Tag, Prefix: dstPrfx}).Error; err != nil {
				tx.Rollback()
				return err
			}
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

			if err := tx.Where(tmpDc).Delete(TpDerivedCharger{
				Tpid:      dCharger.Tpid,
				Direction: dCharger.Direction,
				Tenant:    dCharger.Tenant,
				Category:  dCharger.Category,
				Account:   dCharger.Account,
				Subject:   dCharger.Subject,
			}).Error; err != nil {
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
			if err := tx.Where(&TpAccountAction{Tpid: aa.Tpid, Loadid: aa.Loadid, Tenant: aa.Tenant, Account: aa.Account}).Delete(TpAccountAction{}).Error; err != nil {
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
func (self *SQLStorage) SetSMCost(smc *SMCost) error {
	if smc.CostDetails == nil {
		return nil
	}
	tss, err := json.Marshal(smc.CostDetails)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("Error marshalling timespans to json: %v", err))
		return err
	}
	tx := self.db.Begin()
	cd := &TBLSMCosts{
		Cgrid:       smc.CGRID,
		RunID:       smc.RunID,
		OriginHost:  smc.OriginHost,
		OriginID:    smc.OriginID,
		CostSource:  smc.CostSource,
		CostDetails: string(tss),
		Usage:       smc.Usage,
		CreatedAt:   time.Now(),
	}
	if tx.Save(cd).Error != nil { // Check further since error does not properly reflect duplicates here (sql: no rows in result set)
		tx.Rollback()
		return tx.Error
	}
	tx.Commit()
	return nil
}

// GetSMCosts is used to retrieve one or multiple SMCosts based on filter
func (self *SQLStorage) GetSMCosts(cgrid, runid, originHost, originIDPrefix string) ([]*SMCost, error) {
	var smCosts []*SMCost
	q := self.db.Where(&TBLSMCosts{Cgrid: cgrid, RunID: runid})
	if originIDPrefix != "" {
		q = self.db.Where(&TBLSMCosts{OriginHost: originHost, RunID: runid}).Where(fmt.Sprintf("origin_id LIKE '%s%%'", originIDPrefix))
	}
	results := make([]*TBLSMCosts, 0)
	if err := q.Find(&results).Error; err != nil {
		return nil, err
	}
	for _, result := range results {
		if len(result.CostDetails) == 0 {
			continue
		}
		smc := &SMCost{
			CGRID:       result.Cgrid,
			RunID:       result.RunID,
			OriginHost:  result.OriginHost,
			OriginID:    result.OriginID,
			CostSource:  result.CostSource,
			Usage:       result.Usage,
			CostDetails: &CallCost{},
		}
		if err := json.Unmarshal([]byte(result.CostDetails), smc.CostDetails); err != nil {
			return nil, err
		}
		smCosts = append(smCosts, smc)
	}

	return smCosts, nil
}

func (self *SQLStorage) LogActionTrigger(ubId, source string, at *ActionTrigger, as Actions) (err error) {
	return
}
func (self *SQLStorage) LogActionTiming(source string, at *ActionTiming, as Actions) (err error) {
	return
}

func (self *SQLStorage) SetCDR(cdr *CDR, allowUpdate bool) error {
	extraFields, err := json.Marshal(cdr.ExtraFields)
	if err != nil {
		return err
	}
	tx := self.db.Begin()
	saved := tx.Save(&TBLCDRs{
		Cgrid:           cdr.CGRID,
		RunID:           cdr.RunID,
		OriginHost:      cdr.OriginHost,
		Source:          cdr.Source,
		OriginID:        cdr.OriginID,
		Tor:             cdr.ToR,
		RequestType:     cdr.RequestType,
		Direction:       cdr.Direction,
		Tenant:          cdr.Tenant,
		Category:        cdr.Category,
		Account:         cdr.Account,
		Subject:         cdr.Subject,
		Destination:     cdr.Destination,
		SetupTime:       cdr.SetupTime,
		Pdd:             cdr.PDD.Seconds(),
		AnswerTime:      cdr.AnswerTime,
		Usage:           cdr.Usage.Seconds(),
		Supplier:        cdr.Supplier,
		DisconnectCause: cdr.DisconnectCause,
		ExtraFields:     string(extraFields),
		CostSource:      cdr.CostSource,
		Cost:            cdr.Cost,
		CostDetails:     cdr.CostDetailsJson(),
		AccountSummary:  utils.ToJSON(cdr.AccountSummary),
		ExtraInfo:       cdr.ExtraInfo,
		CreatedAt:       time.Now(),
	})
	if saved.Error != nil {
		tx.Rollback()
		if !allowUpdate {
			return saved.Error
		}
		tx = self.db.Begin()
		updated := tx.Model(&TBLCDRs{}).Where(&TBLCDRs{Cgrid: cdr.CGRID, RunID: cdr.RunID, OriginID: cdr.OriginID}).Updates(
			TBLCDRs{
				OriginHost:      cdr.OriginHost,
				Source:          cdr.Source,
				OriginID:        cdr.OriginID,
				Tor:             cdr.ToR,
				RequestType:     cdr.RequestType,
				Direction:       cdr.Direction,
				Tenant:          cdr.Tenant,
				Category:        cdr.Category,
				Account:         cdr.Account,
				Subject:         cdr.Subject,
				Destination:     cdr.Destination,
				SetupTime:       cdr.SetupTime,
				Pdd:             cdr.PDD.Seconds(),
				AnswerTime:      cdr.AnswerTime,
				Usage:           cdr.Usage.Seconds(),
				Supplier:        cdr.Supplier,
				DisconnectCause: cdr.DisconnectCause,
				ExtraFields:     string(extraFields),
				CostSource:      cdr.CostSource,
				Cost:            cdr.Cost,
				CostDetails:     cdr.CostDetailsJson(),
				AccountSummary:  utils.ToJSON(cdr.AccountSummary),
				ExtraInfo:       cdr.ExtraInfo,
				UpdatedAt:       time.Now(),
			},
		)
		if updated.Error != nil {
			tx.Rollback()
			return updated.Error
		}
	}
	tx.Commit()
	return nil
}

// GetCDRs has ability to remove the selected CDRs, count them or simply return them
// qryFltr.Unscoped will ignore soft deletes or delete records permanently
func (self *SQLStorage) GetCDRs(qryFltr *utils.CDRsFilter, remove bool) ([]*CDR, int64, error) {
	var cdrs []*CDR
	q := self.db.Table(utils.TBL_CDRS).Select("*")
	if qryFltr.Unscoped {
		q = q.Unscoped()
	}
	// Add filters, use in to replace the high number of ORs
	if len(qryFltr.CGRIDs) != 0 {
		q = q.Where("cgrid in (?)", qryFltr.CGRIDs)
	}
	if len(qryFltr.NotCGRIDs) != 0 {
		q = q.Where("cgrid not in (?)", qryFltr.NotCGRIDs)
	}
	if len(qryFltr.RunIDs) != 0 {
		q = q.Where("run_id in (?)", qryFltr.RunIDs)
	}
	if len(qryFltr.NotRunIDs) != 0 {
		q = q.Where("run_id not in (?)", qryFltr.NotRunIDs)
	}
	if len(qryFltr.ToRs) != 0 {
		q = q.Where("tor in (?)", qryFltr.ToRs)
	}
	if len(qryFltr.NotToRs) != 0 {
		q = q.Where("tor not in (?)", qryFltr.NotToRs)
	}
	if len(qryFltr.OriginHosts) != 0 {
		q = q.Where("origin_host in (?)", qryFltr.OriginHosts)
	}
	if len(qryFltr.NotOriginHosts) != 0 {
		q = q.Where("origin_host not in (?)", qryFltr.NotOriginHosts)
	}
	if len(qryFltr.Sources) != 0 {
		q = q.Where("source in (?)", qryFltr.Sources)
	}
	if len(qryFltr.NotSources) != 0 {
		q = q.Where("source not in (?)", qryFltr.NotSources)
	}
	if len(qryFltr.RequestTypes) != 0 {
		q = q.Where("request_type in (?)", qryFltr.RequestTypes)
	}
	if len(qryFltr.NotRequestTypes) != 0 {
		q = q.Where("request_type not in (?)", qryFltr.NotRequestTypes)
	}
	if len(qryFltr.Directions) != 0 {
		q = q.Where("direction in (?)", qryFltr.Directions)
	}
	if len(qryFltr.NotDirections) != 0 {
		q = q.Where("direction not in (?)", qryFltr.NotDirections)
	}
	if len(qryFltr.Tenants) != 0 {
		q = q.Where("tenant in (?)", qryFltr.Tenants)
	}
	if len(qryFltr.NotTenants) != 0 {
		q = q.Where("tenant not in (?)", qryFltr.NotTenants)
	}
	if len(qryFltr.Categories) != 0 {
		q = q.Where("category in (?)", qryFltr.Categories)
	}
	if len(qryFltr.NotCategories) != 0 {
		q = q.Where("category not in (?)", qryFltr.NotCategories)
	}
	if len(qryFltr.Accounts) != 0 {
		q = q.Where("account in (?)", qryFltr.Accounts)
	}
	if len(qryFltr.NotAccounts) != 0 {
		q = q.Where("account not in (?)", qryFltr.NotAccounts)
	}
	if len(qryFltr.Subjects) != 0 {
		q = q.Where("subject in (?)", qryFltr.Subjects)
	}
	if len(qryFltr.NotSubjects) != 0 {
		q = q.Where("subject not in (?)", qryFltr.NotSubjects)
	}
	if len(qryFltr.DestinationPrefixes) != 0 { // A bit ugly but still more readable than scopes provided by gorm
		qIds := bytes.NewBufferString("(")
		for idx, destPrefix := range qryFltr.DestinationPrefixes {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" destination LIKE '%s%%'", destPrefix))
		}
		qIds.WriteString(" )")
		q = q.Where(qIds.String())
	}
	if len(qryFltr.NotDestinationPrefixes) != 0 { // A bit ugly but still more readable than scopes provided by gorm
		qIds := bytes.NewBufferString("(")
		for idx, destPrefix := range qryFltr.NotDestinationPrefixes {
			if idx != 0 {
				qIds.WriteString(" AND")
			}
			qIds.WriteString(fmt.Sprintf(" destination not LIKE '%s%%'", destPrefix))
		}
		qIds.WriteString(" )")
		q = q.Where(qIds.String())
	}
	if len(qryFltr.Suppliers) != 0 {
		q = q.Where("supplier in (?)", qryFltr.Subjects)
	}
	if len(qryFltr.NotSuppliers) != 0 {
		q = q.Where("supplier not in (?)", qryFltr.NotSubjects)
	}
	if len(qryFltr.DisconnectCauses) != 0 {
		q = q.Where("disconnect_cause in (?)", qryFltr.DisconnectCauses)
	}
	if len(qryFltr.NotDisconnectCauses) != 0 {
		q = q.Where("disconnect_cause not in (?)", qryFltr.NotDisconnectCauses)
	}
	if len(qryFltr.Costs) != 0 {
		q = q.Where(utils.TBL_CDRS+".cost in (?)", qryFltr.Costs)
	}
	if len(qryFltr.NotCosts) != 0 {
		q = q.Where(utils.TBL_CDRS+".cost not in (?)", qryFltr.NotCosts)
	}
	if len(qryFltr.ExtraFields) != 0 { // Extra fields searches, implemented as contains in extra field
		qIds := bytes.NewBufferString("(")
		needOr := false
		for field, value := range qryFltr.ExtraFields {
			if needOr {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(` extra_fields LIKE '%%"%s":"%s"%%'`, field, value))
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
			qIds.WriteString(fmt.Sprintf(` extra_fields LIKE '%%"%s":"%s"%%'`, field, value))
			needAnd = true
		}
		qIds.WriteString(" )")
		q = q.Where(qIds.String())
	}
	if qryFltr.OrderIDStart != nil { // Keep backwards compatible by testing 0 value
		q = q.Where(utils.TBL_CDRS+".id >= ?", *qryFltr.OrderIDStart)
	}
	if qryFltr.OrderIDEnd != nil {
		q = q.Where(utils.TBL_CDRS+".id < ?", *qryFltr.OrderIDEnd)
	}
	if qryFltr.SetupTimeStart != nil {
		q = q.Where("setup_time >= ?", qryFltr.SetupTimeStart)
	}
	if qryFltr.SetupTimeEnd != nil {
		q = q.Where("setup_time < ?", qryFltr.SetupTimeEnd)
	}
	if qryFltr.AnswerTimeStart != nil && !qryFltr.AnswerTimeStart.IsZero() { // With IsZero we keep backwards compatible with ApierV1
		q = q.Where("answer_time >= ?", qryFltr.AnswerTimeStart)
	}
	if qryFltr.AnswerTimeEnd != nil && !qryFltr.AnswerTimeEnd.IsZero() {
		q = q.Where("answer_time < ?", qryFltr.AnswerTimeEnd)
	}
	if qryFltr.CreatedAtStart != nil && !qryFltr.CreatedAtStart.IsZero() { // With IsZero we keep backwards compatible with ApierV1
		q = q.Where("created_at >= ?", qryFltr.CreatedAtStart)
	}
	if qryFltr.CreatedAtEnd != nil && !qryFltr.CreatedAtEnd.IsZero() {
		q = q.Where("created_at < ?", qryFltr.CreatedAtEnd)
	}
	if qryFltr.UpdatedAtStart != nil && !qryFltr.UpdatedAtStart.IsZero() { // With IsZero we keep backwards compatible with ApierV1
		q = q.Where("updated_at >= ?", qryFltr.UpdatedAtStart)
	}
	if qryFltr.UpdatedAtEnd != nil && !qryFltr.UpdatedAtEnd.IsZero() {
		q = q.Where("updated_at < ?", qryFltr.UpdatedAtEnd)
	}
	if len(qryFltr.MinUsage) != 0 {
		if minUsage, err := utils.ParseDurationWithSecs(qryFltr.MinUsage); err != nil {
			return nil, 0, err
		} else {
			q = q.Where("usage >= ?", minUsage.Seconds())
		}
	}
	if len(qryFltr.MaxUsage) != 0 {
		if maxUsage, err := utils.ParseDurationWithSecs(qryFltr.MaxUsage); err != nil {
			return nil, 0, err
		} else {
			q = q.Where("usage < ?", maxUsage.Seconds())
		}

	}
	if len(qryFltr.MinPDD) != 0 {
		if minPDD, err := utils.ParseDurationWithSecs(qryFltr.MinPDD); err != nil {
			return nil, 0, err
		} else {
			q = q.Where("pdd >= ?", minPDD.Seconds())
		}

	}
	if len(qryFltr.MaxPDD) != 0 {
		if maxPDD, err := utils.ParseDurationWithSecs(qryFltr.MaxPDD); err != nil {
			return nil, 0, err
		} else {
			q = q.Where("pdd < ?", maxPDD.Seconds())
		}
	}

	if qryFltr.MinCost != nil {
		if qryFltr.MaxCost == nil {
			q = q.Where("cost >= ?", *qryFltr.MinCost)
		} else if *qryFltr.MinCost == 0.0 && *qryFltr.MaxCost == -1.0 { // Special case when we want to skip errors
			q = q.Where("( cost IS NULL OR cost >= 0.0 )")
		} else {
			q = q.Where("cost >= ?", *qryFltr.MinCost)
			q = q.Where("cost < ?", *qryFltr.MaxCost)
		}
	} else if qryFltr.MaxCost != nil {
		if *qryFltr.MaxCost == -1.0 { // Non-rated CDRs
			q = q.Where("cost IS NULL") // Need to include it otherwise all CDRs will be returned
		} else { // Above limited CDRs, since MinCost is empty, make sure we query also NULL cost
			q = q.Where(fmt.Sprintf("( cost IS NULL OR cost < %f )", *qryFltr.MaxCost))
		}
	}
	if qryFltr.Paginator.Limit != nil {
		q = q.Limit(*qryFltr.Paginator.Limit)
	}
	if qryFltr.Paginator.Offset != nil {
		q = q.Offset(*qryFltr.Paginator.Offset)
	}
	if remove { // Remove CDRs instead of querying them
		if err := q.Delete(nil).Error; err != nil {
			q.Rollback()
			return nil, 0, err
		}
	}
	if qryFltr.Count { // Count CDRs
		var cnt int64
		if err := q.Count(&cnt).Error; err != nil {
			//if err := q.Debug().Count(&cnt).Error; err != nil {
			return nil, 0, err
		}
		return nil, cnt, nil
	}
	// Execute query
	results := make([]*TBLCDRs, 0)
	q.Find(&results)

	for _, result := range results {
		extraFieldsMp := make(map[string]string)
		if result.ExtraFields != "" {
			if err := json.Unmarshal([]byte(result.ExtraFields), &extraFieldsMp); err != nil {
				return nil, 0, fmt.Errorf("JSON unmarshal error for cgrid: %s, runid: %v, error: %s", result.Cgrid, result.RunID, err.Error())
			}
		}
		var callCost CallCost
		if result.CostDetails != "" {
			if err := json.Unmarshal([]byte(result.CostDetails), &callCost); err != nil {
				return nil, 0, fmt.Errorf("JSON unmarshal callcost error for cgrid: %s, runid: %v, error: %s", result.Cgrid, result.RunID, err.Error())
			}
		}
		acntSummary, err := NewAccountSummaryFromJSON(result.AccountSummary)
		if err != nil {
			return nil, 0, fmt.Errorf("JSON unmarshal account summary error for cgrid: %s, runid: %v, error: %s", result.Cgrid, result.RunID, err.Error())
		}
		usageDur := time.Duration(result.Usage * utils.NANO_MULTIPLIER)
		pddDur := time.Duration(result.Pdd * utils.NANO_MULTIPLIER)
		storCdr := &CDR{
			CGRID:           result.Cgrid,
			RunID:           result.RunID,
			OrderID:         result.ID,
			OriginHost:      result.OriginHost,
			Source:          result.Source,
			OriginID:        result.OriginID,
			ToR:             result.Tor,
			RequestType:     result.RequestType,
			Direction:       result.Direction,
			Tenant:          result.Tenant,
			Category:        result.Category,
			Account:         result.Account,
			Subject:         result.Subject,
			Destination:     result.Destination,
			SetupTime:       result.SetupTime,
			PDD:             pddDur,
			AnswerTime:      result.AnswerTime,
			Usage:           usageDur,
			Supplier:        result.Supplier,
			DisconnectCause: result.DisconnectCause,
			ExtraFields:     extraFieldsMp,
			CostSource:      result.CostSource,
			Cost:            result.Cost,
			CostDetails:     &callCost,
			AccountSummary:  acntSummary,
			ExtraInfo:       result.ExtraInfo,
		}
		cdrs = append(cdrs, storCdr)
	}
	return cdrs, 0, nil
}

func (self *SQLStorage) GetTPDestinations(tpid, tag string) (uTPDsts []*utils.TPDestination, err error) {
	var tpDests TpDestinations
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("tag = ?", tag)
	}
	if err := q.Find(&tpDests).Error; err != nil {
		return nil, err
	}
	return tpDests.AsTPDestinations(), nil
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

func (self *SQLStorage) GetTpResourceLimits(tpid, tag string) (TpResourceLimits, error) {
	var tpResourceLimits TpResourceLimits
	q := self.db.Where("tpid = ?", tpid)
	if len(tag) != 0 {
		q = q.Where("tag = ?", tag)
	}
	if err := q.Find(&tpResourceLimits).Error; err != nil {
		return nil, err
	}
	return tpResourceLimits, nil
}
