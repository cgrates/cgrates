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

func NewResourceAscendetSorter(spS *SupplierService) *ResourceAscendentSorter {
	return &ResourceAscendentSorter{spS: spS,
		sorting: utils.MetaReas}
}

// ResourceAscendentSorter orders suppliers based on their Resource Usage
type ResourceAscendentSorter struct {
	sorting string
	spS     *SupplierService
}

func (ws *ResourceAscendentSorter) SortSuppliers(prflID string,
	suppls []*Supplier, suplEv *utils.CGREvent, extraOpts *optsGetSuppliers) (sortedSuppls *SortedSuppliers, err error) {
	sortedSuppls = &SortedSuppliers{ProfileID: prflID,
		Sorting:         ws.sorting,
		SortedSuppliers: make([]*SortedSupplier, 0)}
	for _, s := range suppls {
		if len(s.ResourceIDs) == 0 {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> supplier: <%s> - empty ResourceIDs",
					utils.SupplierS, s.ID))
			return nil, utils.NewErrMandatoryIeMissing("ResourceIDs")
		}
		if srtSpl, pass, err := ws.spS.populateSortingData(suplEv, s, extraOpts); err != nil {
			return nil, err
		} else if pass && srtSpl != nil {
			sortedSuppls.SortedSuppliers = append(sortedSuppls.SortedSuppliers, srtSpl)
		}
	}
	sortedSuppls.SortResourceAscendent()
	return
}
