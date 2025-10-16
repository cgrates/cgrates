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

package migrator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentRequestFilter() (err error) {
	mInDB, err := m.GetINConn(utils.MetaFilters)
	if err != nil {
		return err
	}
	dataDB, _, err := mInDB.DataManager().DBConns().GetConn(utils.MetaFilters)
	if err != nil {
		return err
	}
	var ids []string
	ids, err = dataDB.GetKeysForPrefix(context.TODO(), utils.FilterPrefix)
	if err != nil {
		return
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.FilterPrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating filters", id)
		}
		var fl *engine.Filter
		if fl, err = mInDB.DataManager().GetFilter(context.TODO(), tntID[0], tntID[1], false, false,
			utils.NonTransactional); err != nil {
			return
		}
		if m.dryRun || fl == nil {
			continue
		}
		mOutDB, err := m.GetOUTConn(utils.MetaFilters)
		if err != nil {
			return err
		}
		if err = mOutDB.DataManager().SetFilter(context.TODO(), fl, true); err != nil {
			return err
		}
		if err = mInDB.DataManager().RemoveFilter(context.TODO(), tntID[0], tntID[1],
			true); err != nil {
			return err
		}
		m.stats[utils.RQF]++
	}
	return
}

var filterTypes = utils.NewStringSet([]string{utils.MetaRSR, utils.MetaStats, utils.MetaResources,
	utils.MetaNotRSR, utils.MetaNotStatS, utils.MetaNotResources})

func migrateFilterV1(fl *v1Filter) (fltr *engine.Filter) {
	fltr = &engine.Filter{
		Tenant: fl.Tenant,
		ID:     fl.ID,
		Rules:  make([]*engine.FilterRule, len(fl.Rules)),
	}
	for i, rule := range fl.Rules {
		fltr.Rules[i] = &engine.FilterRule{
			Type:    rule.Type,
			Element: rule.FieldName,
			Values:  rule.Values,
		}
		if rule.FieldName == "" ||
			strings.HasPrefix(rule.FieldName, utils.DynamicDataPrefix) ||
			filterTypes.Has(rule.Type) {
			continue
		}
		fltr.Rules[i].Element = utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + rule.FieldName
	}
	return
}

func migrateFilterV2(fl *v1Filter) (fltr *engine.Filter) {
	fltr = &engine.Filter{
		Tenant: fl.Tenant,
		ID:     fl.ID,
		Rules:  make([]*engine.FilterRule, len(fl.Rules)),
	}
	for i, rule := range fl.Rules {
		fltr.Rules[i] = &engine.FilterRule{
			Type:    rule.Type,
			Element: rule.FieldName,
			Values:  rule.Values,
		}
		if rule.FieldName == "" && rule.Type != utils.MetaRSR ||
			strings.HasPrefix(rule.FieldName, utils.DynamicDataPrefix+utils.MetaReq) ||
			strings.HasPrefix(rule.FieldName, utils.DynamicDataPrefix+utils.MetaVars) ||
			strings.HasPrefix(rule.FieldName, utils.DynamicDataPrefix+utils.MetaCgreq) ||
			strings.HasPrefix(rule.FieldName, utils.DynamicDataPrefix+utils.MetaCgrep) ||
			strings.HasPrefix(rule.FieldName, utils.DynamicDataPrefix+utils.MetaRep) ||
			strings.HasPrefix(rule.FieldName, utils.DynamicDataPrefix+utils.MetaAct) {
			continue
		}
		if rule.Type != utils.MetaRSR {
			// in case we found dynamic data prefix we remove it
			if strings.HasPrefix(rule.FieldName, utils.DynamicDataPrefix) {
				fl.Rules[i].FieldName = fl.Rules[i].FieldName[1:]
			}
			fltr.Rules[i].Element = utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + rule.FieldName
		} else {
			for idx, val := range rule.Values {
				if strings.HasPrefix(val, utils.DynamicDataPrefix) {
					// remove dynamic data prefix from fieldName
					val = val[1:]
				}
				fltr.Rules[i].Values[idx] = utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + val
			}
		}
	}
	return
}

func migrateFilterV3(fl *v1Filter) (fltr *engine.Filter) {
	fltr = &engine.Filter{
		Tenant: fl.Tenant,
		ID:     fl.ID,
		Rules:  make([]*engine.FilterRule, len(fl.Rules)),
	}
	for i, rule := range fl.Rules {
		fltr.Rules[i] = &engine.FilterRule{
			Type:    rule.Type,
			Element: rule.FieldName,
			Values:  rule.Values,
		}
	}
	return
}

