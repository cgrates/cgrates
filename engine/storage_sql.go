/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
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
	DataDB
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
	for _, scriptName := range []string{utils.CreateAccountsTablesSQL,
		utils.CreateCDRsTablesSQL, utils.CreateTariffPlanTablesSQL} {
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

// returns all keys in table matching the Tenant and ID
func (sqls *SQLStorage) getAllKeysMatchingTenantID(_ *context.Context, table string, tntID *utils.TenantID) (ids []string, err error) {
	matchingTntID := []utils.TenantID{}
	if err = sqls.db.Table(table).Select("tenant, id").Where("tenant = ? AND id LIKE ?", tntID.Tenant, tntID.ID+"%").
		Find(&matchingTntID).Error; err != nil {
		return nil, err
	}
	ids = make([]string, len(matchingTntID))
	for i, result := range matchingTntID {
		ids[i] = utils.ConcatenatedKey(result.Tenant, result.ID)
	}
	return
}

// GetKeysForPrefix will look for keys matching the prefix given
func (sqls *SQLStorage) GetKeysForPrefix(ctx *context.Context, prefix string) (keys []string, err error) {
	keyLen := len(utils.AccountPrefix)
	if len(prefix) < keyLen {
		return nil, fmt.Errorf("unsupported prefix in GetKeysForPrefix: %q", prefix)
	}
	category := prefix[:keyLen]
	tntID := utils.NewTenantID(prefix[keyLen:])

	switch category {
	case utils.AccountPrefix:
		keys, err = sqls.getAllKeysMatchingTenantID(ctx, utils.TBLAccounts, tntID)
	case utils.IPProfilesPrefix:
		keys, err = sqls.getAllKeysMatchingTenantID(ctx, utils.TBLIPProfiles, tntID)
	case utils.IPAllocationsPrefix:
		keys, err = sqls.getAllKeysMatchingTenantID(ctx, utils.TBLIPAllocations, tntID)
	case utils.ActionProfilePrefix:
		keys, err = sqls.getAllKeysMatchingTenantID(ctx, utils.TBLActionProfiles, tntID)
	case utils.ChargerProfilePrefix:
		keys, err = sqls.getAllKeysMatchingTenantID(ctx, utils.TBLChargerProfiles, tntID)
	case utils.AttributeProfilePrefix:
		keys, err = sqls.getAllKeysMatchingTenantID(ctx, utils.TBLAttributeProfiles, tntID)
	default:
		err = fmt.Errorf("unsupported prefix in GetKeysForPrefix: %q", prefix)
	}
	for i := range keys { // bring the prefix back to match redis style keys to satisfy functions using it
		keys[i] = category + keys[i]
	}
	return keys, err
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

// GetAccountDrv will get the account from the DB matching the tenant and id provided.
// Decimal fields ending in `.0` will be read as whole numbers but still in decimal type.
// (50.0 -> 50)
func (sqls *SQLStorage) GetAccountDrv(ctx *context.Context, tenant, id string) (ap *utils.Account, err error) {
	var result []*AccountJSONMdl
	if err = sqls.db.Model(&AccountJSONMdl{}).Where(&AccountJSONMdl{Tenant: tenant,
		ID: id}).Find(&result).Error; err != nil {
		return
	}
	if len(result) == 0 {
		return nil, utils.ErrNotFound
	}
	return utils.MapStringInterfaceToAccount(result[0].Account)
}

// SetAccountDrv will set in DB the provided Account
func (sqls *SQLStorage) SetAccountDrv(ctx *context.Context, ap *utils.Account) (err error) {
	tx := sqls.db.Begin()
	mdl := &AccountJSONMdl{
		Tenant:  ap.Tenant,
		ID:      ap.ID,
		Account: ap.AsMapStringInterface(),
	}
	if err = tx.Model(&AccountJSONMdl{}).Where(
		AccountJSONMdl{Tenant: mdl.Tenant, ID: mdl.ID}).Delete(
		AccountJSONMdl{Account: mdl.Account}).Error; err != nil {
		tx.Rollback()
		return
	}
	if err = tx.Save(mdl).Error; err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
	return
}

// RemoveAccountDrv will remove from DB the account matching the tenamt and id provided
func (sqls *SQLStorage) RemoveAccountDrv(ctx *context.Context, tenant, id string) (err error) {
	tx := sqls.db.Begin()
	if err = tx.Model(&AccountJSONMdl{}).Where(&AccountJSONMdl{Tenant: tenant, ID: id}).
		Delete(&AccountJSONMdl{}).Error; err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
	return
}

func (sqls *SQLStorage) GetIPProfileDrv(ctx *context.Context, tenant, id string) (*utils.IPProfile, error) {
	var result []*IPProfileMdl
	if err := sqls.db.Model(&IPProfileMdl{}).Where(&IPProfileMdl{Tenant: tenant,
		ID: id}).Find(&result).Error; err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, utils.ErrNotFound
	}
	return utils.MapStringInterfaceToIPProfile(result[0].IPProfile)
}

func (sqls *SQLStorage) SetIPProfileDrv(ctx *context.Context, ipp *utils.IPProfile) error {
	tx := sqls.db.Begin()
	mdl := &IPProfileMdl{
		Tenant:    ipp.Tenant,
		ID:        ipp.ID,
		IPProfile: ipp.AsMapStringInterface(),
	}
	if err := tx.Model(&IPProfileMdl{}).Where(
		IPProfileMdl{Tenant: mdl.Tenant, ID: mdl.ID}).Delete(
		IPProfileMdl{IPProfile: mdl.IPProfile}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Save(mdl).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) RemoveIPProfileDrv(ctx *context.Context, tenant, id string) error {
	tx := sqls.db.Begin()
	if err := tx.Model(&IPProfileMdl{}).Where(&IPProfileMdl{Tenant: tenant, ID: id}).
		Delete(&IPProfileMdl{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) GetIPAllocationsDrv(ctx *context.Context, tenant, id string) (*utils.IPAllocations, error) {
	var result []*IPAllocationMdl
	if err := sqls.db.Model(&IPAllocationMdl{}).Where(&IPAllocationMdl{Tenant: tenant,
		ID: id}).Find(&result).Error; err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, utils.ErrNotFound
	}
	return utils.MapStringInterfaceToIPAllocations(result[0].IPAllocation), nil
}

func (sqls *SQLStorage) SetIPAllocationsDrv(ctx *context.Context, ip *utils.IPAllocations) error {
	tx := sqls.db.Begin()
	mdl := &IPAllocationMdl{
		Tenant:       ip.Tenant,
		ID:           ip.ID,
		IPAllocation: ip.AsMapStringInterface(),
	}
	if err := tx.Model(&IPAllocationMdl{}).Where(
		IPAllocationMdl{Tenant: mdl.Tenant, ID: mdl.ID}).Delete(
		IPAllocationMdl{IPAllocation: mdl.IPAllocation}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Save(mdl).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) RemoveIPAllocationsDrv(ctx *context.Context, tenant, id string) error {
	tx := sqls.db.Begin()
	if err := tx.Model(&IPAllocationMdl{}).Where(&IPAllocationMdl{Tenant: tenant, ID: id}).
		Delete(&IPAllocationMdl{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) GetActionProfileDrv(ctx *context.Context, tenant, id string) (ap *utils.ActionProfile, err error) {
	var result []*ActionProfileJSONMdl
	if err := sqls.db.Model(&ActionProfileJSONMdl{}).Where(&ActionProfileJSONMdl{Tenant: tenant,
		ID: id}).Find(&result).Error; err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, utils.ErrNotFound
	}
	return utils.MapStringInterfaceToActionProfile(result[0].ActionProfile)
}

func (sqls *SQLStorage) SetActionProfileDrv(ctx *context.Context, ap *utils.ActionProfile) (err error) {
	tx := sqls.db.Begin()
	mdl := &ActionProfileJSONMdl{
		Tenant:        ap.Tenant,
		ID:            ap.ID,
		ActionProfile: ap.AsMapStringInterface(),
	}
	if err := tx.Model(&ActionProfileJSONMdl{}).Where(
		ActionProfileJSONMdl{Tenant: mdl.Tenant, ID: mdl.ID}).Delete(
		ActionProfileJSONMdl{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Save(mdl).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) RemoveActionProfileDrv(ctx *context.Context, tenant, id string) (err error) {
	tx := sqls.db.Begin()
	if err := tx.Model(&ActionProfileJSONMdl{}).Where(&ActionProfileJSONMdl{Tenant: tenant, ID: id}).
		Delete(&ActionProfileJSONMdl{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) GetChargerProfileDrv(_ *context.Context, tenant, id string) (cp *utils.ChargerProfile, err error) {
	var result []*ChargerProfileMdl
	if err := sqls.db.Model(&ChargerProfileMdl{}).Where(&ChargerProfileMdl{Tenant: tenant,
		ID: id}).Find(&result).Error; err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, utils.ErrNotFound
	}

	return utils.MapStringInterfaceToChargerProfile(result[0].ChargerProfile)
}

func (sqls *SQLStorage) SetChargerProfileDrv(_ *context.Context, cp *utils.ChargerProfile) (err error) {
	tx := sqls.db.Begin()
	mdl := &ChargerProfileMdl{
		Tenant:         cp.Tenant,
		ID:             cp.ID,
		ChargerProfile: cp.AsMapStringInterface(),
	}
	if err := tx.Model(&ChargerProfileMdl{}).Where(
		ChargerProfileMdl{Tenant: mdl.Tenant, ID: mdl.ID}).Delete(
		ChargerProfileMdl{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Save(mdl).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) RemoveChargerProfileDrv(_ *context.Context, tenant, id string) (err error) {
	tx := sqls.db.Begin()
	if err := tx.Model(&ChargerProfileMdl{}).Where(&ChargerProfileMdl{Tenant: tenant, ID: id}).
		Delete(&ChargerProfileMdl{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (sqls *SQLStorage) GetAttributeProfileDrv(ctx *context.Context, tenant, id string) (ap *utils.AttributeProfile, err error) {
	var result []*AttributeProfileMdl
	if err := sqls.db.Model(&AttributeProfileMdl{}).Where(&AttributeProfileMdl{Tenant: tenant,
		ID: id}).Find(&result).Error; err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, utils.ErrNotFound
	}

	return utils.MapStringInterfaceToAttributeProfile(result[0].AttributeProfile)
}

func (sqls *SQLStorage) SetAttributeProfileDrv(ctx *context.Context, ap *utils.AttributeProfile) (err error) {
	tx := sqls.db.Begin()
	mdl := &AttributeProfileMdl{
		Tenant:           ap.Tenant,
		ID:               ap.ID,
		AttributeProfile: ap.AsMapStringInterface(),
	}
	if err = tx.Model(&AttributeProfileMdl{}).Where(
		AttributeProfileMdl{Tenant: mdl.Tenant, ID: mdl.ID}).Delete(
		AttributeProfileMdl{}).Error; err != nil {
		tx.Rollback()
		return
	}
	if err = tx.Save(mdl).Error; err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
	return
}

func (sqls *SQLStorage) RemoveAttributeProfileDrv(ctx *context.Context, tenant, id string) (err error) {
	tx := sqls.db.Begin()
	if err = tx.Model(&AttributeProfileMdl{}).Where(&AttributeProfileMdl{Tenant: tenant, ID: id}).
		Delete(&AttributeProfileMdl{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return
}

// AddLoadHistory DataDB method not implemented yet
func (sqls *SQLStorage) AddLoadHistory(ldInst *utils.LoadInstance,
	loadHistSize int, transactionID string) error {
	return utils.ErrNotImplemented
}

// Only intended for InternalDB
func (sqls *SQLStorage) BackupConfigDB(backupFolderPath string, zip bool) (err error) {
	return utils.ErrNotImplemented
}

// BackupDataDB used only for InternalDB
func (sqls *SQLStorage) BackupDataDB(backupFolderPath string, zip bool) (err error) {
	return utils.ErrNotImplemented
}

// Will dump everything inside DB to a file, only for InternalDB
func (sqls *SQLStorage) DumpConfigDB() (err error) {
	return utils.ErrNotImplemented
}

// Will dump everything inside DB to a file, only for InternalDB
func (sqls *SQLStorage) DumpDataDB() (err error) {
	return utils.ErrNotImplemented
}

// Will rewrite every dump file of DataDB,  only for InternalDB
func (sqls *SQLStorage) RewriteDataDB() (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) HasDataDrv(ctx *context.Context, category, subject, tenant string) (exists bool, err error) {
	return false, utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) GetLoadHistory(limit int, skipCache bool,
	transactionID string) (loadInsts []*utils.LoadInstance, err error) {
	return nil, utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) GetResourceProfileDrv(ctx *context.Context, tenant, id string) (rsp *utils.ResourceProfile, err error) {
	return nil, utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) SetResourceProfileDrv(ctx *context.Context, rsp *utils.ResourceProfile) (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) RemoveResourceProfileDrv(ctx *context.Context, tenant, id string) (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) GetResourceDrv(ctx *context.Context, tenant, id string) (r *utils.Resource, err error) {
	return nil, utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) SetResourceDrv(ctx *context.Context, r *utils.Resource) (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) RemoveResourceDrv(ctx *context.Context, tenant, id string) (err error) {
	return utils.ErrNotImplemented
}

// GetStatQueueProfileDrv DataDB method not implemented yet
func (sqls *SQLStorage) GetStatQueueProfileDrv(ctx *context.Context, tenant string, id string) (sq *StatQueueProfile, err error) {
	return nil, utils.ErrNotImplemented
}

// SetStatQueueProfileDrv DataDB method not implemented yet
func (sqls *SQLStorage) SetStatQueueProfileDrv(ctx *context.Context, sq *StatQueueProfile) (err error) {
	return utils.ErrNotImplemented
}

// RemStatQueueProfileDrv DataDB method not implemented yet
func (sqls *SQLStorage) RemStatQueueProfileDrv(ctx *context.Context, tenant, id string) (err error) {
	return utils.ErrNotImplemented
}

// GetStatQueueDrv DataDB method not implemented yet
func (sqls *SQLStorage) GetStatQueueDrv(ctx *context.Context, tenant, id string) (sq *StatQueue, err error) {
	return nil, utils.ErrNotImplemented
}

// SetStatQueueDrv DataDB method not implemented yet
func (sqls *SQLStorage) SetStatQueueDrv(ctx *context.Context, ssq *StoredStatQueue, sq *StatQueue) (err error) {
	return utils.ErrNotImplemented
}

// RemStatQueueDrv DataDB method not implemented yet
func (sqls *SQLStorage) RemStatQueueDrv(ctx *context.Context, tenant, id string) (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) SetTrendProfileDrv(ctx *context.Context, sg *utils.TrendProfile) (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) GetTrendProfileDrv(ctx *context.Context, tenant string, id string) (sg *utils.TrendProfile, err error) {
	return nil, utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) RemTrendProfileDrv(ctx *context.Context, tenant string, id string) (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) GetTrendDrv(ctx *context.Context, tenant, id string) (r *utils.Trend, err error) {
	return nil, utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) SetTrendDrv(ctx *context.Context, r *utils.Trend) (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) RemoveTrendDrv(ctx *context.Context, tenant, id string) (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) SetRankingProfileDrv(ctx *context.Context, sg *utils.RankingProfile) (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) GetRankingProfileDrv(ctx *context.Context, tenant string, id string) (sg *utils.RankingProfile, err error) {
	return nil, utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) RemRankingProfileDrv(ctx *context.Context, tenant string, id string) (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) GetRankingDrv(ctx *context.Context, tenant, id string) (rn *utils.Ranking, err error) {
	return nil, utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) SetRankingDrv(_ *context.Context, rn *utils.Ranking) (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) RemoveRankingDrv(ctx *context.Context, tenant, id string) (err error) {
	return utils.ErrNotImplemented
}

// GetThresholdProfileDrv DataDB method not implemented yet
func (sqls *SQLStorage) GetThresholdProfileDrv(ctx *context.Context, tenant, ID string) (tp *ThresholdProfile, err error) {
	return nil, utils.ErrNotImplemented
}

// SetThresholdProfileDrv DataDB method not implemented yet
func (sqls *SQLStorage) SetThresholdProfileDrv(ctx *context.Context, tp *ThresholdProfile) (err error) {
	return utils.ErrNotImplemented
}

// RemThresholdProfileDrv DataDB method not implemented yet
func (sqls *SQLStorage) RemThresholdProfileDrv(ctx *context.Context, tenant, id string) (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) GetThresholdDrv(ctx *context.Context, tenant, id string) (r *Threshold, err error) {
	return nil, utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) SetThresholdDrv(ctx *context.Context, r *Threshold) (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) RemoveThresholdDrv(ctx *context.Context, tenant, id string) (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) GetFilterDrv(ctx *context.Context, tenant, id string) (r *Filter, err error) {
	return nil, utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) SetFilterDrv(ctx *context.Context, r *Filter) (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) RemoveFilterDrv(ctx *context.Context, tenant, id string) (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) GetRouteProfileDrv(ctx *context.Context, tenant, id string) (r *utils.RouteProfile, err error) {
	return nil, utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) SetRouteProfileDrv(ctx *context.Context, r *utils.RouteProfile) (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) RemoveRouteProfileDrv(ctx *context.Context, tenant, id string) (err error) {
	return utils.ErrNotImplemented
}

// GetStorageType returns the storage type that is being used
func (sqls *SQLStorage) GetStorageType() string {
	return utils.MetaMySQL
}

// DataDB method not implemented yet
func (sqls *SQLStorage) GetItemLoadIDsDrv(ctx *context.Context, itemIDPrefix string) (loadIDs map[string]int64, err error) {
	return nil, utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) SetLoadIDsDrv(ctx *context.Context, loadIDs map[string]int64) error {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) RemoveLoadIDsDrv() (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) SetRateProfileDrv(ctx *context.Context, rpp *utils.RateProfile, optOverwrite bool) (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) GetRateProfileDrv(ctx *context.Context, tenant, id string) (rpp *utils.RateProfile, err error) {
	return nil, utils.ErrNotImplemented
}

// GetRateProfileRateIDsDrv DataDB method not implemented yet
func (sqls *SQLStorage) GetRateProfileRatesDrv(ctx *context.Context, tnt, profileID, rtPrfx string, needIDs bool) (rateIDs []string, rates []*utils.Rate, err error) {
	return nil, nil, utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) RemoveRateProfileDrv(ctx *context.Context, tenant, id string, rateIDs *[]string) (err error) {
	return utils.ErrNotImplemented
}

// GetIndexesDrv DataDB method not implemented yet
func (sqls *SQLStorage) GetIndexesDrv(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
	return nil, utils.ErrNotImplemented
}

// SetIndexesDrv DataDB method not implemented yet
func (sqls *SQLStorage) SetIndexesDrv(ctx *context.Context, idxItmType, tntCtx string,
	indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) RemoveIndexesDrv(ctx *context.Context, idxItmType, tntCtx, idxKey string) (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) GetConfigSectionsDrv(ctx *context.Context, nodeID string, sectionIDs []string) (sectionMap map[string][]byte, err error) {
	return nil, utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) SetConfigSectionsDrv(ctx *context.Context, nodeID string, sectionsData map[string][]byte) (err error) {
	return utils.ErrNotImplemented
}

// DataDB method not implemented yet
func (sqls *SQLStorage) RemoveConfigSectionsDrv(ctx *context.Context, nodeID string, sectionIDs []string) (err error) {
	return utils.ErrNotImplemented
}

// ConfigDB method not implemented yet
func (sqls *SQLStorage) GetSection(ctx *context.Context, section string, val any) (err error) {
	return utils.ErrNotImplemented
}

// ConfigDB method not implemented yet
func (sqls *SQLStorage) SetSection(_ *context.Context, section string, jsn any) (err error) {
	return utils.ErrNotImplemented
}

// Only intended for InternalDB
func (sqls *SQLStorage) RewriteConfigDB() (err error) {
	return utils.ErrNotImplemented
}
