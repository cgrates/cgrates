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
	MetaStringPrefix   = "*string_prefix"
	MetaTimings        = "*timings"
	MetaRSRFields      = "*rsr_fields"
	MetaStatS          = "*stats"
	MetaDestinations   = "*destinations"
	MetaMinCapPrefix   = "*min_"
	MetaMaxCapPrefix   = "*max_"
	MetaLessThan       = "*lt"
	MetaLessOrEqual    = "*lte"
	MetaGreaterThan    = "*gt"
	MetaGreaterOrEqual = "*gte"
)

func NewFilterS(cfg *config.CGRConfig, statSChan chan rpcclient.RpcClientConnection, dm *DataManager) *FilterS {
	return &FilterS{statSChan: statSChan, dm: dm}
}

// FilterS is a service used to take decisions in case of filters
// uses lazy connections where necessary to avoid deadlocks on service startup
type FilterS struct {
	cfg        *config.CGRConfig
	statSChan  chan rpcclient.RpcClientConnection // reference towards internal statS connection, used for lazy connect
	statSConns *rpcclient.RpcClientPool
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
	fS.statSConns, err = NewRPCPool(rpcclient.POOL_FIRST, fS.cfg.ConnectAttempts, fS.cfg.Reconnects, fS.cfg.ConnectTimeout, fS.cfg.ReplyTimeout,
		fS.cfg.FilterSCfg().StatSConns, fS.statSChan, fS.cfg.InternalTtl)
	return
}

// PassFiltersForEvent will check all filters wihin filterIDs and require them passing for event
// there should be at least one filter passing, ie: if filters are not active event will fail to pass
func (fS *FilterS) PassFiltersForEvent(tenant string, ev map[string]interface{}, filterIDs []string) (pass bool, err error) {
	var atLeastOneFilterPassing bool
	for _, fltrID := range filterIDs {
		f, err := fS.dm.GetFilter(tenant, fltrID, false, utils.NonTransactional)
		if err != nil {
			return false, err
		}
		if f.ActivationInterval != nil &&
			!f.ActivationInterval.IsActiveAtTime(time.Now()) { // not active
			continue
		}
		for _, fltr := range f.RequestFilters {
			switch fltr.Type {
			case MetaString:
				pass, err = fltr.passString(ev, "")
			case MetaStringPrefix:
				pass, err = fltr.passStringPrefix(ev, "")
			case MetaTimings:
				pass, err = fltr.passTimings(ev, "")
			case MetaDestinations:
				pass, err = fltr.passDestinations(ev, "")
			case MetaRSRFields:
				pass, err = fltr.passRSRFields(ev, "")
			case MetaStatS:
				if err = fS.connStatS(); err != nil {
					return false, err
				}
				pass, err = fltr.passStatS(ev, "", fS.statSConns)
			case MetaLessThan, MetaLessOrEqual, MetaGreaterThan, MetaGreaterOrEqual:
				pass, err = fltr.passGreaterThan(ev, "")
			default:
				return false, fmt.Errorf("tenant: %s filter: %s unsupported filter type: <%s>", tenant, fltrID, fltr.Type)
			}
		}
		if err != nil || !pass {
			return pass, err
		}
		atLeastOneFilterPassing = true
	}
	return atLeastOneFilterPassing, nil
}

type Filter struct {
	Tenant             string
	ID                 string
	RequestFilters     []*RequestFilter
	ActivationInterval *utils.ActivationInterval
}

func (flt *Filter) TenantID() string {
	return utils.ConcatenatedKey(flt.Tenant, flt.ID)
}

// Compile will compile the underlaying request filters where necessary (ie. regexp rules)
func (f *Filter) Compile() (err error) {
	for _, rf := range f.RequestFilters {
		if err = rf.CompileValues(); err != nil {
			return
		}
	}
	return
}

