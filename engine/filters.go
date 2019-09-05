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
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func NewFilterS(cfg *config.CGRConfig,
	statSChan, resSChan, ralSChan chan rpcclient.RpcClientConnection, dm *DataManager) (fS *FilterS) {
	fS = &FilterS{
		dm:  dm,
		cfg: cfg,
	}
	if len(cfg.FilterSCfg().StatSConns) != 0 {
		fS.connStatS(statSChan)
	}
	if len(cfg.FilterSCfg().ResourceSConns) != 0 {
		fS.connResourceS(resSChan)
	}
	if len(cfg.FilterSCfg().RALsConns) != 0 {
		fS.connRALs(ralSChan)
	}
	return
}

// FilterS is a service used to take decisions in case of filters
// uses lazy connections where necessary to avoid deadlocks on service startup
type FilterS struct {
	cfg                               *config.CGRConfig
	statSConns, resSConns, ralSConns  rpcclient.RpcClientConnection
	sSConnMux, rSConnMux, ralSConnMux sync.RWMutex // make sure only one goroutine attempts connecting
	dm                                *DataManager
}

// connStatS returns will connect towards StatS
func (fS *FilterS) connStatS(statSChan chan rpcclient.RpcClientConnection) (err error) {
	fS.sSConnMux.Lock()
	defer fS.sSConnMux.Unlock()
	if fS.statSConns != nil { // connection was populated between locks
		return
	}
	fS.statSConns, err = NewRPCPool(rpcclient.POOL_FIRST,
		fS.cfg.TlsCfg().ClientKey, fS.cfg.TlsCfg().ClientCerificate,
		fS.cfg.TlsCfg().CaCertificate, fS.cfg.GeneralCfg().ConnectAttempts,
		fS.cfg.GeneralCfg().Reconnects, fS.cfg.GeneralCfg().ConnectTimeout,
		fS.cfg.GeneralCfg().ReplyTimeout, fS.cfg.FilterSCfg().StatSConns,
		statSChan, true)
	return
}

// connResourceS returns will connect towards ResourceS
func (fS *FilterS) connResourceS(resSChan chan rpcclient.RpcClientConnection) (err error) {
	fS.rSConnMux.Lock()
	defer fS.rSConnMux.Unlock()
	if fS.resSConns != nil { // connection was populated between locks
		return
	}
	fS.resSConns, err = NewRPCPool(rpcclient.POOL_FIRST,
		fS.cfg.TlsCfg().ClientKey, fS.cfg.TlsCfg().ClientCerificate,
		fS.cfg.TlsCfg().CaCertificate, fS.cfg.GeneralCfg().ConnectAttempts,
		fS.cfg.GeneralCfg().Reconnects, fS.cfg.GeneralCfg().ConnectTimeout,
		fS.cfg.GeneralCfg().ReplyTimeout, fS.cfg.FilterSCfg().ResourceSConns,
		resSChan, true)
	return
}

// connRALs returns will connect towards RALs
func (fS *FilterS) connRALs(ralSChan chan rpcclient.RpcClientConnection) (err error) {
	fS.ralSConnMux.Lock()
	defer fS.ralSConnMux.Unlock()
	if fS.ralSConns != nil { // connection was populated between locks
		return
	}
	fS.ralSConns, err = NewRPCPool(rpcclient.POOL_FIRST,
		fS.cfg.TlsCfg().ClientKey, fS.cfg.TlsCfg().ClientCerificate,
		fS.cfg.TlsCfg().CaCertificate, fS.cfg.GeneralCfg().ConnectAttempts,
		fS.cfg.GeneralCfg().Reconnects, fS.cfg.GeneralCfg().ConnectTimeout,
		fS.cfg.GeneralCfg().ReplyTimeout, fS.cfg.FilterSCfg().RALsConns,
		ralSChan, true)
	return
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
			fieldNameDP, err = fS.getFieldNameDataProvider(ev, fltr.FieldName, tenant)
			if err != nil {
				return pass, err
			}
			fieldValuesDP, err = fS.getFieldValueDataProviders(ev, fltr.Values, tenant)
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
			Type:      ruleSplt[0],
			FieldName: ruleSplt[1],
			Values:    strings.Split(strings.Join(ruleSplt[2:], utils.InInFieldSep), utils.INFIELD_SEP),
		}},
	}
	if err = f.Compile(); err != nil {
		return nil, err
	}
	return
}

type Filter struct {
	Tenant             string
	ID                 string
	Rules              []*FilterRule
	ActivationInterval *utils.ActivationInterval
}

func (flt *Filter) TenantID() string {
	return utils.ConcatenatedKey(flt.Tenant, flt.ID)
}

