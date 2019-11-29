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

package migrator

import (
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentRequestFilter() (err error) {
	var ids []string
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.FilterPrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.FilterPrefix+tenant+":")
		fl, err := m.dmIN.DataManager().GetFilter(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if m.dryRun || fl == nil {
			continue
		}
		if err := m.dmIN.DataManager().RemoveFilter(tenant, idg, utils.NonTransactional); err != nil {
			return err
		}
		if err := m.dmOut.DataManager().SetFilter(fl); err != nil {
			return err
		}
		m.stats[utils.RQF] += 1
	}
	return
}

var filterTypes = utils.NewStringSet([]string{utils.MetaRSR, utils.MetaStatS, utils.MetaResources,
	utils.MetaNotRSR, utils.MetaNotStatS, utils.MetaNotResources})

func migrateFilterV1(fl *engine.Filter) *engine.Filter {
	for i, rule := range fl.Rules {
		if rule.FieldName == "" ||
			strings.HasPrefix(rule.FieldName, utils.DynamicDataPrefix) ||
			filterTypes.Has(rule.Type) {
			continue
		}
		fl.Rules[i].FieldName = utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + rule.FieldName
	}
	return fl
}

func migrateFilterV2(fl *engine.Filter) *engine.Filter {
	for i, rule := range fl.Rules {
		if (rule.FieldName == "" && rule.Type != utils.MetaRSR) ||
			strings.HasPrefix(rule.FieldName, utils.DynamicDataPrefix+utils.MetaReq) ||
			strings.HasPrefix(rule.FieldName, utils.DynamicDataPrefix+utils.MetaVars) ||
			strings.HasPrefix(rule.FieldName, utils.DynamicDataPrefix+utils.MetaCgreq) ||
			strings.HasPrefix(rule.FieldName, utils.DynamicDataPrefix+utils.MetaCgrep) ||
			strings.HasPrefix(rule.FieldName, utils.DynamicDataPrefix+utils.MetaRep) ||
			strings.HasPrefix(rule.FieldName, utils.DynamicDataPrefix+utils.MetaCGRAReq) ||
			strings.HasPrefix(rule.FieldName, utils.DynamicDataPrefix+utils.MetaAct) {
			continue
		}
		if rule.Type != utils.MetaRSR {
			// in case we found dynamic data prefix we remove it
			if strings.HasPrefix(rule.FieldName, utils.DynamicDataPrefix) {
				fl.Rules[i].FieldName = fl.Rules[i].FieldName[1:]
			}
			fl.Rules[i].FieldName = utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + rule.FieldName
		} else {
			for idx, val := range rule.Values {
				if strings.HasPrefix(val, utils.DynamicDataPrefix) {
					// remove dynamic data prefix from fieldName
					val = val[1:]
				}
				fl.Rules[i].Values[idx] = utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + val
			}
		}
	}
	return fl
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
			strings.HasPrefix(ruleSplt[1], utils.DynamicDataPrefix+utils.MetaCGRAReq) ||
			strings.HasPrefix(ruleSplt[1], utils.DynamicDataPrefix+utils.MetaAct)) {
		return fl
	}

	if ruleSplt[0] != utils.MetaRSR {
		if strings.HasPrefix(ruleSplt[1], utils.DynamicDataPrefix) {
			// remove dynamic data prefix from fieldName
			ruleSplt[1] = ruleSplt[1][1:]
		}
		return fmt.Sprintf("%s:~%s:%s", ruleSplt[0], utils.MetaReq+utils.NestingSep+ruleSplt[1], strings.Join(ruleSplt[2:], utils.InInFieldSep))
	} else { // in case of *rsr filter we need to add the prefix at fieldValue
		if strings.HasPrefix(ruleSplt[2], utils.DynamicDataPrefix) {
			// remove dynamic data prefix from fieldName
			ruleSplt[2] = ruleSplt[2][1:]
		}
		return fmt.Sprintf("%s::~%s", ruleSplt[0], utils.MetaReq+utils.NestingSep+strings.Join(ruleSplt[2:], utils.InInFieldSep))
	}

}