func migrateInlineFilter(fl string) string {
	if fl == "" || !strings.HasPrefix(fl, utils.Meta) {
		return fl
	}
	ruleSplt := strings.Split(fl, utils.InInFieldSep)
	if len(ruleSplt) < 3 {
		return fl
	}

	if strings.HasPrefix(ruleSplt[1], utils.DynamicDataPrefix) ||
		filterTypes.Has(ruleSplt[0]) {
		return fl
	}
	return fmt.Sprintf("%s:~%s:%s", ruleSplt[0], utils.MetaReq+utils.NestingSep+ruleSplt[1], strings.Join(ruleSplt[2:], utils.InInFieldSep))
}

func migrateInlineFilterV2(fl string) string {
	if fl == "" || !strings.HasPrefix(fl, utils.Meta) {
		return fl
	}
	ruleSplt := strings.Split(fl, utils.InInFieldSep)
	if len(ruleSplt) < 3 {
		return fl
	}
	if ruleSplt[1] != utils.EmptyString && // no need conversion
		(strings.HasPrefix(ruleSplt[1], utils.DynamicDataPrefix+utils.MetaReq) ||
			strings.HasPrefix(ruleSplt[1], utils.DynamicDataPrefix+utils.MetaVars) ||
			strings.HasPrefix(ruleSplt[1], utils.DynamicDataPrefix+utils.MetaCgreq) ||
			strings.HasPrefix(ruleSplt[1], utils.DynamicDataPrefix+utils.MetaCgrep) ||
			strings.HasPrefix(ruleSplt[1], utils.DynamicDataPrefix+utils.MetaRep) ||
			strings.HasPrefix(ruleSplt[1], utils.DynamicDataPrefix+utils.MetaAct)) {
		return fl
	}

	if ruleSplt[0] != utils.MetaRSR {
		// remove dynamic data prefix from fieldName
		ruleSplt[1] = strings.TrimPrefix(ruleSplt[1], utils.DynamicDataPrefix)
		return fmt.Sprintf("%s:~%s:%s", ruleSplt[0], utils.MetaReq+utils.NestingSep+ruleSplt[1], strings.Join(ruleSplt[2:], utils.InInFieldSep))
	} // in case of *rsr filter we need to add the prefix at fieldValue
	// remove dynamic data prefix from fieldName
	ruleSplt[2] = strings.TrimPrefix(ruleSplt[2], utils.DynamicDataPrefix)
	return fmt.Sprintf("%s::~%s", ruleSplt[0], utils.MetaReq+utils.NestingSep+strings.Join(ruleSplt[2:], utils.InInFieldSep))
}

func (m *Migrator) migrateOthersv1() (err error) {
	if err = m.migrateStatQueueProfileFiltersV1(); err != nil {
		return err
	}
	if err = m.migrateChargerProfileFiltersV1(); err != nil {
		return err
	}
	return
}

func (m *Migrator) migrateRequestFilterV1() (fltr *engine.Filter, err error) {
	mInDB, err := m.GetINConn(utils.MetaFilters)
	if err != nil {
		return nil, err
	}
	var v1Fltr *v1Filter
	if v1Fltr, err = mInDB.getV1Filter(); err != nil {
		return
	}
	if v1Fltr == nil {
		return
	}
	fltr = migrateFilterV1(v1Fltr)
	return
}

func (m *Migrator) migrateOthersV2() (err error) {
	if err = m.migrateStatQueueProfileFiltersV2(); err != nil {
		return fmt.Errorf("Error: <%s> when trying to migrate filter for StatQueueProfiles",
			err.Error())
	}
	if err = m.migrateChargerProfileFiltersV2(); err != nil {
		return fmt.Errorf("Error: <%s> when trying to migrate filter for ChargerProfiles",
			err.Error())
	}
	return
}

func (m *Migrator) migrateRequestFilterV2() (fltr *engine.Filter, err error) {
	mInDB, err := m.GetINConn(utils.MetaFilters)
	if err != nil {
		return nil, err
	}
	var v1Fltr *v1Filter
	if v1Fltr, err = mInDB.getV1Filter(); err != nil {
		return nil, err
	}
	if err == utils.ErrNoMoreData {
		return nil, nil
	}
	fltr = migrateFilterV2(v1Fltr)
	return
}

