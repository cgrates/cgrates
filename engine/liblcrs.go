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

// NewSupplierStrategyDispatcher constructs SupplierStrategyDispatcher
func NewSupplierStrategyDispatcher(lcrS *LCRService) (ssd SupplierStrategyDispatcher, err error) {
	ssd = make(map[string]SuppliersStrategy)
	ssd[utils.MetaStatic] = new(StaticStrategy)
	ssd[utils.MetaLeastCost] = NewLeastCostStrategy(lcrS)
	return
}

// SupplierStrategyHandler will initialize strategies
// and dispatch requests to them
type SupplierStrategyDispatcher map[string]SuppliersStrategy

func (ssd SupplierStrategyDispatcher) OrderSuppliers(strategy string, suppls LCRSuppliers) (err error) {
	sd, has := ssd[strategy]
	if !has {
		return fmt.Errorf("unsupported sort strategy: %s", strategy)
	}
	return sd.OrderSuppliers(suppls)
}

type SuppliersStrategy interface {
	OrderSuppliers(LCRSuppliers) error
}

// NewLeastCostStrategy constructs LeastCostStrategy
func NewLeastCostStrategy(lcrS *LCRService) *LeastCostStrategy {
	return &LeastCostStrategy{lcrS: lcrS}
}

// LeastCostStrategy orders suppliers based on lowest cost
type LeastCostStrategy struct {
	lcrS *LCRService
}

func (lcs *LeastCostStrategy) OrderSuppliers(suppls LCRSuppliers) (err error) {
	return
}

// StaticStrategy orders suppliers based on their weight, no cost involved
type StaticStrategy struct {
}

func (ss *StaticStrategy) OrderSuppliers(suppls LCRSuppliers) (err error) {
	suppls.Sort()
	return
}
