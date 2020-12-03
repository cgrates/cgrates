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

func (self *SQLStorage) ExportGormDB() *gorm.DB {
	return self.db
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
		utils.TBLTPDestinationRates, utils.TBLTPRatingPlans, utils.TBLTPRatingProfiles,
		utils.TBLTPSharedGroups, utils.TBLTPActions, utils.TBLTPActionTriggers,
		utils.TBLTPAccountActions, utils.TBLTPResources, utils.TBLTPStats, utils.TBLTPThresholds,
		utils.TBLTPFilters, utils.SessionCostsTBL, utils.CDRsTBL, utils.TBLTPActionPlans,
		utils.TBLVersions, utils.TBLTPRoutes, utils.TBLTPAttributes, utils.TBLTPChargers,
		utils.TBLTPDispatchers, utils.TBLTPDispatcherHosts,
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
			"(SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s) UNION (SELECT tpid FROM %s)",
			utils.TBLTPTimings,
			utils.TBLTPDestinations,
			utils.TBLTPRates,
			utils.TBLTPDestinationRates,
			utils.TBLTPRatingPlans,
			utils.TBLTPRatingProfiles,
			utils.TBLTPSharedGroups,
			utils.TBLTPActions,
			utils.TBLTPActionTriggers,
			utils.TBLTPAccountActions,
			utils.TBLTPResources,
			utils.TBLTPStats,
			utils.TBLTPThresholds,
			utils.TBLTPFilters,
			utils.TBLTPActionPlans,
			utils.TBLTPRoutes,
			utils.TBLTPAttributes,
			utils.TBLTPChargers,
			utils.TBLTPDispatchers,
			utils.TBLTPDispatcherHosts,
		)
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
func (self *SQLStorage) GetTpTableIds(tpid, table string, distinct utils.TPDistinctIds,
	filters map[string]string, pagination *utils.PaginatorWithSearch) ([]string, error) {
	qry := fmt.Sprintf("SELECT DISTINCT %s FROM %s where tpid='%s'", distinct, table, tpid)
	for key, value := range filters {
		if key != "" && value != "" {
			qry += fmt.Sprintf(" AND %s='%s'", key, value)
		}
	}
	if pagination != nil {
		if len(pagination.Search) != 0 {
			qry += fmt.Sprintf(" AND (%s LIKE '%%%s%%'", distinct[0], pagination.Search)
			for _, d := range distinct[1:] {
				qry += fmt.Sprintf(" OR %s LIKE '%%%s%%'", d, pagination.Search)
			}
			qry += fmt.Sprintf(")")
		}
		if pagination.Paginator != nil {
			if pagination.Limit != nil { // Keep Postgres compatibility by adding offset only when limit defined
				qry += fmt.Sprintf(" LIMIT %d", *pagination.Limit)
				if pagination.Offset != nil {
					qry += fmt.Sprintf(" OFFSET %d", *pagination.Offset)
				}
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
		finalID := vals[0]
		if len(vals) > 1 {
			finalID = strings.Join(vals, utils.CONCATENATED_KEY_SEP)
		}
		ids = append(ids, finalID)
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
			utils.TBLTPDestinationRates, utils.TBLTPRatingPlans, utils.TBLTPRatingProfiles,
			utils.TBLTPSharedGroups, utils.TBLTPActions, utils.TBLTPActionTriggers,
			utils.TBLTPAccountActions, utils.TBLTPResources, utils.TBLTPStats, utils.TBLTPThresholds,
			utils.TBLTPFilters, utils.TBLTPActionPlans, utils.TBLTPRoutes, utils.TBLTPAttributes,
			utils.TBLTPChargers, utils.TBLTPDispatchers, utils.TBLTPDispatcherHosts} {
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
		if err := tx.Where(&TpTimingMdl{Tpid: timing.TPid, Tag: timing.ID}).Delete(TpTimingMdl{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		t := APItoModelTiming(timing)
		if err := tx.Save(&t).Error; err != nil {
			tx.Rollback()
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
		if err := tx.Where(&TpDestinationMdl{Tpid: dst.TPid, Tag: dst.ID}).Delete(TpDestinationMdl{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, d := range APItoModelDestination(dst) {
			if err := tx.Save(&d).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPRates(rs []*utils.TPRateRALs) error {
	if len(rs) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)
	tx := self.db.Begin()
	for _, rate := range rs {
		if found, _ := m[rate.ID]; !found {
			m[rate.ID] = true
			if err := tx.Where(&TpRateMdl{Tpid: rate.TPid, Tag: rate.ID}).Delete(TpRateMdl{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, r := range APItoModelRate(rate) {
			if err := tx.Save(&r).Error; err != nil {
				tx.Rollback()
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
			if err := tx.Where(&TpDestinationRateMdl{Tpid: dRate.TPid, Tag: dRate.ID}).Delete(TpDestinationRateMdl{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, d := range APItoModelDestinationRate(dRate) {
			if err := tx.Save(&d).Error; err != nil {
				tx.Rollback()
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
			if err := tx.Where(&TpRatingPlanMdl{Tpid: rPlan.TPid, Tag: rPlan.ID}).Delete(TpRatingPlanMdl{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, r := range APItoModelRatingPlan(rPlan) {
			if err := tx.Save(&r).Error; err != nil {
				tx.Rollback()
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
		if err := tx.Where(&TpRatingProfileMdl{Tpid: rpf.TPid, Loadid: rpf.LoadId,
			Tenant: rpf.Tenant, Category: rpf.Category,
			Subject: rpf.Subject}).Delete(TpRatingProfileMdl{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, r := range APItoModelRatingProfile(rpf) {
			if err := tx.Save(&r).Error; err != nil {
				tx.Rollback()
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
			if err := tx.Where(&TpSharedGroupMdl{Tpid: sGroup.TPid, Tag: sGroup.ID}).Delete(TpSharedGroupMdl{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, s := range APItoModelSharedGroup(sGroup) {
			if err := tx.Save(&s).Error; err != nil {
				tx.Rollback()
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
			if err := tx.Where(&TpActionMdl{Tpid: a.TPid, Tag: a.ID}).Delete(TpActionMdl{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, sa := range APItoModelAction(a) {
			if err := tx.Save(&sa).Error; err != nil {
				tx.Rollback()
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
			if err := tx.Where(&TpActionPlanMdl{Tpid: aPlan.TPid, Tag: aPlan.ID}).Delete(TpActionPlanMdl{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, a := range APItoModelActionPlan(aPlan) {
			if err := tx.Save(&a).Error; err != nil {
				tx.Rollback()
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
			if err := tx.Where(&TpActionTriggerMdl{Tpid: aTrigger.TPid, Tag: aTrigger.ID}).Delete(TpActionTriggerMdl{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, a := range APItoModelActionTrigger(aTrigger) {
			if err := tx.Save(&a).Error; err != nil {
				tx.Rollback()
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
			if err := tx.Where(&TpAccountActionMdl{Tpid: aa.TPid, Loadid: aa.LoadId, Tenant: aa.Tenant, Account: aa.Account}).Delete(&TpAccountActionMdl{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		sa := APItoModelAccountAction(aa)
		if err := tx.Save(&sa).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPResources(rls []*utils.TPResourceProfile) error {
	if len(rls) == 0 {
		return nil
	}
	tx := self.db.Begin()
	for _, rl := range rls {
		// Remove previous
		if err := tx.Where(&TpResourceMdl{Tpid: rl.TPid, ID: rl.ID}).Delete(TpResourceMdl{}).Error; err != nil {
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

func (self *SQLStorage) SetTPStats(sts []*utils.TPStatProfile) error {
	if len(sts) == 0 {
		return nil
	}
	tx := self.db.Begin()
	for _, stq := range sts {
		// Remove previous
		if err := tx.Where(&TpStatMdl{Tpid: stq.TPid, ID: stq.ID}).Delete(TpStatMdl{}).Error; err != nil {
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

func (self *SQLStorage) SetTPThresholds(ths []*utils.TPThresholdProfile) error {
	if len(ths) == 0 {
		return nil
	}
	tx := self.db.Begin()
	for _, th := range ths {
		// Remove previous
		if err := tx.Where(&TpThresholdMdl{Tpid: th.TPid, ID: th.ID}).Delete(TpThresholdMdl{}).Error; err != nil {
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
		if err := tx.Where(&TpFilterMdl{Tpid: th.TPid, ID: th.ID}).Delete(TpFilterMdl{}).Error; err != nil {
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

func (self *SQLStorage) SetTPRoutes(tpRoutes []*utils.TPRouteProfile) error {
	if len(tpRoutes) == 0 {
		return nil
	}
	tx := self.db.Begin()
	for _, tpRoute := range tpRoutes {
		// Remove previous
		if err := tx.Where(&TpRouteMdl{Tpid: tpRoute.TPid, ID: tpRoute.ID}).Delete(TpRouteMdl{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, mst := range APItoModelTPRoutes(tpRoute) {
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
		if err := tx.Where(&TPAttributeMdl{Tpid: stq.TPid, ID: stq.ID}).Delete(TPAttributeMdl{}).Error; err != nil {
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

func (self *SQLStorage) SetTPChargers(tpCPPs []*utils.TPChargerProfile) error {
	if len(tpCPPs) == 0 {
		return nil
	}
	tx := self.db.Begin()
	for _, cpp := range tpCPPs {
		// Remove previous
		if err := tx.Where(&TPChargerMdl{Tpid: cpp.TPid, ID: cpp.ID}).Delete(TPChargerMdl{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, mst := range APItoModelTPCharger(cpp) {
			if err := tx.Save(&mst).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPDispatcherProfiles(tpDPPs []*utils.TPDispatcherProfile) error {
	if len(tpDPPs) == 0 {
		return nil
	}
	tx := self.db.Begin()
	for _, dpp := range tpDPPs {
		// Remove previous
		if err := tx.Where(&TPDispatcherProfileMdl{Tpid: dpp.TPid, ID: dpp.ID}).Delete(TPDispatcherProfileMdl{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, mst := range APItoModelTPDispatcherProfile(dpp) {
			if err := tx.Save(&mst).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPDispatcherHosts(tpDPPs []*utils.TPDispatcherHost) error {
	if len(tpDPPs) == 0 {
		return nil
	}
	tx := self.db.Begin()
	for _, dpp := range tpDPPs {
		// Remove previous
		if err := tx.Where(&TPDispatcherHostMdl{Tpid: dpp.TPid, ID: dpp.ID}).Delete(TPDispatcherHostMdl{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Save(APItoModelTPDispatcherHost(dpp)).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPRateProfiles(tpDPPs []*utils.TPRateProfile) error {
	if len(tpDPPs) == 0 {
		return nil
	}
	tx := self.db.Begin()
	for _, dpp := range tpDPPs {
		// Remove previous
		if err := tx.Where(&RateProfileMdl{Tpid: dpp.TPid, ID: dpp.ID}).Delete(RateProfileMdl{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, mst := range APItoModelTPRateProfile(dpp) {
			if err := tx.Save(&mst).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) SetTPActionProfiles(tpAps []*utils.TPActionProfile) error {
	if len(tpAps) == 0 {
		return nil
	}
	tx := self.db.Begin()
	for _, tpAp := range tpAps {
		// Remove previous
		if err := tx.Where(&ActionProfileMdl{Tpid: tpAp.TPid, Tenant: tpAp.Tenant, ID: tpAp.ID}).Delete(ActionProfileMdl{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, mst := range APItoModelTPActionProfile(tpAp) {
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
	cd := &SessionCostsSQL{
		Cgrid:       smc.CGRID,
		RunID:       smc.RunID,
		OriginHost:  smc.OriginHost,
		OriginID:    smc.OriginID,
		CostSource:  smc.CostSource,
		CostDetails: utils.ToJSON(smc.CostDetails),
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
	var rmParam *SessionCostsSQL
	if smc != nil {
		rmParam = &SessionCostsSQL{Cgrid: smc.CGRID,
			RunID: smc.RunID}
	}
	if err := tx.Where(rmParam).Delete(SessionCostsSQL{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (self *SQLStorage) RemoveSMCosts(qryFltr *utils.SMCostFilter) error {
	q := self.db.Table(utils.SessionCostsTBL).Select("*")
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
	if len(qryFltr.OriginHosts) != 0 {
		q = q.Where("origin_host in (?)", qryFltr.OriginHosts)
	}
	if len(qryFltr.NotOriginHosts) != 0 {
		q = q.Where("origin_host not in (?)", qryFltr.NotOriginHosts)
	}
	if len(qryFltr.CostSources) != 0 {
		q = q.Where("costsource in (?)", qryFltr.CostSources)
	}
	if len(qryFltr.NotCostSources) != 0 {
		q = q.Where("costsource not in (?)", qryFltr.NotCostSources)
	}
	if qryFltr.CreatedAt.Begin != nil {
		q = q.Where("created_at >= ?", qryFltr.CreatedAt.Begin)
	}
	if qryFltr.CreatedAt.End != nil {
		q = q.Where("created_at < ?", qryFltr.CreatedAt.End)
	}
	if qryFltr.Usage.Min != nil {
		if self.db.Dialect().GetName() == utils.MYSQL { // MySQL needs escaping for usage
			q = q.Where("`usage` >= ?", qryFltr.Usage.Min.Nanoseconds())
		} else {
			q = q.Where("usage >= ?", qryFltr.Usage.Min.Nanoseconds())
		}
	}
	if qryFltr.Usage.Max != nil {
		if self.db.Dialect().GetName() == utils.MYSQL { // MySQL needs escaping for usage
			q = q.Where("`usage` < ?", qryFltr.Usage.Max.Nanoseconds())
		} else {
			q = q.Where("usage < ?", qryFltr.Usage.Max.Nanoseconds())
		}
	}
	if err := q.Delete(nil).Error; err != nil {
		q.Rollback()
		return err
	}
	return nil
}

// GetSMCosts is used to retrieve one or multiple SMCosts based on filter
func (self *SQLStorage) GetSMCosts(cgrid, runid, originHost, originIDPrefix string) ([]*SMCost, error) {
	var smCosts []*SMCost
	filter := &SessionCostsSQL{}
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
	results := make([]*SessionCostsSQL, 0)
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
			CostDetails: new(EventCost),
		}
		if err := json.Unmarshal([]byte(result.CostDetails), smc.CostDetails); err != nil {
			return nil, err
		}
		smc.CostDetails.initCache()
		smCosts = append(smCosts, smc)
	}
	if len(smCosts) == 0 {
		return smCosts, utils.ErrNotFound
	}
	return smCosts, nil
}

func (self *SQLStorage) SetCDR(cdr *CDR, allowUpdate bool) error {
	tx := self.db.Begin()
	cdrSql := cdr.AsCDRsql()
	cdrSql.CreatedAt = time.Now()
	saved := tx.Save(cdrSql)
	if saved.Error != nil {
		tx.Rollback()
		if !allowUpdate {
			if strings.Contains(saved.Error.Error(), "1062") || strings.Contains(saved.Error.Error(), "duplicate key") { // returns 1062/pq when key is duplicated
				return utils.ErrExists
			}
			return saved.Error
		}
		tx = self.db.Begin()
		cdrSql.UpdatedAt = time.Now()
		updated := tx.Model(&CDRsql{}).Where(
			&CDRsql{Cgrid: cdr.CGRID, RunID: cdr.RunID, OriginID: cdr.OriginID}).Updates(cdrSql.AsMapStringInterface())
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
	if qryFltr.AnswerTimeStart != nil && !qryFltr.AnswerTimeStart.IsZero() { // With IsZero we keep backwards compatible with APIerSv1
		q = q.Where("answer_time >= ?", qryFltr.AnswerTimeStart)
	}
	if qryFltr.AnswerTimeEnd != nil && !qryFltr.AnswerTimeEnd.IsZero() {
		q = q.Where("answer_time < ?", qryFltr.AnswerTimeEnd)
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
	if qryFltr.CreatedAtStart != nil && !qryFltr.CreatedAtStart.IsZero() { // With IsZero we keep backwards compatible with APIerSv1
		q = q.Where("created_at >= ?", qryFltr.CreatedAtStart)
	}
	if qryFltr.CreatedAtEnd != nil && !qryFltr.CreatedAtEnd.IsZero() {
		q = q.Where("created_at < ?", qryFltr.CreatedAtEnd)
	}
	if qryFltr.UpdatedAtStart != nil && !qryFltr.UpdatedAtStart.IsZero() { // With IsZero we keep backwards compatible with APIerSv1
		q = q.Where("updated_at >= ?", qryFltr.UpdatedAtStart)
	}
	if qryFltr.UpdatedAtEnd != nil && !qryFltr.UpdatedAtEnd.IsZero() {
		q = q.Where("updated_at < ?", qryFltr.UpdatedAtEnd)
	}
	if qryFltr.OrderBy != "" {
		var orderVal string
		separateVals := strings.Split(qryFltr.OrderBy, utils.INFIELD_SEP)
		switch separateVals[0] {
		case utils.OrderID:
			orderVal = "id"
		case utils.AnswerTime:
			orderVal = "answer_time"
		case utils.SetupTime:
			orderVal = "setup_time"
		case utils.Usage:
			if self.db.Dialect().GetName() == utils.MYSQL {
				orderVal = "`usage`"
			} else {
				orderVal = "usage"
			}
		case utils.Cost:
			orderVal = "cost"
		default:
			return nil, 0, fmt.Errorf("Invalid value : %s", separateVals[0])
		}
		if len(separateVals) == 2 && separateVals[1] == "desc" {
			orderVal += " DESC"
		}
		q = q.Order(orderVal)
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
			cdr.CostDetails.initCache()
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

func (self *SQLStorage) GetTPRates(tpid, id string) ([]*utils.TPRateRALs, error) {
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

func (self *SQLStorage) GetTPResources(tpid, tenant, id string) ([]*utils.TPResourceProfile, error) {
	var rls TpResources
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("id = ?", id)
	}
	if len(tenant) != 0 {
		q = q.Where("tenant = ?", tenant)
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

func (self *SQLStorage) GetTPStats(tpid, tenant, id string) ([]*utils.TPStatProfile, error) {
	var sts TpStats
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("id = ?", id)
	}
	if len(tenant) != 0 {
		q = q.Where("tenant = ?", tenant)
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

func (self *SQLStorage) GetTPThresholds(tpid, tenant, id string) ([]*utils.TPThresholdProfile, error) {
	var ths TpThresholds
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("id = ?", id)
	}
	if len(tenant) != 0 {
		q = q.Where("tenant = ?", tenant)
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

func (self *SQLStorage) GetTPFilters(tpid, tenant, id string) ([]*utils.TPFilterProfile, error) {
	var ths TpFilterS
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("id = ?", id)
	}
	if len(tenant) != 0 {
		q = q.Where("tenant = ?", tenant)
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

func (self *SQLStorage) GetTPRoutes(tpid, tenant, id string) ([]*utils.TPRouteProfile, error) {
	var tpRoutes TPRoutes
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("id = ?", id)
	}
	if len(tenant) != 0 {
		q = q.Where("tenant = ?", tenant)
	}
	if err := q.Find(&tpRoutes).Error; err != nil {
		return nil, err
	}
	aTpRoutes := tpRoutes.AsTPRouteProfile()
	if len(aTpRoutes) == 0 {
		return aTpRoutes, utils.ErrNotFound
	}
	return aTpRoutes, nil
}

func (self *SQLStorage) GetTPAttributes(tpid, tenant, id string) ([]*utils.TPAttributeProfile, error) {
	var sps TPAttributes
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("id = ?", id)
	}
	if len(tenant) != 0 {
		q = q.Where("tenant = ?", tenant)
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

func (self *SQLStorage) GetTPChargers(tpid, tenant, id string) ([]*utils.TPChargerProfile, error) {
	var cpps TPChargers
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("id = ?", id)
	}
	if len(tenant) != 0 {
		q = q.Where("tenant = ?", tenant)
	}
	if err := q.Find(&cpps).Error; err != nil {
		return nil, err
	}
	arls := cpps.AsTPChargers()
	if len(arls) == 0 {
		return arls, utils.ErrNotFound
	}
	return arls, nil
}

func (self *SQLStorage) GetTPDispatcherProfiles(tpid, tenant, id string) ([]*utils.TPDispatcherProfile, error) {
	var dpps TPDispatcherProfiles
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("id = ?", id)
	}
	if len(tenant) != 0 {
		q = q.Where("tenant = ?", tenant)
	}
	if err := q.Find(&dpps).Error; err != nil {
		return nil, err
	}
	arls := dpps.AsTPDispatcherProfiles()
	if len(arls) == 0 {
		return arls, utils.ErrNotFound
	}
	return arls, nil
}

func (self *SQLStorage) GetTPDispatcherHosts(tpid, tenant, id string) ([]*utils.TPDispatcherHost, error) {
	var dpps TPDispatcherHosts
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("id = ?", id)
	}
	if len(tenant) != 0 {
		q = q.Where("tenant = ?", tenant)
	}
	if err := q.Find(&dpps).Error; err != nil {
		return nil, err
	}
	arls := dpps.AsTPDispatcherHosts()
	if len(arls) == 0 {
		return arls, utils.ErrNotFound
	}
	return arls, nil
}

func (self *SQLStorage) GetTPRateProfiles(tpid, tenant, id string) ([]*utils.TPRateProfile, error) {
	var dpps RateProfileMdls
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("id = ?", id)
	}
	if len(tenant) != 0 {
		q = q.Where("tenant = ?", tenant)
	}
	if err := q.Find(&dpps).Error; err != nil {
		return nil, err
	}
	arls := dpps.AsTPRateProfile()
	if len(arls) == 0 {
		return arls, utils.ErrNotFound
	}
	return arls, nil
}

func (self *SQLStorage) GetTPActionProfiles(tpid, tenant, id string) ([]*utils.TPActionProfile, error) {
	var dpps ActionProfileMdls
	q := self.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("id = ?", id)
	}
	if len(tenant) != 0 {
		q = q.Where("tenant = ?", tenant)
	}
	if err := q.Find(&dpps).Error; err != nil {
		return nil, err
	}
	arls := dpps.AsTPActionProfile()
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