func (m *Migrator) migrateRequestFilterV1() (err error) {
	var ids []string
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.FilterPrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.FilterPrefix+tenant+":")
		fl, err := m.dmIN.DataManager().GetFilter(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if m.dryRun || fl == nil {
			continue
		}
		if err := m.dmOut.DataManager().SetFilter(migrateFilterV1(fl)); err != nil {
			return err
		}
		m.stats[utils.RQF] += 1
	}
	if err = m.migrateResourceProfileFiltersV1(); err != nil {
		return err
	}
	if err = m.migrateStatQueueProfileFiltersV1(); err != nil {
		return err
	}
	if err = m.migrateThresholdsProfileFiltersV1(); err != nil {
		return err
	}
	if err = m.migrateSupplierProfileFiltersV1(); err != nil {
		return err
	}
	if err = m.migrateAttributeProfileFiltersV1(); err != nil {
		return err
	}
	if err = m.migrateChargerProfileFiltersV1(); err != nil {
		return err
	}
	if err = m.migrateDispatcherProfileFiltersV1(); err != nil {
		return err
	}
	vrs := engine.Versions{utils.RQF: 2}
	if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating Filters version into dataDB", err.Error()))
	}
	return
}

func (m *Migrator) migrateRequestFilterV2() (err error) {
	var ids []string
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.FilterPrefix)
	if err != nil {
		return fmt.Errorf("Error: <%s> when getting filter IDs for migration", err.Error())
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.FilterPrefix+tenant+":")
		fl, err := m.dmIN.DataManager().GetFilter(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {
			return fmt.Errorf("Error: <%s> when getting filter with tenant: <%s> and id: <%s> for migration",
				err.Error(), tenant, idg)
		}
		if m.dryRun || fl == nil {
			continue
		}
		if err := m.dmOut.DataManager().SetFilter(migrateFilterV2(fl)); err != nil {
			return fmt.Errorf("Error: <%s> when setting filter with tenant: <%s> and id: <%s> after migration",
				err.Error(), tenant, idg)
		}
		m.stats[utils.RQF] += 1
	}
	if err = m.migrateResourceProfileFiltersV2(); err != nil {
		return fmt.Errorf("Error: <%s> when trying to migrate filter for ResourceProfiles",
			err.Error())
	}
	if err = m.migrateStatQueueProfileFiltersV2(); err != nil {
		return fmt.Errorf("Error: <%s> when trying to migrate filter for StatQueueProfiles",
			err.Error())
	}
	if err = m.migrateThresholdsProfileFiltersV2(); err != nil {
		return fmt.Errorf("Error: <%s> when trying to migrate filter for ThresholdProfiles",
			err.Error())
	}
	if err = m.migrateSupplierProfileFiltersV2(); err != nil {
		return fmt.Errorf("Error: <%s> when trying to migrate filter for SupplierProfiles",
			err.Error())
	}
	if err = m.migrateAttributeProfileFiltersV2(); err != nil {
		return fmt.Errorf("Error: <%s> when trying to migrate filter for AttributeProfiles",
			err.Error())
	}
	if err = m.migrateChargerProfileFiltersV2(); err != nil {
		return fmt.Errorf("Error: <%s> when trying to migrate filter for ChargerProfiles",
			err.Error())
	}
	if err = m.migrateDispatcherProfileFiltersV2(); err != nil {
		return fmt.Errorf("Error: <%s> when trying to migrate filter for DispatcherProfiles",
			err.Error())
	}
	vrs := engine.Versions{utils.RQF: engine.CurrentDataDBVersions()[utils.RQF]}
	if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating Filters version into dataDB", err.Error()))
	}
	return
}

func (m *Migrator) migrateFilters() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	vrs, err = m.dmIN.DataManager().DataDB().GetVersions("")
	if err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when querying oldDataDB for versions", err.Error()))
	} else if len(vrs) == 0 {
		return utils.NewCGRError(utils.Migrator,
			utils.MandatoryIEMissingCaps,
			utils.UndefinedVersion,
			"version number is not defined for ActionTriggers model")
	}
	switch vrs[utils.RQF] {
	case 2:
		if err = m.migrateRequestFilterV2(); err != nil {
			return err
		}
	case 1:
		if err = m.migrateRequestFilterV1(); err != nil {
			return err
		}
	case current[utils.RQF]:
		if m.sameDataDB {
			break
		}
		if err = m.migrateCurrentRequestFilter(); err != nil {
			return err
		}
	}
	return m.ensureIndexesDataDB(engine.ColFlt)
}

