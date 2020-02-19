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
	"fmt"
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// NewFilterS initializtes the filter service
func NewFilterS(cfg *config.CGRConfig, connMgr *ConnManager, dm *DataManager) (fS *FilterS) {
	fS = &FilterS{
		dm:      dm,
		cfg:     cfg,
		connMgr: connMgr,
	}

	return
}

// FilterS is a service used to take decisions in case of filters
// uses lazy connections where necessary to avoid deadlocks on service startup
type FilterS struct {
	cfg     *config.CGRConfig
	dm      *DataManager
	connMgr *ConnManager
}

// Pass will check all filters wihin filterIDs and require them passing for dataProvider
// there should be at least one filter passing, ie: if filters are not active event will fail to pass
// receives the event as DataProvider so we can accept undecoded data (ie: HttpRequest)
func (fS *FilterS) Pass(tenant string, filterIDs []string,
	ev config.DataProvider) (pass bool, err error) {
	if len(filterIDs) == 0 {
		return true, nil
	}
	for _, fltrID := range filterIDs {
		f, err := GetFilter(fS.dm, tenant, fltrID,
			true, true, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				err = utils.ErrPrefixNotFound(fltrID)
			}
			return false, err
		}
		if f.ActivationInterval != nil &&
			!f.ActivationInterval.IsActiveAtTime(time.Now()) { // not active
			continue
		}
		dDP := newDynamicDP(fS.cfg, fS.connMgr, tenant, ev)
		for _, fltr := range f.Rules {
			if pass, err = fltr.Pass(dDP); err != nil || !pass {
				return pass, err
			}
		}
		pass = true
	}
	return
}

// NewFilterFromInline parses an inline rule into a compiled Filter
func NewFilterFromInline(tenant, inlnRule string) (f *Filter, err error) {
	ruleSplt := strings.SplitN(inlnRule, utils.InInFieldSep, 3)
	if len(ruleSplt) != 3 {
		return nil, fmt.Errorf("inline parse error for string: <%s>", inlnRule)
	}
	var vals []string
	if ruleSplt[2] != utils.EmptyString {
		vals = strings.Split(ruleSplt[2], utils.INFIELD_SEP)
	}
	f = &Filter{
		Tenant: tenant,
		ID:     inlnRule,
		Rules: []*FilterRule{{
			Type:    ruleSplt[0],
			Element: ruleSplt[1],
			Values:  vals,
		}},
	}
	if err = f.Compile(); err != nil {
		return nil, err
	}
	return
}

// Filter structure to define a basic filter
type Filter struct {
	Tenant             string
	ID                 string
	Rules              []*FilterRule
	ActivationInterval *utils.ActivationInterval
}

type FilterWithArgDispatcher struct {
	*Filter
	*utils.ArgDispatcher
}

// TenantID returns the tenant wit the ID
func (fltr *Filter) TenantID() string {
	return utils.ConcatenatedKey(fltr.Tenant, fltr.ID)
}

// Compile will compile the underlaying request filters where necessary (ie. regexp rules)
func (fltr *Filter) Compile() (err error) {
	for _, rf := range fltr.Rules {
		if err = rf.CompileValues(); err != nil {
			return
		}
	}
	return
}

var supportedFiltersType *utils.StringSet = utils.NewStringSet([]string{utils.MetaString, utils.MetaPrefix, utils.MetaSuffix,
	utils.MetaTimings, utils.MetaRSR, utils.MetaDestinations,
	utils.MetaEmpty, utils.MetaExists, utils.MetaLessThan, utils.MetaLessOrEqual,
	utils.MetaGreaterThan, utils.MetaGreaterOrEqual, utils.MetaEqual,
	utils.MetaNotEqual})
var needsFieldName *utils.StringSet = utils.NewStringSet([]string{utils.MetaString, utils.MetaPrefix,
	utils.MetaSuffix, utils.MetaTimings, utils.MetaDestinations, utils.MetaLessThan,
	utils.MetaEmpty, utils.MetaExists, utils.MetaLessOrEqual, utils.MetaGreaterThan,
	utils.MetaGreaterOrEqual, utils.MetaEqual, utils.MetaNotEqual})
var needsValues *utils.StringSet = utils.NewStringSet([]string{utils.MetaString, utils.MetaPrefix,
	utils.MetaSuffix, utils.MetaTimings, utils.MetaRSR, utils.MetaDestinations,
	utils.MetaLessThan, utils.MetaLessOrEqual, utils.MetaGreaterThan, utils.MetaGreaterOrEqual,
	utils.MetaEqual, utils.MetaNotEqual})

