/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

const (
	LCR_STRATEGY_STATIC        = "*static"
	LCR_STRATEGY_LOWEST        = "*lowest_cost"
	LCR_STRATEGY_HIGHEST       = "*highest_cost"
	LCR_STRATEGY_QOS_THRESHOLD = "*qos_threshold"
	LCR_STRATEGY_QOS           = "*qos"
)

// A request for LCR, used in APIer and SM where we need to expose it
type LcrRequest struct {
	Direction   string
	Tenant      string
	Category    string
	Account     string
	Subject     string
	Destination string
	StartTime   string
	Duration    string
}

func (self *LcrRequest) AsCallDescriptor() (*CallDescriptor, error) {
	if len(self.Account) == 0 || len(self.Destination) == 0 {
		return nil, errors.New(utils.ERR_MANDATORY_IE_MISSING)
	}
	// Set defaults
	if len(self.Direction) == 0 {
		self.Direction = utils.OUT
	}
	if len(self.Tenant) == 0 {
		self.Tenant = config.CgrConfig().DefaultTenant
	}
	if len(self.Category) == 0 {
		self.Category = config.CgrConfig().DefaultCategory
	}
	if len(self.Subject) == 0 {
		self.Subject = self.Account
	}
	var timeStart time.Time
	var err error
	if len(self.StartTime) == 0 {
		timeStart = time.Now()
	} else if timeStart, err = utils.ParseTimeDetectLayout(self.StartTime); err != nil {
		return nil, err
	}
	var callDur time.Duration
	if len(self.Duration) == 0 {
		callDur = time.Duration(1) * time.Minute
	} else if callDur, err = utils.ParseDurationWithSecs(self.Duration); err != nil {
		return nil, err
	}
	return &CallDescriptor{
		Direction:   self.Direction,
		Tenant:      self.Tenant,
		Category:    self.Category,
		Account:     self.Account,
		Subject:     self.Subject,
		Destination: self.Destination,
		TimeStart:   timeStart,
		TimeEnd:     timeStart.Add(callDur),
	}, nil
}

// A LCR reply, used in APIer and SM where we need to expose it
type LcrReply struct {
	DestinationId string
	RPCategory    string
	Strategy      string
	Suppliers     []*LcrSupplier
}

// One supplier out of LCR reply
type LcrSupplier struct {
	Supplier string
	Cost     float64
	QOS      map[string]float64
}

type LCR struct {
	Direction   string
	Tenant      string
	Category    string
	Account     string
	Subject     string
	Activations []*LCRActivation
}
type LCRActivation struct {
	ActivationTime time.Time
	Entries        []*LCREntry
}
type LCREntry struct {
	DestinationId  string
	RPCategory     string
	Strategy       string
	StrategyParams string
	Weight         float64
	precision      int
}

type LCRCost struct {
	Entry         *LCREntry
	SupplierCosts []*LCRSupplierCost
}

type LCRSupplierCost struct {
	Supplier      string
	Cost          float64
	Duration      time.Duration
	Error         string // Not error due to JSON automatic serialization into struct
	QOS           map[string]float64
	qosSortParams []string
}

func (lcr *LCR) GetId() string {
	return utils.LCRKey(lcr.Direction, lcr.Tenant, lcr.Category, lcr.Account, lcr.Subject)
}

func (lcr *LCR) Len() int {
	return len(lcr.Activations)
}

func (lcr *LCR) Swap(i, j int) {
	lcr.Activations[i], lcr.Activations[j] = lcr.Activations[j], lcr.Activations[i]
}

func (lcr *LCR) Less(i, j int) bool {
	return lcr.Activations[i].ActivationTime.Before(lcr.Activations[j].ActivationTime)
}

func (lcr *LCR) Sort() {
	sort.Sort(lcr)
}

func (le *LCREntry) GetQOSLimits() (minASR, maxASR float64, minACD, maxACD, minTCD, maxTCD time.Duration, minACC, maxACC, minTCC, maxTCC float64) {
	// MIN_ASR;MAX_ASR;MIN_ACD;MAX_ACD;MIN_TCD;MAX_TCD;MIN_ACC;MAX_ACC;MIN_TCC;MAX_TCC
	minASR, maxASR, minACD, maxACD, minTCD, maxTCD, minACC, maxACC, minTCC, maxTCC = -1, -1, -1, -1, -1, -1, -1, -1, -1, -1
	params := strings.Split(le.StrategyParams, utils.INFIELD_SEP)
	if len(params) == 10 {
		var err error
		if minASR, err = strconv.ParseFloat(params[0], 64); err != nil {
			minASR = -1
		}
		if maxASR, err = strconv.ParseFloat(params[1], 64); err != nil {
			maxASR = -1
		}
		if minACD, err = time.ParseDuration(params[2]); err != nil {
			minACD = -1
		}
		if maxACD, err = time.ParseDuration(params[3]); err != nil {
			maxACD = -1
		}
		if minTCD, err = time.ParseDuration(params[4]); err != nil {
			minTCD = -1
		}
		if maxTCD, err = time.ParseDuration(params[5]); err != nil {
			maxTCD = -1
		}
		if minACC, err = strconv.ParseFloat(params[6], 64); err != nil {
			minACC = -1
		}
		if maxACC, err = strconv.ParseFloat(params[7], 64); err != nil {
			maxACC = -1
		}
		if minTCC, err = strconv.ParseFloat(params[8], 64); err != nil {
			minTCC = -1
		}
		if maxTCC, err = strconv.ParseFloat(params[9], 64); err != nil {
			maxTCC = -1
		}
	}
	return
}

