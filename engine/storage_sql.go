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
	"database/sql"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
	"gorm.io/gorm"
)

type SQLImpl interface {
	extraFieldsExistsQry(string) string
	extraFieldsValueQry(string, string) string
	notExtraFieldsExistsQry(string) string
	notExtraFieldsValueQry(string, string) string
	valueQry(string, string, string, []string, bool) []string // will query for every type of filtering in case of needed
	cdrIDQuery(string) string                                 // will get the unique *cdrID for every CDR
	existField(string, string) string                         // will query for every element on json type if the field exists
}

type SQLStorage struct {
	DB *sql.DB
	db *gorm.DB
	StorDB
	SQLImpl
}

func (sqls *SQLStorage) Close() {
	sqls.DB.Close()
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
	if _, err := sqls.DB.Query(fmt.Sprintf("SELECT 1 FROM %s", utils.CDRsTBL)); err != nil {
		return err
	}
	return nil
}

func (sqls *SQLStorage) SelectDatabase(dbName string) (err error) {
	return
}

func (sqls *SQLStorage) GetKeysForPrefix(ctx *context.Context, prefix string) ([]string, error) {
	return nil, utils.ErrNotImplemented
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
		if _, err := sqls.DB.Exec(qry); err != nil {
			return err
		}
	}
	return nil
}

func (sqls *SQLStorage) IsDBEmpty() (resp bool, err error) {
	for _, tbl := range []string{utils.CDRsTBL, utils.TBLVersions} {
		if sqls.db.Migrator().HasTable(tbl) {
			return false, nil
		}
	}
	return true, nil
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

func (sqls *SQLStorage) SetCDR(_ *context.Context, cdr *utils.CGREvent, allowUpdate bool) error {
	tx := sqls.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	cdrTable := &utils.CDRSQLTable{
		Tenant:    cdr.Tenant,
		Opts:      cdr.APIOpts,
		Event:     cdr.Event,
		CreatedAt: time.Now(),
	}
	saved := tx.Save(cdrTable)
	if saved.Error != nil {
		tx.Rollback()
		if !allowUpdate {
			if strings.Contains(saved.Error.Error(), "1062") || strings.Contains(saved.Error.Error(), "duplicate key") { // returns 1062/pq when key is duplicated
				return utils.ErrExists
			}
			return saved.Error
		}
		tx = sqls.db.Begin()
		if tx.Error != nil {
			return tx.Error
		}

		cdrID := utils.IfaceAsString(cdr.APIOpts[utils.MetaCDRID])
		updated := tx.Model(&utils.CDRSQLTable{}).Where(
			sqls.cdrIDQuery(cdrID)).Updates(
			utils.CDRSQLTable{Opts: cdr.APIOpts, Event: cdr.Event, UpdatedAt: time.Now()})
		if updated.Error != nil {
			tx.Rollback()
			return updated.Error
		}
	}
	tx.Commit()
	return nil
}

// GetCDRs has ability to get the filtered CDRs, count them or simply return them
// qryFltr.Unscoped will ignore soft deletes or delete records permanently
func (sqls *SQLStorage) GetCDRs(ctx *context.Context, qryFltr []*Filter, opts map[string]any) ([]*utils.CDR, error) {
	q := sqls.db.Table(utils.CDRsTBL)
	var excludedCdrQueryFilterTypes []*FilterRule
	for _, fltr := range qryFltr {
		for _, rule := range fltr.Rules {
			if !cdrQueryFilterTypes.Has(rule.Type) || checkNestedFields(rule.Element, rule.Values) {
				excludedCdrQueryFilterTypes = append(excludedCdrQueryFilterTypes, rule)
				continue
			}
			var elem, field string
			switch {
			case strings.HasPrefix(rule.Element, utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep):
				elem = "event"
				field = strings.TrimPrefix(rule.Element, utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep)
			case strings.HasPrefix(rule.Element, utils.DynamicDataPrefix+utils.MetaOpts+utils.NestingSep):
				elem = "opts"
				field = strings.TrimPrefix(rule.Element, utils.DynamicDataPrefix+utils.MetaOpts+utils.NestingSep)
			}
			var count int64
			if _ = sqls.db.Table(utils.CDRsTBL).Where(
				sqls.existField(elem, field)).Count(&count); count > 0 &&
				(rule.Type == utils.MetaNotExists ||
					rule.Type == utils.MetaNotString) {
				continue
			}
			conditions := sqls.valueQry(rule.Type, elem, field, rule.Values, strings.HasPrefix(rule.Type, utils.MetaNot))
			q.Where(strings.Join(conditions, " OR "))
		}
	}

	limit, offset, maxItems, err := utils.GetPaginateOpts(opts)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve paginator opts: %w", err)
	}
	if maxItems < limit+offset {
		return nil, fmt.Errorf("sum of limit and offset exceeds maxItems")
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}

	// Execute query
	results := make([]*utils.CDRSQLTable, 0)
	if err = q.Find(&results).Error; err != nil {
		return nil, err
	}

	//convert into CDR
	cdrs := make([]*utils.CDR, 0, len(results))
	for _, val := range results {
		cdr := &utils.CDR{
			Tenant:    val.Tenant,
			Opts:      val.Opts,
			Event:     val.Event,
			CreatedAt: val.CreatedAt,
			UpdatedAt: val.UpdatedAt,
			DeletedAt: val.DeletedAt,
		}
		// here we wil do our filtration, meaning that we will filter those cdrs who cannot be filtered in the databes eg: *ai, *rsr..
		if len(excludedCdrQueryFilterTypes) != 0 {
			var pass bool
			dP := cdr.CGREvent().AsDataProvider()
			for _, fltr := range excludedCdrQueryFilterTypes {
				if pass, err = fltr.Pass(ctx, dP); err != nil {
					return nil, err
				} else if !pass {
					break
				}
			}
			// if the cdr passed the filtration, get it as result, else continue
			if !pass {
				continue
			}
		}
		cdrs = append(cdrs, cdr)
	}
	if len(cdrs) == 0 {
		return nil, utils.ErrNotFound
	}
	return cdrs, nil
}