func (m *Migrator) migrateResourceProfileFiltersV1() (err error) {
	var ids []string
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.ResourceProfilesPrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.ResourceProfilesPrefix+tenant+":")
		res, err := m.dmIN.DataManager().GetResourceProfile(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if m.dryRun || res == nil {
			continue
		}
		for i, fl := range res.FilterIDs {
			res.FilterIDs[i] = migrateInlineFilter(fl)
		}
		if err := m.dmOut.DataManager().SetResourceProfile(res, true); err != nil {
			return err
		}
		m.stats[utils.RQF] += 1
	}
	return
}

func (m *Migrator) migrateStatQueueProfileFiltersV1() (err error) {
	var ids []string
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.StatQueueProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.StatQueueProfilePrefix+tenant+":")
		sgs, err := m.dmIN.DataManager().GetStatQueueProfile(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if sgs == nil || m.dryRun {
			continue
		}
		for i, fl := range sgs.FilterIDs {
			sgs.FilterIDs[i] = migrateInlineFilter(fl)
		}
		if err = m.dmOut.DataManager().SetStatQueueProfile(sgs, true); err != nil {
			return err
		}
		m.stats[utils.RQF] += 1
	}
	return
}

func (m *Migrator) migrateThresholdsProfileFiltersV1() (err error) {
	var ids []string
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.ThresholdProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.ThresholdProfilePrefix+tenant+":")
		ths, err := m.dmIN.DataManager().GetThresholdProfile(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if ths == nil || m.dryRun {
			continue
		}
		for i, fl := range ths.FilterIDs {
			ths.FilterIDs[i] = migrateInlineFilter(fl)
		}
		if err := m.dmOut.DataManager().SetThresholdProfile(ths, true); err != nil {
			return err
		}
		m.stats[utils.RQF] += 1
	}
	return
}

func (m *Migrator) migrateSupplierProfileFiltersV1() (err error) {
	var ids []string
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.SupplierProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.SupplierProfilePrefix)
		splp, err := m.dmIN.DataManager().GetSupplierProfile(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if splp == nil || m.dryRun {
			continue
		}
		for i, fl := range splp.FilterIDs {
			splp.FilterIDs[i] = migrateInlineFilter(fl)
		}
		if err := m.dmOut.DataManager().SetSupplierProfile(splp, true); err != nil {
			return err
		}
		m.stats[utils.RQF] += 1
	}
	return
}

func (m *Migrator) migrateAttributeProfileFiltersV1() (err error) {
	var ids []string
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.AttributeProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.AttributeProfilePrefix+tenant+":")
		attrPrf, err := m.dmIN.DataManager().GetAttributeProfile(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if attrPrf == nil || m.dryRun {
			continue
		}
		for i, fl := range attrPrf.FilterIDs {
			attrPrf.FilterIDs[i] = migrateInlineFilter(fl)
		}
		for i, attr := range attrPrf.Attributes {
			for j, fl := range attr.FilterIDs {
				attrPrf.Attributes[i].FilterIDs[j] = migrateInlineFilter(fl)
			}
		}
		if err := m.dmOut.DataManager().SetAttributeProfile(attrPrf, true); err != nil {
			return err
		}
		m.stats[utils.RQF] += 1
	}
	return
}

func (m *Migrator) migrateChargerProfileFiltersV1() (err error) {
	var ids []string
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.ChargerProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.ChargerProfilePrefix+tenant+":")
		cpp, err := m.dmIN.DataManager().GetChargerProfile(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if cpp == nil || m.dryRun {
			continue
		}
		for i, fl := range cpp.FilterIDs {
			cpp.FilterIDs[i] = migrateInlineFilter(fl)
		}
		if err := m.dmOut.DataManager().SetChargerProfile(cpp, true); err != nil {
			return err
		}
		m.stats[utils.RQF] += 1
	}
	return
}

