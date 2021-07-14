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
	"os"
	"path"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
	"gorm.io/gorm"
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

func (sqls *SQLStorage) Close() {
	sqls.Db.Close()
	// sqls.db
}

func (sqls *SQLStorage) ExportGormDB() *gorm.DB {
	return sqls.db
}

func (sqls *SQLStorage) Flush(scriptsPath string) (err error) {
	for _, scriptName := range []string{utils.CreateCDRsTablesSQL, utils.CreateTariffPlanTablesSQL} {
		if err := sqls.CreateTablesFromScript(path.Join(scriptsPath, scriptName)); err != nil {
			return err
		}
	}
	if _, err := sqls.Db.Query(fmt.Sprintf("SELECT 1 FROM %s", utils.CDRsTBL)); err != nil {
		return err
	}
	return nil
}

func (rs *SQLStorage) SelectDatabase(dbName string) (err error) {
	return
}

func (sqls *SQLStorage) GetKeysForPrefix(prefix string) ([]string, error) {
	return nil, utils.ErrNotImplemented
}

func (SQLStorage) RemoveKeysForPrefix(string) error {
	return utils.ErrNotImplemented
}

func (sqls *SQLStorage) CreateTablesFromScript(scriptPath string) error {
	fileContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return err
	}
	qries := strings.Split(string(fileContent), ";") // Script has normally multiple queries separate by ';' go driver does not understand this so we handle it here
	for _, qry := range qries {
		qry = strings.TrimSpace(qry) // Avoid empty queries
		if len(qry) == 0 {
			continue
		}
		if _, err := sqls.Db.Exec(qry); err != nil {
			return err
		}
	}
	return nil
}

func (sqls *SQLStorage) IsDBEmpty() (resp bool, err error) {
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
		if sqls.db.Migrator().HasTable(tbl) {
			return false, nil
		}

	}
	return true, nil
}

// update
// Return a list with all TPids defined in the system, even if incomplete, isolated in some table.
func (sqls *SQLStorage) GetTpIds(colName string) ([]string, error) {
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
	rows, err = sqls.Db.Query(qryStr)
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
func (sqls *SQLStorage) GetTpTableIds(tpid, table string, distinct utils.TPDistinctIds,
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
			qry += ")"
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
	rows, err := sqls.Db.Query(qry)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	ids := []string{}
	i := 0
	for rows.Next() {
		i++ //Keep here a reference so we know we got at least one

		cols, err := rows.Columns() // Get the column names; remember to check err
		if err != nil {
			return nil, err
		}
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
			finalID = strings.Join(vals, utils.ConcatenatedKeySep)
		}
		ids = append(ids, finalID)
	}
	if i == 0 {
		return nil, nil
	}
	return ids, nil
}