// Compile will compile the underlaying request filters where necessary (ie. regexp rules)
func (f *Filter) Compile() (err error) {
	for _, rf := range f.Rules {
		if err = rf.CompileValues(); err != nil {
			return
		}
	}
	return
}

var supportedFiltersType *utils.StringSet = utils.NewStringSet([]string{utils.MetaString, utils.MetaPrefix, utils.MetaSuffix,
	utils.MetaTimings, utils.MetaRSR, utils.MetaStatS, utils.MetaDestinations,
	utils.MetaEmpty, utils.MetaExists, utils.MetaLessThan, utils.MetaLessOrEqual,
	utils.MetaGreaterThan, utils.MetaGreaterOrEqual, utils.MetaResources, utils.MetaEqual,
	utils.MetaAccount, utils.MetaNotEqual})
var needsFieldName *utils.StringSet = utils.NewStringSet([]string{utils.MetaString, utils.MetaPrefix,
	utils.MetaSuffix, utils.MetaTimings, utils.MetaDestinations, utils.MetaLessThan,
	utils.MetaEmpty, utils.MetaExists, utils.MetaLessOrEqual, utils.MetaGreaterThan,
	utils.MetaGreaterOrEqual, utils.MetaEqual, utils.MetaNotEqual})
var needsValues *utils.StringSet = utils.NewStringSet([]string{utils.MetaString, utils.MetaPrefix,
	utils.MetaSuffix, utils.MetaTimings, utils.MetaRSR, utils.MetaDestinations,
	utils.MetaLessThan, utils.MetaLessOrEqual, utils.MetaGreaterThan, utils.MetaGreaterOrEqual,
	utils.MetaEqual, utils.MetaNotEqual})

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
		return nil, fmt.Errorf("FieldName is mandatory for Type: %s", rfType)
	}
	if len(vals) == 0 && needsValues.Has(rType) {
		return nil, fmt.Errorf("Values is mandatory for Type: %s", rfType)
	}
	rf := &FilterRule{
		Type:      rfType,
		FieldName: fieldName,
		Values:    vals,
		negative:  utils.BoolPointer(negative),
	}
	if err := rf.CompileValues(); err != nil {
		return nil, err
	}
	return rf, nil
}

//itemFilter is used for *stats and *resources filter type
type itemFilter struct {
	ItemID      string
	FilterType  string
	FilterValue string
}

// FilterRule filters requests coming into various places
// Pass rule: default negative, one mathing rule should pass the filter
type FilterRule struct {
	Type          string            // Filter type (*string, *timing, *rsr_filters, *stats, *lt, *lte, *gt, *gte)
	FieldName     string            // Name of the field providing us the Values to check (used in case of some )
	Values        []string          // Filter definition
	rsrFields     config.RSRParsers // Cache here the RSRFilter Values
	negative      *bool
	statItems     []*itemFilter // Cached compiled itemFilter out of Values
	resourceItems []*itemFilter // Cached compiled itemFilter out of Values
	accountItems  []*itemFilter // Cached compiled itemFilter out of Values
}

