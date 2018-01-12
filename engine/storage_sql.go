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

type SQLImpl interface {
	extraFieldsExistsQry(string) string
	extraFieldsValueQry(string, string) string
	notExtraFieldsExistsQry(string) string
	notExtraFieldsValueQry(string, string) string
}

type SQLStorage struct {
	Db *sql.DB
	db *gorm.DB
	StorDB
	SQLImpl
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
	if _, err := self.Db.Query(fmt.Sprintf("SELECT 1 FROM %s", utils.CDRsTBL)); err != nil {
		return err
	}
	if err := SetDBVersions(self); err != nil {
		return err
	}
	return nil
}

func (rs *SQLStorage) SelectDatabase(dbName string) (err error) {
	return
}

func (self *SQLStorage) GetKeysForPrefix(prefix string) ([]string, error) {
	return nil, utils.ErrNotImplemented
}

func (self *SQLStorage) RebuildReverseForPrefix(prefix string) error {
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

func (self *SQLStorage) IsDBEmpty() (resp bool, err error) {
	tbls := []string{
		utils.TBLTPTimings, utils.TBLTPDestinations, utils.TBLTPRates,
		utils.TBLTPDestinationRates, utils.TBLTPRatingPlans, utils.TBLTPRateProfiles,
		utils.TBLTPSharedGroups, utils.TBLTPCdrStats, utils.TBLTPLcrs, utils.TBLTPActions,
		utils.TBLTPActionTriggers, utils.TBLTPAccountActions, utils.TBLTPDerivedChargers, utils.TBLTPUsers,
		utils.TBLTPAliases, utils.TBLTPResources, utils.TBLTPStats, utils.TBLTPThresholds,
		utils.TBLTPFilters, utils.SMCostsTBL, utils.CDRsTBL, utils.TBLTPActionPlans,
		utils.TBLVersions, utils.TBLTPSuppliers, utils.TBLTPAttributes,
	}
	for _, tbl := range tbls {
		if self.db.HasTable(tbl) {
			return false, nil
		}

	}
	return true, nil
}

// update
// Return a list with all TPids defined in the system, even if incomplete, isolated in some table.
func (self *SQLStorage) GetTpIds(colName string) ([]string, error) {
	var rows *sql.Rows
	var err error
	qryStr := fmt.Sprintf(" (SELECT tpid FROM %s)", colName)
	if colName == "" {
		qryStr = fmt.Sprintf(
			"(SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s)",
			utils.TBLTPTimings,
			utils.TBLTPDestinations,
			utils.TBLTPRates,
			utils.TBLTPDestinationRates,
			utils.TBLTPRatingPlans,
			utils.TBLTPRateProfiles,
			utils.TBLTPSharedGroups,
			utils.TBLTPCdrStats,
			utils.TBLTPLcrs,
			utils.TBLTPActions,
			utils.TBLTPActionTriggers,
			utils.TBLTPAccountActions,
			utils.TBLTPDerivedChargers,
			utils.TBLTPUsers,
			utils.TBLTPAliases,
			utils.TBLTPResources,
			utils.TBLTPStats,
			utils.TBLTPThresholds,
			utils.TBLTPFilters,
			utils.TBLTPActionPlans,
			utils.TBLTPSuppliers,
			utils.TBLTPAttributes)
	}
	rows, err = self.Db.Query(qryStr)
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
		for _, tblName := range []string{utils.TBLTPTimings, utils.TBLTPDestinations, utils.TBLTPRates,
			utils.TBLTPDestinationRates, utils.TBLTPRatingPlans, utils.TBLTPRateProfiles, utils.TBLTPSharedGroups,
			utils.TBLTPCdrStats, utils.TBLTPLcrs, utils.TBLTPActions, utils.TBLTPActionPlans, utils.TBLTPActionTriggers,
			utils.TBLTPAccountActions, utils.TBLTPDerivedChargers, utils.TBLTPAliases, utils.TBLTPUsers,
			utils.TBLTPResources, utils.TBLTPStats, utils.TBLTPFilters, utils.TBLTPSuppliers, utils.TBLTPAttributes} {
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

func (self *SQLStorage) SetTPTimings(timings []*utils.ApierTPTiming) error {
	if len(timings) == 0 {
		return nil
	}

	tx := self.db.Begin()
	for _, timing := range timings {
		if err := tx.Where(&TpTiming{Tpid: timing.TPid, Tag: timing.ID}).Delete(TpTiming{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		t := APItoModelTiming(timing)
		if err := tx.Save(&t).Error; err != nil {
			return err
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
		if err := tx.Where(&TpDestination{Tpid: dst.TPid, Tag: dst.ID}).Delete(TpDestination{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, d := range APItoModelDestination(dst) {
			if err := tx.Save(&d).Error; err != nil {
				return err
			}
		}
		// for _, dstPrfx := range dst.Prefixes {
		// 	if err := tx.Save(&TpDestination{Tpid: dst.TPid, Tag: dst.Tag, Prefix: dstPrfx}).Error; err != nil {
		// 		tx.Rollback()
		// 		return err
		// 	}
		// }
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPRates(rs []*utils.TPRate) error {
	if len(rs) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)
	tx := self.db.Begin()
	for _, rate := range rs {
		if found, _ := m[rate.ID]; !found {
			m[rate.ID] = true
			if err := tx.Where(&TpRate{Tpid: rate.TPid, Tag: rate.ID}).Delete(TpRate{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, r := range APItoModelRate(rate) {
			if err := tx.Save(&r).Error; err != nil {
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPDestinationRates(drs []*utils.TPDestinationRate) error {
	if len(drs) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)
	tx := self.db.Begin()
	for _, dRate := range drs {
		if found, _ := m[dRate.ID]; !found {
			m[dRate.ID] = true
			if err := tx.Where(&TpDestinationRate{Tpid: dRate.TPid, Tag: dRate.ID}).Delete(TpDestinationRate{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, d := range APItoModelDestinationRate(dRate) {
			if err := tx.Save(&d).Error; err != nil {
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPRatingPlans(rps []*utils.TPRatingPlan) error {
	if len(rps) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)
	tx := self.db.Begin()
	for _, rPlan := range rps {
		if found, _ := m[rPlan.ID]; !found {
			m[rPlan.ID] = true
			if err := tx.Where(&TpRatingPlan{Tpid: rPlan.TPid, Tag: rPlan.ID}).Delete(TpRatingPlan{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, r := range APItoModelRatingPlan(rPlan) {
			if err := tx.Save(&r).Error; err != nil {
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPRatingProfiles(rpfs []*utils.TPRatingProfile) error {
	if len(rpfs) == 0 {
		return nil //Nothing to set
	}
	tx := self.db.Begin()
	for _, rpf := range rpfs {
		if err := tx.Where(&TpRatingProfile{Tpid: rpf.TPid, Loadid: rpf.LoadId, Direction: rpf.Direction, Tenant: rpf.Tenant, Category: rpf.Category, Subject: rpf.Subject}).Delete(TpRatingProfile{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, r := range APItoModelRatingProfile(rpf) {
			if err := tx.Save(&r).Error; err != nil {
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPSharedGroups(sgs []*utils.TPSharedGroups) error {
	if len(sgs) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)
	tx := self.db.Begin()
	for _, sGroup := range sgs {
		if found, _ := m[sGroup.ID]; !found {
			m[sGroup.ID] = true
			if err := tx.Where(&TpSharedGroup{Tpid: sGroup.TPid, Tag: sGroup.ID}).Delete(TpSharedGroup{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, s := range APItoModelSharedGroup(sGroup) {
			if err := tx.Save(&s).Error; err != nil {
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPCdrStats(css []*utils.TPCdrStats) error {
	if len(css) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)
	tx := self.db.Begin()
	for _, cStat := range css {
		if found, _ := m[cStat.ID]; !found {
			m[cStat.ID] = true
			if err := tx.Where(&TpCdrstat{Tpid: cStat.TPid, Tag: cStat.ID}).Delete(TpCdrstat{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, c := range APItoModelCdrStat(cStat) {
			if err := tx.Save(&c).Error; err != nil {
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPDerivedChargers(sgs []*utils.TPDerivedChargers) error {
	if len(sgs) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)
	tx := self.db.Begin()
	for _, dCharger := range sgs {
		dcKey := dCharger.GetDerivedChargersKey()
		if found, _ := m[dcKey]; !found {
			m[dcKey] = true
			if err := tx.Where(TpDerivedCharger{
				Tpid:      dCharger.TPid,
				Direction: dCharger.Direction,
				Tenant:    dCharger.Tenant,
				Category:  dCharger.Category,
				Account:   dCharger.Account,
				Subject:   dCharger.Subject,
			}).Delete(TpDerivedCharger{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, d := range APItoModelDerivedCharger(dCharger) {
			if err := tx.Save(&d).Error; err != nil {
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPLCRs(lcrs []*utils.TPLcrRules) error {
	if len(lcrs) == 0 {
		return nil //Nothing to set
	}
	tx := self.db.Begin()
	for _, lcr := range lcrs {
		if err := tx.Where(&TpLcrRule{
			Tpid:      lcr.TPid,
			Direction: lcr.Direction,
			Tenant:    lcr.Tenant,
			Category:  lcr.Category,
			Account:   lcr.Account,
			Subject:   lcr.Subject,
		}).Delete(TpLcrRule{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, l := range APItoModelLcrRule(lcr) {
			if err := tx.Save(&l).Error; err != nil {
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPActions(acts []*utils.TPActions) error {
	if len(acts) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)
	tx := self.db.Begin()
	for _, a := range acts {
		if found, _ := m[a.ID]; !found {
			m[a.ID] = true
			if err := tx.Where(&TpAction{Tpid: a.TPid, Tag: a.ID}).Delete(TpAction{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, sa := range APItoModelAction(a) {
			if err := tx.Save(&sa).Error; err != nil {
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPActionPlans(ats []*utils.TPActionPlan) error {
	if len(ats) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)

	tx := self.db.Begin()
	for _, aPlan := range ats {
		if found, _ := m[aPlan.ID]; !found {
			m[aPlan.ID] = true
			if err := tx.Where(&TpActionPlan{Tpid: aPlan.TPid, Tag: aPlan.ID}).Delete(TpActionPlan{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, a := range APItoModelActionPlan(aPlan) {
			if err := tx.Save(&a).Error; err != nil {
				return err
			}
		}
	}
	r := tx.Commit()
	return r.Error
}

func (self *SQLStorage) SetTPActionTriggers(ats []*utils.TPActionTriggers) error {
	if len(ats) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)
	tx := self.db.Begin()
	for _, aTrigger := range ats {
		if found, _ := m[aTrigger.ID]; !found {
			m[aTrigger.ID] = true
			if err := tx.Where(&TpActionTrigger{Tpid: aTrigger.TPid, Tag: aTrigger.ID}).Delete(TpActionTrigger{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, a := range APItoModelActionTrigger(aTrigger) {
			if err := tx.Save(&a).Error; err != nil {
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

// Sets a group of account actions. Map key has the role of grouping within a tpid
func (self *SQLStorage) SetTPAccountActions(aas []*utils.TPAccountActions) error {
	if len(aas) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)

	tx := self.db.Begin()
	for _, aa := range aas {
		if found, _ := m[aa.GetId()]; !found {
			m[aa.GetId()] = true
			if err := tx.Where(&TpAccountAction{Tpid: aa.TPid, Loadid: aa.LoadId, Tenant: aa.Tenant, Account: aa.Account}).Delete(&TpAccountAction{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		sa := APItoModelAccountAction(aa)
		if err := tx.Save(&sa).Error; err != nil {
			return err
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPResources(rls []*utils.TPResource) error {
	if len(rls) == 0 {
		return nil
	}
	tx := self.db.Begin()
	for _, rl := range rls {
		// Remove previous
		if err := tx.Where(&TpResource{Tpid: rl.TPid, ID: rl.ID}).Delete(TpResource{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, mrl := range APItoModelResource(rl) {
			if err := tx.Save(&mrl).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPStats(sts []*utils.TPStats) error {
	if len(sts) == 0 {
		return nil
	}
	tx := self.db.Begin()
	for _, stq := range sts {
		// Remove previous
		if err := tx.Where(&TpStats{Tpid: stq.TPid, ID: stq.ID}).Delete(TpStats{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, mst := range APItoModelStats(stq) {
			if err := tx.Save(&mst).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPThresholds(ths []*utils.TPThreshold) error {
	if len(ths) == 0 {
		return nil
	}
	tx := self.db.Begin()
	for _, th := range ths {
		// Remove previous
		if err := tx.Where(&TpThreshold{Tpid: th.TPid, ID: th.ID}).Delete(TpThreshold{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, mst := range APItoModelTPThreshold(th) {
			if err := tx.Save(&mst).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPFilters(ths []*utils.TPFilterProfile) error {
	if len(ths) == 0 {
		return nil
	}
	tx := self.db.Begin()
	for _, th := range ths {
		// Remove previous
		if err := tx.Where(&TpFilter{Tpid: th.TPid, ID: th.ID}).Delete(TpFilter{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, mst := range APItoModelTPFilter(th) {
			if err := tx.Save(&mst).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPSuppliers(tpSPs []*utils.TPSupplierProfile) error {
	if len(tpSPs) == 0 {
		return nil
	}
	tx := self.db.Begin()
	for _, stq := range tpSPs {
		// Remove previous
		if err := tx.Where(&TpSupplier{Tpid: stq.TPid, ID: stq.ID}).Delete(TpSupplier{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, mst := range APItoModelTPSuppliers(stq) {
			if err := tx.Save(&mst).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPAttributes(tpAttrs []*utils.TPAttributeProfile) error {
	if len(tpAttrs) == 0 {
		return nil
	}
	tx := self.db.Begin()
	for _, stq := range tpAttrs {
		// Remove previous
		if err := tx.Where(&TPAttribute{Tpid: stq.TPid, ID: stq.ID}).Delete(TPAttribute{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, mst := range APItoModelTPAttribute(stq) {
			if err := tx.Save(&mst).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetSMCost(smc *SMCost) error {
	if smc.CostDetails == nil {
		return nil
	}
	tx := self.db.Begin()
	cd := &SMCostSQL{
		Cgrid:       smc.CGRID,
		RunID:       smc.RunID,
		OriginHost:  smc.OriginHost,
		OriginID:    smc.OriginID,
		CostSource:  smc.CostSource,
		CostDetails: smc.CostDetails.AsJSON(),
		Usage:       smc.Usage.Nanoseconds(),
		CreatedAt:   time.Now(),
	}
	if tx.Save(cd).Error != nil { // Check further since error does not properly reflect duplicates here (sql: no rows in result set)
		tx.Rollback()
		return tx.Error
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) RemoveSMCost(smc *SMCost) error {
	tx := self.db.Begin()

	if err := tx.Where(&SMCostSQL{Cgrid: smc.CGRID, RunID: smc.RunID}).Delete(SMCost{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// GetSMCosts is used to retrieve one or multiple SMCosts based on filter
func (self *SQLStorage) GetSMCosts(cgrid, runid, originHost, originIDPrefix string) ([]*SMCost, error) {
	var smCosts []*SMCost
	filter := &SMCostSQL{}
	if cgrid != "" {
		filter.Cgrid = cgrid
	}
	if runid != "" {
		filter.RunID = runid
	}
	if originHost != "" {
		filter.OriginHost = originHost
	}
	q := self.db.Where(filter)
	if originIDPrefix != "" {
		q = self.db.Where(filter).Where(fmt.Sprintf("origin_id LIKE '%s%%'", originIDPrefix))
	}
	results := make([]*SMCostSQL, 0)
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
			Usage:       time.Duration(result.Usage),
			CostDetails: &CallCost{},
		}
		if err := json.Unmarshal([]byte(result.CostDetails), smc.CostDetails); err != nil {
			return nil, err
		}
		smCosts = append(smCosts, smc)
	}
	if len(smCosts) == 0 {
		return smCosts, utils.ErrNotFound
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
	tx := self.db.Begin()
	cdrSql := cdr.AsCDRsql()
	cdrSql.CreatedAt = time.Now()
	saved := tx.Save(cdrSql)
	if saved.Error != nil {
		tx.Rollback()
		if !allowUpdate {
			return saved.Error
		}
		tx = self.db.Begin()
		cdrSql.UpdatedAt = time.Now()
		updated := tx.Model(&CDRsql{}).Where(
			&CDRsql{Cgrid: cdr.CGRID, RunID: cdr.RunID, OriginID: cdr.OriginID}).Updates(cdrSql)
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
	q := self.db.Table(utils.CDRsTBL).Select("*")
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
	if len(qryFltr.OriginIDs) != 0 {
		q = q.Where("origin_id in (?)", qryFltr.OriginIDs)
	}
	if len(qryFltr.NotOriginIDs) != 0 {
		q = q.Where("origin_id not in (?)", qryFltr.NotOriginIDs)
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
	if len(qryFltr.Costs) != 0 {
		q = q.Where(utils.CDRsTBL+".cost in (?)", qryFltr.Costs)
	}
	if len(qryFltr.NotCosts) != 0 {
		q = q.Where(utils.CDRsTBL+".cost not in (?)", qryFltr.NotCosts)
	}
	if len(qryFltr.ExtraFields) != 0 { // Extra fields searches, implemented as contains in extra field
		qIds := bytes.NewBufferString("(")
		needOr := false
		for field, value := range qryFltr.ExtraFields {
			if needOr {
				qIds.WriteString(" OR")
			}
			if value == utils.MetaExists {
				qIds.WriteString(self.SQLImpl.extraFieldsExistsQry(field))
			} else {
				qIds.WriteString(self.SQLImpl.extraFieldsValueQry(field, value))
			}
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
				qIds.WriteString(" AND")
			}
			if value == utils.MetaExists {
				qIds.WriteString(self.SQLImpl.notExtraFieldsExistsQry(field))
			} else {
				qIds.WriteString(self.SQLImpl.notExtraFieldsValueQry(field, value))
			}
			needAnd = true
		}
		qIds.WriteString(" )")
		q = q.Where(qIds.String())
	}
	if qryFltr.OrderIDStart != nil { // Keep backwards compatible by testing 0 value
		q = q.Where(utils.CDRsTBL+".id >= ?", *qryFltr.OrderIDStart)
	}
	if qryFltr.OrderIDEnd != nil {
		q = q.Where(utils.CDRsTBL+".id < ?", *qryFltr.OrderIDEnd)
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
		minUsage, err := utils.ParseDurationWithNanosecs(qryFltr.MinUsage)
		if err != nil {
			return nil, 0, err
		}
		if self.db.Dialect().GetName() == utils.MYSQL { // MySQL needs escaping for usage
			q = q.Where("`usage` >= ?", minUsage.Nanoseconds())
		} else {
			q = q.Where("usage >= ?", minUsage.Nanoseconds())
		}
	}
	if len(qryFltr.MaxUsage) != 0 {
		maxUsage, err := utils.ParseDurationWithNanosecs(qryFltr.MaxUsage)
		if err != nil {
			return nil, 0, err
		}
		if self.db.Dialect().GetName() == utils.MYSQL { // MySQL needs escaping for usage
			q = q.Where("`usage` < ?", maxUsage.Nanoseconds())
		} else {
			q = q.Where("usage < ?", maxUsage.Nanoseconds())
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
	results := make([]*CDRsql, 0)
	if err := q.Find(&results).Error; err != nil {
		return nil, 0, err
	}
	for _, result := range results {
		if cdr, err := NewCDRFromSQL(result); err != nil {
			return nil, 0, err
		} else {
			cdrs = append(cdrs, cdr)
		}
	}
	if len(cdrs) == 0 && !remove {
		return cdrs, 0, utils.ErrNotFound
	}
	return cdrs, 0, nil
}

func (self *SQLStorage) GetTPDestinations(tpid, id string) (uTPDsts []*utils.TPDestination, err error) {
	var tpDests TpDestinations
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("tag = ?", id)
	}
	if err := q.Find(&tpDests).Error; err != nil {
		return nil, err
	}
	if len(tpDests.AsTPDestinations()) == 0 {
		return tpDests.AsTPDestinations(), utils.ErrNotFound
	}
	return tpDests.AsTPDestinations(), nil
}

func (self *SQLStorage) GetTPRates(tpid, id string) ([]*utils.TPRate, error) {
	var tpRates TpRates
	q := self.db.Where("tpid = ?", tpid).Order("id")
	if len(id) != 0 {
		q = q.Where("tag = ?", id)
	}
	if err := q.Find(&tpRates).Error; err != nil {
		return nil, err
	}
	if rs, err := tpRates.AsTPRates(); err != nil {
		return nil, err
	} else {
		if len(rs) == 0 {
			return rs, utils.ErrNotFound
		}
		return rs, nil
	}
}

func (self *SQLStorage) GetTPDestinationRates(tpid, id string, pagination *utils.Paginator) ([]*utils.TPDestinationRate, error) {
	var tpDestinationRates TpDestinationRates
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("tag = ?", id)
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
	if drs, err := tpDestinationRates.AsTPDestinationRates(); err != nil {
		return nil, err
	} else {
		if len(drs) == 0 {
			return drs, utils.ErrNotFound
		}
		return drs, nil
	}
}

func (self *SQLStorage) GetTPTimings(tpid, id string) ([]*utils.ApierTPTiming, error) {
	var tpTimings TpTimings
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("tag = ?", id)
	}
	if err := q.Find(&tpTimings).Error; err != nil {
		return nil, err
	}
	ts := tpTimings.AsTPTimings()
	if len(ts) == 0 {
		return ts, utils.ErrNotFound
	}
	return ts, nil
}

func (self *SQLStorage) GetTPRatingPlans(tpid, id string, pagination *utils.Paginator) ([]*utils.TPRatingPlan, error) {
	var tpRatingPlans TpRatingPlans
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("tag = ?", id)
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
	if rps, err := tpRatingPlans.AsTPRatingPlans(); err != nil {
		return nil, err
	} else {
		if len(rps) == 0 {
			return rps, utils.ErrNotFound
		}
		return rps, nil
	}
}

func (self *SQLStorage) GetTPRatingProfiles(filter *utils.TPRatingProfile) ([]*utils.TPRatingProfile, error) {
	var tpRpfs TpRatingProfiles
	q := self.db.Where("tpid = ?", filter.TPid)
	if len(filter.LoadId) != 0 {
		q = q.Where("loadid = ?", filter.LoadId)
	}
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
	if err := q.Find(&tpRpfs).Error; err != nil {
		return nil, err
	}
	if rps, err := tpRpfs.AsTPRatingProfiles(); err != nil {
		return nil, err
	} else {
		if len(rps) == 0 {
			return rps, utils.ErrNotFound
		}
		return rps, nil
	}
}

func (self *SQLStorage) GetTPSharedGroups(tpid, id string) ([]*utils.TPSharedGroups, error) {
	var tpShareGroups TpSharedGroups
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("tag = ?", id)
	}
	if err := q.Find(&tpShareGroups).Error; err != nil {
		return nil, err
	}
	if sgs, err := tpShareGroups.AsTPSharedGroups(); err != nil {
		return nil, err
	} else {
		if len(sgs) == 0 {
			return sgs, utils.ErrNotFound
		}
		return sgs, nil
	}
}

func (self *SQLStorage) GetTPLCRs(filter *utils.TPLcrRules) ([]*utils.TPLcrRules, error) {
	var tpLcrRules TpLcrRules
	q := self.db.Where("tpid = ?", filter.TPid)
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
	if err := q.Find(&tpLcrRules).Error; err != nil {
		return nil, err
	}
	if lrs, err := tpLcrRules.AsTPLcrRules(); err != nil {
		return nil, err
	} else {
		if len(lrs) == 0 {
			return lrs, utils.ErrNotFound
		}
		return lrs, nil
	}
}

func (self *SQLStorage) GetTPActions(tpid, id string) ([]*utils.TPActions, error) {
	var tpActions TpActions
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("tag = ?", id)
	}
	if err := q.Find(&tpActions).Error; err != nil {
		return nil, err
	}
	if as, err := tpActions.AsTPActions(); err != nil {
		return nil, err
	} else {
		if len(as) == 0 {
			return as, utils.ErrNotFound
		}
		return as, nil
	}
}

func (self *SQLStorage) GetTPActionTriggers(tpid, id string) ([]*utils.TPActionTriggers, error) {
	var tpActionTriggers TpActionTriggers
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("tag = ?", id)
	}
	if err := q.Find(&tpActionTriggers).Error; err != nil {
		return nil, err
	}
	if ats, err := tpActionTriggers.AsTPActionTriggers(); err != nil {
		return nil, err
	} else {
		if len(ats) == 0 {
			return ats, utils.ErrNotFound
		}
		return ats, nil
	}
}

func (self *SQLStorage) GetTPActionPlans(tpid, id string) ([]*utils.TPActionPlan, error) {
	var tpActionPlans TpActionPlans
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("tag = ?", id)
	}
	if err := q.Find(&tpActionPlans).Error; err != nil {
		return nil, err
	}
	if aps, err := tpActionPlans.AsTPActionPlans(); err != nil {
		return nil, err
	} else {
		if len(aps) == 0 {
			return aps, utils.ErrNotFound
		}
		return aps, nil
	}
}

func (self *SQLStorage) GetTPAccountActions(filter *utils.TPAccountActions) ([]*utils.TPAccountActions, error) {
	var tpAccActs TpAccountActions
	q := self.db.Where("tpid = ?", filter.TPid)
	if len(filter.LoadId) != 0 {
		q = q.Where("loadid = ?", filter.LoadId)
	}
	if len(filter.Tenant) != 0 {
		q = q.Where("tenant = ?", filter.Tenant)
	}
	if len(filter.Account) != 0 {
		q = q.Where("account = ?", filter.Account)
	}
	if err := q.Find(&tpAccActs).Error; err != nil {
		return nil, err
	}
	if aas, err := tpAccActs.AsTPAccountActions(); err != nil {
		return nil, err
	} else {
		if len(aas) == 0 {
			return aas, utils.ErrNotFound
		}
		return aas, nil
	}
}

func (self *SQLStorage) GetTPDerivedChargers(filter *utils.TPDerivedChargers) ([]*utils.TPDerivedChargers, error) {
	var tpDerivedChargers TpDerivedChargers
	q := self.db.Where("tpid = ?", filter.TPid)
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
	if len(filter.LoadId) != 0 {
		q = q.Where("loadid = ?", filter.LoadId)
	}
	if err := q.Find(&tpDerivedChargers).Error; err != nil {
		return nil, err
	}
	if dcs, err := tpDerivedChargers.AsTPDerivedChargers(); err != nil {
		return nil, err
	} else {
		if len(dcs) == 0 {
			return dcs, utils.ErrNotFound
		}
		return dcs, nil
	}
}

func (self *SQLStorage) GetTPCdrStats(tpid, id string) ([]*utils.TPCdrStats, error) {
	var tpCdrStats TpCdrStats
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("tag = ?", id)
	}
	if err := q.Find(&tpCdrStats).Error; err != nil {
		return nil, err
	}
	if css, err := tpCdrStats.AsTPCdrStats(); err != nil {
		return nil, err
	} else {
		if len(css) == 0 {
			return css, utils.ErrNotFound
		}
		return css, nil
	}
}

func (self *SQLStorage) SetTPUsers(users []*utils.TPUsers) error {
	if len(users) == 0 {
		return nil
	}
	m := make(map[string]bool)
	tx := self.db.Begin()
	for _, user := range users {
		if found, _ := m[user.Tenant]; !found {
			m[user.Tenant] = true
			if err := tx.Where(&TpUser{Tpid: user.TPid, Tenant: user.Tenant, UserName: user.UserName}).Delete(&TpUser{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, u := range APItoModelUsers(user) {
			if err := tx.Save(&u).Error; err != nil {
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) GetTPUsers(filter *utils.TPUsers) ([]*utils.TPUsers, error) {
	var tpUsers TpUsers
	q := self.db.Where("tpid = ?", filter.TPid)
	if len(filter.Tenant) != 0 {
		q = q.Where("tenant = ?", filter.Tenant)
	}
	if len(filter.UserName) != 0 {
		q = q.Where("user_name = ?", filter.UserName)
	}
	if err := q.Find(&tpUsers).Error; err != nil {
		return nil, err
	}
	if us, err := tpUsers.AsTPUsers(); err != nil {
		return nil, err
	} else {
		if len(us) == 0 {
			return us, utils.ErrNotFound
		}
		return us, nil
	}
}

func (self *SQLStorage) SetTPAliases(aliases []*utils.TPAliases) error {
	if len(aliases) == 0 {
		return nil
	}
	m := make(map[string]bool)
	tx := self.db.Begin()
	for _, alias := range aliases {
		if found, _ := m[alias.GetId()]; !found {
			m[alias.GetId()] = true
			if err := tx.Where(&TpAlias{
				Tpid:      alias.TPid,
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
		for _, a := range APItoModelAliases(alias) {
			if err := tx.Save(&a).Error; err != nil {
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) GetTPAliases(filter *utils.TPAliases) ([]*utils.TPAliases, error) {
	var tpAliases TpAliases
	q := self.db.Where("tpid = ?", filter.TPid)
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
	if as, err := tpAliases.AsTPAliases(); err != nil {
		return nil, err
	} else {
		if len(as) == 0 {
			return as, utils.ErrNotFound
		}
		return as, nil
	}
}

func (self *SQLStorage) GetTPResources(tpid, id string) ([]*utils.TPResource, error) {
	var rls TpResources
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("id = ?", id)
	}
	if err := q.Find(&rls).Error; err != nil {
		return nil, err
	}
	arls := rls.AsTPResources()
	if len(arls) == 0 {
		return arls, utils.ErrNotFound
	}
	return arls, nil
}

func (self *SQLStorage) GetTPStats(tpid, id string) ([]*utils.TPStats, error) {
	var sts TpStatsS
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("id = ?", id)
	}
	if err := q.Find(&sts).Error; err != nil {
		return nil, err
	}
	asts := sts.AsTPStats()
	if len(asts) == 0 {
		return asts, utils.ErrNotFound
	}
	return asts, nil
}

func (self *SQLStorage) GetTPThresholds(tpid, id string) ([]*utils.TPThreshold, error) {
	var ths TpThresholdS
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("id = ?", id)
	}
	if err := q.Find(&ths).Error; err != nil {
		return nil, err
	}
	aths := ths.AsTPThreshold()
	if len(aths) == 0 {
		return aths, utils.ErrNotFound
	}
	return aths, nil
}

func (self *SQLStorage) GetTPFilters(tpid, id string) ([]*utils.TPFilterProfile, error) {
	var ths TpFilterS
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("id = ?", id)
	}
	if err := q.Find(&ths).Error; err != nil {
		return nil, err
	}
	aths := ths.AsTPFilter()
	if len(aths) == 0 {
		return aths, utils.ErrNotFound
	}
	return aths, nil
}

func (self *SQLStorage) GetTPSuppliers(tpid, id string) ([]*utils.TPSupplierProfile, error) {
	var sps TpSuppliers
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("id = ?", id)
	}
	if err := q.Find(&sps).Error; err != nil {
		return nil, err
	}
	arls := sps.AsTPSuppliers()
	if len(arls) == 0 {
		return arls, utils.ErrNotFound
	}
	return arls, nil
}

func (self *SQLStorage) GetTPAttributes(tpid, id string) ([]*utils.TPAttributeProfile, error) {
	var sps TPAttributes
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("id = ?", id)
	}
	if err := q.Find(&sps).Error; err != nil {
		return nil, err
	}
	arls := sps.AsTPAttributes()
	if len(arls) == 0 {
		return arls, utils.ErrNotFound
	}
	return arls, nil
}

// GetVersions returns slice of all versions or a specific version if tag is specified
func (self *SQLStorage) GetVersions(itm string) (vrs Versions, err error) {
	q := self.db.Model(&TBLVersion{})
	if itm != utils.TBLVersions && itm != "" {
		q = self.db.Where(&TBLVersion{Item: itm})
	}
	var verModels []*TBLVersion
	if err = q.Find(&verModels).Error; err != nil {
		return
	}
	vrs = make(Versions)
	for _, verModel := range verModels {
		vrs[verModel.Item] = verModel.Version
	}
	if len(vrs) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

// RemoveVersions will remove specific versions out of storage
func (self *SQLStorage) RemoveVersions(vrs Versions) (err error) {
	if len(vrs) == 0 { // Remove all if no key provided
		err = self.db.Delete(TBLVersion{}).Error
		return
	}
	tx := self.db.Begin()
	for key := range vrs {
		if err = tx.Where(&TBLVersion{Item: key}).Delete(TBLVersion{}).Error; err != nil {
			tx.Rollback()
			return
		}
	}
	tx.Commit()
	return
}