func (sqls *SQLStorage) RemTpData(table, tpid string, args map[string]string) error {
	tx := sqls.db.Begin()

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

func (sqls *SQLStorage) SetTPTimings(timings []*utils.ApierTPTiming) error {
	if len(timings) == 0 {
		return nil
	}

	tx := sqls.db.Begin()
	for _, timing := range timings {
		if err := tx.Where(&TimingMdl{Tpid: timing.TPid, Tag: timing.ID}).Delete(TimingMdl{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		t := APItoModelTiming(timing)
		if err := tx.Create(&t).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) SetTPDestinations(dests []*utils.TPDestination) error {
	if len(dests) == 0 {
		return nil
	}
	tx := sqls.db.Begin()
	for _, dst := range dests {
		// Remove previous
		if err := tx.Where(&DestinationMdl{Tpid: dst.TPid, Tag: dst.ID}).Delete(DestinationMdl{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, d := range APItoModelDestination(dst) {
			if err := tx.Create(&d).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) SetTPRates(rs []*utils.TPRateRALs) error {
	if len(rs) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)
	tx := sqls.db.Begin()
	for _, rate := range rs {
		if !m[rate.ID] {
			m[rate.ID] = true
			if err := tx.Where(&RateMdl{Tpid: rate.TPid, Tag: rate.ID}).Delete(RateMdl{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, r := range APItoModelRate(rate) {
			if err := tx.Create(&r).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) SetTPDestinationRates(drs []*utils.TPDestinationRate) error {
	if len(drs) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)
	tx := sqls.db.Begin()
	for _, dRate := range drs {
		if !m[dRate.ID] {
			m[dRate.ID] = true
			if err := tx.Where(&DestinationRateMdl{Tpid: dRate.TPid, Tag: dRate.ID}).Delete(DestinationRateMdl{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, d := range APItoModelDestinationRate(dRate) {
			if err := tx.Create(&d).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) SetTPRatingPlans(rps []*utils.TPRatingPlan) error {
	if len(rps) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)
	tx := sqls.db.Begin()
	for _, rPlan := range rps {
		if !m[rPlan.ID] {
			m[rPlan.ID] = true
			if err := tx.Where(&RatingPlanMdl{Tpid: rPlan.TPid, Tag: rPlan.ID}).Delete(RatingPlanMdl{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, r := range APItoModelRatingPlan(rPlan) {
			if err := tx.Create(&r).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) SetTPRatingProfiles(rpfs []*utils.TPRatingProfile) error {
	if len(rpfs) == 0 {
		return nil //Nothing to set
	}
	tx := sqls.db.Begin()
	for _, rpf := range rpfs {
		if err := tx.Where(&RatingProfileMdl{Tpid: rpf.TPid, Loadid: rpf.LoadId,
			Tenant: rpf.Tenant, Category: rpf.Category,
			Subject: rpf.Subject}).Delete(RatingProfileMdl{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, r := range APItoModelRatingProfile(rpf) {
			if err := tx.Create(&r).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) SetTPSharedGroups(sgs []*utils.TPSharedGroups) error {
	if len(sgs) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)
	tx := sqls.db.Begin()
	for _, sGroup := range sgs {
		if !m[sGroup.ID] {
			m[sGroup.ID] = true
			if err := tx.Where(&SharedGroupMdl{Tpid: sGroup.TPid, Tag: sGroup.ID}).Delete(SharedGroupMdl{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, s := range APItoModelSharedGroup(sGroup) {
			if err := tx.Create(&s).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) SetTPActions(acts []*utils.TPActions) error {
	if len(acts) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)
	tx := sqls.db.Begin()
	for _, a := range acts {
		if !m[a.ID] {
			m[a.ID] = true
			if err := tx.Where(&ActionMdl{Tpid: a.TPid, Tag: a.ID}).Delete(ActionMdl{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, sa := range APItoModelAction(a) {
			if err := tx.Create(&sa).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) SetTPActionPlans(ats []*utils.TPActionPlan) error {
	if len(ats) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)

	tx := sqls.db.Begin()
	for _, aPlan := range ats {
		if !m[aPlan.ID] {
			m[aPlan.ID] = true
			if err := tx.Where(&ActionPlanMdl{Tpid: aPlan.TPid, Tag: aPlan.ID}).Delete(ActionPlanMdl{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, a := range APItoModelActionPlan(aPlan) {
			if err := tx.Create(&a).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	r := tx.Commit()
	return r.Error
}

func (sqls *SQLStorage) SetTPActionTriggers(ats []*utils.TPActionTriggers) error {
	if len(ats) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)
	tx := sqls.db.Begin()
	for _, aTrigger := range ats {
		if !m[aTrigger.ID] {
			m[aTrigger.ID] = true
			if err := tx.Where(&ActionTriggerMdl{Tpid: aTrigger.TPid, Tag: aTrigger.ID}).Delete(ActionTriggerMdl{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, a := range APItoModelActionTrigger(aTrigger) {
			if err := tx.Create(&a).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

// Sets a group of account actions. Map key has the role of grouping within a tpid
func (sqls *SQLStorage) SetTPAccountActions(aas []*utils.TPAccountActions) error {
	if len(aas) == 0 {
		return nil //Nothing to set
	}
	m := make(map[string]bool)

	tx := sqls.db.Begin()
	for _, aa := range aas {
		if !m[aa.GetId()] {
			m[aa.GetId()] = true
			if err := tx.Where(&AccountActionMdl{Tpid: aa.TPid, Loadid: aa.LoadId, Tenant: aa.Tenant, Account: aa.Account}).Delete(&AccountActionMdl{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		sa := APItoModelAccountAction(aa)
		if err := tx.Create(&sa).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) SetTPResources(rls []*utils.TPResourceProfile) error {
	if len(rls) == 0 {
		return nil
	}
	tx := sqls.db.Begin()
	for _, rl := range rls {
		// Remove previous
		if err := tx.Where(&ResourceMdl{Tpid: rl.TPid, ID: rl.ID}).Delete(ResourceMdl{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, mrl := range APItoModelResource(rl) {
			if err := tx.Create(&mrl).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) SetTPStats(sts []*utils.TPStatProfile) error {
	if len(sts) == 0 {
		return nil
	}
	tx := sqls.db.Begin()
	for _, stq := range sts {
		// Remove previous
		if err := tx.Where(&StatMdl{Tpid: stq.TPid, ID: stq.ID}).Delete(StatMdl{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, mst := range APItoModelStats(stq) {
			if err := tx.Create(&mst).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) SetTPThresholds(ths []*utils.TPThresholdProfile) error {
	if len(ths) == 0 {
		return nil
	}
	tx := sqls.db.Begin()
	for _, th := range ths {
		// Remove previous
		if err := tx.Where(&ThresholdMdl{Tpid: th.TPid, ID: th.ID}).Delete(ThresholdMdl{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, mst := range APItoModelTPThreshold(th) {
			if err := tx.Create(&mst).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) SetTPFilters(ths []*utils.TPFilterProfile) error {
	if len(ths) == 0 {
		return nil
	}
	tx := sqls.db.Begin()
	for _, th := range ths {
		// Remove previous
		if err := tx.Where(&FilterMdl{Tpid: th.TPid, ID: th.ID}).Delete(FilterMdl{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, mst := range APItoModelTPFilter(th) {
			if err := tx.Create(&mst).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) SetTPRoutes(tpRoutes []*utils.TPRouteProfile) error {
	if len(tpRoutes) == 0 {
		return nil
	}
	tx := sqls.db.Begin()
	for _, tpRoute := range tpRoutes {
		// Remove previous
		if err := tx.Where(&RouteMdl{Tpid: tpRoute.TPid, ID: tpRoute.ID}).Delete(RouteMdl{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, mst := range APItoModelTPRoutes(tpRoute) {
			if err := tx.Create(&mst).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) SetTPAttributes(tpAttrs []*utils.TPAttributeProfile) error {
	if len(tpAttrs) == 0 {
		return nil
	}
	tx := sqls.db.Begin()
	for _, stq := range tpAttrs {
		// Remove previous
		if err := tx.Where(&AttributeMdl{Tpid: stq.TPid, ID: stq.ID}).Delete(AttributeMdl{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, mst := range APItoModelTPAttribute(stq) {
			if err := tx.Create(&mst).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) SetTPChargers(tpCPPs []*utils.TPChargerProfile) error {
	if len(tpCPPs) == 0 {
		return nil
	}
	tx := sqls.db.Begin()
	for _, cpp := range tpCPPs {
		// Remove previous
		if err := tx.Where(&ChargerMdl{Tpid: cpp.TPid, ID: cpp.ID}).Delete(ChargerMdl{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, mst := range APItoModelTPCharger(cpp) {
			if err := tx.Create(&mst).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) SetTPDispatcherProfiles(tpDPPs []*utils.TPDispatcherProfile) error {
	if len(tpDPPs) == 0 {
		return nil
	}
	tx := sqls.db.Begin()
	for _, dpp := range tpDPPs {
		// Remove previous
		if err := tx.Where(&DispatcherProfileMdl{Tpid: dpp.TPid, ID: dpp.ID}).Delete(DispatcherProfileMdl{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, mst := range APItoModelTPDispatcherProfile(dpp) {
			if err := tx.Create(&mst).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) SetTPDispatcherHosts(tpDPPs []*utils.TPDispatcherHost) error {
	if len(tpDPPs) == 0 {
		return nil
	}
	tx := sqls.db.Begin()
	for _, dpp := range tpDPPs {
		// Remove previous
		if err := tx.Where(&DispatcherHostMdl{Tpid: dpp.TPid, ID: dpp.ID}).Delete(DispatcherHostMdl{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Create(APItoModelTPDispatcherHost(dpp)).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) SetSMCost(smc *SMCost) error {
	if smc.CostDetails == nil {
		return nil
	}
	tx := sqls.db.Begin()
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

func (sqls *SQLStorage) RemoveSMCost(smc *SMCost) error {
	tx := sqls.db.Begin()
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

func (sqls *SQLStorage) RemoveSMCosts(qryFltr *utils.SMCostFilter) error {
	q := sqls.db.Table(utils.SessionCostsTBL).Select("*")
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
		if sqls.db.Dialector.Name() == utils.MySQL { // MySQL needs escaping for usage
			q = q.Where("`usage` >= ?", qryFltr.Usage.Min.Nanoseconds())
		} else {
			q = q.Where("usage >= ?", qryFltr.Usage.Min.Nanoseconds())
		}
	}
	if qryFltr.Usage.Max != nil {
		if sqls.db.Dialector.Name() == utils.MySQL { // MySQL needs escaping for usage
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
func (sqls *SQLStorage) GetSMCosts(cgrid, runid, originHost, originIDPrefix string) ([]*SMCost, error) {
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
	q := sqls.db.Where(filter)
	if originIDPrefix != "" {
		q = sqls.db.Where(filter).Where(fmt.Sprintf("origin_id LIKE '%s%%'", originIDPrefix))
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

func (sqls *SQLStorage) SetCDR(cdr *CDR, allowUpdate bool) error {
	tx := sqls.db.Begin()
	cdrSQL := cdr.AsCDRsql()
	cdrSQL.CreatedAt = time.Now()
	saved := tx.Save(cdrSQL)
	if saved.Error != nil {
		tx.Rollback()
		if !allowUpdate {
			if strings.Contains(saved.Error.Error(), "1062") || strings.Contains(saved.Error.Error(), "duplicate key") { // returns 1062/pq when key is duplicated
				return utils.ErrExists
			}
			return saved.Error
		}
		tx = sqls.db.Begin()
		cdrSQL.UpdatedAt = time.Now()
		updated := tx.Model(&CDRsql{}).Where(
			&CDRsql{Cgrid: cdr.CGRID, RunID: cdr.RunID, OriginID: cdr.OriginID}).Updates(cdrSQL.AsMapStringInterface())
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
func (sqls *SQLStorage) GetCDRs(qryFltr *utils.CDRsFilter, remove bool) ([]*CDR, int64, error) {
	var cdrs []*CDR
	q := sqls.db.Table(utils.CDRsTBL)
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
				qIds.WriteString(sqls.SQLImpl.extraFieldsExistsQry(field))
			} else {
				qIds.WriteString(sqls.SQLImpl.extraFieldsValueQry(field, value))
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
				qIds.WriteString(sqls.SQLImpl.notExtraFieldsExistsQry(field))
			} else {
				qIds.WriteString(sqls.SQLImpl.notExtraFieldsValueQry(field, value))
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
		separateVals := strings.Split(qryFltr.OrderBy, utils.InfieldSep)
		switch separateVals[0] {
		case utils.OrderID:
			orderVal = "id"
		case utils.AnswerTime:
			orderVal = "answer_time"
		case utils.SetupTime:
			orderVal = "setup_time"
		case utils.Usage:
			if sqls.db.Dialector.Name() == utils.MySQL {
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
		if sqls.db.Dialector.Name() == utils.MySQL { // MySQL needs escaping for usage
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
		if sqls.db.Dialector.Name() == utils.MySQL { // MySQL needs escaping for usage
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

func (sqls *SQLStorage) GetTPDestinations(tpid, id string) (uTPDsts []*utils.TPDestination, err error) {
	var tpDests DestinationMdls
	q := sqls.db.Where("tpid = ?", tpid)
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

func (sqls *SQLStorage) GetTPRates(tpid, id string) ([]*utils.TPRateRALs, error) {
	var tpRates RateMdls
	q := sqls.db.Where("tpid = ?", tpid).Order("id")
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

func (sqls *SQLStorage) GetTPDestinationRates(tpid, id string, pagination *utils.Paginator) ([]*utils.TPDestinationRate, error) {
	var tpDestinationRates DestinationRateMdls
	q := sqls.db.Where("tpid = ?", tpid)
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
	drs := tpDestinationRates.AsTPDestinationRates()
	if len(drs) == 0 {
		return drs, utils.ErrNotFound
	}
	return drs, nil
}

func (sqls *SQLStorage) GetTPTimings(tpid, id string) ([]*utils.ApierTPTiming, error) {
	var tpTimings TimingMdls
	q := sqls.db.Where("tpid = ?", tpid)
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

func (sqls *SQLStorage) GetTPRatingPlans(tpid, id string, pagination *utils.Paginator) ([]*utils.TPRatingPlan, error) {
	var tpRatingPlans RatingPlanMdls
	q := sqls.db.Where("tpid = ?", tpid)
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
	rps := tpRatingPlans.AsTPRatingPlans()

	if len(rps) == 0 {
		return rps, utils.ErrNotFound
	}
	return rps, nil
}

func (sqls *SQLStorage) GetTPRatingProfiles(filter *utils.TPRatingProfile) ([]*utils.TPRatingProfile, error) {
	var tpRpfs RatingProfileMdls
	q := sqls.db.Where("tpid = ?", filter.TPid)
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
	rps := tpRpfs.AsTPRatingProfiles()
	if len(rps) == 0 {
		return rps, utils.ErrNotFound
	}
	return rps, nil
}

func (sqls *SQLStorage) GetTPSharedGroups(tpid, id string) ([]*utils.TPSharedGroups, error) {
	var tpShareGroups SharedGroupMdls
	q := sqls.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("tag = ?", id)
	}
	if err := q.Find(&tpShareGroups).Error; err != nil {
		return nil, err
	}
	sgs := tpShareGroups.AsTPSharedGroups()

	if len(sgs) == 0 {
		return sgs, utils.ErrNotFound
	}
	return sgs, nil
}

func (sqls *SQLStorage) GetTPActions(tpid, id string) ([]*utils.TPActions, error) {
	var tpActions ActionMdls
	q := sqls.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("tag = ?", id)
	}
	if err := q.Find(&tpActions).Error; err != nil {
		return nil, err
	}
	as := tpActions.AsTPActions()
	if len(as) == 0 {
		return as, utils.ErrNotFound
	}
	return as, nil
}

func (sqls *SQLStorage) GetTPActionTriggers(tpid, id string) ([]*utils.TPActionTriggers, error) {
	var tpActionTriggers ActionTriggerMdls
	q := sqls.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("tag = ?", id)
	}
	if err := q.Find(&tpActionTriggers).Error; err != nil {
		return nil, err
	}
	ats := tpActionTriggers.AsTPActionTriggers()
	if len(ats) == 0 {
		return ats, utils.ErrNotFound
	}
	return ats, nil
}

func (sqls *SQLStorage) GetTPActionPlans(tpid, id string) ([]*utils.TPActionPlan, error) {
	var tpActionPlans ActionPlanMdls
	q := sqls.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("tag = ?", id)
	}
	if err := q.Find(&tpActionPlans).Error; err != nil {
		return nil, err
	}
	aps := tpActionPlans.AsTPActionPlans()
	if len(aps) == 0 {
		return aps, utils.ErrNotFound
	}
	return aps, nil
}

func (sqls *SQLStorage) GetTPAccountActions(filter *utils.TPAccountActions) ([]*utils.TPAccountActions, error) {
	var tpAccActs AccountActionMdls
	q := sqls.db.Where("tpid = ?", filter.TPid)
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
	aas := tpAccActs.AsTPAccountActions()
	if len(aas) == 0 {
		return aas, utils.ErrNotFound
	}
	return aas, nil
}

func (sqls *SQLStorage) GetTPResources(tpid, tenant, id string) ([]*utils.TPResourceProfile, error) {
	var rls ResourceMdls
	q := sqls.db.Where("tpid = ?", tpid)
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

func (sqls *SQLStorage) GetTPStats(tpid, tenant, id string) ([]*utils.TPStatProfile, error) {
	var sts StatMdls
	q := sqls.db.Where("tpid = ?", tpid)
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

func (sqls *SQLStorage) GetTPThresholds(tpid, tenant, id string) ([]*utils.TPThresholdProfile, error) {
	var ths ThresholdMdls
	q := sqls.db.Where("tpid = ?", tpid)
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

func (sqls *SQLStorage) GetTPFilters(tpid, tenant, id string) ([]*utils.TPFilterProfile, error) {
	var ths FilterMdls
	q := sqls.db.Where("tpid = ?", tpid)
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

func (sqls *SQLStorage) GetTPRoutes(tpid, tenant, id string) ([]*utils.TPRouteProfile, error) {
	var tpRoutes RouteMdls
	q := sqls.db.Where("tpid = ?", tpid)
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

func (sqls *SQLStorage) GetTPAttributes(tpid, tenant, id string) ([]*utils.TPAttributeProfile, error) {
	var sps AttributeMdls
	q := sqls.db.Where("tpid = ?", tpid)
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

func (sqls *SQLStorage) GetTPChargers(tpid, tenant, id string) ([]*utils.TPChargerProfile, error) {
	var cpps ChargerMdls
	q := sqls.db.Where("tpid = ?", tpid)
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

func (sqls *SQLStorage) GetTPDispatcherProfiles(tpid, tenant, id string) ([]*utils.TPDispatcherProfile, error) {
	var dpps DispatcherProfileMdls
	q := sqls.db.Where("tpid = ?", tpid)
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

func (sqls *SQLStorage) GetTPDispatcherHosts(tpid, tenant, id string) ([]*utils.TPDispatcherHost, error) {
	var dpps DispatcherHostMdls
	q := sqls.db.Where("tpid = ?", tpid)
	if len(id) != 0 {
		q = q.Where("id = ?", id)
	}
	if len(tenant) != 0 {
		q = q.Where("tenant = ?", tenant)
	}
	if err := q.Find(&dpps).Error; err != nil {
		return nil, err
	}
	arls, err := dpps.AsTPDispatcherHosts()
	if err != nil {
		return nil, err
	}
	if len(arls) == 0 {
		return arls, utils.ErrNotFound
	}
	return arls, nil
}

// GetVersions returns slice of all versions or a specific version if tag is specified
func (sqls *SQLStorage) GetVersions(itm string) (vrs Versions, err error) {
	q := sqls.db.Model(&TBLVersion{})
	if itm != utils.TBLVersions && itm != "" {
		q = sqls.db.Where(&TBLVersion{Item: itm})
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
func (sqls *SQLStorage) RemoveVersions(vrs Versions) (err error) {
	if len(vrs) == 0 { // Remove all if no key provided
		err = sqls.db.Delete(TBLVersion{}).Error
		return
	}
	tx := sqls.db.Begin()
	for key := range vrs {
		if err = tx.Where(&TBLVersion{Item: key}).Delete(TBLVersion{}).Error; err != nil {
			tx.Rollback()
			return
		}
	}
	tx.Commit()
	return
}
