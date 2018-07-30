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
	"strings"

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

func (lcs *QOSSupplierSorter) SortSuppliers(prflID string, suppls []*Supplier,
	ev *utils.CGREvent, extraOpts *optsGetSuppliers) (sortedSuppls *SortedSuppliers, err error) {
	sortedSuppls = &SortedSuppliers{ProfileID: prflID,
		Sorting:         lcs.sorting,
		SortedSuppliers: make([]*SortedSupplier, 0)}
	for _, s := range suppls {
		metricSupp, err := lcs.spS.statMetrics(s.StatIDs, ev.Tenant) //create metric map for suppier s
		if err != nil {
			if extraOpts.ignoreErrors {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> profile: %s ignoring supplier with ID: %s, err: %s",
						utils.SupplierS, prflID, s.ID, err.Error()))
				continue
			}
			return nil, err
		}

		srtData := map[string]float64{
			utils.Weight: s.Weight,
		}
		for _, metric := range extraOpts.sortingParameters {
			hasMetric := false                         //check if metricSupp have sortingParameter
			for keyWithID, value := range metricSupp { //transfer data from metric into srtData
				if metric == strings.Split(keyWithID, utils.InInFieldSep)[0] {
					if val, hasKey := srtData[metric]; !hasKey ||
						(metric == utils.MetaPDD && val < value) || //worst values
						(metric != utils.MetaPDD && val > value) {
						srtData[metric] = value
						hasMetric = true
					}
				}
			}
			if !hasMetric { //if not have populate with default value
				switch metric {
				default:
					srtData[metric] = -1
				case utils.MetaPDD:
					srtData[metric] = 1000000
				}
			}
		}

		sortingData := map[string]interface{}{
			utils.Weight: s.Weight,
		}
		for k, v := range metricSupp {
			sortingData[k] = v
		}
		sortedSuppls.SortedSuppliers = append(sortedSuppls.SortedSuppliers,
			&SortedSupplier{
				SupplierID:         s.ID,
				SortingData:        sortingData,
				SupplierParameters: s.SupplierParameters,
				worstStats:         srtData,
			},
		)
	}
	if len(sortedSuppls.SortedSuppliers) == 0 {
		return nil, utils.ErrNotFound
	}
	sortedSuppls.SortQOS(extraOpts.sortingParameters)
	return
}
