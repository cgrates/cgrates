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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

// newAbstractBalance constructs an abstractBalanceOperator
func newAbstractBalanceOperator(acntID string, blnCfg *utils.Balance,
	cncrtBlncs []*concreteBalance,
	fltrS *engine.FilterS, connMgr *engine.ConnManager,
	attrSConns, rateSConns []string) balanceOperator {
	return &abstractBalance{acntID, blnCfg, cncrtBlncs, fltrS, connMgr, attrSConns, rateSConns}
}

// abstractBalance is the operator for *abstract balance type
type abstractBalance struct {
	acntID     string
	blnCfg     *utils.Balance
	cncrtBlncs []*concreteBalance // paying balances
	fltrS      *engine.FilterS
	connMgr    *engine.ConnManager
	attrSConns []string
	rateSConns []string
}

// debitAbstracts implements the balanceOperator interface
func (aB *abstractBalance) debitAbstracts(usage *decimal.Big,
	cgrEv *utils.CGREvent) (ec *utils.EventCharges, err error) {

	evNm := utils.MapStorage{
		utils.MetaOpts: cgrEv.APIOpts,
		utils.MetaReq:  cgrEv.Event,
	}

	// pass the general balance filters
	var pass bool
	if pass, err = aB.fltrS.Pass(context.TODO(), cgrEv.Tenant, aB.blnCfg.FilterIDs, evNm); err != nil {
		return
	} else if !pass {
		return nil, utils.ErrFilterNotPassingNoCaps
	}

	dbtUnits := new(decimal.Big).Copy(usage)

	// balanceLimit
	var hasLmt bool
	var blncLmt *utils.Decimal
	if blncLmt, err = balanceLimit(aB.blnCfg.Opts); err != nil {
		return
	}
	if blncLmt != nil && blncLmt.Cmp(decimal.New(0, 0)) != 0 {
		aB.blnCfg.Units.Big = utils.SubstractBig(aB.blnCfg.Units.Big, blncLmt.Big)
		hasLmt = true
	}

	// costIncrement
	var costIcrm *utils.CostIncrement
	if costIcrm, err = costIncrement(aB.blnCfg.CostIncrements, aB.fltrS,
		cgrEv.Tenant, evNm); err != nil {
		return
	}
	// unitFactor
	var uF *utils.UnitFactor
	if uF, err = unitFactor(aB.blnCfg.UnitFactors, aB.fltrS, cgrEv.Tenant, evNm); err != nil {
		return
	}
	var hasUF bool
	if uF != nil && uF.Factor.Cmp(decimal.New(1, 0)) != 0 {
		dbtUnits = utils.MultiplyBig(dbtUnits, uF.Factor.Big)
		hasUF = true
	}
	// balance smaller than usage, correct usage if the balance has limit
	if aB.blnCfg.Units.Big.Cmp(dbtUnits) == -1 && blncLmt != nil {
		// decrease the usage to match the maximum increments
		// will use special rounding to 0 since otherwise we go negative (ie: 0.05 as increment)
		maxDbt := new(decimal.Big).Copy(aB.blnCfg.Units.Big)
		if hasUF {
			maxDbt = utils.DivideBig(maxDbt, uF.Factor.Big)
		}
		usage = roundUnitsWithIncrements(maxDbt, costIcrm.Increment.Big)
	}

	/*
		if costIcrm.RecurrentFee.Cmp(decimal.New(0, 0)) == 0 &&
			(costIcrm.FixedFee == nil ||
				costIcrm.FixedFee.Cmp(decimal.New(0, 0)) == 0) {

			acntID := utils.UUIDSha1Prefix()
			ec.Accounting[acntID] = &utils.AccountCharge{
				AccountID:    aB.acntID,
				BalanceID:    aB.blnCfg.ID,
				Units:        &utils.Decimal{usage},
				BalanceLimit: blncLmt,
				UnitFactorID: ufID,
				RatingID:     ratingID,
			}
			ec.ChargingIntervals = []*utils.ChargingInterval{
				{
					Increments: []*utils.ChargingIncrement{
						{
							Units:           &utils.Decimal{usage},
							AccountChargeID: acntID,
							CompressFactor:  1,
						},
					},
					CompressFactor: 1,
				},
			}

	*/

	var ecCost *utils.EventCharges
	if (costIcrm.FixedFee != nil &&
		costIcrm.FixedFee.Cmp(decimal.New(0, 0)) != 0) ||
		(costIcrm.RecurrentFee != nil &&
			costIcrm.RecurrentFee.Cmp(decimal.New(0, 0)) != 0) {

		// attempt to debit usage with cost
		if ecCost, err = maxDebitAbstractsFromConcretes(usage,
			aB.acntID, aB.cncrtBlncs,
			aB.connMgr, cgrEv,
			aB.attrSConns, aB.blnCfg.AttributeIDs,
			aB.rateSConns, aB.blnCfg.RateProfileIDs,
			costIcrm); err != nil {
			return
		}

	}

	if ecCost != nil {
		usage = ecCost.Abstracts.Big
		dbtUnits = ecCost.Abstracts.Big
		if hasUF {
			dbtUnits = utils.MultiplyBig(dbtUnits, uF.Factor.Big)
		}
	}
	if dbtUnits.Cmp(decimal.New(0, 0)) != 0 {
		aB.blnCfg.Units.Big = utils.SubstractBig(aB.blnCfg.Units.Big, dbtUnits)
	}
	if hasLmt { // put back the limit
		aB.blnCfg.Units.Big = utils.SumBig(aB.blnCfg.Units.Big, blncLmt.Big)
	}

	// EvenCharges building
	ec = utils.NewEventCharges()
	ec.Abstracts = &utils.Decimal{usage}
	if ecCost != nil {
		ec.Concretes = ecCost.Concretes
	}
	// UnitFactors
	var ufID string
	if hasUF {
		ufID = utils.UUIDSha1Prefix()
		ec.UnitFactors[ufID] = uF
	}
	// RatingID
	var ratingID string
	if costIcrm != nil {
		ratingID = utils.UUIDSha1Prefix()
		ec.Rating[ratingID] = &utils.RateSInterval{
			Increments: []*utils.RateSIncrement{
				{
					Rate: &utils.Rate{
						ID: utils.MetaCostIncrement,
						IntervalRates: []*utils.IntervalRate{
							{
								FixedFee:     costIcrm.FixedFee,
								RecurrentFee: costIcrm.RecurrentFee,
							},
						},
					},
					CompressFactor: 1,
				},
			},
			CompressFactor: 1,
		}
	} else { // take it from first increment, not copying since it will be done bellow
		ratingID = ecCost.Accounting[ecCost.Charges[0].ChargingID].RatingID
	}
	// AccountingID
	acntID := utils.UUIDSha1Prefix()
	ec.Accounting[acntID] = &utils.AccountCharge{
		AccountID:    aB.acntID,
		BalanceID:    aB.blnCfg.ID,
		BalanceLimit: blncLmt,
		UnitFactorID: ufID,
		RatingID:     ratingID,
	}
	if ecCost != nil {
		for _, ival := range ecCost.Charges {
			ec.Accounting[acntID].JoinedChargeIDs = append(ec.Accounting[acntID].JoinedChargeIDs, ival.ChargingID)
			ec.Accounting[ival.ChargingID] = ecCost.Accounting[ival.ChargingID]
			// Copy the unitFactor data
			if ecCost.Accounting[ival.ChargingID].UnitFactorID != utils.EmptyString {
				ec.UnitFactors[ecCost.Accounting[ival.ChargingID].UnitFactorID] = ecCost.UnitFactors[ecCost.Accounting[ival.ChargingID].UnitFactorID]
			}
			// Copy the Rating data
			if ecCost.Accounting[ival.ChargingID].RatingID != utils.EmptyString {
				ec.Rating[ecCost.Accounting[ival.ChargingID].RatingID] = ecCost.Rating[ecCost.Accounting[ival.ChargingID].RatingID]
			}
		}
	}
	ec.Charges = []*utils.ChargeEntry{
		{
			ChargingID:     acntID,
			CompressFactor: 1,
		},
	}
	return
}

// debitConcretes implements the balanceOperator interface
func (aB *abstractBalance) debitConcretes(_ *decimal.Big,
	_ *utils.CGREvent) (ec *utils.EventCharges, err error) {
	return nil, utils.ErrNotImplemented
}