// Separate method to compile RSR fields
func (rf *FilterRule) CompileValues() (err error) {
	switch rf.Type {
	case utils.MetaRSR, utils.MetaNotRSR:
		if rf.rsrFields, err = config.NewRSRParsersFromSlice(rf.Values, true); err != nil {
			return
		}
	case utils.MetaStatS, utils.MetaNotStatS:
		//value for filter of type *stats needs to be in the following form:
		//*gt#acd:StatID:ValueOfMetric
		rf.statItems = make([]*itemFilter, len(rf.Values))
		for i, val := range rf.Values {
			valSplt := strings.Split(val, utils.InInFieldSep)
			if len(valSplt) != 3 {
				return fmt.Errorf("Value %s needs to contain at least 3 items", val)
			}
			// valSplt[0] filter type with metric
			// valSplt[1] id of the statQueue
			// valSplt[2] value to compare
			rf.statItems[i] = &itemFilter{
				FilterType:  valSplt[0],
				ItemID:      valSplt[1],
				FilterValue: valSplt[2],
			}
		}
	case utils.MetaResources, utils.MetaNotResources:
		//value for filter of type *resources needs to be in the following form:
		//*gt:ResourceID:ValueOfUsage
		rf.resourceItems = make([]*itemFilter, len(rf.Values))
		for i, val := range rf.Values {
			valSplt := strings.Split(val, utils.InInFieldSep)
			if len(valSplt) != 3 {
				return fmt.Errorf("Value %s needs to contain at least 3 items", val)
			}
			// valSplt[0] filter type
			// valSplt[1] id of the AccountID.FieldToUsed
			// valSplt[2] value to compare
			rf.resourceItems[i] = &itemFilter{
				FilterType:  valSplt[0],
				ItemID:      valSplt[1],
				FilterValue: valSplt[2],
			}
		}
	case utils.MetaAccount:
		//value for filter of type *accounts needs to be in the following form:
		//*gt:AccountID:ValueOfUsage
		rf.accountItems = make([]*itemFilter, len(rf.Values))
		for i, val := range rf.Values {
			valSplt := strings.Split(val, utils.InInFieldSep)
			if len(valSplt) != 3 {
				return fmt.Errorf("Value %s needs to contain at least 3 items", val)
			}
			// valSplt[0] filter type
			// valSplt[1] id of the Resource
			// valSplt[2] value to compare
			rf.accountItems[i] = &itemFilter{
				FilterType:  valSplt[0],
				ItemID:      valSplt[1],
				FilterValue: valSplt[2],
			}
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
	strVal, err := config.GetDynamicString(fltr.FieldName, fielNameDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	for i, val := range fltr.Values {
		sval, err := config.GetDynamicString(val, fieldValuesDP[i])
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
	_, err := config.GetDynamicInterface(fltr.FieldName, fielNameDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (fltr *FilterRule) passEmpty(fielNameDP config.DataProvider) (bool, error) {
	val, err := config.GetDynamicInterface(fltr.FieldName, fielNameDP)
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
	strVal, err := config.GetDynamicString(fltr.FieldName, fielNameDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	for i, prfx := range fltr.Values {
		prfx, err := config.GetDynamicString(prfx, fieldValuesDP[i])
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
	strVal, err := config.GetDynamicString(fltr.FieldName, fielNameDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	for i, prfx := range fltr.Values {
		prfx, err := config.GetDynamicString(prfx, fieldValuesDP[i])
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
	dst, err := config.GetDynamicString(fltr.FieldName, fielNameDP)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	for _, p := range utils.SplitPrefix(dst, MIN_PREFIX_MATCH) {
		if destIDs, err := dm.DataDB().GetReverseDestination(p, false, utils.NonTransactional); err == nil {
			for _, dID := range destIDs {
				for i, valDstID := range fltr.Values {
					valDstID, err := config.GetDynamicString(valDstID, fieldValuesDP[i])
					if err != nil {
						continue
					}
					if valDstID == dID {
						return true, nil
					}
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

// func (fltr *FilterRule) passStatS(dP config.DataProvider,
// 	stats rpcclient.RpcClientConnection, tenant string) (bool, error) {
// 	if stats == nil || reflect.ValueOf(stats).IsNil() {
// 		return false, errors.New("Missing StatS information")
// 	}
// 	for _, statItem := range fltr.statItems {
// 		statValues := make(map[string]float64)
// 		if err := stats.Call(utils.StatSv1GetQueueFloatMetrics,
// 			&utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{Tenant: tenant, ID: statItem.ItemID}}, &statValues); err != nil {
// 			return false, err
// 		}
// 		//convert statValues to map[string]interface{}
// 		ifaceStatValues := make(map[string]interface{})
// 		for key, val := range statValues {
// 			ifaceStatValues[key] = val
// 		}
// 		//convert ifaceStatValues into a NavigableMap so we can send it to passGreaterThan
// 		nM := config.NewNavigableMap(ifaceStatValues)
// 		//split the type in exact 2 parts
// 		//special cases like *gt#sum#Usage
// 		fltrType := strings.SplitN(statItem.FilterType, utils.STATS_CHAR, 2)
// 		if len(fltrType) < 2 {
// 			return false, errors.New(fmt.Sprintf("<%s> Invalid format for filter of type *stats", utils.FilterS))
// 		}
// 		//compose the newFilter
// 		fltr, err := NewFilterRule(fltrType[0],
// 			utils.DynamicDataPrefix+utils.Meta+fltrType[1], []string{statItem.FilterValue})
// 		if err != nil {
// 			return false, err
// 		}
// 		//send it to passGreaterThan
// 		if val, err := fltr.passGreaterThan(nM); err != nil || !val {
// 			//in case of error return false and error
// 			//and in case of not pass return false and nil
// 			return false, err
// 		}
// 	}
// 	return true, nil
// }

func (fltr *FilterRule) passGreaterThan(fielNameDP config.DataProvider, fieldValuesDP []config.DataProvider) (bool, error) {
	fldIf, err := config.GetDynamicInterface(fltr.FieldName, fielNameDP)
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
		sval, err := config.GetDynamicInterface(val, fieldValuesDP[i])
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

// func (fltr *FilterRule) passResourceS(dP config.DataProvider,
// 	resourceS rpcclient.RpcClientConnection, tenant string) (bool, error) {
// 	if resourceS == nil || reflect.ValueOf(resourceS).IsNil() {
// 		return false, errors.New("Missing ResourceS information")
// 	}
// 	for _, resItem := range fltr.resourceItems {
// 		//take total usage for resource
// 		var reply Resource
// 		if err := resourceS.Call(utils.ResourceSv1GetResource,
// 			&utils.TenantID{Tenant: tenant, ID: resItem.ItemID}, &reply); err != nil {
// 			return false, err
// 		}
// 		data := map[string]interface{}{
// 			utils.Usage: reply.totalUsage(),
// 		}
// 		//convert data into a NavigableMap so we can send it to passGreaterThan
// 		nM := config.NewNavigableMap(data)
// 		//compose the newFilter
// 		fltr, err := NewFilterRule(resItem.FilterType,
// 			utils.DynamicDataPrefix+utils.Usage, []string{resItem.FilterValue})
// 		if err != nil {
// 			return false, err
// 		}
// 		// send it to passGreaterThan
// 		if val, err := fltr.passGreaterThan(nM); err != nil || !val {
// 			//in case of error return false and error
// 			//and in case of not pass return false and nil
// 			return false, err
// 		}
// 	}
// 	return true, nil
// }

// func (fltr *FilterRule) passAccountS(dP config.DataProvider,
// 	accountS rpcclient.RpcClientConnection, tenant string) (bool, error) {
// 	if accountS == nil || reflect.ValueOf(accountS).IsNil() {
// 		return false, errors.New("Missing AccountS information")
// 	}
// 	for _, accItem := range fltr.accountItems {
// 		//split accItem.ItemID in two accountID and actual filter
// 		//AccountID.BalanceMap.*monetary[0].Value
// 		splittedString := strings.SplitN(accItem.ItemID, utils.NestingSep, 2)
// 		accID := splittedString[0]
// 		filterID := splittedString[1]
// 		var reply Account
// 		if err := accountS.Call(utils.ApierV2GetAccount,
// 			&utils.AttrGetAccount{Tenant: tenant, Account: accID}, &reply); err != nil {
// 			return false, err
// 		}
// 		//compose the newFilter
// 		fltr, err := NewFilterRule(accItem.FilterType,
// 			utils.DynamicDataPrefix+filterID, []string{accItem.FilterValue})
// 		if err != nil {
// 			return false, err
// 		}
// 		dP, _ := reply.AsNavigableMap(nil)
// 		if val, err := fltr.Pass(dP, nil, tenant); err != nil || !val {
// 			//in case of error return false and error
// 			//and in case of not pass return false and nil
// 			return false, err
// 		}
// 	}
// 	return true, nil
// }

func (fltr *FilterRule) passEqualTo(fielNameDP config.DataProvider, fieldValuesDP []config.DataProvider) (bool, error) {
	fldIf, err := config.GetDynamicInterface(fltr.FieldName, fielNameDP)
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
		sval, err := config.GetDynamicInterface(val, fieldValuesDP[i])
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

func (fS *FilterS) getFieldNameDataProvider(initialDP config.DataProvider, fieldName string, tenant string) (dp config.DataProvider, err error) {
	switch {
	case strings.HasPrefix(fieldName, utils.MetaAccounts):
		//construct dataProvider from account and set it furthder
		var account *Account
		//extract the AccountID from fieldName
		if err = fS.ralSConns.Call(utils.ApierV2GetAccount,
			&utils.AttrGetAccount{Tenant: tenant, Account: "completeHereWithID"}, &account); err != nil {
			return
		}
		dp = config.NewObjectDP(account)
	case strings.HasPrefix(fieldName, utils.MetaResources):
	case strings.HasPrefix(fieldName, utils.MetaStats):
	default:
		dp = initialDP
	}
	return
}

func (fS *FilterS) getFieldValueDataProviders(initialDP config.DataProvider, values []string, tenant string) (dp []config.DataProvider, err error) {
	dp = make([]config.DataProvider, len(values))
	for i, val := range values {
		switch {
		case strings.HasPrefix(val, utils.MetaAccounts):
			var account *Account
			//extract the AccountID from fieldName
			if err = fS.ralSConns.Call(utils.ApierV2GetAccount,
				&utils.AttrGetAccount{Tenant: tenant, Account: "completeHereWithID"}, &account); err != nil {
				return
			}
			dp[i] = config.NewObjectDP(account)
		case strings.HasPrefix(val, utils.MetaResources):
		case strings.HasPrefix(val, utils.MetaStats):
		default:
			dp[i] = initialDP
		}
	}
	return
}
