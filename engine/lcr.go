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
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

const (
	LCR_STRATEGY_STATIC        = "*static"
	LCR_STRATEGY_LOWEST        = "*lowest_cost"
	LCR_STRATEGY_HIGHEST       = "*highest_cost"
	LCR_STRATEGY_QOS_THRESHOLD = "*qos_threshold"
	LCR_STRATEGY_QOS           = "*qos"
	LCR_STRATEGY_LOAD          = "*load_distribution"

	// used for load distribution sorting
	RAND_LIMIT          = 99
	LOW_PRIORITY_LIMIT  = 100
	MED_PRIORITY_LIMIT  = 200
	HIGH_PRIORITY_LIMIT = 300
)

// A request for LCR, used in APIer and SM where we need to expose it
type LcrRequest struct {
	Direction    string
	Tenant       string
	Category     string
	Account      string
	Subject      string
	Destination  string
	SetupTime    string
	Duration     string
	IgnoreErrors bool
	ExtraFields  map[string]string
	*LCRFilter
	*utils.Paginator
}

type LCRFilter struct {
	MinCost *float64
	MaxCost *float64
}

func (self *LcrRequest) AsCallDescriptor(timezone string) (*CallDescriptor, error) {
	if len(self.Account) == 0 || len(self.Destination) == 0 {
		return nil, utils.ErrMandatoryIeMissing
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
	if len(self.SetupTime) == 0 {
		timeStart = time.Now()
	} else if timeStart, err = utils.ParseTimeDetectLayout(self.SetupTime, timezone); err != nil {
		return nil, err
	}
	var callDur time.Duration
	if len(self.Duration) == 0 {
		callDur = time.Duration(1) * time.Minute
	} else if callDur, err = utils.ParseDurationWithSecs(self.Duration); err != nil {
		return nil, err
	}
	cd := &CallDescriptor{
		Direction:   self.Direction,
		Tenant:      self.Tenant,
		Category:    self.Category,
		Account:     self.Account,
		Subject:     self.Subject,
		Destination: self.Destination,
		TimeStart:   timeStart,
		TimeEnd:     timeStart.Add(callDur),
	}
	if self.ExtraFields != nil {
		cd.ExtraFields = make(map[string]string)
	}
	for key, val := range self.ExtraFields {
		cd.ExtraFields[key] = val
	}
	return cd, nil
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
	Supplier       string
	Cost           float64
	Duration       time.Duration
	Error          string // Not error due to JSON automatic serialization into struct
	QOS            map[string]float64
	qosSortParams  []string
	supplierQueues []*StatsQueue // used for load distribution
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

func (le *LCREntry) GetQOSLimits() (minASR, maxASR float64, minPDD, maxPDD, minACD, maxACD, minTCD, maxTCD time.Duration, minACC, maxACC, minTCC, maxTCC, minDDC, maxDDC float64) {
	// MIN_ASR;MAX_ASR;MIN_PDD;MAX_PDD;MIN_ACD;MAX_ACD;MIN_TCD;MAX_TCD;MIN_ACC;MAX_ACC;MIN_TCC;MAX_TCC;MIN_DDC;MAX_DDC
	minASR, maxASR, minPDD, maxPDD, minACD, maxACD, minTCD, maxTCD, minACC, maxACC, minTCC, maxTCC, minDDC, maxDDC = -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1
	params := strings.Split(le.StrategyParams, utils.INFIELD_SEP)
	if len(params) == 14 {
		var err error
		if minASR, err = strconv.ParseFloat(params[0], 64); err != nil {
			minASR = -1
		}
		if maxASR, err = strconv.ParseFloat(params[1], 64); err != nil {
			maxASR = -1
		}
		if minPDD, err = utils.ParseDurationWithSecs(params[2]); err != nil {
			minPDD = -1
		}
		if maxPDD, err = utils.ParseDurationWithSecs(params[3]); err != nil {
			maxPDD = -1
		}
		if minACD, err = utils.ParseDurationWithSecs(params[4]); err != nil {
			minACD = -1
		}
		if maxACD, err = utils.ParseDurationWithSecs(params[5]); err != nil {
			maxACD = -1
		}
		if minTCD, err = utils.ParseDurationWithSecs(params[6]); err != nil {
			minTCD = -1
		}
		if maxTCD, err = utils.ParseDurationWithSecs(params[7]); err != nil {
			maxTCD = -1
		}
		if minACC, err = strconv.ParseFloat(params[8], 64); err != nil {
			minACC = -1
		}
		if maxACC, err = strconv.ParseFloat(params[9], 64); err != nil {
			maxACC = -1
		}
		if minTCC, err = strconv.ParseFloat(params[10], 64); err != nil {
			minTCC = -1
		}
		if maxTCC, err = strconv.ParseFloat(params[11], 64); err != nil {
			maxTCC = -1
		}
		if minDDC, err = strconv.ParseFloat(params[12], 64); err != nil {
			minDDC = -1
		}
		if maxDDC, err = strconv.ParseFloat(params[13], 64); err != nil {
			maxDDC = -1
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
		return []string{ASR, PDD, ACD, TCD, ACC, TCC, DDC} // Default QoS stats if none configured
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
		if x, ok := CacheGet(utils.DESTINATION_PREFIX + p); ok {

			destIds := x.(map[string]struct{})
			for dId := range destIds {
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
	case LCR_STRATEGY_LOAD:
		lc.SortLoadDistribution()
		sort.Sort(HighestSupplierCostSorter(lc.SupplierCosts))
	}
}

func (lc *LCRCost) SortLoadDistribution() {
	// find the time window that is common to all qeues
	scoreBoard := make(map[time.Duration]int) // register TimeWindow across suppliers

	var winnerTimeWindow time.Duration
	maxScore := 0
	for _, supCost := range lc.SupplierCosts {
		timeWindowFlag := make(map[time.Duration]bool) // flags appearance in same supplier
		for _, sq := range supCost.supplierQueues {
			if !timeWindowFlag[sq.conf.TimeWindow] {
				timeWindowFlag[sq.conf.TimeWindow] = true
				scoreBoard[sq.conf.TimeWindow]++
			}
			if scoreBoard[sq.conf.TimeWindow] > maxScore {
				maxScore = scoreBoard[sq.conf.TimeWindow]
				winnerTimeWindow = sq.conf.TimeWindow
			}
		}
	}
	supplierQueues := make(map[*LCRSupplierCost]*StatsQueue)
	for _, supCost := range lc.SupplierCosts {
		for _, sq := range supCost.supplierQueues {
			if sq.conf.TimeWindow == winnerTimeWindow {
				supplierQueues[supCost] = sq
				break
			}
		}
	}
	/*for supplier, sq := range supplierQueues {
		log.Printf("Useful supplier qeues: %s %v", supplier, sq.conf.TimeWindow)
	}*/
	// if all have less than ratio return random order
	// if some have a cdr count not divisible by ratio return them first and all ordered by cdr times, oldest first
	// if all have a multiple of ratio return in the order of cdr times, oldest first

	// first put them in one of the above categories
	haveRatiolessSuppliers := false
	for supCost, sq := range supplierQueues {
		ratio := lc.GetSupplierRatio(supCost.Supplier)
		if ratio == -1 {
			supCost.Cost = -1
			haveRatiolessSuppliers = true
			continue
		}
		cdrCount := len(sq.Cdrs)
		if cdrCount < ratio {
			supCost.Cost = float64(LOW_PRIORITY_LIMIT + rand.Intn(RAND_LIMIT))
			continue
		}
		if cdrCount%ratio == 0 {
			supCost.Cost = float64(MED_PRIORITY_LIMIT+rand.Intn(RAND_LIMIT)) + (time.Now().Sub(sq.Cdrs[len(sq.Cdrs)-1].SetupTime).Seconds() / RAND_LIMIT)
			continue
		} else {
			supCost.Cost = float64(HIGH_PRIORITY_LIMIT+rand.Intn(RAND_LIMIT)) + (time.Now().Sub(sq.Cdrs[len(sq.Cdrs)-1].SetupTime).Seconds() / RAND_LIMIT)
			continue
		}
	}
	if haveRatiolessSuppliers {
		var filteredSupplierCost []*LCRSupplierCost
		for _, supCost := range lc.SupplierCosts {
			if supCost.Cost != -1 {
				filteredSupplierCost = append(filteredSupplierCost, supCost)
			}
		}
		lc.SupplierCosts = filteredSupplierCost
	}
}

// used in load distribution strategy only
// receives a long supplier id and will return the ratio found in strategy params
func (lc *LCRCost) GetSupplierRatio(supplier string) int {
	// parse strategy params
	ratios := make(map[string]int)
	params := strings.Split(lc.Entry.StrategyParams, utils.INFIELD_SEP)
	for _, param := range params {
		ratioSlice := strings.Split(param, utils.CONCATENATED_KEY_SEP)
		if len(ratioSlice) != 2 {
			utils.Logger.Warning(fmt.Sprintf("bad format in load distribution strategy param: %s", lc.Entry.StrategyParams))
			continue
		}
		p, err := strconv.Atoi(ratioSlice[1])
		if err != nil {
			utils.Logger.Warning(fmt.Sprintf("bad format in load distribution strategy param: %s", lc.Entry.StrategyParams))
			continue
		}
		ratios[ratioSlice[0]] = p
	}
	parts := strings.Split(supplier, utils.CONCATENATED_KEY_SEP)
	if len(parts) > 0 {
		supplierSubject := parts[len(parts)-1]
		if ratio, found := ratios[supplierSubject]; found {
			return ratio
		}
		if ratio, found := ratios[utils.META_DEFAULT]; found {
			return ratio
		}
	}
	if len(ratios) == 0 {
		return 1 // use random/last cdr date sorting
	}
	return -1 // exclude missing suppliers
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
			utils.Logger.Err(fmt.Sprintf("LCR_ERROR: supplier <%s>, error <%s>", supplCost.Supplier, supplCost.Error))
		}
	}
}

func (lc *LCRCost) SuppliersSlice() ([]string, error) {
	if lc.Entry == nil {
		return nil, utils.ErrNotFound
	}
	supps := []string{}
	for _, supplCost := range lc.SupplierCosts {
		if supplCost.Error != "" {
			continue // Do not add the supplier with cost errors to list of suppliers available
		}
		if dtcs, err := utils.NewDTCSFromRPKey(supplCost.Supplier); err != nil {
			return nil, err
		} else if len(dtcs.Subject) != 0 {
			supps = append(supps, dtcs.Subject)
		}
	}
	if len(supps) == 0 {
		return nil, utils.ErrNotFound
	}
	return supps, nil
}

// Returns a list of suppliers separated via
func (lc *LCRCost) SuppliersString() (string, error) {
	supps, err := lc.SuppliersSlice()
	if err != nil {
		return "", err
	}
	supplStr := ""
	for idx, suppl := range supps {
		if idx != 0 {
			supplStr += utils.FIELDS_SEP
		}
		supplStr += suppl
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
		// skip to next param
		if qoss[i].QOS[param] == qoss[j].QOS[param] {
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
	}
	return false
}
