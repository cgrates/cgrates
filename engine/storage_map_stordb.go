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
	"github.com/cgrates/cgrates/utils"
)

//implement LoadReader interface
func (ms *MapStorage) GetTpIds(colName string) (ids []string, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTpTableIds(tpid, table string, distinct utils.TPDistinctIds,
	filters map[string]string, paginator *utils.Paginator) (ids []string, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPTimings(tpid, id string) (timings []*utils.ApierTPTiming, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPDestinations(tpid, id string) (dsts []*utils.TPDestination, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPRates(tpid, id string) (rates []*utils.TPRate, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPDestinationRates(tpid, id string,
	paginator *utils.Paginator) (dRates []*utils.TPDestinationRate, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPRatingPlans(string, string, *utils.Paginator) (rPlans []*utils.TPRatingPlan, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPRatingProfiles(filter *utils.TPRatingProfile) (rProfiles []*utils.TPRatingProfile, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPSharedGroups(tpid, id string) (sGroups []*utils.TPSharedGroups, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPUsers(filter *utils.TPUsers) (users []*utils.TPUsers, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPAliases(filter *utils.TPAliases) (aliases []*utils.TPAliases, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPDerivedChargers(*utils.TPDerivedChargers) (dCharges []*utils.TPDerivedChargers, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPActions(tpid, id string) (actions []*utils.TPActions, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPActionPlans(tpid, id string) (aPlans []*utils.TPActionPlan, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPActionTriggers(tpid, id string) (aTriggers []*utils.TPActionTriggers, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPAccountActions(filter *utils.TPAccountActions) (accounts []*utils.TPAccountActions, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPResources(tpid, id string) (resources []*utils.TPResource, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPStats(tpid, id string) (stats []*utils.TPStats, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPThresholds(tpid, id string) (ths []*utils.TPThreshold, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPFilters(tpid, id string) (fltrs []*utils.TPFilterProfile, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPSuppliers(tpid, id string) (supps []*utils.TPSupplierProfile, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPAttributes(tpid, id string) (attrs []*utils.TPAttributeProfile, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPChargers(tpid, id string) (attrs []*utils.TPChargerProfile, err error) {
	return nil, utils.ErrNotImplemented
}

//implement LoadWriter interface
func (ms *MapStorage) RemTpData(table, tpid string, args map[string]string) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPTimings(timings []*utils.ApierTPTiming) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPDestinations(dests []*utils.TPDestination) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPRates(rates []*utils.TPRate) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPDestinationRates(dRates []*utils.TPDestinationRate) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPRatingPlans(ratingPlans []*utils.TPRatingPlan) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPRatingProfiles(ratingProfiles []*utils.TPRatingProfile) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPSharedGroups(groups []*utils.TPSharedGroups) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPUsers(users []*utils.TPUsers) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPAliases(aliases []*utils.TPAliases) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPDerivedChargers(dc []*utils.TPDerivedChargers) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPActions(acts []*utils.TPActions) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPActionPlans(aPlans []*utils.TPActionPlan) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPActionTriggers(aTriggers []*utils.TPActionTriggers) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPAccountActions(accActions []*utils.TPAccountActions) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPResources(resources []*utils.TPResource) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPStats(stats []*utils.TPStats) (err error) {
	return utils.ErrNotImplemented
}

func (ms *MapStorage) SetTPThresholds(thresholds []*utils.TPThreshold) (err error) {
	return utils.ErrNotImplemented
}

func (ms *MapStorage) SetTPFilters(filters []*utils.TPFilterProfile) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPSuppliers(suppliers []*utils.TPSupplierProfile) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPAttributes(attributes []*utils.TPAttributeProfile) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPChargers(attributes []*utils.TPChargerProfile) (err error) {
	return utils.ErrNotImplemented
}

//implement CdrStorage interface
func (ms *MapStorage) SetCDR(cdr *CDR, allowUpdate bool) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) RemoveSMCost(smc *SMCost) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) GetCDRs(filter *utils.CDRsFilter, remove bool) (cdrs []*CDR, count int64, err error) {
	return nil, 0, utils.ErrNotImplemented
}

func (ms *MapStorage) GetSMCosts(cgrid, runid, originHost, originIDPrfx string) (smCosts []*SMCost, err error) {
	return nil, utils.ErrNotImplemented
}
