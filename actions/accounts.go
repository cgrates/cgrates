// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package actions

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// actSetBalance will update the account
type actSetBalance struct {
	config  *config.CGRConfig
	connMgr *engine.ConnManager
	fltrS   *engine.FilterS
	aCfg    *utils.APAction
	tnt     string
	reset   bool
}

func (aL *actSetBalance) id() string {
	return aL.aCfg.ID
}

func (aL *actSetBalance) cfg() *utils.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actSetBalance) execute(ctx *context.Context, data utils.MapStorage, trgID string) (err error) {
	var conns []string
	conns, err = engine.GetConnIDs(ctx, aL.config.ActionSCfg().Conns, utils.MetaAccounts, aL.tnt, data, nil, aL.fltrS)
	if err != nil {
		return
	}
	if len(conns) == 0 {
		return fmt.Errorf("no connection with AccountS")
	}

	weights := make(map[string]float64)   // stores sorting weights by Diktat ID
	diktats := make([]*utils.APDiktat, 0) // list of diktats which have *balancePath in opts, will be weight sorted later
	for _, diktat := range aL.cfg().Diktats {
		if pass, err := aL.fltrS.Pass(ctx, aL.tnt, diktat.FilterIDs, data); err != nil {
			return err
		} else if !pass {
			continue
		}
		weight, err := engine.WeightFromDynamics(ctx, diktat.Weights, aL.fltrS, aL.tnt, data)
		if err != nil {
			return err
		}
		weights[diktat.ID] = weight
		diktats = append(diktats, diktat)
	}
	// Sort by weight (higher values first).
	slices.SortFunc(diktats, func(a, b *utils.APDiktat) int {
		return cmp.Compare(weights[b.ID], weights[a.ID])
	})
	args := &utils.ArgsActSetBalance{
		Tenant:    aL.tnt,
		AccountID: trgID,
		Reset:     aL.reset,
		Diktats:   make([]*utils.BalDiktat, 0),
		APIOpts:   aL.cfg().Opts,
	}
	for _, actD := range diktats {
		var rsr utils.RSRParsers
		if rsr, err = actD.RSRValues(); err != nil {
			return
		}
		var val string
		if val, err = rsr.ParseDataProvider(data); err != nil {
			return
		}
		args.Diktats = append(args.Diktats, &utils.BalDiktat{
			Path:  utils.IfaceAsString(actD.Opts[utils.MetaBalancePath]),
			Value: val,
		})
		if blocker, err := engine.BlockerFromDynamics(ctx, actD.Blockers, aL.fltrS, aL.tnt, data); err != nil {
			return err
		} else if blocker {
			break
		}
	}
	var rply string
	return aL.connMgr.Call(ctx, conns,
		utils.AccountSv1ActionSetBalance, args, &rply)
}

// actRemBalance will remove multiple balances from account
type actRemBalance struct {
	config  *config.CGRConfig
	connMgr *engine.ConnManager
	fltrS   *engine.FilterS
	aCfg    *utils.APAction
	tnt     string
}

func (aL *actRemBalance) id() string {
	return aL.aCfg.ID
}

func (aL *actRemBalance) cfg() *utils.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actRemBalance) execute(ctx *context.Context, data utils.MapStorage, trgID string) (err error) {
	var conns []string
	conns, err = engine.GetConnIDs(ctx, aL.config.ActionSCfg().Conns, utils.MetaAccounts, aL.tnt, data, nil, aL.fltrS)
	if len(conns) == 0 {
		return fmt.Errorf("no connection with AccountS")
	}

	weights := make(map[string]float64)   // stores sorting weights by Diktat ID
	diktats := make([]*utils.APDiktat, 0) // list of diktats which have *balancePath in opts, will be weight sorted later
	for _, diktat := range aL.cfg().Diktats {
		if pass, err := aL.fltrS.Pass(ctx, aL.tnt, diktat.FilterIDs, data); err != nil {
			return err
		} else if !pass {
			continue
		}
		weight, err := engine.WeightFromDynamics(ctx, diktat.Weights, aL.fltrS, aL.tnt, data)
		if err != nil {
			return err
		}
		weights[diktat.ID] = weight
		diktats = append(diktats, diktat)
	}
	// Sort by weight (higher values first).
	slices.SortFunc(diktats, func(a, b *utils.APDiktat) int {
		return cmp.Compare(weights[b.ID], weights[a.ID])
	})
	args := &utils.ArgsActRemoveBalances{
		Tenant:     aL.tnt,
		AccountID:  trgID,
		BalanceIDs: make([]string, len(diktats)),
		APIOpts:    aL.cfg().Opts,
	}
	for i, actD := range diktats {
		args.BalanceIDs[i] = utils.IfaceAsString(actD.Opts[utils.MetaBalancePath])
		if blocker, err := engine.BlockerFromDynamics(ctx, actD.Blockers, aL.fltrS, aL.tnt, data); err != nil {
			return err
		} else if blocker {
			break
		}
	}
	var rply string
	return aL.connMgr.Call(ctx, conns,
		utils.AccountSv1ActionRemoveBalance, args, &rply)
}