func (m *Migrator) migrateRequestFilterV3() (fltr *engine.Filter, err error) {
	mInDB, err := m.GetINConn(utils.MetaFilters)
	if err != nil {
		return nil, err
	}
	var v1Fltr *v1Filter
	if v1Fltr, err = mInDB.getV1Filter(); err != nil {
		return nil, err
	}
	if v1Fltr == nil {
		return
	}
	fltr = migrateFilterV3(v1Fltr)
	return
}

func (m *Migrator) migrateFilters() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	if vrs, err = m.getVersions(utils.RQF); err != nil {
		return
	}
	migrated := true
	migratedFrom := 0
	var v4Fltr *engine.Filter
	var fltr *engine.Filter
	for {
		version := vrs[utils.RQF]
		migratedFrom = int(version)
		for {
			switch version {
			default:
				return fmt.Errorf("Unsupported version %v", version)
			case current[utils.RQF]:
				migrated = false
				if m.sameDataDB {
					break
				}
				if err = m.migrateCurrentRequestFilter(); err != nil {
					return
				}
			case 1:
				if v4Fltr, err = m.migrateRequestFilterV1(); err != nil && err != utils.ErrNoMoreData {
					return
				}
				version = 4
			case 2:
				if v4Fltr, err = m.migrateRequestFilterV2(); err != nil && err != utils.ErrNoMoreData {
					return
				}
				version = 4
			case 3:
				if v4Fltr, err = m.migrateRequestFilterV3(); err != nil && err != utils.ErrNoMoreData {
					return
				}
				version = 4
			case 4: // in case we change the structure to the filters please update the geing method from this version
				if fltr, err = m.migrateRequestFilterV4(v4Fltr); err != nil && err != utils.ErrNoMoreData {
					return
				}
				version = 5
			}
			if version == current[utils.RQF] || err == utils.ErrNoMoreData {
				break
			}
		}
		if err == utils.ErrNoMoreData || !migrated {
			err = nil
			break
		}
		if !m.dryRun {
			if err = m.setFilterv5WithoutCompile(fltr); err != nil {
				return fmt.Errorf("Error: <%s> when setting filter with tenant: <%s> and id: <%s> after migration",
					err.Error(), fltr.Tenant, fltr.ID)
			}
		}
		m.stats[utils.RQF]++
	}
	if m.dryRun || !migrated {
		return
	}

	switch migratedFrom {
	case 1:
		if err = m.migrateOthersv1(); err != nil {
			return
		}
	case 2:
		if err = m.migrateOthersV2(); err != nil {
			return
		}
	}

	if err = m.setVersions(utils.RQF); err != nil {
		return
	}
	return m.ensureIndexesDataDB(engine.ColFlt)
}

func (m *Migrator) migrateStatQueueProfileFiltersV1() (err error) {
	mInDB, err := m.GetINConn(utils.MetaStatQueueProfiles)
	if err != nil {
		return err
	}
	dataDB, _, err := mInDB.DataManager().DBConns().GetConn(utils.MetaStatQueueProfiles)
	if err != nil {
		return err
	}
	var ids []string
	ids, err = dataDB.GetKeysForPrefix(context.TODO(), utils.StatQueueProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.StatQueueProfilePrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating filter for statQueueProfile", id)
		}
		mInDB, err := m.GetINConn(utils.MetaStatQueueProfiles)
		if err != nil {
			return err
		}
		sgs, err := mInDB.DataManager().GetStatQueueProfile(context.TODO(), tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if sgs == nil || m.dryRun {
			continue
		}
		for i, fl := range sgs.FilterIDs {
			sgs.FilterIDs[i] = migrateInlineFilter(fl)
		}
		mOutDB, err := m.GetOUTConn(utils.MetaStatQueueProfiles)
		if err != nil {
			return err
		}
		if err = mOutDB.DataManager().SetStatQueueProfile(context.TODO(), sgs, true); err != nil {
			return err
		}
		m.stats[utils.RQF]++
	}
	return
}

func (m *Migrator) migrateChargerProfileFiltersV1() (err error) {
	mInDB, err := m.GetINConn(utils.MetaChargerProfiles)
	if err != nil {
		return err
	}
	dataDB, _, err := mInDB.DataManager().DBConns().GetConn(utils.MetaChargerProfiles)
	if err != nil {
		return err
	}
	var ids []string
	ids, err = dataDB.GetKeysForPrefix(context.TODO(), utils.ChargerProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.ChargerProfilePrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating filter for chragerProfile", id)
		}
		cpp, err := mInDB.DataManager().GetChargerProfile(context.TODO(), tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if cpp == nil || m.dryRun {
			continue
		}
		for i, fl := range cpp.FilterIDs {
			cpp.FilterIDs[i] = migrateInlineFilter(fl)
		}
		mOutDB, err := m.GetOUTConn(utils.MetaChargerProfiles)
		if err != nil {
			return err
		}
		if err := mOutDB.DataManager().SetChargerProfile(context.TODO(), cpp, true); err != nil {
			return err
		}
		m.stats[utils.RQF]++
	}
	return
}

