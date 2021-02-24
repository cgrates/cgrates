/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package accounts

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// actSetAccount updates the balances base on the diktas
func actSetAccount(dm *engine.DataManager, tnt, acntID string, diktats []*utils.BalDiktat, reset bool) (err error) {
	var qAcnt *utils.AccountProfile
	if qAcnt, err = dm.GetAccountProfile(tnt, acntID); err != nil {
		return
	}
	for _, dk := range diktats {
		// check if we have a valid path(e.g. *balance.Test.ID)
		path := strings.Split(dk.Path, utils.NestingSep)
		if len(path) == 0 {
			return utils.ErrWrongPath
		}
		if path[0] != utils.MetaBalance {
			return utils.ErrWrongPath
		}
		if len(path) < 3 {
			return utils.ErrWrongPath
		}
		bal, has := qAcnt.Balances[path[1]]
		if !has {
			// no balance for that ID create one
			bal = &utils.Balance{ID: path[1]}
			qAcnt.Balances[path[1]] = bal
		}
		if err = actSetBalance(bal, path[2:], dk.Value, reset); err != nil {
			return
		}
	}
	return dm.SetAccountProfile(qAcnt, false)
}

// actSetBalance will set the field at path from balance with value
// value is string as the value received from action is string
// the balance must not be nil
func actSetBalance(bal *utils.Balance, path []string, value string, reset bool) (err error) {
	// check if we have path past *balance
	if len(path) == 0 {
		return utils.ErrWrongPath
	}
	// select what field is update based on the first value from path
	// special case for CostIncrements and UnitFactors
	// that are converted from string similar to how are loaded from CSVs
	switch path[0] {
	case utils.ID:
		bal.ID = value
	case utils.FilterIDs:
		bal.FilterIDs = utils.NewStringSet(strings.Split(value, utils.InfieldSep)).AsSlice()
	case utils.Weights:
		bal.Weights, err = utils.NewDynamicWeightsFromString(value, utils.InfieldSep, utils.ANDSep)
	case utils.Type:
		bal.Type = value
	case utils.Units:
		var z *utils.Decimal
		if z, err = utils.NewDecimalFromString(value); err != nil {
			return
		}
		// do not overwrite the  Units if the action is *add_balance
		// this flag makes the difference between the *add_balance and *set_balance actions
		if !reset && bal.Units != nil {
			bal.Units.Add(bal.Units.Big, z.Big)
		} else {
			bal.Units = z
		}
	case utils.UnitFactors:
		// just recreate them from strinf
		bal.UnitFactors, err = actNewUnitFactorsFromString(value)
	case utils.Opts:
		err = utils.MapStorage(bal.Opts).Set(path[1:], value)
	case utils.CostIncrements:
		// just recreate them from strinf
		bal.CostIncrements, err = actNewCostIncrementsFromString(value)
	case utils.AttributeIDs:
		bal.AttributeIDs = strings.Split(value, utils.InfieldSep)
	case utils.RateProfileIDs:
		bal.RateProfileIDs = utils.NewStringSet(strings.Split(value, utils.InfieldSep)).AsSlice()
	default:
		// we modify the UnitFactors explicit
		// e.g. *balance.TEST.UnitFactors[0].Factor
		if strings.HasPrefix(path[0], utils.UnitFactors) {
			return actSetUnitFactor(bal.UnitFactors, path, value)
		}

		// we modify the CostIncrements explicit
		// e.g. *balance.TEST.CostIncrements[0].Increment
		if strings.HasPrefix(path[0], utils.CostIncrements) {
			return actSetCostIncrement(bal.CostIncrements, path, value)
		}
		// not a valid path
		err = utils.ErrWrongPath
	}
	return
}

// actNewUnitFactorsFromString converts a string to a list of UnitFactors
// similar to the how the TP are loaded from CSV
func actNewUnitFactorsFromString(value string) (units []*utils.UnitFactor, err error) {
	sls := strings.Split(value, utils.InfieldSep)
	if len(sls)%2 != 0 {
		return nil, fmt.Errorf("invlid key: <%s> for BalanceUnitFactors", value)
	}
	units = make([]*utils.UnitFactor, 0, len(sls)/2)

	for j := 0; j < len(sls); j += 2 {
		var z *utils.Decimal
		if z, err = utils.NewDecimalFromString(sls[j+1]); err != nil {
			return
		}
		units = append(units, &utils.UnitFactor{
			FilterIDs: strings.Split(sls[j], utils.ANDSep),
			Factor:    z,
		})
	}
	return
}

