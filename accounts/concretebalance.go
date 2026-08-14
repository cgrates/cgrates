// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package accounts

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// cloneUnitsFromConcretes returns cloned units from the concrete balances passed as parameters
func cloneUnitsFromConcretes(cBs []*concreteBalance) (clnedUnts []*utils.Decimal) {
	if cBs == nil {
		return
	}
	clnedUnts = make([]*utils.Decimal, len(cBs))
	for i := range cBs {
		clnedUnts[i] = cBs[i].blnCfg.Units.Clone()
	}
	return
}

// restoreUnitsFromClones will restore the units from the clones
func restoreUnitsFromClones(cBs []*concreteBalance, clnedUnts []*utils.Decimal) {
	for i, clnedUnt := range clnedUnts {
		cBs[i].blnCfg.Units = clnedUnt
	}
}

// newConcreteBalance constructs a concreteBalanceOperator
func newConcreteBalanceOperator(ctx *context.Context, cfg *config.CGRConfig, acntID string, blnCfg *utils.Balance,
	fltrS *engine.FilterS, connMgr *engine.ConnManager,
	attrSConns, rateSConns []string) balanceOperator {
	return &concreteBalance{cfg, acntID, blnCfg, fltrS, connMgr, ctx, attrSConns, rateSConns}
}

// concreteBalance is the operator for *concrete balance type
type concreteBalance struct {
	cfg        *config.CGRConfig
	acntID     string
	blnCfg     *utils.Balance
	fltrS      *engine.FilterS
	connMgr    *engine.ConnManager
	ctx        *context.Context
	attrSConns []string
	rateSConns []string
}

// id implements the balanceOperator interface
func (cB *concreteBalance) id() string {
	return cB.blnCfg.ID
}