func (m *Migrator) migrateStatQueueProfileFiltersV2() (err error) {
	mInDB, err := m.GetINConn(utils.MetaStatQueueProfiles)
	if err != nil {
		return err
	}
	dataDB, _, err := mInDB.DataManager().DBConns().GetConn(utils.MetaStatQueueProfiles)
	if err != nil {
		return err
	}
	var ids []string
	ids, err = dataDB.GetKeysForPrefix(context.TODO(), utils.StatQueueProfilePrefix)
	if err != nil {
		return fmt.Errorf("error: <%s> when getting statQueue profile IDs", err.Error())
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.StatQueueProfilePrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating filter for statQueueProfile", id)
		}
		sgs, err := mInDB.DataManager().GetStatQueueProfile(context.TODO(), tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return fmt.Errorf("error: <%s> when getting statQueue profile with tenant: <%s> and id: <%s>",
				err.Error(), tntID[0], tntID[1])
		}
		if sgs == nil || m.dryRun {
			continue
		}
		for i, fl := range sgs.FilterIDs {
			sgs.FilterIDs[i] = migrateInlineFilterV2(fl)
		}
		mOutDB, err := m.GetOUTConn(utils.MetaStatQueueProfiles)
		if err != nil {
			return err
		}
		if err = mOutDB.DataManager().SetStatQueueProfile(context.TODO(), sgs, true); err != nil {
			return fmt.Errorf("error: <%s> when setting statQueue profile with tenant: <%s> and id: <%s>",
				err.Error(), tntID[0], tntID[1])
		}
		m.stats[utils.RQF]++
	}
	return
}

func (m *Migrator) migrateChargerProfileFiltersV2() (err error) {
	mInDB, err := m.GetINConn(utils.MetaChargerProfiles)
	if err != nil {
		return err
	}
	dataDB, _, err := mInDB.DataManager().DBConns().GetConn(utils.MetaChargerProfiles)
	if err != nil {
		return err
	}
	var ids []string
	ids, err = dataDB.GetKeysForPrefix(context.TODO(), utils.ChargerProfilePrefix)
	if err != nil {
		return fmt.Errorf("error: <%s> when getting charger profile IDs", err)
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.ChargerProfilePrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating filter for chargerProfile", id)
		}
		cpp, err := mInDB.DataManager().GetChargerProfile(context.TODO(), tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return fmt.Errorf("error: <%s> when getting charger profile with tenant: <%s> and id: <%s>",
				err.Error(), tntID[0], tntID[1])
		}
		if cpp == nil || m.dryRun {
			continue
		}
		for i, fl := range cpp.FilterIDs {
			cpp.FilterIDs[i] = migrateInlineFilterV2(fl)
		}
		mOutDB, err := m.GetOUTConn(utils.MetaChargerProfiles)
		if err != nil {
			return err
		}
		if err := mOutDB.DataManager().SetChargerProfile(context.TODO(), cpp, true); err != nil {
			return fmt.Errorf("error: <%s> when setting charger profile with tenant: <%s> and id: <%s>",
				err.Error(), tntID[0], tntID[1])
		}
		m.stats[utils.RQF]++
	}
	return
}

type v1Filter struct {
	Tenant             string
	ID                 string
	Rules              []*v1FilterRule
	ActivationInterval *utils.ActivationInterval
}

type v1FilterRule struct {
	Type      string           // Filter type (*string,  *rsr_filters, *stats, *lt, *lte, *gt, *gte)
	FieldName string           // Name of the field providing us the Values to check (used in case of some )
	Values    []string         // Filter definition
	rsrFields utils.RSRParsers // Cache here the RSRFilter Values
	negative  *bool
}

