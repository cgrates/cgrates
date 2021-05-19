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

	"github.com/cgrates/cgrates/utils"
)

func NewLoadDistributionSorter(spS *SupplierService) *LoadDistributionSorter {
	return &LoadDistributionSorter{spS: spS,
		sorting: utils.MetaLoad}
}

// ResourceAscendentSorter orders suppliers based on their Resource Usage
type LoadDistributionSorter struct {
	sorting string
	spS     *SupplierService
}

func (ws *LoadDistributionSorter) SortSuppliers(prflID string,
	suppls []*Supplier, suplEv *utils.CGREvent, extraOpts *optsGetSuppliers, argDsp *utils.ArgDispatcher) (sortedSuppls *SortedSuppliers, err error) {
	sortedSuppls = &SortedSuppliers{ProfileID: prflID,
		Sorting:         ws.sorting,
		SortedSuppliers: make([]*SortedSupplier, 0)}
	for _, s := range suppls {
		// we should have at least 1 statID defined for counting CDR (a.k.a *sum:1)
		if len(s.StatIDs) == 0 {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> supplier: <%s> - empty StatIDs",
					utils.SupplierS, s.ID))
			return nil, utils.NewErrMandatoryIeMissing("StatIDs")
		}
		if srtSpl, pass, err := ws.spS.populateSortingData(suplEv, s, extraOpts, argDsp); err != nil {
			return nil, err
		} else if pass && srtSpl != nil {
			// Add the ratio in SortingData so we can used it later in SortLoadDistribution
			srtSpl.SortingData[utils.Ratio] = s.cacheSupplier[utils.MetaRatio].(float64)
			sortedSuppls.SortedSuppliers = append(sortedSuppls.SortedSuppliers, srtSpl)
		}
	}
	sortedSuppls.SortLoadDistribution()
	return
}
