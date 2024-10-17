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
	"regexp"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
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

// Pass will check all filters within filterIDs and require them passing for dataProvider
// there should be at least one filter passing, ie: if filters are not active event will fail to pass
// receives the event as DataProvider so we can accept undecoded data (ie: HttpRequest)
func (fS *FilterS) Pass(tenant string, filterIDs []string,
	ev utils.DataProvider) (pass bool, err error) {
	if len(filterIDs) == 0 {
		return true, nil
	}
	dDP := newDynamicDP(fS.cfg.FilterSCfg().ResourceSConns, fS.cfg.FilterSCfg().StatSConns,
		fS.cfg.FilterSCfg().ApierSConns, fS.cfg.FilterSCfg().TrendSConns, tenant, ev)
	for _, fltrID := range filterIDs {
		f, err := fS.dm.GetFilter(tenant, fltrID,
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
			if pass, err = fltr.Pass(dDP); err != nil || !pass {
				return pass, err
			}
		}
		pass = true
	}
	return
}

// checkPrefix verify if the value has as prefix one of the prefixes
func checkPrefix(value string, prefixes []string) (hasPrefix bool) {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			hasPrefix = true
			break
		}
	}
	if !hasPrefix {
		return false
	}
	return
}

// verifyPrefixes verify the Element and the Values from FilterRule if has as prefix one of the prefixes
func verifyPrefixes(rule *FilterRule, prefixes []string) (hasPrefix bool) {
	if strings.HasPrefix(rule.Element, utils.DynamicDataPrefix) {
		if hasPrefix = checkPrefix(rule.Element, prefixes); !hasPrefix {
			return
		}
	}
	for _, value := range rule.Values {
		hasPrefix = false // reset hasPrefix
		if strings.HasPrefix(value, utils.DynamicDataPrefix) {
			if hasPrefix = checkPrefix(value, prefixes); !hasPrefix {
				return
			}
		}
	}
	return true
}

// LazyPass is almost the same as Pass except that it verify if the
// Element of the Values from FilterRules has as prefix one of the pathPrfxs
func (fS *FilterS) LazyPass(tenant string, filterIDs []string,
	ev utils.DataProvider, pathPrfxs []string) (pass bool, lazyCheckRules []*FilterRule, err error) {
	if len(filterIDs) == 0 {
		return true, nil, nil
	}
	pass = true
	dDP := newDynamicDP(fS.cfg.FilterSCfg().ResourceSConns, fS.cfg.FilterSCfg().StatSConns,
		fS.cfg.FilterSCfg().ApierSConns, fS.cfg.FilterSCfg().TrendSConns, tenant, ev)
	for _, fltrID := range filterIDs {
		var f *Filter
		f, err = fS.dm.GetFilter(tenant, fltrID,
			true, true, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				err = utils.ErrPrefixNotFound(fltrID)
			}
			return
		}
		if f.ActivationInterval != nil &&
			!f.ActivationInterval.IsActiveAtTime(time.Now()) { // not active
			continue
		}

		for _, rule := range f.Rules {
			if !verifyPrefixes(rule, pathPrfxs) {
				lazyCheckRules = append(lazyCheckRules, rule)
				continue
			}
			if pass, err = rule.Pass(dDP); err != nil || !pass {
				return
			}
		}
	}
	return
}

func splitDynFltrValues(val, sep string) (vals []string) {
	startIdx := strings.IndexByte(val, utils.RSRDynStartChar)
	endIdx := strings.IndexByte(val, utils.RSRDynEndChar)
	if startIdx == -1 || endIdx == -1 {
		return strings.Split(val, sep)
	}

	vals = strings.Split(val[:startIdx], sep)
	vals[len(vals)-1] += val[startIdx : endIdx+1]
	val = val[endIdx+1:]
	if len(val) == 0 {
		return
	}
	valsEnd := splitDynFltrValues(val, sep)
	vals[len(vals)-1] += valsEnd[0]
	return append(vals, valsEnd[1:]...)
}