// debitAbstracts implements the balanceOperator interface
// it will mainly debit the aUnits out of a single concrete balance
// the abstract
func (cB *concreteBalance) debitAbstracts(ctx *context.Context, aUnits *utils.Decimal,
	cgrEv *utils.CGREvent, dbted *utils.Decimal) (ec *utils.EventCharges, err error) {
	evNm := cgrEv.AsDataProvider()
	// pass the general balance filters
	var pass bool
	if pass, err = cB.fltrS.Pass(ctx, cgrEv.Tenant, cB.blnCfg.FilterIDs, evNm); err != nil {
		return
	} else if !pass {
		return nil, utils.ErrFilterNotPassingNoCaps
	}
	// costIncrement
	var costIcrm *utils.CostIncrement
	if costIcrm, err = costIncrement(ctx, cB.blnCfg.CostIncrements,
		cB.fltrS, cgrEv.Tenant, evNm); err != nil {
		return
	}
	var ecCncrt *utils.EventCharges
	if ecCncrt, err = maxDebitAbstractsFromConcretes(ctx, aUnits,
		cB.acntID, []*concreteBalance{cB},
		cB.connMgr, cgrEv,
		cB.attrSConns, cB.blnCfg.AttributeIDs,
		cB.rateSConns, cB.blnCfg.RateProfileIDs,
		costIcrm, dbted, cB.cfg.AccountSCfg().MaxIterations); err != nil {
		return
	} else if ecCncrt.Abstracts.Compare(utils.NewDecimal(0, 0)) == 0 { // no debit performed
		return
	}
	ec = utils.NewEventCharges()
	ec.Abstracts = ecCncrt.Abstracts
	ec.Concretes = ecCncrt.Concretes
	// RatingID
	var ratingID, rateID string
	if costIcrm != nil {
		ratingID = utils.UUIDSha1Prefix()
		rateID = utils.UUIDSha1Prefix()
		ec.Rating[ratingID] = &utils.RateSInterval{
			Increments: []*utils.RateSIncrement{
				{
					RateID:         rateID,
					CompressFactor: 1,
				},
			},
			CompressFactor: 1,
		}
		ec.Rates[rateID] = &utils.IntervalRate{
			FixedFee:     costIcrm.FixedFee,
			RecurrentFee: costIcrm.RecurrentFee,
		}
	} else { // take it from first increment
		ratingID = ecCncrt.Accounting[ecCncrt.Charges[0].ChargingID].RatingID
		ec.Rating[ratingID] = ecCncrt.Rating[ratingID]
		for _, incr := range ecCncrt.Rating[ratingID].Increments {
			ec.Rates[incr.RateID] = ecCncrt.Rates[incr.RateID]
		}
	}
	// AccountingID
	acntID := utils.UUIDSha1Prefix()
	ec.Accounting[acntID] = &utils.AccountCharge{
		AccountID: cB.acntID,
		BalanceID: utils.MetaMockAbstract,
		Units:     ec.Abstracts,
		RatingID:  ratingID,
	}
	for _, cE := range ecCncrt.Charges {
		if ecCncrt.Accounting[cE.ChargingID].UnitFactorID != utils.EmptyString {
			ec.UnitFactors[ecCncrt.Accounting[cE.ChargingID].UnitFactorID] = ecCncrt.UnitFactors[ecCncrt.Accounting[cE.ChargingID].UnitFactorID]
		}
		ec.Accounting[cE.ChargingID] = ecCncrt.Accounting[cE.ChargingID]
		ec.Accounting[cE.ChargingID].RatingID = utils.EmptyString // should not be needed since we have used it above
		ec.Accounting[acntID].JoinedChargeIDs = append(ec.Accounting[acntID].JoinedChargeIDs, cE.ChargingID)
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
// it will attempt to debit the amount of concrete units out of this single concrete balance
func (cB *concreteBalance) debitConcretes(ctx *context.Context, cUnits *utils.Decimal,
	cgrEv *utils.CGREvent, debited *utils.Decimal) (ec *utils.EventCharges, err error) {
	evNm := cgrEv.AsDataProvider()
	// pass the general balance filters
	var pass bool
	if pass, err = cB.fltrS.Pass(ctx, cgrEv.Tenant, cB.blnCfg.FilterIDs, evNm); err != nil {
		return
	} else if !pass {
		return nil, utils.ErrFilterNotPassingNoCaps
	}

	// unitFactor
	var uF *utils.UnitFactor
	if uF, err = unitFactor(ctx, cB.blnCfg.UnitFactors, cB.fltrS, cgrEv.Tenant, evNm); err != nil {
		return
	}
	var hasUF bool
	if uF != nil && uF.Factor.Compare(utils.NewDecimal(1, 0)) != 0 {
		hasUF = true
		cUnits = utils.MultiplyDecimal(cUnits, uF.Factor)
	}

	// balanceLimit
	var hasLmt bool
	var blncLmt *utils.Decimal
	if blncLmt, err = balanceLimit(cB.blnCfg.Opts); err != nil {
		return
	}
	if blncLmt != nil && blncLmt.Compare(utils.NewDecimal(0, 0)) != 0 {
		cB.blnCfg.Units = utils.SubstractDecimal(cB.blnCfg.Units, blncLmt)
		hasLmt = true
	}
	var dbted *utils.Decimal
	if cB.blnCfg.Units.Compare(cUnits) <= 0 && blncLmt != nil { // balance smaller than debit and limited
		dbted = cB.blnCfg.Units
		cB.blnCfg.Units = blncLmt
	} else {
		cB.blnCfg.Units = utils.SubstractDecimal(cB.blnCfg.Units, cUnits)
		if hasLmt { // put back the limit
			cB.blnCfg.Units = utils.SumDecimal(cB.blnCfg.Units, blncLmt)
		}
		dbted = cUnits
	}
	if hasUF {
		dbted = utils.DivideDecimal(dbted, uF.Factor)
	}
	if dbted.Compare(utils.NewDecimal(0, 0)) == 0 {
		return // no event cost for 0 debit
	}
	// EventCharges
	ec = utils.NewEventCharges()
	ec.Concretes = dbted
	// UnitFactors
	var ufID string
	if hasUF {
		ufID = utils.UUIDSha1Prefix()
		ec.UnitFactors[ufID] = uF
	}
	acntID := utils.UUIDSha1Prefix()
	ec.Accounting[acntID] = &utils.AccountCharge{
		AccountID:    cB.acntID,
		BalanceID:    cB.blnCfg.ID,
		Units:        dbted,
		BalanceLimit: blncLmt,
		UnitFactorID: ufID,
	}
	ec.Charges = []*utils.ChargeEntry{
		{
			ChargingID:     acntID,
			CompressFactor: 1,
		},
	}

	return
}

// getBalanceCfg will return the balance
func (cB *concreteBalance) balanceCfg() *utils.Balance {
	return cB.blnCfg
}