func (m *Migrator) migrateDispatcherProfileFiltersV1() (err error) {
	var ids []string
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.DispatcherProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.DispatcherProfilePrefix+tenant+":")
		dpp, err := m.dmIN.DataManager().GetDispatcherProfile(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if dpp == nil || m.dryRun {
			continue
		}
		for i, fl := range dpp.FilterIDs {
			dpp.FilterIDs[i] = migrateInlineFilter(fl)
		}
		if err := m.dmOut.DataManager().SetDispatcherProfile(dpp, true); err != nil {
			return err
		}
		m.stats[utils.RQF] += 1
	}
	return
}

// migrate filters from v2 to v3 for items
func (m *Migrator) migrateResourceProfileFiltersV2() (err error) {
	var ids []string
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.ResourceProfilesPrefix)
	if err != nil {
		return fmt.Errorf("error: <%s> when getting resource profile IDs", err.Error())
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.ResourceProfilesPrefix+tenant+":")
		res, err := m.dmIN.DataManager().GetResourceProfile(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {
			return fmt.Errorf("error: <%s> when getting resource profile with tenant: <%s> and id: <%s>",
				err.Error(), tenant, idg)
		}
		if m.dryRun || res == nil {
			continue
		}
		for i, fl := range res.FilterIDs {
			res.FilterIDs[i] = migrateInlineFilterV2(fl)
		}
		if err := m.dmOut.DataManager().SetResourceProfile(res, true); err != nil {
			return fmt.Errorf("error: <%s> when setting resource profile with tenant: <%s> and id: <%s>",
				err.Error(), tenant, idg)
		}
		m.stats[utils.RQF] += 1
	}
	return
}

func (m *Migrator) migrateStatQueueProfileFiltersV2() (err error) {
	var ids []string
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.StatQueueProfilePrefix)
	if err != nil {
		return fmt.Errorf("error: <%s> when getting statQueue profile IDs", err.Error())
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.StatQueueProfilePrefix+tenant+":")
		sgs, err := m.dmIN.DataManager().GetStatQueueProfile(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {
			return fmt.Errorf("error: <%s> when getting statQueue profile with tenant: <%s> and id: <%s>",
				err.Error(), tenant, idg)
		}
		if sgs == nil || m.dryRun {
			continue
		}
		for i, fl := range sgs.FilterIDs {
			sgs.FilterIDs[i] = migrateInlineFilterV2(fl)
		}
		if err = m.dmOut.DataManager().SetStatQueueProfile(sgs, true); err != nil {
			return fmt.Errorf("error: <%s> when setting statQueue profile with tenant: <%s> and id: <%s>",
				err.Error(), tenant, idg)
		}
		m.stats[utils.RQF] += 1
	}
	return
}

func (m *Migrator) migrateThresholdsProfileFiltersV2() (err error) {
	var ids []string
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.ThresholdProfilePrefix)
	if err != nil {
		return fmt.Errorf("error: <%s> when getting threshold profile IDs", err)
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.ThresholdProfilePrefix+tenant+":")
		ths, err := m.dmIN.DataManager().GetThresholdProfile(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {
			return fmt.Errorf("error: <%s> when getting threshold profile with tenant: <%s> and id: <%s>",
				err.Error(), tenant, idg)
		}
		if ths == nil || m.dryRun {
			continue
		}
		for i, fl := range ths.FilterIDs {
			ths.FilterIDs[i] = migrateInlineFilterV2(fl)
		}
		if err := m.dmOut.DataManager().SetThresholdProfile(ths, true); err != nil {
			return fmt.Errorf("error: <%s> when setting threshold profile with tenant: <%s> and id: <%s>",
				err.Error(), tenant, idg)
		}
		m.stats[utils.RQF] += 1
	}
	return
}