// NewFilterFromInline parses an inline rule into a compiled Filter
func NewFilterFromInline(tenant, inlnRule string) (f *Filter, err error) {
	ruleSplt := utils.SplitPath(inlnRule, utils.InInFieldSep[0], 3)
	if len(ruleSplt) != 3 {
		return nil, fmt.Errorf("inline parse error for string: <%s>", inlnRule)
	}
	var vals []string
	if ruleSplt[2] != utils.EmptyString {
		vals = splitDynFltrValues(ruleSplt[2], utils.PipeSep)
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

// FilterWithOpts the arguments for the replication
type FilterWithAPIOpts struct {
	*Filter
	APIOpts map[string]any
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

var supportedFiltersType utils.StringSet = utils.NewStringSet([]string{
	utils.MetaString, utils.MetaContains, utils.MetaPrefix, utils.MetaSuffix,
	utils.MetaTimings, utils.MetaRSR, utils.MetaDestinations, utils.MetaHTTP,
	utils.MetaEmpty, utils.MetaExists, utils.MetaLessThan, utils.MetaLessOrEqual,
	utils.MetaGreaterThan, utils.MetaGreaterOrEqual, utils.MetaEqual,
	utils.MetaIPNet, utils.MetaAPIBan, utils.MetaSentryPeer, utils.MetaActivationInterval,
	utils.MetaRegex})
var needsFieldName utils.StringSet = utils.NewStringSet([]string{
	utils.MetaString, utils.MetaContains, utils.MetaPrefix, utils.MetaSuffix,
	utils.MetaTimings, utils.MetaRSR, utils.MetaDestinations, utils.MetaLessThan,
	utils.MetaEmpty, utils.MetaExists, utils.MetaLessOrEqual, utils.MetaGreaterThan,
	utils.MetaGreaterOrEqual, utils.MetaEqual, utils.MetaIPNet, utils.MetaAPIBan, utils.MetaSentryPeer,
	utils.MetaActivationInterval,
	utils.MetaRegex})
var needsValues utils.StringSet = utils.NewStringSet([]string{utils.MetaString, utils.MetaContains, utils.MetaPrefix,
	utils.MetaSuffix, utils.MetaTimings, utils.MetaRSR, utils.MetaDestinations,
	utils.MetaLessThan, utils.MetaLessOrEqual, utils.MetaGreaterThan, utils.MetaGreaterOrEqual,
	utils.MetaEqual, utils.MetaIPNet, utils.MetaAPIBan, utils.MetaSentryPeer, utils.MetaActivationInterval,
	utils.MetaRegex})

// NewFilterRule returns a new filter
func NewFilterRule(rfType, fieldName string, vals []string) (*FilterRule, error) {
	var negative bool
	rType := rfType
	if strings.HasPrefix(rfType, utils.MetaNot) {
		rType = utils.Meta + strings.TrimPrefix(rfType, utils.MetaNot)
		negative = true
	}
	if strings.HasPrefix(rfType, utils.MetaHTTP) {
		rType = utils.MetaHTTP
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
// Pass rule: default negative, one matching rule should pass the filter
type FilterRule struct {
	Type        string            // Filter type (*string, *timing, *rsr_filters, *stats, *lt, *lte, *gt, *gte)
	Element     string            // Name of the field providing us the Values to check (used in case of some )
	Values      []string          // Filter definition
	rsrValues   config.RSRParsers // Cache here the
	rsrElement  *config.RSRParser // Cache here the
	rsrFilters  utils.RSRFilters  // Cache here the RSRFilter Values
	regexValues []*regexp.Regexp
	negative    *bool
}

// CompileValues compiles RSR fields
func (fltr *FilterRule) CompileValues() (err error) {
	switch fltr.Type {
	case utils.MetaRegex, utils.MetaNotRegex:
		fltr.regexValues = make([]*regexp.Regexp, len(fltr.Values))
		for i, val := range fltr.Values {
			if fltr.regexValues[i], err = regexp.Compile(val); err != nil {
				return
			}
		}
	case utils.MetaRSR, utils.MetaNotRSR:
		if fltr.rsrFilters, err = utils.ParseRSRFiltersFromSlice(fltr.Values); err != nil {
			return
		}
	case utils.MetaExists, utils.MetaNotExists, utils.MetaEmpty, utils.MetaNotEmpty: // only the element is builded
	case utils.MetaActivationInterval, utils.MetaNotActivationInterval:
		fltr.rsrValues = make(config.RSRParsers, len(fltr.Values))
		for i, strVal := range fltr.Values {
			if fltr.rsrValues[i], err = config.NewRSRParser(strVal); err != nil {
				return
			}
		}
	default:
		if fltr.rsrValues, err = config.NewRSRParsersFromSlice(fltr.Values); err != nil {
			return
		}
	}
	if fltr.rsrElement, err = config.NewRSRParser(fltr.Element); err != nil {
		return
	} else if fltr.rsrElement == nil {
		return fmt.Errorf("empty RSRParser in rule: <%s>", fltr.Element)
	}
	return
}

// Pass is the method which should be used from outside.
func (fltr *FilterRule) Pass(dDP utils.DataProvider) (result bool, err error) {
	if fltr.negative == nil {
		fltr.negative = utils.BoolPointer(strings.HasPrefix(fltr.Type, utils.MetaNot))
	}

	switch fltr.Type {
	case utils.MetaString, utils.MetaNotString:
		result, err = fltr.passString(dDP)
	case utils.MetaContains, utils.MetaNotContains:
		result, err = fltr.passContains(dDP)
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
	case utils.MetaIPNet, utils.MetaNotIPNet:
		result, err = fltr.passIPNet(dDP)
	case utils.MetaAPIBan, utils.MetaNotAPIBan:
		result, err = fltr.passAPIBan(dDP)
	case utils.MetaSentryPeer, utils.MetaNotSentryPeer:
		result, err = fltr.passSentryPeer(dDP)
	case utils.MetaActivationInterval, utils.MetaNotActivationInterval:
		result, err = fltr.passActivationInterval(dDP)
	case utils.MetaRegex, utils.MetaNotRegex:
		result, err = fltr.passRegex(dDP)
	default:
		if strings.HasPrefix(fltr.Type, utils.MetaHTTP) && strings.Index(fltr.Type, "#") == len(utils.MetaHTTP) {
			result, err = fltr.passHttp(dDP)
			break
		}
		err = utils.ErrPrefixNotErrNotImplemented(fltr.Type)
	}
	if err != nil {
		return false, err
	}
	return result != *(fltr.negative), nil
}

func (fltr *FilterRule) passString(dDP utils.DataProvider) (bool, error) {
	strVal, err := fltr.rsrElement.ParseDataProvider(dDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	for _, val := range fltr.rsrValues {
		sval, err := val.ParseDataProvider(dDP)
		if err != nil {
			continue
		}
		if strVal == sval {
			return true, nil
		}
	}
	return false, nil
}

func (fltr *FilterRule) passContains(dDP utils.DataProvider) (bool, error) {
	strVal, err := fltr.rsrElement.ParseDataProvider(dDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	for _, val := range fltr.rsrValues {
		sval, err := val.ParseDataProvider(dDP)
		if err != nil {
			continue
		}
		if strings.Contains(strVal, sval) {
			return true, nil
		}
	}
	return false, nil
}

func (fltr *FilterRule) passExists(dDP utils.DataProvider) (bool, error) {
	path, err := fltr.rsrElement.CompileDynRule(dDP)
	if err != nil {
		return false, err
	}
	if _, err := utils.DPDynamicInterface(path, dDP); err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (fltr *FilterRule) passEmpty(dDP utils.DataProvider) (bool, error) {
	path, err := fltr.rsrElement.CompileDynRule(dDP)
	if err != nil {
		return false, err
	}
	val, err := utils.DPDynamicInterface(path, dDP)
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

func (fltr *FilterRule) passStringPrefix(dDP utils.DataProvider) (bool, error) {
	strVal, err := fltr.rsrElement.ParseDataProvider(dDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	for _, prfxVal := range fltr.rsrValues {
		prfx, err := prfxVal.ParseDataProvider(dDP)
		if err != nil {
			continue
		}
		if strings.HasPrefix(strVal, prfx) {
			return true, nil
		}
	}
	return false, nil
}

func (fltr *FilterRule) passStringSuffix(dDP utils.DataProvider) (bool, error) {
	strVal, err := fltr.rsrElement.ParseDataProvider(dDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	for _, prfxVal := range fltr.rsrValues {
		prfx, err := prfxVal.ParseDataProvider(dDP)
		if err != nil {
			continue
		}
		if strings.HasSuffix(strVal, prfx) {
			return true, nil
		}
	}
	return false, nil
}

func (fltr *FilterRule) passTimings(dDP utils.DataProvider) (bool, error) {
	tmVal, err := fltr.rsrElement.ParseDataProviderWithInterfaces(dDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}

	tmTime, err := utils.IfaceAsTime(tmVal, config.CgrConfig().GeneralCfg().DefaultTimezone)
	if err != nil {
		return false, err
	}

	for _, valTmIDVal := range fltr.rsrValues {
		valTmID, err := valTmIDVal.ParseDataProvider(dDP)
		if err != nil {
			return false, err
		}
		var tm utils.TPTiming
		if err = connMgr.Call(context.TODO(), config.CgrConfig().FilterSCfg().ApierSConns,
			utils.APIerSv1GetTiming,
			&utils.ArgsGetTimingID{ID: valTmID}, &tm); err != nil {
			continue
		}
		ritm := &RITiming{
			ID:        tm.ID,
			Years:     tm.Years,
			Months:    tm.Months,
			MonthDays: tm.MonthDays,
			WeekDays:  tm.WeekDays,
			StartTime: tm.StartTime,
			EndTime:   tm.EndTime,
		}
		if ritm.IsActiveAt(tmTime) {
			return true, nil
		}
	}
	return false, nil
}

func (fltr *FilterRule) passDestinations(dDP utils.DataProvider) (bool, error) {
	dst, err := fltr.rsrElement.ParseDataProvider(dDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	for _, p := range utils.SplitPrefix(dst, MIN_PREFIX_MATCH) {
		var destIDs []string
		if err = connMgr.Call(context.TODO(), config.CgrConfig().FilterSCfg().ApierSConns,
			utils.APIerSv1GetReverseDestination,
			&p, &destIDs); err != nil {
			continue
		}
		for _, dID := range destIDs {
			for _, valDstIDVal := range fltr.rsrValues {
				valDstID, err := valDstIDVal.ParseDataProvider(dDP)
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

func (fltr *FilterRule) passRSR(dDP utils.DataProvider) (bool, error) {
	fld, err := fltr.rsrElement.ParseDataProvider(dDP)
	if err != nil {
		if err == utils.ErrNotFound {
			match := fltr.rsrFilters.FilterRules() == "^$"
			return match, nil
		}
		return false, err
	}
	match := fltr.rsrFilters.Pass(fld, false)
	return match, nil
}

func (fltr *FilterRule) passGreaterThan(dDP utils.DataProvider) (bool, error) {
	fldStr, err := fltr.rsrElement.ParseDataProviderWithInterfaces(dDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	fldIf := utils.StringToInterface(fldStr)
	orEqual := fltr.Type == utils.MetaGreaterOrEqual ||
		fltr.Type == utils.MetaLessThan
	for _, val := range fltr.rsrValues {
		sval, err := val.ParseDataProviderWithInterfaces(dDP)
		if err != nil {
			continue
		}
		if gte, err := utils.GreaterThan(fldIf, utils.StringToInterface(sval), orEqual); err != nil {
			return false, err
		} else if (utils.MetaGreaterThan == fltr.Type || utils.MetaGreaterOrEqual == fltr.Type) && gte {
			return true, nil
		} else if (utils.MetaLessThan == fltr.Type || utils.MetaLessOrEqual == fltr.Type) && !gte {
			return true, nil
		}
	}
	return false, nil
}

func (fltr *FilterRule) passEqualTo(dDP utils.DataProvider) (bool, error) {
	fldStr, err := fltr.rsrElement.ParseDataProviderWithInterfaces(dDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	fldIf := utils.StringToInterface(fldStr)
	for _, val := range fltr.rsrValues {
		sval, err := val.ParseDataProviderWithInterfaces(dDP)
		if err != nil {
			continue
		}
		if eq, err := utils.EqualTo(fldIf, utils.StringToInterface(sval)); err != nil {
			return false, err
		} else if eq {
			return true, nil
		}
	}
	return false, nil
}

func (fltr *FilterRule) passIPNet(dDP utils.DataProvider) (bool, error) {
	strVal, err := fltr.rsrElement.ParseDataProvider(dDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	ip := net.ParseIP(strVal)
	if ip == nil {
		return false, nil
	}

	for _, val := range fltr.rsrValues {
		sval, err := val.ParseDataProvider(dDP)
		if err != nil {
			continue
		}
		_, ipNet, err := net.ParseCIDR(sval)
		if err != nil {
			continue
		}
		if ipNet.Contains(ip) {
			return true, nil
		}
	}
	return false, nil
}

func (fltr *FilterRule) passAPIBan(dDP utils.DataProvider) (bool, error) {
	strVal, err := fltr.rsrElement.ParseDataProvider(dDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	if fltr.Values[0] != utils.MetaAll &&
		fltr.Values[0] != utils.MetaSingle { // force only valid values
		return false, fmt.Errorf("invalid value for apiban filter: <%s>", fltr.Values[0])
	}
	return dm.GetAPIBan(strVal, config.CgrConfig().APIBanCfg().Keys, fltr.Values[0] != utils.MetaAll, true, true)
}

func (fltr *FilterRule) passSentryPeer(dDP utils.DataProvider) (bool, error) {
	strVal, err := fltr.rsrElement.ParseDataProvider(dDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	if fltr.Values[0] != utils.MetaNumber && fltr.Values[0] != utils.MetaIp {
		return false, fmt.Errorf("invalid value for sentrypeer filter: <%s>", fltr.Values[0])
	}
	return GetSentryPeer(strVal, config.CgrConfig().SentryPeerCfg(), fltr.Values[0])
}

func parseTime(rsr *config.RSRParser, dDp utils.DataProvider) (_ time.Time, err error) {
	var str string
	if str, err = rsr.ParseDataProvider(dDp); err != nil {
		return
	}
	return utils.ParseTimeDetectLayout(str, config.CgrConfig().GeneralCfg().DefaultTimezone)
}

func (fltr *FilterRule) passActivationInterval(dDp utils.DataProvider) (bool, error) {
	timeVal, err := parseTime(fltr.rsrElement, dDp)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}

	if len(fltr.rsrValues) == 2 {
		endTime, err := parseTime(fltr.rsrValues[1], dDp)
		if err != nil {
			return false, err
		}
		if fltr.rsrValues[0] == nil {
			return timeVal.Before(endTime), nil
		}
		startTime, err := parseTime(fltr.rsrValues[0], dDp)
		if err != nil {
			return false, err
		}
		return startTime.Before(timeVal) && timeVal.Before(endTime), nil
	}
	startTime, err := parseTime(fltr.rsrValues[0], dDp)
	if err != nil {
		return false, err
	}
	return startTime.Before(timeVal), nil
}

func verifyInlineFilterS(fltrs []string) (err error) {
	for _, fl := range fltrs {
		if strings.HasPrefix(fl, utils.Meta) {
			if _, err = NewFilterFromInline(utils.EmptyString, fl); err != nil {
				return
			}
		}
	}
	return
}

func CheckFilter(fltr *Filter) (err error) {
	for _, rls := range fltr.Rules {
		valFunc := utils.IsPathValid
		if rls.Type == utils.MetaEmpty || rls.Type == utils.MetaExists {
			valFunc = utils.IsPathValidForExporters
		}
		if err = valFunc(rls.Element); err != nil {
			return fmt.Errorf("%s for filter <%v>", err, fltr) //encapsulated error
		}
		for _, val := range rls.Values {
			if err = valFunc(val); err != nil {
				return fmt.Errorf("%s for filter <%v>", err, fltr) //encapsulated error
			}
		}
	}
	return nil
}

func (fltr *FilterRule) passRegex(dDP utils.DataProvider) (bool, error) {
	strVal, err := fltr.rsrElement.ParseDataProvider(dDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	for _, val := range fltr.regexValues {
		if val.MatchString(strVal) {
			return true, nil
		}
	}
	return false, nil
}

func (fltr *FilterRule) passHttp(dDP utils.DataProvider) (bool, error) {
	strVal, err := fltr.rsrElement.ParseDataProvider(dDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return filterHTTP(fltr.Type, dDP, fltr.Element, strVal)

}
