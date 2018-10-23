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
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

const (
	MetaString         = "*string"
	MetaPrefix         = "*prefix"
	MetaTimings        = "*timings"
	MetaRSR            = "*rsr"
	MetaStatS          = "*stats"
	MetaDestinations   = "*destinations"
	MetaMinCapPrefix   = "*min_"
	MetaMaxCapPrefix   = "*max_"
	MetaLessThan       = "*lt"
	MetaLessOrEqual    = "*lte"
	MetaGreaterThan    = "*gt"
	MetaGreaterOrEqual = "*gte"
)

func NewFilterS(cfg *config.CGRConfig,
	statSChan chan rpcclient.RpcClientConnection, dm *DataManager) *FilterS {
	return &FilterS{statSChan: statSChan, dm: dm, cfg: cfg}
}

// FilterS is a service used to take decisions in case of filters
// uses lazy connections where necessary to avoid deadlocks on service startup
type FilterS struct {
	cfg        *config.CGRConfig
	statSChan  chan rpcclient.RpcClientConnection // reference towards internal statS connection, used for lazy connect
	statSConns rpcclient.RpcClientConnection
	sSConnMux  sync.RWMutex // make sure only one goroutine attempts connecting
	dm         *DataManager
}

// connStatS returns will connect towards StatS
func (fS *FilterS) connStatS() (err error) {
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
		fS.statSChan, fS.cfg.GeneralCfg().InternalTtl)
	return
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
		f, err := fS.dm.GetFilter(tenant, fltrID,
			true, true, utils.NonTransactional)
		if err != nil {
			return false, err
		}
		if f.ActivationInterval != nil &&
			!f.ActivationInterval.IsActiveAtTime(time.Now()) { // not active
			continue
		}
		for _, fltr := range f.Rules {
			if pass, err = fltr.Pass(ev, fS.statSConns); err != nil || !pass {
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
	if len(ruleSplt) != 3 {
		return nil, fmt.Errorf("inline parse error for string: <%s>", inlnRule)
	}
	f = &Filter{
		Tenant: tenant,
		ID:     inlnRule,
		Rules: []*FilterRule{
			{
				Type:      ruleSplt[0],
				FieldName: ruleSplt[1],
				Values:    strings.Split(ruleSplt[2], utils.INFIELD_SEP)}},
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

func NewFilterRule(rfType, fieldName string, vals []string) (*FilterRule, error) {
	if !utils.IsSliceMember([]string{MetaString, MetaPrefix,
		MetaTimings, MetaRSR, MetaStatS, MetaDestinations,
		MetaLessThan, MetaLessOrEqual, MetaGreaterThan, MetaGreaterOrEqual}, rfType) {
		return nil, fmt.Errorf("Unsupported filter Type: %s", rfType)
	}
	if fieldName == "" && utils.IsSliceMember([]string{MetaString,
		MetaPrefix, MetaTimings, MetaDestinations, MetaLessThan,
		MetaLessOrEqual, MetaGreaterThan, MetaGreaterOrEqual}, rfType) {
		return nil, fmt.Errorf("FieldName is mandatory for Type: %s", rfType)
	}
	if len(vals) == 0 && utils.IsSliceMember([]string{MetaString, MetaPrefix,
		MetaTimings, MetaRSR, MetaDestinations, MetaDestinations, MetaLessThan,
		MetaLessOrEqual, MetaGreaterThan, MetaGreaterOrEqual}, rfType) {
		return nil, fmt.Errorf("Values is mandatory for Type: %s", rfType)
	}
	rf := &FilterRule{Type: rfType, FieldName: fieldName, Values: vals}
	if err := rf.CompileValues(); err != nil {
		return nil, err
	}
	return rf, nil
}

type RFStatSThreshold struct {
	QueueID        string
	ThresholdType  string
	ThresholdValue float64
}

// FilterRule filters requests coming into various places
// Pass rule: default negative, one mathing rule should pass the filter
type FilterRule struct {
	Type            string              // Filter type (*string, *timing, *rsr_filters, *stats, *lt, *lte, *gt, *gte)
	FieldName       string              // Name of the field providing us the Values to check (used in case of some )
	Values          []string            // Filter definition
	rsrFields       config.RSRParsers   // Cache here the RSRFilter Values
	statSThresholds []*RFStatSThreshold // Cached compiled RFStatsThreshold out of Values
}

// Separate method to compile RSR fields
func (rf *FilterRule) CompileValues() (err error) {
	if rf.Type == MetaRSR {
		if rf.rsrFields, err = config.NewRSRParsersFromSlice(rf.Values, true); err != nil {
			return
		}
	} else if rf.Type == MetaStatS {
		rf.statSThresholds = make([]*RFStatSThreshold, len(rf.Values))
		for i, val := range rf.Values {
			valSplt := strings.Split(val, utils.InInFieldSep)
			if len(valSplt) != 3 {
				return fmt.Errorf("Value %s needs to contain at least 3 items", val)
			}
			st := &RFStatSThreshold{QueueID: valSplt[0], ThresholdType: valSplt[1]}
			if len(st.ThresholdType) < len(MetaMinCapPrefix)+1 {
				return fmt.Errorf("Value %s contains a unsupported ThresholdType format", val)
			} else if !strings.HasPrefix(st.ThresholdType, MetaMinCapPrefix) &&
				!strings.HasPrefix(st.ThresholdType, MetaMaxCapPrefix) {
				return fmt.Errorf("Value %s contains unsupported ThresholdType prefix", val)
			}
			if tv, err := strconv.ParseFloat(valSplt[2], 64); err != nil {
				return err
			} else {
				st.ThresholdValue = tv
			}
			rf.statSThresholds[i] = st
		}
	}
	return
}

// Pass is the method which should be used from outside.
func (fltr *FilterRule) Pass(dP config.DataProvider, rpcClnt rpcclient.RpcClientConnection) (bool, error) {
	switch fltr.Type {
	case MetaString:
		return fltr.passString(dP)
	case MetaPrefix:
		return fltr.passStringPrefix(dP)
	case MetaTimings:
		return fltr.passTimings(dP)
	case MetaDestinations:
		return fltr.passDestinations(dP)
	case MetaRSR:
		return fltr.passRSR(dP)
	case MetaStatS:
		return fltr.passStatS(dP, rpcClnt)
	case MetaLessThan, MetaLessOrEqual, MetaGreaterThan, MetaGreaterOrEqual:
		return fltr.passGreaterThan(dP)
	default:
		return false, utils.ErrNotImplemented
	}
}

func (fltr *FilterRule) passString(dP config.DataProvider) (bool, error) {
	strVal, err := dP.FieldAsString(strings.Split(fltr.FieldName, utils.NestingSep))
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	for _, val := range fltr.Values {
		if strVal == val {
			return true, nil
		}
	}
	return false, nil
}

func (fltr *FilterRule) passStringPrefix(dP config.DataProvider) (bool, error) {
	strVal, err := dP.FieldAsString(strings.Split(fltr.FieldName, utils.NestingSep))
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	for _, prfx := range fltr.Values {
		if strings.HasPrefix(strVal, prfx) {
			return true, nil
		}
	}
	return false, nil
}

// ToDo when Timings will be available in DataDb
func (fltr *FilterRule) passTimings(dP config.DataProvider) (bool, error) {
	return false, utils.ErrNotImplemented
}

func (fltr *FilterRule) passDestinations(dP config.DataProvider) (bool, error) {
	dst, err := dP.FieldAsString(strings.Split(fltr.FieldName, utils.NestingSep))
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	for _, p := range utils.SplitPrefix(dst, MIN_PREFIX_MATCH) {
		if destIDs, err := dm.DataDB().GetReverseDestination(p, false, utils.NonTransactional); err == nil {
			for _, dID := range destIDs {
				for _, valDstID := range fltr.Values {
					if valDstID == dID {
						return true, nil
					}
				}
			}
		}
	}
	return false, nil
}

func (fltr *FilterRule) passRSR(dP config.DataProvider) (bool, error) {
	_, err := fltr.rsrFields.ParseDataProviderWithInterfaces(dP, utils.NestingSep)
	if err != nil {
		if err == utils.ErrNotFound || err == utils.ErrFilterNotPassingNoCaps {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (fltr *FilterRule) passStatS(dP config.DataProvider,
	stats rpcclient.RpcClientConnection) (bool, error) {
	if stats == nil || reflect.ValueOf(stats).IsNil() {
		return false, errors.New("Missing StatS information")
	}
	for _, threshold := range fltr.statSThresholds {
		statValues := make(map[string]float64)
		if err := stats.Call("StatSV1.GetFloatMetrics", threshold.QueueID, &statValues); err != nil {
			return false, err
		}
		val, hasIt := statValues[utils.Meta+threshold.ThresholdType[len(MetaMinCapPrefix):]]
		if !hasIt {
			continue
		}
		if strings.HasPrefix(threshold.ThresholdType, MetaMinCapPrefix) &&
			val >= threshold.ThresholdValue {
			return true, nil
		} else if strings.HasPrefix(threshold.ThresholdType, MetaMaxCapPrefix) &&
			val < threshold.ThresholdValue {
			return true, nil
		}
	}
	return false, nil
}

func (fltr *FilterRule) passGreaterThan(dP config.DataProvider) (bool, error) {
	fldIf, err := dP.FieldAsInterface(strings.Split(fltr.FieldName, utils.NestingSep))
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
		orEqual := false
		if fltr.Type == MetaGreaterOrEqual ||
			fltr.Type == MetaLessThan {
			orEqual = true
		}
		if gte, err := utils.GreaterThan(fldIf, utils.StringToInterface(val), orEqual); err != nil {
			return false, err
		} else if utils.IsSliceMember([]string{MetaGreaterThan, MetaGreaterOrEqual}, fltr.Type) && gte {
			return true, nil
		} else if !gte && utils.IsSliceMember([]string{MetaLessThan, MetaLessOrEqual}, fltr.Type) && !gte {
			return true, nil
		}
	}
	return false, nil
}