// NewFilterRule returns a new filter
func NewFilterRule(rfType, fieldName string, vals []string) (*FilterRule, error) {
	var negative bool
	rType := rfType
	if strings.HasPrefix(rfType, utils.MetaNot) {
		rType = "*" + strings.TrimPrefix(rfType, utils.MetaNot)
		negative = true
	}
	if !supportedFiltersType.Has(rType) {
		return nil, fmt.Errorf("Unsupported filter Type: %s", rfType)
	}
	if fieldName == "" && needsFieldName.Has(rType) {
		return nil, fmt.Errorf("Element is mandatory for Type: %s", rfType)
	}
	if len(vals) == 0 && needsValues.Has(rType) {
		return nil, fmt.Errorf("Values is mandatory for Type: %s", rfType)
	}
	rf := &FilterRule{
		Type:     rfType,
		Element:  fieldName,
		Values:   vals,
		negative: utils.BoolPointer(negative),
	}
	if err := rf.CompileValues(); err != nil {
		return nil, err
	}
	return rf, nil
}

// FilterRule filters requests coming into various places
// Pass rule: default negative, one mathing rule should pass the filter
type FilterRule struct {
	Type      string            // Filter type (*string, *timing, *rsr_filters, *stats, *lt, *lte, *gt, *gte)
	Element   string            // Name of the field providing us the Values to check (used in case of some )
	Values    []string          // Filter definition
	rsrFields config.RSRParsers // Cache here the RSRFilter Values
	negative  *bool
}

// CompileValues compiles RSR fields
func (fltr *FilterRule) CompileValues() (err error) {
	switch fltr.Type {
	case utils.MetaRSR, utils.MetaNotRSR:
		if fltr.rsrFields, err = config.NewRSRParsersFromSlice(fltr.Values, true); err != nil {
			return
		}
	case utils.MetaExists, utils.MetaNotExists:
		if len(fltr.Values) != 0 {
			if fltr.rsrFields, err = config.NewRSRParsersFromSlice(fltr.Values, true); err != nil {
				return
			}
		}
	}
	return
}

// Pass is the method which should be used from outside.
func (fltr *FilterRule) Pass(dDP config.DataProvider) (result bool, err error) {
	if fltr.negative == nil {
		fltr.negative = utils.BoolPointer(strings.HasPrefix(fltr.Type, utils.MetaNot))
	}

	switch fltr.Type {
	case utils.MetaString, utils.MetaNotString:
		result, err = fltr.passString(dDP)
	case utils.MetaEmpty, utils.MetaNotEmpty:
		result, err = fltr.passEmpty(dDP)
	case utils.MetaExists, utils.MetaNotExists:
		result, err = fltr.passExists(dDP)
	case utils.MetaPrefix, utils.MetaNotPrefix:
		result, err = fltr.passStringPrefix(dDP)
	case utils.MetaSuffix, utils.MetaNotSuffix:
		result, err = fltr.passStringSuffix(dDP)
	case utils.MetaTimings, utils.MetaNotTimings:
		result, err = fltr.passTimings(dDP)
	case utils.MetaDestinations, utils.MetaNotDestinations:
		result, err = fltr.passDestinations(dDP)
	case utils.MetaRSR, utils.MetaNotRSR:
		result, err = fltr.passRSR(dDP)
	case utils.MetaLessThan, utils.MetaLessOrEqual, utils.MetaGreaterThan, utils.MetaGreaterOrEqual:
		result, err = fltr.passGreaterThan(dDP)
	case utils.MetaEqual, utils.MetaNotEqual:
		result, err = fltr.passEqualTo(dDP)
	default:
		err = utils.ErrPrefixNotErrNotImplemented(fltr.Type)
	}
	if err != nil {
		return false, err
	}
	return result != *(fltr.negative), nil
}

