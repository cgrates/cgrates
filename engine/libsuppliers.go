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
	"sort"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// SupplierReply represents one supplier in
type SortedSupplier struct {
	SupplierID         string
	SupplierParameters string
	SortingData        map[string]interface{} // store here extra info like cost or stats
}

// SuppliersReply is returned as part of GetSuppliers call
type SortedSuppliers struct {
	ProfileID       string            // Profile matched
	Sorting         string            // Sorting algorithm
	SortedSuppliers []*SortedSupplier // list of supplier IDs and SortingData data
}

// SupplierIDs returns list of suppliers
func (sSpls *SortedSuppliers) SupplierIDs() (sIDs []string) {
	sIDs = make([]string, len(sSpls.SortedSuppliers))
	for i, spl := range sSpls.SortedSuppliers {
		sIDs[i] = spl.SupplierID
	}
	return
}

// SupplierIDs returns list of suppliers
func (sSpls *SortedSuppliers) SuppliersWithParams() (sPs []string) {
	sPs = make([]string, len(sSpls.SortedSuppliers))
	for i, spl := range sSpls.SortedSuppliers {
		sPs[i] = spl.SupplierID
		if spl.SupplierParameters != "" {
			sPs[i] += utils.InInFieldSep + spl.SupplierParameters
		}
	}
	return
}

// SortWeight is part of sort interface, sort based on Weight
func (sSpls *SortedSuppliers) SortWeight() {
	sort.Slice(sSpls.SortedSuppliers, func(i, j int) bool {
		return sSpls.SortedSuppliers[i].SortingData[utils.Weight].(float64) > sSpls.SortedSuppliers[j].SortingData[utils.Weight].(float64)
	})
}

// SortLeastCost is part of sort interface,
// sort ascendent based on Cost with fallback on Weight
func (sSpls *SortedSuppliers) SortLeastCost() {
	sort.Slice(sSpls.SortedSuppliers, func(i, j int) bool {
		if sSpls.SortedSuppliers[i].SortingData[utils.Cost].(float64) == sSpls.SortedSuppliers[j].SortingData[utils.Cost].(float64) {
			return sSpls.SortedSuppliers[i].SortingData[utils.Weight].(float64) > sSpls.SortedSuppliers[j].SortingData[utils.Weight].(float64)
		}
		return sSpls.SortedSuppliers[i].SortingData[utils.Cost].(float64) < sSpls.SortedSuppliers[j].SortingData[utils.Cost].(float64)
	})
}

// SortHighestCost is part of sort interface,
// sort descendent based on Cost with fallback on Weight
func (sSpls *SortedSuppliers) SortHighestCost() {
	sort.Slice(sSpls.SortedSuppliers, func(i, j int) bool {
		if sSpls.SortedSuppliers[i].SortingData[utils.Cost].(float64) == sSpls.SortedSuppliers[j].SortingData[utils.Cost].(float64) {
			return sSpls.SortedSuppliers[i].SortingData[utils.Weight].(float64) > sSpls.SortedSuppliers[j].SortingData[utils.Weight].(float64)
		}
		return sSpls.SortedSuppliers[i].SortingData[utils.Cost].(float64) > sSpls.SortedSuppliers[j].SortingData[utils.Cost].(float64)
	})
}

// SortQOS is part of sort interface,
// sort based on Stats

//map[*acd][]float64{10,20} => 10
//map[*tcd][]float64{40,50} => 40
//map[*pdd][]float64{40,50} => 50
func (sSpls *SortedSuppliers) SortQOS(params []string) {
	//sort the metrics before sorting the suppliers
	for _, val := range sSpls.SortedSuppliers {
		for _, iface := range val.SortingData {
			if castedVal, canCast := iface.(SplStatMetrics); !canCast {
				castedVal.Sort()
			}
		}
	}
	//sort suppliers
	sort.Slice(sSpls.SortedSuppliers, func(i, j int) bool {
		for _, param := range params {
			//in case we have the same value for the current param we skip to the next one
			if sSpls.SortedSuppliers[i].SortingData[param].(SplStatMetrics)[0].MetricValue == sSpls.SortedSuppliers[j].SortingData[param].(SplStatMetrics)[0].MetricValue {
				continue
			}
			switch sSpls.SortedSuppliers[i].SortingData[param].(SplStatMetrics)[0].metricType {
			default:
				if sSpls.SortedSuppliers[i].SortingData[param].(SplStatMetrics)[0].MetricValue > sSpls.SortedSuppliers[j].SortingData[param].(SplStatMetrics)[0].MetricValue {
					return true
				}
				return false
			case utils.MetaPDD: //in case of pdd the smalles value if the best
				if sSpls.SortedSuppliers[i].SortingData[param].(SplStatMetrics)[0].MetricValue < sSpls.SortedSuppliers[j].SortingData[param].(SplStatMetrics)[0].MetricValue {
					return true
				}
				return false
			}

		}
		//in case that we have the same value for all params we sort base on weight
		return sSpls.SortedSuppliers[i].SortingData[utils.Weight].(float64) > sSpls.SortedSuppliers[j].SortingData[utils.Weight].(float64)
	})
}

