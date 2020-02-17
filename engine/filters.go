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
	var fieldNameDP config.DataProvider
	var fieldValuesDP []config.DataProvider
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
		for _, fltr := range f.Rules {
			fieldNameDP, err = fS.getFieldNameDataProvider(ev, fltr.Element, tenant)
			if err != nil {
				return pass, err
			}
			fieldValuesDP, err = fS.getFieldValuesDataProviders(ev, fltr.Values, tenant)
			if err != nil {
				return pass, err
			}
			if pass, err = fltr.Pass(fieldNameDP, fieldValuesDP); err != nil || !pass {
				return pass, err
			}
		}
		pass = true
	}
	return
}

// NewFilterFromInline parses an inline rule into a compiled Filter
func NewFilterFromInline(tenant, inlnRule string) (f *Filter, err error) {
	ruleSplt := strings.Split(inlnRule, utils.InInFieldSep)
	if len(ruleSplt) < 3 {
		return nil, fmt.Errorf("inline parse error for string: <%s>", inlnRule)
	}
	f = &Filter{
		Tenant: tenant,
		ID:     inlnRule,
		Rules: []*FilterRule{{
			Type:    ruleSplt[0],
			Element: ruleSplt[1],
			Values:  strings.Split(strings.Join(ruleSplt[2:], utils.InInFieldSep), utils.INFIELD_SEP),
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
	}
	return
}

// Pass is the method which should be used from outside.
func (fltr *FilterRule) Pass(fieldNameDP config.DataProvider,
	fieldValuesDP []config.DataProvider) (result bool, err error) {
	if fltr.negative == nil {
		fltr.negative = utils.BoolPointer(strings.HasPrefix(fltr.Type, utils.MetaNot))
	}

	switch fltr.Type {
	case utils.MetaString, utils.MetaNotString:
		result, err = fltr.passString(fieldNameDP, fieldValuesDP)
	case utils.MetaEmpty, utils.MetaNotEmpty:
		result, err = fltr.passEmpty(fieldNameDP)
	case utils.MetaExists, utils.MetaNotExists:
		result, err = fltr.passExists(fieldNameDP)
	case utils.MetaPrefix, utils.MetaNotPrefix:
		result, err = fltr.passStringPrefix(fieldNameDP, fieldValuesDP)
	case utils.MetaSuffix, utils.MetaNotSuffix:
		result, err = fltr.passStringSuffix(fieldNameDP, fieldValuesDP)
	case utils.MetaTimings, utils.MetaNotTimings:
		result, err = fltr.passTimings(fieldNameDP, fieldValuesDP)
	case utils.MetaDestinations, utils.MetaNotDestinations:
		result, err = fltr.passDestinations(fieldNameDP, fieldValuesDP)
	case utils.MetaRSR, utils.MetaNotRSR:
		result, err = fltr.passRSR(fieldValuesDP)
	case utils.MetaLessThan, utils.MetaLessOrEqual, utils.MetaGreaterThan, utils.MetaGreaterOrEqual:
		result, err = fltr.passGreaterThan(fieldNameDP, fieldValuesDP)
	case utils.MetaEqual, utils.MetaNotEqual:
		result, err = fltr.passEqualTo(fieldNameDP, fieldValuesDP)
	default:
		err = utils.ErrPrefixNotErrNotImplemented(fltr.Type)
	}
	if err != nil {
		return false, err
	}
	return result != *(fltr.negative), nil
}

func (fltr *FilterRule) passString(fielNameDP config.DataProvider, fieldValuesDP []config.DataProvider) (bool, error) {
	strVal, err := config.DPDynamicString(fltr.Element, fielNameDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	for i, val := range fltr.Values {
		sval, err := config.DPDynamicString(val, fieldValuesDP[i])
		if err != nil {
			continue
		}
		if strVal == sval {
			return true, nil
		}
	}
	return false, nil
}

func (fltr *FilterRule) passExists(fielNameDP config.DataProvider) (bool, error) {
	_, err := config.DPDynamicInterface(fltr.Element, fielNameDP)
	if err != nil {
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

func (fltr *FilterRule) passStringPrefix(fielNameDP config.DataProvider, fieldValuesDP []config.DataProvider) (bool, error) {
	strVal, err := config.DPDynamicString(fltr.Element, fielNameDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	for i, prfx := range fltr.Values {
		prfx, err := config.DPDynamicString(prfx, fieldValuesDP[i])
		if err != nil {
			continue
		}
		if strings.HasPrefix(strVal, prfx) {
			return true, nil
		}
	}
	return false, nil
}

func (fltr *FilterRule) passStringSuffix(fielNameDP config.DataProvider, fieldValuesDP []config.DataProvider) (bool, error) {
	strVal, err := config.DPDynamicString(fltr.Element, fielNameDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	for i, prfx := range fltr.Values {
		prfx, err := config.DPDynamicString(prfx, fieldValuesDP[i])
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
func (fltr *FilterRule) passTimings(fielNameDP config.DataProvider, fieldValuesDP []config.DataProvider) (bool, error) {
	return false, utils.ErrNotImplemented
}

func (fltr *FilterRule) passDestinations(fielNameDP config.DataProvider, fieldValuesDP []config.DataProvider) (bool, error) {
	dst, err := config.DPDynamicString(fltr.Element, fielNameDP)
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
			for i, valDstID := range fltr.Values {
				valDstID, err := config.DPDynamicString(valDstID, fieldValuesDP[i])
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

func (fltr *FilterRule) passRSR(fieldValuesDP []config.DataProvider) (bool, error) {
	_, err := fltr.rsrFields.ParseDataProviderWithInterfaces(fieldValuesDP[0], utils.NestingSep)
	if err != nil {
		if err == utils.ErrNotFound || err == utils.ErrFilterNotPassingNoCaps {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (fltr *FilterRule) passGreaterThan(fielNameDP config.DataProvider, fieldValuesDP []config.DataProvider) (bool, error) {
	fldIf, err := config.DPDynamicInterface(fltr.Element, fielNameDP)
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
	for i, val := range fltr.Values {
		sval, err := config.DPDynamicInterface(val, fieldValuesDP[i])
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

func (fltr *FilterRule) passEqualTo(fielNameDP config.DataProvider, fieldValuesDP []config.DataProvider) (bool, error) {
	fldIf, err := config.DPDynamicInterface(fltr.Element, fielNameDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	if fldStr, castStr := fldIf.(string); castStr { // attempt converting string since deserialization fails here (ie: time.Time fields)
		fldIf = utils.StringToInterface(fldStr)
	}
	for i, val := range fltr.Values {
		sval, err := config.DPDynamicInterface(val, fieldValuesDP[i])
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

func (fS *FilterS) getFieldNameDataProvider(initialDP config.DataProvider,
	fieldName string, tenant string) (dp config.DataProvider, err error) {
	switch {
	case strings.HasPrefix(fieldName, utils.DynamicDataPrefix+utils.MetaAccounts):
		// sample of fieldName : ~*accounts.1001.BalanceMap.*monetary[0].Value
		// split the field name in 3 parts
		// fieldNameType (~*accounts), accountID(1001) and quried part (BalanceMap.*monetary[0].Value)
		splitFldName := strings.SplitN(fieldName, utils.NestingSep, 3)
		if len(splitFldName) != 3 {
			return nil, fmt.Errorf("invalid fieldname <%s>", fieldName)
		}
		var account Account
		if err = fS.connMgr.Call(fS.cfg.FilterSCfg().ApierSConns, nil, utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: tenant, Account: splitFldName[1]}, &account); err != nil {
			return
		}
		//construct dataProvider from account and set it furthder
		dp = config.NewObjectDP(account, []string{utils.MetaAccounts, splitFldName[1]})
	case strings.HasPrefix(fieldName, utils.DynamicDataPrefix+utils.MetaResources):
		// sample of fieldName : ~*resources.ResourceID.Field
		splitFldName := strings.SplitN(fieldName, utils.NestingSep, 3)
		if len(splitFldName) != 3 {
			return nil, fmt.Errorf("invalid fieldname <%s>", fieldName)
		}
		var reply *Resource
		if err := fS.connMgr.Call(fS.cfg.FilterSCfg().ResourceSConns, nil, utils.ResourceSv1GetResource, &utils.TenantIDWithArgDispatcher{
			TenantID: &utils.TenantID{Tenant: tenant, ID: splitFldName[1]}}, &reply); err != nil {
			return nil, err
		}
		dp = config.NewObjectDP(reply, []string{utils.MetaResources, reply.ID})
	case strings.HasPrefix(fieldName, utils.DynamicDataPrefix+utils.MetaStats):
		// sample of fieldName : ~*stats.StatID.*acd
		splitFldName := strings.SplitN(fieldName, utils.NestingSep, 3)
		if len(splitFldName) != 3 {
			return nil, fmt.Errorf("invalid fieldname <%s>", fieldName)
		}
		var statValues map[string]float64

		if err := fS.connMgr.Call(fS.cfg.FilterSCfg().StatSConns, nil, utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{Tenant: tenant, ID: splitFldName[1]}},
			&statValues); err != nil {
			return nil, err
		}
		evNm := config.NewNavigableMap(nil)
		for k, v := range statValues {
			evNm.Set([]string{utils.MetaStats, splitFldName[1], k}, v, false, false)
		}
		dp = evNm
	case strings.HasPrefix(fieldName, utils.DynamicDataPrefix+utils.MetaReq),
		strings.HasPrefix(fieldName, utils.DynamicDataPrefix+utils.MetaVars),
		strings.HasPrefix(fieldName, utils.DynamicDataPrefix+utils.MetaCgreq),
		strings.HasPrefix(fieldName, utils.DynamicDataPrefix+utils.MetaCgrep),
		strings.HasPrefix(fieldName, utils.DynamicDataPrefix+utils.MetaRep),
		strings.HasPrefix(fieldName, utils.DynamicDataPrefix+utils.MetaCGRAReq),
		strings.HasPrefix(fieldName, utils.DynamicDataPrefix+utils.MetaAct),
		strings.HasPrefix(fieldName, utils.DynamicDataPrefix+utils.MetaEC):
		dp = initialDP
	// don't need to take out the prefix because the navigable map have ~*req prefix
	case fieldName == utils.EmptyString:
	default:
		return nil, utils.ErrPrefixNotFound(fmt.Sprintf(" data provider prefix for <%s>", fieldName))
	}
	return
}

func (fS *FilterS) getFieldValuesDataProviders(initialDP config.DataProvider,
	values []string, tenant string) (dp []config.DataProvider, err error) {
	dp = make([]config.DataProvider, len(values))
	for i := range values {
		if dp[i], err = fS.getFieldValueDataProvider(initialDP, values[i], tenant); err != nil {
			return
		}
	}
	return
}

func (fS *FilterS) getFieldValueDataProvider(initialDP config.DataProvider,
	fieldValue string, tenant string) (dp config.DataProvider, err error) {
	switch {
	case strings.HasPrefix(fieldValue, utils.DynamicDataPrefix+utils.MetaAccounts):
		// sample of fieldName : ~*accounts.1001.BalanceMap.*monetary[0].Value
		// split the field name in 3 parts
		// fieldNameType (~*accounts), accountID(1001) and quried part (BalanceMap.*monetary[0].Value)
		splitFldName := strings.SplitN(fieldValue, utils.NestingSep, 3)
		if len(splitFldName) != 3 {
			return nil, fmt.Errorf("invalid fieldname <%s>", fieldValue)
		}
		var account *Account
		if err = fS.connMgr.Call(fS.cfg.FilterSCfg().ApierSConns, nil, utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: tenant, Account: splitFldName[1]}, &account); err != nil {
			return
		}
		//construct dataProvider from account and set it furthder
		dp = config.NewObjectDP(account, []string{utils.MetaAccounts, account.ID})
	case strings.HasPrefix(fieldValue, utils.DynamicDataPrefix+utils.MetaResources):
		// sample of fieldName : ~*resources.ResourceID.Field
		splitFldName := strings.SplitN(fieldValue, utils.NestingSep, 3)
		if len(splitFldName) != 3 {
			return nil, fmt.Errorf("invalid fieldname <%s>", fieldValue)
		}
		var reply *Resource
		if err := fS.connMgr.Call(fS.cfg.FilterSCfg().ResourceSConns, nil, utils.ResourceSv1GetResource,
			&utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{
				Tenant: tenant, ID: splitFldName[1]}}, &reply); err != nil {
			return nil, err
		}
		dp = config.NewObjectDP(reply, []string{utils.MetaResources, reply.ID})
	case strings.HasPrefix(fieldValue, utils.DynamicDataPrefix+utils.MetaStats):
		// sample of fieldName : ~*resources.ResourceID.Field
		splitFldName := strings.SplitN(fieldValue, utils.NestingSep, 3)
		if len(splitFldName) != 3 {
			return nil, fmt.Errorf("invalid fieldname <%s>", fieldValue)
		}
		var statValues map[string]float64

		if err := fS.connMgr.Call(fS.cfg.FilterSCfg().StatSConns, nil, utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{Tenant: tenant, ID: splitFldName[1]}},
			&statValues); err != nil {
			return nil, err
		}
		ifaceMetric := make(map[string]interface{})
		for k, v := range statValues {
			ifaceMetric[k] = v
		}
		evNm := config.NewNavigableMap(nil)
		evNm.Set([]string{utils.MetaStats, splitFldName[1]}, ifaceMetric, false, false)
		dp = evNm
	case strings.HasPrefix(fieldValue, utils.DynamicDataPrefix+utils.MetaReq),
		strings.HasPrefix(fieldValue, utils.DynamicDataPrefix+utils.MetaVars),
		strings.HasPrefix(fieldValue, utils.DynamicDataPrefix+utils.MetaCgreq),
		strings.HasPrefix(fieldValue, utils.DynamicDataPrefix+utils.MetaCgrep),
		strings.HasPrefix(fieldValue, utils.DynamicDataPrefix+utils.MetaRep),
		strings.HasPrefix(fieldValue, utils.DynamicDataPrefix+utils.MetaCGRAReq),
		strings.HasPrefix(fieldValue, utils.DynamicDataPrefix+utils.MetaAct):
		dp = initialDP
	default: // in case of constant we give an empty DataProvider ( empty navigable map )
		dp = config.NewNavigableMap(nil)
	}

	return
}
