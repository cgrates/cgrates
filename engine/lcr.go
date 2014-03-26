/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"sort"
	"time"

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/utils"
)

const (
	LCR_STRATEGY_STATIC  = "*static"
	LCR_STRATEGY_LOWEST  = "*lowest"
	LCR_STRATEGY_HIGHEST = "*highest"
)

type LCR struct {
	Tenant      string
	Customer    string
	Direction   string
	Activations []*LCRActivation
}
type LCRActivation struct {
	ActivationTime time.Time
	Entries        []*LCREntry
}
type LCREntry struct {
	DestinationId string
	TOR           string
	Strategy      string
	Suppliers     string
	Weight        float64
	precision     int
}

type LCRCost struct {
	TimeSpans []*LCRTimeSpan
}

type LCRTimeSpan struct {
	StartTime     time.Time
	SupplierCosts []*LCRSupplierCost
	Entry         *LCREntry
}

type LCRSupplierCost struct {
	Supplier string
	Cost     float64
	Error    error
}

func (lcr *LCR) GetId() string {
	return fmt.Sprintf("%s:%s:%s", lcr.Direction, lcr.Tenant, lcr.Customer)
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

type LCREntriesSorter []*LCREntry

func (es LCREntriesSorter) Len() int {
	return len(es)
}

func (es LCREntriesSorter) Swap(i, j int) {
	es[i], es[j] = es[j], es[i]
}

func (es LCREntriesSorter) Less(j, i int) bool {
	return es[i].precision < es[j].precision ||
		(es[i].precision == es[j].precision && es[i].Weight < es[j].Weight)

}

func (es LCREntriesSorter) Sort() {
	sort.Sort(es)
}

func (lcra *LCRActivation) GetLCREntryForPrefix(destination string) *LCREntry {
	var potentials LCREntriesSorter
	for _, p := range utils.SplitPrefix(destination, MIN_PREFIX_MATCH) {
		if x, err := cache2go.GetCached(DESTINATION_PREFIX + p); err == nil {
			destIds := x.([]string)
			for _, dId := range destIds {
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

func (lts *LCRTimeSpan) Sort() {
	switch lts.Entry.Strategy {
	case LCR_STRATEGY_LOWEST:
		sort.Sort(LowestSupplierCostSorter(lts.SupplierCosts))
	case LCR_STRATEGY_HIGHEST:
		sort.Sort(HighestSupplierCostSorter(lts.SupplierCosts))
	}
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