// Digest returns list of supplierIDs + parameters for easier outside access
// format suppl1:suppl1params,suppl2:suppl2params
func (sSpls *SortedSuppliers) Digest() string {
	return strings.Join(sSpls.SuppliersWithParams(), utils.FIELDS_SEP)
}

type SupplierWithParams struct {
	SupplierName   string
	SupplierParams string
}

// SuppliersSorter is the interface which needs to be implemented by supplier sorters
type SuppliersSorter interface {
	SortSuppliers(string, []*Supplier, *utils.CGREvent, *optsGetSuppliers) (*SortedSuppliers, error)
}

// NewSupplierSortDispatcher constructs SupplierSortDispatcher
func NewSupplierSortDispatcher(lcrS *SupplierService) (ssd SupplierSortDispatcher, err error) {
	ssd = make(map[string]SuppliersSorter)
	ssd[utils.MetaWeight] = NewWeightSorter(lcrS)
	ssd[utils.MetaLeastCost] = NewLeastCostSorter(lcrS)
	ssd[utils.MetaHighestCost] = NewHighestCostSorter(lcrS)
	ssd[utils.MetaQOS] = NewQOSSupplierSorter(lcrS)
	return
}

// SupplierStrategyHandler will initialize strategies
// and dispatch requests to them
type SupplierSortDispatcher map[string]SuppliersSorter

func (ssd SupplierSortDispatcher) SortSuppliers(prflID, strategy string,
	suppls []*Supplier, suplEv *utils.CGREvent, extraOpts *optsGetSuppliers) (sortedSuppls *SortedSuppliers, err error) {
	sd, has := ssd[strategy]
	if !has {
		return nil, fmt.Errorf("unsupported sorting strategy: %s", strategy)
	}
	return sd.SortSuppliers(prflID, suppls, suplEv, extraOpts)
}

//StatMetric used to store the statID and the metric value
type SplStatMetric struct {
	StatID      string
	MetricValue float64

	metricType string
}

//StatMetrics  is a sortable list of StatMetric
type SplStatMetrics []*SplStatMetric

// Sort is part of sort interface, sort based on Weight
func (sm SplStatMetrics) Sort() {
	sort.Slice(sm, func(i, j int) bool {
		switch sm[i].metricType {
		default:
			return sm[i].MetricValue < sm[j].MetricValue
		case utils.MetaPDD: // in case of PDD we take the greater value
			return sm[i].MetricValue > sm[j].MetricValue
		}
	})
}

// newRADataProvider constructs a DataProvider
func newSplStsDP(req map[string]SplStatMetrics) (dP config.DataProvider) {
	for key, _ := range req {
		req[key].Sort()
	}
	dP = &splStsDP{req: req, cache: config.NewNavigableMap(nil)}
	return
}

type splStsDP struct {
	req   map[string]SplStatMetrics
	cache *config.NavigableMap
}

// String is part of engine.DataProvider interface
// when called, it will display the already parsed values out of cache
func (sm *splStsDP) String() string {
	return ""
}

// FieldAsInterface is part of engine.DataProvider interface
func (sm *splStsDP) FieldAsInterface(fldPath []string) (data interface{}, err error) {
	fmt.Println("enter here ??? ")
	if data, err = sm.cache.FieldAsInterface(fldPath); err != nil {
		if err != utils.ErrNotFound { // item found in cache
			return
		}
		err = nil // cancel previous err
	} else {
		return // data found in cache
	}
	data = sm.req[fldPath[0]][0].MetricValue
	sm.cache.Set(fldPath, data, false, false)
	return
}

// FieldAsString is part of engine.DataProvider interface
func (sm *splStsDP) FieldAsString(fldPath []string) (data string, err error) {
	var valIface interface{}
	valIface, err = sm.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	data, err = utils.IfaceAsString(valIface)
	return
}

// AsNavigableMap is part of engine.DataProvider interface
func (sm *splStsDP) AsNavigableMap([]*config.FCTemplate) (
	nm *config.NavigableMap, err error) {
	return nil, utils.ErrNotImplemented
}

// RemoteHost is part of engine.DataProvider interface
func (sm *splStsDP) RemoteHost() net.Addr {
	return nil
}