func (m *Migrator) migrateSupplierProfileFiltersV2() (err error) {
	var ids []string
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.SupplierProfilePrefix)
	if err != nil {
		return fmt.Errorf("error: <%s> when getting supplier profile IDs", err)
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.SupplierProfilePrefix)
		splp, err := m.dmIN.DataManager().GetSupplierProfile(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {
			return fmt.Errorf("error: <%s> when getting supplier profile with tenant: <%s> and id: <%s>",
				err.Error(), tenant, idg)
		}
		if splp == nil || m.dryRun {
			continue
		}
		for i, fl := range splp.FilterIDs {
			splp.FilterIDs[i] = migrateInlineFilterV2(fl)
		}
		if err := m.dmOut.DataManager().SetSupplierProfile(splp, true); err != nil {
			return fmt.Errorf("error: <%s> when setting supplier profile with tenant: <%s> and id: <%s>",
				err.Error(), tenant, idg)
		}
		m.stats[utils.RQF] += 1
	}
	return
}

func (m *Migrator) migrateAttributeProfileFiltersV2() (err error) {
	var ids []string
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.AttributeProfilePrefix)
	if err != nil {
		return fmt.Errorf("error: <%s> when getting attribute profile IDs", err)
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.AttributeProfilePrefix+tenant+":")
		attrPrf, err := m.dmIN.DataManager().GetAttributeProfile(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {
			return fmt.Errorf("error: <%s> when getting attribute profile with tenant: <%s> and id: <%s>",
				err.Error(), tenant, idg)
		}
		if attrPrf == nil || m.dryRun {
			continue
		}
		for i, fl := range attrPrf.FilterIDs {
			attrPrf.FilterIDs[i] = migrateInlineFilterV2(fl)
		}
		for i, attr := range attrPrf.Attributes {
			for j, fl := range attr.FilterIDs {
				attrPrf.Attributes[i].FilterIDs[j] = migrateInlineFilterV2(fl)
			}
		}
		if err := m.dmOut.DataManager().SetAttributeProfile(attrPrf, true); err != nil {
			return fmt.Errorf("error: <%s> when setting attribute profile with tenant: <%s> and id: <%s>",
				err.Error(), tenant, idg)
		}
		m.stats[utils.RQF] += 1
	}
	return
}

func (m *Migrator) migrateChargerProfileFiltersV2() (err error) {
	var ids []string
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.ChargerProfilePrefix)
	if err != nil {
		return fmt.Errorf("error: <%s> when getting charger profile IDs", err)
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.ChargerProfilePrefix+tenant+":")
		cpp, err := m.dmIN.DataManager().GetChargerProfile(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {
			return fmt.Errorf("error: <%s> when getting charger profile with tenant: <%s> and id: <%s>",
				err.Error(), tenant, idg)
		}
		if cpp == nil || m.dryRun {
			continue
		}
		for i, fl := range cpp.FilterIDs {
			cpp.FilterIDs[i] = migrateInlineFilterV2(fl)
		}
		if err := m.dmOut.DataManager().SetChargerProfile(cpp, true); err != nil {
			return fmt.Errorf("error: <%s> when setting charger profile with tenant: <%s> and id: <%s>",
				err.Error(), tenant, idg)
		}
		m.stats[utils.RQF] += 1
	}
	return
}

func (m *Migrator) migrateDispatcherProfileFiltersV2() (err error) {
	var ids []string
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.DispatcherProfilePrefix)
	if err != nil {
		return fmt.Errorf("error: <%s> when getting dispatcher profile IDs", err)
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.DispatcherProfilePrefix+tenant+":")
		dpp, err := m.dmIN.DataManager().GetDispatcherProfile(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {
			return fmt.Errorf("error: <%s> when getting dispatcher profile with tenant: <%s> and id: <%s>",
				err.Error(), tenant, idg)
		}
		if dpp == nil || m.dryRun {
			continue
		}
		for i, fl := range dpp.FilterIDs {
			dpp.FilterIDs[i] = migrateInlineFilterV2(fl)
		}
		if err := m.dmOut.DataManager().SetDispatcherProfile(dpp, true); err != nil {
			return fmt.Errorf("error: <%s> when setting dispatcher profile with tenant: <%s> and id: <%s>",
				err.Error(), tenant, idg)
		}
		m.stats[utils.RQF] += 1
	}
	return
}