func (m *Migrator) migrateRequestFilterV4(v4Fltr *engine.Filter) (fltr *engine.Filter, err error) {
	if v4Fltr == nil {
		// read data from DataDB
		mInDB, err := m.GetINConn(utils.MetaFilters)
		if err != nil {
			return nil, err
		}
		v4Fltr, err = mInDB.getV4Filter()
		if err != nil {
			return nil, err
		}
	}
	//migrate
	fltr = &engine.Filter{
		Tenant: v4Fltr.Tenant,
		ID:     v4Fltr.ID,
		Rules:  make([]*engine.FilterRule, 0, len(v4Fltr.Rules)),
	}
	for _, rule := range v4Fltr.Rules {
		if rule.Type != utils.MetaRSR &&
			rule.Type != utils.MetaNotRSR {
			fltr.Rules = append(fltr.Rules, rule)
			continue
		}
		for _, val := range rule.Values {
			el, vals, err := migrateRSRFilterV4(val)
			if err != nil {
				return nil, fmt.Errorf("%s for filter<%s>", err.Error(), fltr.TenantID())
			}
			if len(vals) == 0 { // is not a filter so we ignore this value
				continue
			}
			fltr.Rules = append(fltr.Rules, &engine.FilterRule{
				Type:    rule.Type,
				Element: el,
				Values:  vals,
			})
		}
	}
	return
}

var (
	spltRgxp = regexp.MustCompile(`:s\/`)
)

func migrateRSRFilterV4(rsr string) (el string, vals []string, err error) {
	if !strings.HasSuffix(rsr, utils.FilterValEnd) { // is not a filter so we ignore this value
		return
	}
	fltrStart := strings.Index(rsr, utils.FilterValStart)
	if fltrStart < 1 {
		err = fmt.Errorf("invalid RSRFilter start rule in string: <%s> ", rsr)
		return
	}
	vals = strings.Split(rsr[fltrStart+1:len(rsr)-1], utils.ANDSep)
	el = rsr[:fltrStart]

	if idxConverters := strings.Index(el, "{*"); idxConverters != -1 { // converters in the string
		if !strings.HasSuffix(el, "}") {
			err = fmt.Errorf("invalid converter terminator in rule: <%s>", el)
			return
		}
		el = el[:idxConverters]
	}
	if !strings.HasPrefix(el, utils.DynamicDataPrefix) ||
		len(el) == 1 { // special case when RSR is defined as static attribute
		return
	}
	// dynamic content via attributeNames
	el = spltRgxp.Split(el, -1)[0]
	return
}

func migrateInlineFilterV4(v4fltIDs []string) (fltrIDs []string, err error) {
	fltrIDs = make([]string, 0, len(v4fltIDs))
	for _, v4flt := range v4fltIDs {
		var fltr string
		if strings.HasPrefix(v4flt, utils.MetaRSR) {
			fltr = utils.MetaRSR + utils.InInFieldSep
			v4flt = strings.TrimPrefix(v4flt, utils.MetaRSR+utils.InInFieldSep+utils.InInFieldSep)
		} else if strings.HasPrefix(v4flt, utils.MetaNotRSR) {
			fltr = utils.MetaNotRSR + utils.InInFieldSep
			v4flt = strings.TrimPrefix(v4flt, utils.MetaNotRSR+utils.InInFieldSep+utils.InInFieldSep)
		} else {
			fltrIDs = append(fltrIDs, v4flt)
		}
		for _, val := range strings.Split(v4flt, utils.InfieldSep) {
			el, vals, err := migrateRSRFilterV4(val)
			if err != nil {
				return nil, err
			}
			if len(vals) == 0 { // is not a filter so we ignore this value
				continue
			}

			fltrIDs = append(fltrIDs, fltr+el+utils.InInFieldSep+
				strings.Join(vals, utils.PipeSep))
		}
	}
	return
}

// setFilterv5WithoutCompile we need a method that get's the filter from DataDB without compiling the filter rules
func (m *Migrator) setFilterv5WithoutCompile(fltr *engine.Filter) (err error) {
	mOutDB, err := m.GetOUTConn(utils.MetaFilters)
	if err != nil {
		return err
	}
	dataDB, _, err := mOutDB.DataManager().DBConns().GetConn(utils.MetaFilters)
	if err != nil {
		return err
	}
	var oldFlt *engine.Filter
	if oldFlt, err = dataDB.GetFilterDrv(context.TODO(), fltr.Tenant, fltr.ID); err != nil &&
		err != utils.ErrNotFound {
		return
	}
	if err = dataDB.SetFilterDrv(context.TODO(), fltr); err != nil {
		return
	}
	return engine.UpdateFilterIndex(context.TODO(), mOutDB.DataManager(), oldFlt, fltr)
}