func (fltr *FilterRule) passString(dDP config.DataProvider) (bool, error) {
	strVal, err := config.DPDynamicString(fltr.Element, dDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	for _, val := range fltr.Values {
		sval, err := config.DPDynamicString(val, dDP)
		if err != nil {
			continue
		}
		if strVal == sval {
			return true, nil
		}
	}
	return false, nil
}

func (fltr *FilterRule) passExists(dDP config.DataProvider) (bool, error) {
	var err error
	path := fltr.Element
	if fltr.rsrFields != nil {
		if path, err = fltr.rsrFields.ParseDataProviderWithInterfaces(dDP, utils.NestingSep); err != nil {
			return false, err
		}
	}
	if _, err = config.DPDynamicInterface(path, dDP); err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (fltr *FilterRule) passEmpty(fielNameDP config.DataProvider) (bool, error) {
	val, err := config.DPDynamicInterface(fltr.Element, fielNameDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return true, nil
		}
		return false, err
	}
	if val == nil {
		return true, nil
	}
	rval := reflect.ValueOf(val)
	if rval.Type().Kind() == reflect.Ptr {
		if rval.IsNil() {
			return true, nil
		}
		rval = rval.Elem()
	}
	switch rval.Type().Kind() {
	case reflect.String:
		return rval.Interface() == "", nil
	case reflect.Slice:
		return rval.Len() == 0, nil
	case reflect.Map:
		return len(rval.MapKeys()) == 0, nil
	default:
		return false, nil
	}
}

func (fltr *FilterRule) passStringPrefix(dDP config.DataProvider) (bool, error) {
	strVal, err := config.DPDynamicString(fltr.Element, dDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	for _, prfx := range fltr.Values {
		prfx, err := config.DPDynamicString(prfx, dDP)
		if err != nil {
			continue
		}
		if strings.HasPrefix(strVal, prfx) {
			return true, nil
		}
	}
	return false, nil
}

func (fltr *FilterRule) passStringSuffix(dDP config.DataProvider) (bool, error) {
	strVal, err := config.DPDynamicString(fltr.Element, dDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	for _, prfx := range fltr.Values {
		prfx, err := config.DPDynamicString(prfx, dDP)
		if err != nil {
			continue
		}
		if strings.HasSuffix(strVal, prfx) {
			return true, nil
		}
	}
	return false, nil
}

// ToDo when Timings will be available in DataDb
func (fltr *FilterRule) passTimings(dDP config.DataProvider) (bool, error) {
	return false, utils.ErrNotImplemented
}

func (fltr *FilterRule) passDestinations(dDP config.DataProvider) (bool, error) {
	dst, err := config.DPDynamicString(fltr.Element, dDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	for _, p := range utils.SplitPrefix(dst, MIN_PREFIX_MATCH) {
		var destIDs []string
		if err = connMgr.Call(config.CgrConfig().FilterSCfg().ApierSConns, nil, utils.APIerSv1GetReverseDestination,
			p, &destIDs); err != nil {
			continue
		}
		for _, dID := range destIDs {
			for _, valDstID := range fltr.Values {
				valDstID, err := config.DPDynamicString(valDstID, dDP)
				if err != nil {
					continue
				}
				if valDstID == dID {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func (fltr *FilterRule) passRSR(dDP config.DataProvider) (bool, error) {
	_, err := fltr.rsrFields.ParseDataProviderWithInterfaces(dDP, utils.NestingSep)
	if err != nil {
		if err == utils.ErrNotFound || err == utils.ErrFilterNotPassingNoCaps {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (fltr *FilterRule) passGreaterThan(dDP config.DataProvider) (bool, error) {
	fldIf, err := config.DPDynamicInterface(fltr.Element, dDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	if fldStr, castStr := fldIf.(string); castStr { // attempt converting string since deserialization fails here (ie: time.Time fields)
		fldIf = utils.StringToInterface(fldStr)
	}
	orEqual := false
	if fltr.Type == utils.MetaGreaterOrEqual ||
		fltr.Type == utils.MetaLessThan {
		orEqual = true
	}
	for _, val := range fltr.Values {
		sval, err := config.DPDynamicInterface(val, dDP)
		if err != nil {
			continue
		}
		if gte, err := utils.GreaterThan(fldIf, sval, orEqual); err != nil {
			return false, err
		} else if utils.SliceHasMember([]string{utils.MetaGreaterThan, utils.MetaGreaterOrEqual}, fltr.Type) && gte {
			return true, nil
		} else if utils.SliceHasMember([]string{utils.MetaLessThan, utils.MetaLessOrEqual}, fltr.Type) && !gte {
			return true, nil
		}
	}
	return false, nil
}

func (fltr *FilterRule) passEqualTo(dDP config.DataProvider) (bool, error) {
	fldIf, err := config.DPDynamicInterface(fltr.Element, dDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	if fldStr, castStr := fldIf.(string); castStr { // attempt converting string since deserialization fails here (ie: time.Time fields)
		fldIf = utils.StringToInterface(fldStr)
	}
	for _, val := range fltr.Values {
		sval, err := config.DPDynamicInterface(val, dDP)
		if err != nil {
			continue
		}
		if eq, err := utils.EqualTo(fldIf, sval); err != nil {
			return false, err
		} else if eq {
			return true, nil
		}
	}
	return false, nil
}

func newDynamicDP(cfg *config.CGRConfig, connMgr *ConnManager,
	tenant string, initialDP config.DataProvider) *dynamicDP {
	return &dynamicDP{
		cfg:       cfg,
		connMgr:   connMgr,
		tenant:    tenant,
		initialDP: initialDP,
		cache:     config.NewNavigableMap(nil),
	}
}

type dynamicDP struct {
	cfg       *config.CGRConfig
	connMgr   *ConnManager
	tenant    string
	initialDP config.DataProvider

	cache *config.NavigableMap
}

func (dDP *dynamicDP) String() string { return utils.ToJSON(dDP) }

func (dDP *dynamicDP) FieldAsString(fldPath []string) (string, error) {
	val, err := dDP.FieldAsInterface(fldPath)
	if err != nil {
		return "", err
	}
	return utils.IfaceAsString(val), nil
}
func (dDP *dynamicDP) AsNavigableMap([]*config.FCTemplate) (*config.NavigableMap, error) {
	return nil, utils.ErrNotImplemented
}
func (dDP *dynamicDP) RemoteHost() net.Addr {
	return utils.LocalAddr()
}

var initialDPPrefixes = utils.NewStringSet([]string{utils.MetaReq, utils.MetaVars,
	utils.MetaCgreq, utils.MetaCgrep, utils.MetaRep, utils.MetaCGRAReq,
	utils.MetaAct, utils.MetaEC})

func (dDP *dynamicDP) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	if len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}
	if initialDPPrefixes.Has(fldPath[0]) {
		return dDP.initialDP.FieldAsInterface(fldPath)
	}
	val, err = dDP.cache.FieldAsInterface(fldPath)
	if err == utils.ErrNotFound { // in case not found in cache try to populate it
		return dDP.fieldAsInterface(fldPath)
	}
	return
}

func (dDP *dynamicDP) fieldAsInterface(fldPath []string) (val interface{}, err error) {
	if len(fldPath) < 2 {
		return nil, fmt.Errorf("invalid fieldname <%s>", fldPath)
	}
	switch fldPath[0] {
	case utils.MetaAccounts:
		// sample of fieldName : ~*accounts.1001.BalanceMap.*monetary[0].Value
		// split the field name in 3 parts
		// fieldNameType (~*accounts), accountID(1001) and quried part (BalanceMap.*monetary[0].Value)

		var account Account
		if err = dDP.connMgr.Call(dDP.cfg.FilterSCfg().ApierSConns, nil, utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: dDP.tenant, Account: fldPath[1]}, &account); err != nil {
			return
		}
		//construct dataProvider from account and set it furthder
		dp := config.NewObjectDP(account)
		dDP.cache.Set(fldPath[:2], dp, false, false)
		return dp.FieldAsInterface(fldPath[2:])
	case utils.MetaResources:
		// sample of fieldName : ~*resources.ResourceID.Field
		var reply *Resource
		if err := dDP.connMgr.Call(dDP.cfg.FilterSCfg().ResourceSConns, nil, utils.ResourceSv1GetResource,
			&utils.TenantID{Tenant: dDP.tenant, ID: fldPath[1]}, &reply); err != nil {
			return nil, err
		}
		dp := config.NewObjectDP(reply)
		dDP.cache.Set(fldPath[:2], dp, false, false)
		return dp.FieldAsInterface(fldPath[2:])
	case utils.MetaStats:
		// sample of fieldName : ~*stats.StatID.*acd
		var statValues map[string]float64

		if err := dDP.connMgr.Call(dDP.cfg.FilterSCfg().StatSConns, nil, utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{Tenant: dDP.tenant, ID: fldPath[1]}},
			&statValues); err != nil {
			return nil, err
		}
		for k, v := range statValues {
			dDP.cache.Set([]string{utils.MetaStats, fldPath[1], k}, v, false, false)
		}
		return dDP.cache.FieldAsInterface(fldPath)
	default: // in case of constant we give an empty DataProvider ( empty navigable map )
	}
	return nil, utils.ErrNotFound
}