func (sqls *SQLStorage) RemoveCDRs(ctx *context.Context, qryFltr []*Filter) (err error) {
	q := sqls.db.Table(utils.CDRsTBL)
	var excludedCdrQueryFilterTypes []*FilterRule
	for _, fltr := range qryFltr {
		for _, rule := range fltr.Rules {
			if !cdrQueryFilterTypes.Has(rule.Type) || checkNestedFields(rule.Element, rule.Values) {
				excludedCdrQueryFilterTypes = append(excludedCdrQueryFilterTypes, rule)
				continue
			}
			var elem, field string
			switch {
			case strings.HasPrefix(rule.Element, utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep):
				elem = "event"
				field = strings.TrimPrefix(rule.Element, utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep)
			case strings.HasPrefix(rule.Element, utils.DynamicDataPrefix+utils.MetaOpts+utils.NestingSep):
				elem = "opts"
				field = strings.TrimPrefix(rule.Element, utils.DynamicDataPrefix+utils.MetaOpts+utils.NestingSep)
			}
			var count int64
			if _ = sqls.db.Table(utils.CDRsTBL).Where(
				sqls.existField(elem, field)).Count(&count); count > 0 &&
				(rule.Type == utils.MetaNotExists ||
					rule.Type == utils.MetaNotString) {
				continue
			}
			conditions := sqls.valueQry(rule.Type, elem, field, rule.Values, strings.HasPrefix(rule.Type, utils.MetaNot))
			q.Where(strings.Join(conditions, " OR "))
		}
	}
	// if we do not have any filters that cannot be queried in database, just delete all the results (e.g. *rsr, *ai, *cronexp ..))
	if len(excludedCdrQueryFilterTypes) == 0 {
		if err = q.Delete(nil).Error; err != nil {
			q.Rollback()
			return err
		}
		return
	}
	// in the other case, if we have such filters, check the results based on those filters
	results := make([]*utils.CDRSQLTable, 0)
	if err = q.Find(&results).Error; err != nil {
		return
	}
	// this means nothing in database matched, so we will not check the filtration process
	if len(results) == 0 {
		return
	}
	// keep the result for quering with other filter type that are not allowed in database
	q = sqls.db.Table(utils.CDRsTBL)          // reset the query
	remCdr := make([]string, 0, len(results)) // we will keep the *cdrID of every CDR taht matched the those filters
	for _, cdr := range results {
		if len(excludedCdrQueryFilterTypes) != 0 {
			newCdr := &utils.CDR{
				Tenant: cdr.Tenant,
				Opts:   cdr.Opts,
				Event:  cdr.Event,
			}
			var pass bool
			dP := newCdr.CGREvent().AsDataProvider()
			// check if the filter pass
			for _, fltr := range excludedCdrQueryFilterTypes {
				if pass, err = fltr.Pass(ctx, dP); err != nil {
					return err
				} else if !pass {
					break
				}
			}
			if pass {
				// if the filters passed, remove the CDR by it's *cdrID
				remCdr = append(remCdr, sqls.cdrIDQuery(utils.IfaceAsString(newCdr.Opts[utils.MetaCDRID])))
			}
		}
	}
	// this means nothing PASSED trough filtration process, so nothing will be deleted
	if len(remCdr) == 0 {
		return
	}
	q.Where(strings.Join(remCdr, " OR "))
	if err = q.Delete(nil).Error; err != nil {
		q.Rollback()
		return err
	}
	return
}