// actNewCostIncrementsFromString converts a string to a list of CostIncrements
// similar to the how the TP are loaded from CSV
func actNewCostIncrementsFromString(value string) (costs []*utils.CostIncrement, err error) {
	sls := strings.Split(value, utils.InfieldSep)
	if len(sls)%4 != 0 {
		return nil, fmt.Errorf("invlid key: <%s> for BalanceCostIncrements", value)
	}
	costs = make([]*utils.CostIncrement, 0, len(sls)/4)
	for j := 0; j < len(sls); j += 4 {
		cost := &utils.CostIncrement{
			FilterIDs: strings.Split(sls[j], utils.ANDSep),
		}
		if incrementStr := sls[j+1]; incrementStr != utils.EmptyString {
			if cost.Increment, err = utils.NewDecimalFromString(incrementStr); err != nil {
				return
			}
		}
		if fixedFeeStr := sls[j+2]; fixedFeeStr != utils.EmptyString {
			if cost.FixedFee, err = utils.NewDecimalFromString(fixedFeeStr); err != nil {
				return
			}
		}
		if recurrentFeeStr := sls[j+3]; recurrentFeeStr != utils.EmptyString {
			if cost.RecurrentFee, err = utils.NewDecimalFromString(recurrentFeeStr); err != nil {
				return
			}
		}
		costs = append(costs, cost)
	}
	return
}

// actSetUnitFactor will update the UnitFactors
func actSetUnitFactor(uFs []*utils.UnitFactor, path []string, value string) (err error) {
	pathVal := path[0][11:]
	lp := len(pathVal)
	// check path requierments
	// exact 2 elements
	// and the first element have an index between brackets
	if lp != 2 ||
		pathVal[0] != '[' ||
		pathVal[lp-1] != ']' {
		return utils.ErrWrongPath
	}
	pathVal = pathVal[1 : lp-1]
	var idx int
	// convert the index from string to int
	if idx, err = strconv.Atoi(pathVal); err != nil {
		return
	}
	if len(uFs) == idx { // special case add a new unitFactor
		uFs = append(uFs, &utils.UnitFactor{})
	} else if len(uFs) < idx { // make sure we are in slice range
		return utils.ErrWrongPath
	}

	switch path[1] {
	case utils.FilterIDs:
		uFs[idx].FilterIDs = utils.NewStringSet(strings.Split(value, utils.InfieldSep)).AsSlice()
	case utils.Factor:
		uFs[idx].Factor, err = utils.NewDecimalFromString(value)
	default:
		err = utils.ErrWrongPath
	}
	return
}

func actSetCostIncrement(cIs []*utils.CostIncrement, path []string, value string) (err error) {
	pathVal := path[0][14:]
	lp := len(pathVal)
	// check path requierments
	// exact 2 elements
	// and the first element have an index between brackets
	if lp != 2 ||
		pathVal[0] != '[' ||
		pathVal[lp-1] != ']' {
		return utils.ErrWrongPath
	}
	pathVal = pathVal[1 : lp-1]
	var idx int
	// convert the index from string to int
	if idx, err = strconv.Atoi(pathVal); err != nil {
		return
	}
	if len(cIs) == idx { // special case add a new CostIncrement
		cIs = append(cIs, &utils.CostIncrement{})
	} else if len(cIs) < idx { // make sure we are in slice range
		return utils.ErrWrongPath
	}

	switch path[0] {
	case utils.FilterIDs:
		cIs[idx].FilterIDs = utils.NewStringSet(strings.Split(value, utils.InfieldSep)).AsSlice()
	case utils.Increment:
		cIs[idx].Increment, err = utils.NewDecimalFromString(value)
	case utils.FixedFee:
		cIs[idx].FixedFee, err = utils.NewDecimalFromString(value)
	case utils.RecurrentFee:
		cIs[idx].RecurrentFee, err = utils.NewDecimalFromString(value)
	default:
		err = utils.ErrWrongPath
	}
	return
}
