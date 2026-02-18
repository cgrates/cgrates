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
	"fmt"
	"net"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/cron"
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
func (fS *FilterS) Pass(ctx *context.Context, tenant string, filterIDs []string,
	ev utils.DataProvider) (pass bool, err error) {
	if len(filterIDs) == 0 {
		return true, nil
	}
	dDP := NewDynamicDP(ctx, fS.connMgr, fS.cfg.FilterSCfg().ResourceSConns, fS.cfg.FilterSCfg().StatSConns,
		fS.cfg.FilterSCfg().AccountSConns, fS.cfg.FilterSCfg().TrendSConns, fS.cfg.FilterSCfg().RankingSConns, tenant, ev)
	for _, fltrID := range filterIDs {
		f, err := fS.dm.GetFilter(ctx, tenant, fltrID,
			true, true, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				err = utils.ErrPrefixNotFound(fltrID)
			}
			return false, err
		}
		for _, fltr := range f.Rules {
			if pass, err = fltr.Pass(ctx, dDP); err != nil || !pass {
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
func (fS *FilterS) LazyPass(ctx *context.Context, tenant string, filterIDs []string,
	ev utils.DataProvider, pathPrfxs []string) (pass bool, lazyCheckRules []*FilterRule, err error) {
	if len(filterIDs) == 0 {
		return true, nil, nil
	}
	pass = true
	dDP := NewDynamicDP(ctx, fS.connMgr, fS.cfg.FilterSCfg().ResourceSConns, fS.cfg.FilterSCfg().StatSConns,
		fS.cfg.FilterSCfg().AccountSConns, fS.cfg.FilterSCfg().TrendSConns, fS.cfg.FilterSCfg().RankingSConns, tenant, ev)
	for _, fltrID := range filterIDs {
		var f *Filter
		f, err = fS.dm.GetFilter(ctx, tenant, fltrID,
			true, true, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				err = utils.ErrPrefixNotFound(fltrID)
			}
			return
		}
		for _, rule := range f.Rules {
			if !verifyPrefixes(rule, pathPrfxs) {
				lazyCheckRules = append(lazyCheckRules, rule)
				continue
			}
			if pass, err = rule.Pass(ctx, dDP); err != nil || !pass {
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

type ArgsFiltersMatch struct {
	*utils.CGREvent
	FilterIDs []string
}

// Filter structure to define a basic filter
type Filter struct {
	Tenant string
	ID     string
	Rules  []*FilterRule
}

// Clone method for Filter
func (fltr *Filter) Clone() *Filter {
	if fltr == nil {
		return nil
	}
	clone := &Filter{
		Tenant: fltr.Tenant,
		ID:     fltr.ID,
	}
	if fltr.Rules != nil {
		clone.Rules = make([]*FilterRule, len(fltr.Rules))
		for i, rule := range fltr.Rules {
			clone.Rules[i] = rule.Clone()
		}
	}
	return clone
}

// CacheClone returns a clone of Filter used by ltcache CacheCloner
func (fltr *Filter) CacheClone() any {
	return fltr.Clone()
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

var (
	cdrQueryFilterTypes = utils.NewStringSet([]string{
		utils.MetaString, utils.MetaNotString,
		utils.MetaContains, utils.MetaNotContains,
		utils.MetaGreaterThan, utils.MetaGreaterOrEqual,
		utils.MetaLessThan, utils.MetaLessOrEqual,
		utils.MetaExists, utils.MetaNotExists,
		utils.MetaPrefix, utils.MetaNotPrefix,
		utils.MetaSuffix, utils.MetaNotSuffix,
		utils.MetaEmpty, utils.MetaNotEmpty,
		utils.MetaRegex, utils.MetaNotRegex,
		utils.MetaEqual, utils.MetaNotEqual,
		utils.MetaNever})

	supportedFiltersType utils.StringSet = utils.NewStringSet([]string{
		utils.MetaString, utils.MetaNotString,
		utils.MetaContains, utils.MetaNotContains,
		utils.MetaPrefix, utils.MetaNotPrefix,
		utils.MetaSuffix, utils.MetaNotSuffix,
		utils.MetaCronExp, utils.MetaNotCronExp,
		utils.MetaRSR, utils.MetaNotRSR,
		utils.MetaEmpty, utils.MetaNotEmpty,
		utils.MetaExists, utils.MetaNotExists,
		utils.MetaLessThan, utils.MetaLessOrEqual,
		utils.MetaGreaterThan, utils.MetaGreaterOrEqual,
		utils.MetaEqual, utils.MetaNotEqual,
		utils.MetaIPNet, utils.MetaNotIPNet,
		utils.MetaAPIBan, utils.MetaNotAPIBan,
		utils.MetaSentryPeer, utils.MetaNotSentryPeer,
		utils.MetaActivationInterval, utils.MetaNotActivationInterval,
		utils.MetaRegex, utils.MetaNotRegex,
		utils.MetaNever})

	needsFieldName utils.StringSet = utils.NewStringSet([]string{
		utils.MetaString, utils.MetaNotString,
		utils.MetaContains, utils.MetaNotContains,
		utils.MetaPrefix, utils.MetaNotPrefix,
		utils.MetaSuffix, utils.MetaNotSuffix,
		utils.MetaCronExp, utils.MetaNotCronExp,
		utils.MetaRSR, utils.MetaNotRSR,
		utils.MetaLessThan, utils.MetaLessThan,
		utils.MetaEmpty, utils.MetaNotEmpty,
		utils.MetaExists, utils.MetaNotExists,
		utils.MetaLessThan, utils.MetaLessOrEqual,
		utils.MetaGreaterThan, utils.MetaGreaterOrEqual,
		utils.MetaEqual, utils.MetaNotEqual,
		utils.MetaIPNet, utils.MetaNotIPNet,
		utils.MetaAPIBan, utils.MetaNotAPIBan,
		utils.MetaSentryPeer, utils.MetaNotSentryPeer,
		utils.MetaActivationInterval, utils.MetaNotActivationInterval,
		utils.MetaRegex, utils.MetaNotRegex})

	needsValues utils.StringSet = utils.NewStringSet([]string{
		utils.MetaString, utils.MetaNotString,
		utils.MetaContains, utils.MetaNotContains,
		utils.MetaPrefix, utils.MetaNotPrefix,
		utils.MetaSuffix, utils.MetaNotSuffix,
		utils.MetaCronExp, utils.MetaNotCronExp,
		utils.MetaRSR, utils.MetaNotRSR,
		utils.MetaEqual, utils.MetaNotEqual,
		utils.MetaIPNet, utils.MetaNotIPNet,
		utils.MetaAPIBan, utils.MetaNotAPIBan,
		utils.MetaSentryPeer, utils.MetaNotSentryPeer,
		utils.MetaActivationInterval, utils.MetaNotActivationInterval,
		utils.MetaRegex, utils.MetaNotRegex})
)

// NewFilterRule returns a new filter
func NewFilterRule(rfType, fieldName string, vals []string) (*FilterRule, error) {
	var negative bool
	rType := rfType
	if strings.HasPrefix(rfType, utils.MetaNot) {
		rType = utils.Meta + strings.TrimPrefix(rfType, utils.MetaNot)
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
// Pass rule: default negative, one matching rule should pass the filter
type FilterRule struct {
	Type        string           // Filter type (*string, *timing, *rsr_filters, *stats, *lt, *lte, *gt, *gte)
	Element     string           // Name of the field providing us the Values to check (used in case of some )
	Values      []string         // Filter definition
	rsrValues   utils.RSRParsers // Cache here the
	rsrElement  *utils.RSRParser // Cache here the
	rsrFilters  utils.RSRFilters // Cache here the RSRFilter Values
	regexValues []*regexp.Regexp
	negative    *bool
}

// Clone method for FilterRule
func (fltr *FilterRule) Clone() *FilterRule {
	if fltr == nil {
		return nil
	}
	clone := &FilterRule{
		Type:    fltr.Type,
		Element: fltr.Element,
	}
	if fltr.Values != nil {
		clone.Values = make([]string, len(fltr.Values))
		copy(clone.Values, fltr.Values)
	}
	if fltr.rsrValues != nil {
		clone.rsrValues = make(utils.RSRParsers, len(fltr.rsrValues))
		copy(clone.rsrValues, fltr.rsrValues)
	}
	if fltr.negative != nil {
		clone.negative = new(bool)
		*clone.negative = *fltr.negative
	}
	if fltr.rsrFilters != nil {
		fltr.rsrFilters = make(utils.RSRFilters, len(fltr.rsrFilters))
		for _, filter := range fltr.rsrFilters {
			clone.rsrFilters = append(clone.rsrFilters, filter.Clone())
		}
	}
	if fltr.regexValues != nil {
		clone.regexValues = make([]*regexp.Regexp, len(fltr.regexValues))
		for i, regex := range fltr.regexValues {
			clone.regexValues[i] = regex.Copy()
		}
	}
	return clone
}

// IsValid checks whether a filter rule is valid or not
func (fltr *FilterRule) IsValid() bool {
	// Type must be specified
	if fltr.Type == utils.EmptyString {
		return false
	}
	// Element must be specified only when the type is different from *never
	if fltr.Element == utils.EmptyString {
		return fltr.Type == utils.MetaNever
	}
	if len(fltr.Values) == 0 && !slices.Contains([]string{utils.MetaExists, utils.MetaNotExists,
		utils.MetaEmpty, utils.MetaNotEmpty}, fltr.Type) {
		return false
	}
	return true
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
	case utils.MetaExists, utils.MetaNotExists, utils.MetaEmpty, utils.MetaNotEmpty: // only the element is built
	case utils.MetaActivationInterval, utils.MetaNotActivationInterval:
		fltr.rsrValues = make(utils.RSRParsers, len(fltr.Values))
		for i, strVal := range fltr.Values {
			if fltr.rsrValues[i], err = utils.NewRSRParser(strVal); err != nil {
				return
			}
		}
	case utils.MetaNever: //return since there is not need for the values to be compiled in this case
		return
	default:
		if fltr.rsrValues, err = utils.NewRSRParsersFromSlice(fltr.Values); err != nil {
			return
		}
	}
	if fltr.rsrElement, err = utils.NewRSRParser(fltr.Element); err != nil {
		return
	} else if fltr.rsrElement == nil {
		return fmt.Errorf("empty RSRParser in rule: <%s>", fltr.Element)
	}
	return
}

// Pass is the method which should be used from outside.
func (fltr *FilterRule) Pass(ctx *context.Context, dDP utils.DataProvider) (result bool, err error) {
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
	case utils.MetaCronExp, utils.MetaNotCronExp:
		result, err = fltr.passCronExp(ctx, dDP)
	case utils.MetaRSR, utils.MetaNotRSR:
		result, err = fltr.passRSR(dDP)
	case utils.MetaLessThan, utils.MetaLessOrEqual, utils.MetaGreaterThan, utils.MetaGreaterOrEqual:
		result, err = fltr.passGreaterThan(dDP)
	case utils.MetaEqual, utils.MetaNotEqual:
		result, err = fltr.passEqualTo(dDP)
	case utils.MetaIPNet, utils.MetaNotIPNet:
		result, err = fltr.passIPNet(dDP)
	case utils.MetaAPIBan, utils.MetaNotAPIBan:
		result, err = fltr.passAPIBan(ctx, dDP)
	case utils.MetaSentryPeer, utils.MetaNotSentryPeer:
		result, err = fltr.passSentryPeer(ctx, dDP)
	case utils.MetaActivationInterval, utils.MetaNotActivationInterval:
		result, err = fltr.passActivationInterval(dDP)
	case utils.MetaRegex, utils.MetaNotRegex:
		result, err = fltr.passRegex(dDP)
	case utils.MetaNever:
		result, err = fltr.passNever(dDP)
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
	return result != *fltr.negative, nil
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
	case reflect.Slice:
		return rval.Len() == 0, nil
	case reflect.Map:
		return len(rval.MapKeys()) == 0, nil
	default:
		return rval.IsZero(), nil
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

func (fltr *FilterRule) passCronExp(ctx *context.Context, dDP utils.DataProvider) (bool, error) {
	tm, err := fltr.rsrElement.ParseDataProvider(dDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	tmTime, err := utils.IfaceAsTime(tm, config.CgrConfig().GeneralCfg().DefaultTimezone)
	if err != nil {
		return false, err
	}

	// tmTime = tmTime.Truncate(time.Second)
	tmTime = tmTime.Truncate(time.Minute)
	tmBefore := tmTime.Add(-time.Second)

	for _, valCronIDVal := range fltr.rsrValues {
		valTmID, err := valCronIDVal.ParseDataProvider(dDP)
		if err != nil {
			continue
		}
		exp, err := cron.ParseStandard(valTmID)
		if err != nil {
			continue
		}
		if exp.Next(tmBefore) == tmTime {
			return true, nil
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

func (fltr *FilterRule) passAPIBan(ctx *context.Context, dDP utils.DataProvider) (bool, error) {
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
	return GetAPIBan(ctx, strVal, config.CgrConfig().APIBanCfg().Keys, fltr.Values[0] != utils.MetaAll, true, true)
}

func (fltr *FilterRule) passSentryPeer(ctx *context.Context, dDP utils.DataProvider) (bool, error) {
	strVal, err := fltr.rsrElement.ParseDataProvider(dDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	if fltr.Values[0] != utils.MetaNumber && fltr.Values[0] != utils.MetaIP {
		return false, fmt.Errorf("invalid value for sentrypeer filter: <%s>", fltr.Values[0])
	}
	return GetSentryPeer(ctx, strVal, config.CgrConfig().SentryPeerCfg(), fltr.Values[0])
}

func parseTime(rsr *utils.RSRParser, dDp utils.DataProvider) (_ time.Time, err error) {
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
			if rls.Type == utils.MetaEmpty || rls.Type == utils.MetaNotEmpty ||
				rls.Type == utils.MetaExists || rls.Type == utils.MetaNotExists &&
				val != utils.EmptyString {
				return fmt.Errorf("value of filter <%s> is not empty <%s>",
					fltr.ID, val)
			}
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

func (fltr *FilterRule) passNever(dDP utils.DataProvider) (bool, error) {
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

func (fltr *Filter) Set(path []string, val any, newBranch bool) (err error) {
	switch len(path) {
	default:
		return utils.ErrWrongPath
	case 1:
		switch path[0] {
		default:
			return utils.ErrWrongPath
		case utils.Tenant:
			fltr.Tenant = utils.IfaceAsString(val)
		case utils.ID:
			fltr.ID = utils.IfaceAsString(val)
		}
	case 2:
		if path[0] != utils.Rules {
			return utils.ErrWrongPath
		}
		if len(fltr.Rules) == 0 || newBranch {
			fltr.Rules = append(fltr.Rules, new(FilterRule))
		}
		switch path[1] {
		case utils.Type:
			fltr.Rules[len(fltr.Rules)-1].Type = utils.IfaceAsString(val)
		case utils.Element:
			fltr.Rules[len(fltr.Rules)-1].Element = utils.IfaceAsString(val)
		case utils.Values:
			fltr.Rules[len(fltr.Rules)-1].Values, err = utils.IfaceAsStringSlice(val)
		default:
			return utils.ErrWrongPath
		}
	}
	return
}

func (fltr *Filter) Compress() {
	newRules := make([]*FilterRule, 0, len(fltr.Rules))
	for i, flt := range fltr.Rules {
		if i == 0 ||
			newRules[len(newRules)-1].Type != flt.Type ||
			newRules[len(newRules)-1].Element != flt.Element {
			newRules = append(newRules, flt)
			continue
		}
		newRules[len(newRules)-1].Values = append(newRules[len(newRules)-1].Values, flt.Values...)
	}
	fltr.Rules = newRules
}

func (fltr *Filter) Merge(v2 any) {
	vi := v2.(*Filter)
	if len(vi.Tenant) != 0 {
		fltr.Tenant = vi.Tenant
	}
	if len(vi.ID) != 0 {
		fltr.ID = vi.ID
	}
	for _, rule := range vi.Rules {
		if rule.Type != utils.EmptyString {
			fltr.Rules = append(fltr.Rules, rule)
		}
	}
}

func (fltr *Filter) String() string { return utils.ToJSON(fltr) }
func (fltr *Filter) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = fltr.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(val), nil
}
func (fltr *Filter) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) == 1 {
		switch fldPath[0] {
		default:
			fld, idx := utils.GetPathIndex(fldPath[0])
			if fld == utils.Rules &&
				idx != nil &&
				*idx < len(fltr.Rules) {
				return fltr.Rules[*idx], nil
			}
			return nil, utils.ErrNotFound
		case utils.Tenant:
			return fltr.Tenant, nil
		case utils.ID:
			return fltr.ID, nil
		}
	}
	if len(fldPath) == 0 ||
		!strings.HasPrefix(fldPath[0], utils.Rules) ||
		fldPath[0][5] != '[' ||
		fldPath[0][len(fldPath[0])-1] != ']' {
		return nil, utils.ErrNotFound
	}
	var idx int
	if idx, err = strconv.Atoi(fldPath[0][6 : len(fldPath[0])-1]); err != nil {
		return
	}
	if idx >= len(fltr.Rules) {
		return nil, utils.ErrNotFound
	}
	return fltr.Rules[idx].FieldAsInterface(fldPath[1:])
}

// AsMapStringInterface converts Filter struct to map[string]any
func (fltr *Filter) AsMapStringInterface() map[string]any {
	if fltr == nil {
		return nil
	}
	return map[string]any{
		utils.Tenant: fltr.Tenant,
		utils.ID:     fltr.ID,
		utils.Rules:  fltr.Rules,
	}
}

// MapStringInterfaceToFilter converts map[string]any to Filter struct
func MapStringInterfaceToFilter(m map[string]any) (*Filter, error) {
	fltr := &Filter{}

	if v, ok := m[utils.Tenant].(string); ok {
		fltr.Tenant = v
	}
	if v, ok := m[utils.ID].(string); ok {
		fltr.ID = v
	}
	fltr.Rules = InterfaceToRules(m[utils.Rules])
	return fltr, nil
}

// InterfaceToRules converts interface to []*FilterRule
func InterfaceToRules(v any) []*FilterRule {
	if v == nil {
		return nil
	}
	if val, ok := v.([]any); !ok {
		return nil
	} else {
		result := make([]*FilterRule, 0, len(val))
		for _, item := range val {
			if filters, ok := item.(map[string]any); ok {
				result = append(result, MapStringInterfaceToFilterRule(filters))
			}
		}
		return result
	}
}

// MapStringInterfaceToFilterRule converts map[string]any to *FilterRule
func MapStringInterfaceToFilterRule(m map[string]any) *FilterRule {
	if m == nil {
		return nil
	}
	rule := &FilterRule{}

	if v, ok := m[utils.Type].(string); ok {
		rule.Type = v
	}

	if v, ok := m[utils.Element].(string); ok {
		rule.Element = v
	}

	rule.Values = utils.InterfaceToStringSlice(m[utils.Values])
	return rule
}

func (fltr *FilterRule) String() string { return utils.ToJSON(fltr) }
func (fltr *FilterRule) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = fltr.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(val), nil
}
func (fltr *FilterRule) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	switch fldPath[0] {
	default:
		fld, idx := utils.GetPathIndex(fldPath[0])
		if fld == utils.Values &&
			idx != nil &&
			*idx < len(fltr.Values) {
			return fltr.Values[*idx], nil
		}
		return nil, utils.ErrNotFound
	case utils.Type:
		return fltr.Type, nil
	case utils.Element:
		return fltr.Element, nil
	case utils.Values:
		return fltr.Values, nil
	}
}

// GetFilters retrieves and compiles the filters identified by filterIDs for the specified tenant.
func GetFilters(ctx *context.Context, filterIDs []string, tenant string,
	dm *DataManager) ([]*Filter, error) {

	fltrs := make([]*Filter, 0, len(filterIDs))
	for _, fltrID := range filterIDs {
		fltr, err := dm.GetFilter(ctx, tenant, fltrID, true, true, utils.NonTransactional)
		if err != nil {
			return nil, fmt.Errorf("retrieving filter %s failed: %w", fltrID, err)
		}
		if err = fltr.Compile(); err != nil {
			return nil, fmt.Errorf("compiling filter %s failed: %w", fltrID, err)
		}
		fltrs = append(fltrs, fltr)
	}
	return fltrs, nil
}

// Will return all items the element, e.g.  "~*req", "cost_details", "Charges[0]". "RatingID"
func (fltr *FilterRule) ElementItems() []string {
	return strings.Split(fltr.Element, utils.NestingSep)
}

// Creates mysql conditions used in WHERE statement out of filters
func (fltr *FilterRule) FilterToSQLQuery() (conditions []string) {
	var firstItem string   // Excluding ~*req, hold the first item of an element, left empty if no more than 1 item in element. e.g. "cost_details" out of ~*req.cost_details.Charges[0].RatingID or "" out of ~*req.answer_time
	var restOfItems string // Excluding ~*req, hold the rest of the items past the first one. If only 1 item in all element, holds that item. e.g. "Charges[0].RatingID" out of ~*req.cost_details.Charges[0].RatingID or "answer_time" out of ~*req.answer_time
	not := strings.HasPrefix(fltr.Type, utils.MetaNot)
	elementItems := fltr.ElementItems()[1:] // exclude first item: ~*req
	for i := range elementItems {           // encapsulate with "" strings starting with *
		if strings.HasPrefix(elementItems[i], utils.Meta) {
			elementItems[i] = "\"" + elementItems[i] + "\""
		}
	}
	if len(elementItems) > 1 {
		firstItem = elementItems[0]
		restOfItems = strings.Join(elementItems[1:], utils.NestingSep)
	} else {
		restOfItems = elementItems[0]
	}

	// here are for the filters that their values are empty: *exists, *notexists, *empty, *notempty..
	if len(fltr.Values) == 0 {
		switch fltr.Type {
		case utils.MetaExists, utils.MetaNotExists:
			if not { // not existing means Column IS NULL
				if firstItem == utils.EmptyString {
					conditions = append(conditions, fmt.Sprintf("%s IS NULL", restOfItems))
					return
				}
				queryPart := fmt.Sprintf("JSON_VALUE(%s, '$.%s') IS NULL", firstItem, restOfItems)
				if strings.HasPrefix(restOfItems, `"*`) {
					queryPart = fmt.Sprintf("JSON_UNQUOTE(%s)", queryPart)
				}
				conditions = append(conditions, queryPart)
				return
			}
			// existing means Column IS NOT NULL
			if firstItem == utils.EmptyString {
				conditions = append(conditions, fmt.Sprintf("%s IS NOT NULL", restOfItems))
				return
			}
			queryPart := fmt.Sprintf("JSON_VALUE(%s, '$.%s') IS NOT NULL", firstItem, restOfItems)
			if strings.HasPrefix(restOfItems, `"*`) {
				queryPart = fmt.Sprintf("JSON_UNQUOTE(%s)", queryPart)
			}
			conditions = append(conditions, queryPart)
		case utils.MetaEmpty, utils.MetaNotEmpty:
			if not {
				if firstItem == utils.EmptyString {
					conditions = append(conditions, fmt.Sprintf("%s != ''", restOfItems))
					return
				}
				queryPart := fmt.Sprintf("JSON_VALUE(%s, '$.%s') != ''", firstItem, restOfItems)
				if strings.HasPrefix(restOfItems, `"*`) {
					queryPart = fmt.Sprintf("JSON_UNQUOTE(%s)", queryPart)
				}
				conditions = append(conditions, queryPart)
				return
			}
			if firstItem == utils.EmptyString {
				conditions = append(conditions, fmt.Sprintf("%s == ''", restOfItems))
				return
			}
			queryPart := fmt.Sprintf("JSON_VALUE(%s, '$.%s') == ''", firstItem, restOfItems)
			if strings.HasPrefix(restOfItems, `"*`) {
				queryPart = fmt.Sprintf("JSON_UNQUOTE(%s)", queryPart)
			}
			conditions = append(conditions, queryPart)
		}
		return
	}
	// here are for the filters that can have more than one value: *string, *prefix, *suffix ..
	for _, value := range fltr.Values {
		switch value { // in case we have boolean values, it should be queried over 1 or 0
		case "true":
			value = "1"
		case "false":
			value = "0"
		}
		var singleCond string
		switch fltr.Type {
		case utils.MetaString, utils.MetaNotString, utils.MetaEqual, utils.MetaNotEqual:
			if not {
				if firstItem == utils.EmptyString {
					conditions = append(conditions, fmt.Sprintf("%s != '%s'", restOfItems, value))
					continue
				}
				queryPart := fmt.Sprintf("JSON_VALUE(%s, '$.%s') != '%s'",
					firstItem, restOfItems, value)
				if strings.HasPrefix(restOfItems, `"*`) {
					queryPart = fmt.Sprintf("JSON_UNQUOTE(%s)", queryPart)
				}
				conditions = append(conditions, queryPart)
				continue
			}
			if firstItem == utils.EmptyString {
				singleCond = fmt.Sprintf("%s = '%s'", restOfItems, value)
			} else {
				queryPart := fmt.Sprintf("JSON_VALUE(%s, '$.%s') = '%s'", firstItem, restOfItems, value)
				if strings.HasPrefix(restOfItems, `"*`) {
					queryPart = fmt.Sprintf("JSON_UNQUOTE(%s)", queryPart)
				}
				singleCond = queryPart
			}
		case utils.MetaLessThan, utils.MetaLessOrEqual, utils.MetaGreaterThan, utils.MetaGreaterOrEqual:
			parsedValAny := utils.StringToInterface(value)
			switch fltr.Type {
			case utils.MetaGreaterOrEqual:
				if firstItem == utils.EmptyString {
					singleCond = fmt.Sprintf("%s >= '%v'", restOfItems, parsedValAny)
				} else {
					queryPart := fmt.Sprintf("JSON_VALUE(%s, '$.%s') >= '%v'", firstItem, restOfItems, parsedValAny)
					if strings.HasPrefix(restOfItems, `"*`) {
						queryPart = fmt.Sprintf("JSON_UNQUOTE(%s)", queryPart)
					}
					singleCond = queryPart
				}
			case utils.MetaGreaterThan:
				if firstItem == utils.EmptyString {
					singleCond = fmt.Sprintf("%s > '%v'", restOfItems, parsedValAny)
				} else {
					queryPart := fmt.Sprintf("JSON_VALUE(%s, '$.%s') > '%v'", firstItem, restOfItems, parsedValAny)
					if strings.HasPrefix(restOfItems, `"*`) {
						queryPart = fmt.Sprintf("JSON_UNQUOTE(%s)", queryPart)
					}
					singleCond = queryPart
				}
			case utils.MetaLessOrEqual:
				if firstItem == utils.EmptyString {
					singleCond = fmt.Sprintf("%s <= '%v'", restOfItems, parsedValAny)
				} else {
					queryPart := fmt.Sprintf("JSON_VALUE(%s, '$.%s') <= '%v'", firstItem, restOfItems, parsedValAny)
					if strings.HasPrefix(restOfItems, `"*`) {
						queryPart = fmt.Sprintf("JSON_UNQUOTE(%s)", queryPart)
					}
					singleCond = queryPart
				}
			case utils.MetaLessThan:
				if firstItem == utils.EmptyString {
					singleCond = fmt.Sprintf("%s < '%v'", restOfItems, parsedValAny)
				} else {
					queryPart := fmt.Sprintf("JSON_VALUE(%s, '$.%s') < '%v'", firstItem, restOfItems, parsedValAny)
					if strings.HasPrefix(restOfItems, `"*`) {
						queryPart = fmt.Sprintf("JSON_UNQUOTE(%s)", queryPart)
					}
					singleCond = queryPart
				}
			}
		case utils.MetaPrefix, utils.MetaNotPrefix:
			if not {
				if firstItem == utils.EmptyString {
					conditions = append(conditions, fmt.Sprintf("%s NOT LIKE '%s%%'", restOfItems, value))
					continue
				}
				queryPart := fmt.Sprintf("JSON_VALUE(%s, '$.%s') NOT LIKE '%s%%'", firstItem, restOfItems, value)
				if strings.HasPrefix(restOfItems, `"*`) {
					queryPart = fmt.Sprintf("JSON_UNQUOTE(%s)", queryPart)
				}
				conditions = append(conditions, queryPart)
				continue
			}
			if firstItem == utils.EmptyString {
				singleCond = fmt.Sprintf("%s LIKE '%s%%'", restOfItems, value)
			} else {
				queryPart := fmt.Sprintf("JSON_VALUE(%s, '$.%s') LIKE '%s%%'", firstItem, restOfItems, value)
				if strings.HasPrefix(restOfItems, `"*`) {
					queryPart = fmt.Sprintf("JSON_UNQUOTE(%s)", queryPart)
				}
				singleCond = queryPart
			}
		case utils.MetaSuffix, utils.MetaNotSuffix:
			if not {
				if firstItem == utils.EmptyString {
					conditions = append(conditions, fmt.Sprintf("%s NOT LIKE '%%%s'", restOfItems, value))
					continue
				}
				queryPart := fmt.Sprintf("JSON_VALUE(%s, '$.%s') NOT LIKE '%%%s'", firstItem, restOfItems, value)
				if strings.HasPrefix(restOfItems, `"*`) {
					queryPart = fmt.Sprintf("JSON_UNQUOTE(%s)", queryPart)
				}
				conditions = append(conditions, queryPart)
				continue
			}
			if firstItem == utils.EmptyString {
				singleCond = fmt.Sprintf("%s LIKE '%%%s'", restOfItems, value)
			} else {
				queryPart := fmt.Sprintf("JSON_VALUE(%s, '$.%s') LIKE '%%%s'", firstItem, restOfItems, value)
				if strings.HasPrefix(restOfItems, `"*`) {
					queryPart = fmt.Sprintf("JSON_UNQUOTE(%s)", queryPart)
				}
				singleCond = queryPart
			}
		case utils.MetaRegex, utils.MetaNotRegex:
			if not {
				if firstItem == utils.EmptyString {
					conditions = append(conditions, fmt.Sprintf("%s NOT REGEXP '%s'", restOfItems, value))
					continue
				}
				queryPart := fmt.Sprintf("JSON_VALUE(%s, '$.%s') NOT REGEXP '%s'", firstItem, restOfItems, value)
				if strings.HasPrefix(restOfItems, `"*`) {
					queryPart = fmt.Sprintf("JSON_UNQUOTE(%s)", queryPart)
				}
				conditions = append(conditions, queryPart)
				continue
			}
			if firstItem == utils.EmptyString {
				singleCond = fmt.Sprintf("%s REGEXP '%s'", restOfItems, value)
			} else {
				queryPart := fmt.Sprintf("JSON_VALUE(%s, '$.%s') REGEXP '%s'", firstItem, restOfItems, value)
				if strings.HasPrefix(restOfItems, `"*`) {
					queryPart = fmt.Sprintf("JSON_UNQUOTE(%s)", queryPart)
				}
				singleCond = queryPart
			}
		}
		conditions = append(conditions, singleCond)
	}
	return
}
