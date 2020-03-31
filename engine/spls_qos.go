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
	"github.com/cgrates/cgrates/utils"
)

func NewQOSSupplierSorter(spS *SupplierService) *QOSSupplierSorter {
	return &QOSSupplierSorter{spS: spS,
		sorting: utils.MetaQOS}
}

// QOSSorter sorts suppliers based on stats
type QOSSupplierSorter struct {
	sorting string
	spS     *SupplierService
}

func (qos *QOSSupplierSorter) SortSuppliers(prflID string, suppls []*Supplier,
	ev *utils.CGREvent, extraOpts *optsGetSuppliers) (sortedSuppls *SortedSuppliers, err error) {
	sortedSuppls = &SortedSuppliers{ProfileID: prflID,
		Sorting:         qos.sorting,
		SortedSuppliers: make([]*SortedSupplier, 0)}
	for _, s := range suppls {
		if srtSpl, pass, err := qos.spS.populateSortingData(ev, s, extraOpts); err != nil {
			return nil, err
		} else if pass && srtSpl != nil {
			sortedSuppls.SortedSuppliers = append(sortedSuppls.SortedSuppliers, srtSpl)
		}
	}
	sortedSuppls.SortQOS(extraOpts.sortingParameters)
	return
}