func (le *LCREntry) GetParams() []string {
	// ASR;ACD
	params := strings.Split(le.StrategyParams, utils.INFIELD_SEP)
	// eliminate empty strings
	var cleanParams []string
	for _, p := range params {
		p = strings.TrimSpace(p)
		if p != "" {
			cleanParams = append(cleanParams, p)
		}
	}
	if len(cleanParams) == 0 && le.Strategy == LCR_STRATEGY_QOS {
		return []string{ASR, ACD, TCD, ACC, TCC} // Default QoS stats if none configured
	}
	return cleanParams
}

type LCREntriesSorter []*LCREntry

func (es LCREntriesSorter) Len() int {
	return len(es)
}

func (es LCREntriesSorter) Swap(i, j int) {
	es[i], es[j] = es[j], es[i]
}

// we need the best earlyer in the list
func (es LCREntriesSorter) Less(j, i int) bool {
	return es[i].Weight < es[j].Weight ||
		(es[i].Weight == es[j].Weight && es[i].precision < es[j].precision)

}

func (es LCREntriesSorter) Sort() {
	sort.Sort(es)
}

func (lcra *LCRActivation) GetLCREntryForPrefix(destination string) *LCREntry {
	var potentials LCREntriesSorter
	for _, p := range utils.SplitPrefix(destination, MIN_PREFIX_MATCH) {
		if x, err := cache2go.GetCached(DESTINATION_PREFIX + p); err == nil {

			destIds := x.(map[interface{}]struct{})
			for idId := range destIds {
				dId := idId.(string)
				for _, entry := range lcra.Entries {
					if entry.DestinationId == dId {
						entry.precision = len(p)
						potentials = append(potentials, entry)
					}
				}
			}
		}
	}
	if len(potentials) > 0 {
		// sort by precision and weight
		potentials.Sort()
		return potentials[0]
	}
	// return the *any entry if it exists
	for _, entry := range lcra.Entries {
		if entry.DestinationId == utils.ANY {
			return entry
		}
	}
	return nil
}

func (lc *LCRCost) Sort() {
	switch lc.Entry.Strategy {
	case LCR_STRATEGY_LOWEST, LCR_STRATEGY_QOS_THRESHOLD:
		sort.Sort(LowestSupplierCostSorter(lc.SupplierCosts))
	case LCR_STRATEGY_HIGHEST:
		sort.Sort(HighestSupplierCostSorter(lc.SupplierCosts))
	case LCR_STRATEGY_QOS:
		sort.Sort(QOSSorter(lc.SupplierCosts))
	}
}

func (lc *LCRCost) HasErrors() bool {
	for _, supplCost := range lc.SupplierCosts {
		if len(supplCost.Error) != 0 {
			return true
		}
	}
	return false
}

func (lc *LCRCost) LogErrors() {
	for _, supplCost := range lc.SupplierCosts {
		if len(supplCost.Error) != 0 {
			Logger.Err(fmt.Sprintf("LCR_ERROR: supplier <%s>, error <%s>", supplCost.Supplier, supplCost.Error))
		}
	}
}

// Returns a list of suppliers separated via
func (lc *LCRCost) SuppliersString() (string, error) {
	supplStr := ""
	if lc.Entry == nil {
		return "", errors.New(utils.ERR_NOT_FOUND)
	}
	for idx, supplCost := range lc.SupplierCosts {
		if dtcs, err := utils.NewDTCSFromRPKey(supplCost.Supplier); err != nil {
			return "", err
		} else {
			if idx != 0 {
				supplStr += utils.FIELDS_SEP
			}
			supplStr += dtcs.Subject
		}
	}
	return supplStr, nil
}

type LowestSupplierCostSorter []*LCRSupplierCost

func (lscs LowestSupplierCostSorter) Len() int {
	return len(lscs)
}

func (lscs LowestSupplierCostSorter) Swap(i, j int) {
	lscs[i], lscs[j] = lscs[j], lscs[i]
}

func (lscs LowestSupplierCostSorter) Less(i, j int) bool {
	return lscs[i].Cost < lscs[j].Cost
}

type HighestSupplierCostSorter []*LCRSupplierCost

func (hscs HighestSupplierCostSorter) Len() int {
	return len(hscs)
}

func (hscs HighestSupplierCostSorter) Swap(i, j int) {
	hscs[i], hscs[j] = hscs[j], hscs[i]
}

func (hscs HighestSupplierCostSorter) Less(i, j int) bool {
	return hscs[i].Cost > hscs[j].Cost
}

type QOSSorter []*LCRSupplierCost

func (qoss QOSSorter) Len() int {
	return len(qoss)
}

func (qoss QOSSorter) Swap(i, j int) {
	qoss[i], qoss[j] = qoss[j], qoss[i]
}

func (qoss QOSSorter) Less(i, j int) bool {
	for _, param := range qoss[i].qosSortParams {
		// if one of the supplier is missing the qos parram skip to next one
		if _, exists := qoss[i].QOS[param]; !exists {
			continue
		}
		if _, exists := qoss[j].QOS[param]; !exists {
			continue
		}
		// -1 is the best
		if qoss[j].QOS[param] == -1 {
			return false
		}
		// more is better
		if qoss[i].QOS[param] == -1 || qoss[i].QOS[param] > qoss[j].QOS[param] {
			return true
		}
		// skip to next param
		if qoss[i].QOS[param] == qoss[j].QOS[param] {
			continue
		}
	}
	return false
}