func NewRequestFilter(rfType, fieldName string, vals []string) (*RequestFilter, error) {
	if !utils.IsSliceMember([]string{MetaStringPrefix, MetaTimings, MetaRSRFields, MetaStatS, MetaDestinations,
		MetaLessThan, MetaLessOrEqual, MetaGreaterThan, MetaGreaterOrEqual}, rfType) {
		return nil, fmt.Errorf("Unsupported filter Type: %s", rfType)
	}
	if fieldName == "" && utils.IsSliceMember([]string{MetaStringPrefix, MetaTimings, MetaDestinations,
		MetaLessThan, MetaLessOrEqual, MetaGreaterThan, MetaGreaterOrEqual}, rfType) {
		return nil, fmt.Errorf("FieldName is mandatory for Type: %s", rfType)
	}
	if len(vals) == 0 && utils.IsSliceMember([]string{MetaStringPrefix, MetaTimings, MetaRSRFields,
		MetaDestinations, MetaDestinations, MetaLessThan, MetaLessOrEqual, MetaGreaterThan, MetaGreaterOrEqual}, rfType) {
		return nil, fmt.Errorf("Values is mandatory for Type: %s", rfType)
	}
	rf := &RequestFilter{Type: rfType, FieldName: fieldName, Values: vals}
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

// RequestFilter filters requests coming into various places
// Pass rule: default negative, one mathing rule should pass the filter
type RequestFilter struct {
	Type               string   // Filter type (*string, *timing, *rsr_filters, *stats, *lt, *lte, *gt, *gte)
	FieldName          string   // Name of the field providing us the Values to check (used in case of some )
	Values             []string // Filter definition
	ActivationInterval *utils.ActivationInterval
	rsrFields          utils.RSRFields     // Cache here the RSRFilter Values
	statSThresholds    []*RFStatSThreshold // Cached compiled RFStatsThreshold out of Values
}

// Separate method to compile RSR fields
func (rf *RequestFilter) CompileValues() (err error) {
	if rf.Type == MetaRSRFields {
		if rf.rsrFields, err = utils.ParseRSRFieldsFromSlice(rf.Values); err != nil {
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
			} else if !strings.HasPrefix(st.ThresholdType, MetaMinCapPrefix) && !strings.HasPrefix(st.ThresholdType, MetaMaxCapPrefix) {
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
func (fltr *RequestFilter) Pass(req interface{}, extraFieldsLabel string, rpcClnt rpcclient.RpcClientConnection) (bool, error) {
	switch fltr.Type {
	case MetaString:
		return fltr.passString(req, extraFieldsLabel)
	case MetaStringPrefix:
		return fltr.passStringPrefix(req, extraFieldsLabel)
	case MetaTimings:
		return fltr.passTimings(req, extraFieldsLabel)
	case MetaDestinations:
		return fltr.passDestinations(req, extraFieldsLabel)
	case MetaRSRFields:
		return fltr.passRSRFields(req, extraFieldsLabel)
	case MetaStatS:
		return fltr.passStatS(req, extraFieldsLabel, rpcClnt)
	case MetaLessThan, MetaLessOrEqual, MetaGreaterThan, MetaGreaterOrEqual:
		return fltr.passGreaterThan(req, extraFieldsLabel)
	default:
		return false, utils.ErrNotImplemented
	}
}

func (fltr *RequestFilter) passString(req interface{}, extraFieldsLabel string) (bool, error) {
	strVal, err := utils.ReflectFieldAsString(req, fltr.FieldName, extraFieldsLabel)
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

func (fltr *RequestFilter) passStringPrefix(req interface{}, extraFieldsLabel string) (bool, error) {
	strVal, err := utils.ReflectFieldAsString(req, fltr.FieldName, extraFieldsLabel)
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
func (fltr *RequestFilter) passTimings(req interface{}, extraFieldsLabel string) (bool, error) {
	return false, utils.ErrNotImplemented
}

func (fltr *RequestFilter) passDestinations(req interface{}, extraFieldsLabel string) (bool, error) {
	dst, err := utils.ReflectFieldAsString(req, fltr.FieldName, extraFieldsLabel)
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

func (fltr *RequestFilter) passRSRFields(req interface{}, extraFieldsLabel string) (bool, error) {
	for _, rsrFld := range fltr.rsrFields {
		if strVal, err := utils.ReflectFieldAsString(req, rsrFld.Id, extraFieldsLabel); err != nil {
			if err == utils.ErrNotFound {
				return false, nil
			}
			return false, err
		} else if rsrFld.FilterPasses(strVal) {
			return true, nil
		}
	}
	return false, nil
}

func (fltr *RequestFilter) passStatS(req interface{}, extraFieldsLabel string, stats rpcclient.RpcClientConnection) (bool, error) {
	if stats == nil || reflect.ValueOf(stats).IsNil() {
		return false, errors.New("Missing StatS information")
	}
	for _, threshold := range fltr.statSThresholds {
		statValues := make(map[string]float64)
		if err := stats.Call("StatSV1.GetFloatMetrics", threshold.QueueID, &statValues); err != nil {
			return false, err
		}
		val, hasIt := statValues[utils.MetaPrefix+threshold.ThresholdType[len(MetaMinCapPrefix):]]
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

func (fltr *RequestFilter) passGreaterThan(req interface{}, extraFieldsLabel string) (bool, error) {
	fldIf, err := utils.ReflectFieldInterface(req, fltr.FieldName, extraFieldsLabel)
	if err != nil {
		if err == utils.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	for _, val := range fltr.Values {
		ifVal := utils.StringToInterface(val)
		orEqual := false
		if fltr.Type == MetaGreaterOrEqual ||
			fltr.Type == MetaLessThan {
			orEqual = true
		}
		if gte, err := utils.GreaterThan(fldIf, ifVal, orEqual); err != nil {
			return false, err
		} else if utils.IsSliceMember([]string{MetaGreaterThan, MetaGreaterOrEqual}, fltr.Type) && gte {
			return true, nil
		} else if !gte && utils.IsSliceMember([]string{MetaLessThan, MetaLessOrEqual}, fltr.Type) && !gte {
			return true, nil
		}
	}
	return false, nil
}
