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

// NewSupplierSortDispatcher constructs SupplierSortDispatcher
func NewSupplierSortDispatcher(lcrS *LCRService) (ssd SupplierSortDispatcher, err error) {
	ssd = make(map[string]SuppliersSorting)
	ssd[utils.MetaWeight] = new(WeightStrategy)
	ssd[utils.MetaLeastCost] = NewLeastCostStrategy(lcrS)
	return
}

// SupplierStrategyHandler will initialize strategies
// and dispatch requests to them
type SupplierSortDispatcher map[string]SuppliersSorting

func (ssd SupplierSortDispatcher) SortSuppliers(strategy string, suppls LCRSuppliers) (err error) {
	sd, has := ssd[strategy]
	if !has {
		return fmt.Errorf("unsupported sort strategy: %s", strategy)
	}
	return sd.SortSuppliers(suppls)
}

type SuppliersSorting interface {
	SortSuppliers(LCRSuppliers) error
}

// NewLeastCostStrategy constructs LeastCostStrategy
func NewLeastCostStrategy(lcrS *LCRService) *LeastCostStrategy {
	return &LeastCostStrategy{lcrS: lcrS}
}

// LeastCostStrategy orders suppliers based on lowest cost
type LeastCostStrategy struct {
	lcrS *LCRService
}

func (lcs *LeastCostStrategy) SortSuppliers(suppls LCRSuppliers) (err error) {
	return
}

// WeightStrategy orders suppliers based on their weight, no cost involved
type WeightStrategy struct {
}

func (ss *WeightStrategy) SortSuppliers(suppls LCRSuppliers) (err error) {
	suppls.Sort()
	return
}
