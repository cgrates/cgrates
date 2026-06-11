/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package stats

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

// V1ProcessEvent implements StatV1 method for processing an Event
func (sS *StatS) V1ProcessEvent(ctx *context.Context, args *utils.CGREvent, reply *[]string) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cfg.GeneralCfg().DefaultTenant
	}
	var ids []string
	if ids, err = sS.processEvent(ctx, tnt, args); err != nil {
		return
	}
	*reply = ids
	return
}

// V1GetStatQueuesForEvent implements StatV1 method for processing an Event
func (sS *StatS) V1GetStatQueuesForEvent(ctx *context.Context, args *utils.CGREvent, reply *[]string) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cfg.GeneralCfg().DefaultTenant
	}
	sQs, unlock, err := sS.matchingStatQueuesForEvent(ctx, tnt, args)
	if err != nil {
		return err
	}
	defer unlock()
	*reply = getStatQueueIDs(sQs)
	return
}

// V1GetStatQueue returns a StatQueue object
func (sS *StatS) V1GetStatQueue(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *utils.StatQueue) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cfg.GeneralCfg().DefaultTenant
	}
	// make sure statQueue is locked at process level
	lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout,
		utils.StatQueueLockKey(tnt, args.ID))
	defer guardian.Guardian.UnguardIDs(lkID)
	sq, err := sS.getStatQueue(ctx, tnt, args.ID)
	if err != nil {
		return err
	}
	*reply = *sq.Clone() // clone so the reply is marshaled safely after the lock is released
	return
}

// V1GetQueueStringMetrics returns the metrics of a Queue as string values
func (sS *StatS) V1GetQueueStringMetrics(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *map[string]string) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cfg.GeneralCfg().DefaultTenant
	}
	// make sure statQueue is locked at process level
	lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout,
		utils.StatQueueLockKey(tnt, args.ID))
	defer guardian.Guardian.UnguardIDs(lkID)
	sq, err := sS.getStatQueue(ctx, tnt, args.ID)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	var rnd int
	if rnd, err = engine.GetIntOpts(ctx, tnt, engine.MapEvent{utils.Tenant: tnt, "*opts": map[string]any{}}, nil, sS.fltrS,
		sS.cfg.StatSCfg().Opts.RoundingDecimals,
		utils.OptsRoundingDecimals); err != nil {
		return
	}
	metrics := make(map[string]string, len(sq.SQMetrics))
	for metricID, metric := range sq.SQMetrics {
		metrics[metricID] = metric.GetStringValue(rnd)
	}
	*reply = metrics
	return
}

// V1GetQueueFloatMetrics returns the metrics as float64 values
func (sS *StatS) V1GetQueueFloatMetrics(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *map[string]float64) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cfg.GeneralCfg().DefaultTenant
	}
	// make sure statQueue is locked at process level
	lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout,
		utils.StatQueueLockKey(tnt, args.ID))
	defer guardian.Guardian.UnguardIDs(lkID)
	sq, err := sS.getStatQueue(ctx, tnt, args.ID)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	metrics := make(map[string]float64, len(sq.SQMetrics))
	for metricID, metric := range sq.SQMetrics {
		val := metric.GetValue()
		metrics[metricID] = -1
		if val != utils.DecimalNaN {
			metrics[metricID], _ = val.Float64()
		}
	}
	*reply = metrics
	return
}

// V1GetQueueDecimalMetrics returns the metrics as decimal values
func (sS *StatS) V1GetQueueDecimalMetrics(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *map[string]*utils.Decimal) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cfg.GeneralCfg().DefaultTenant
	}
	// make sure statQueue is locked at process level
	lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout,
		utils.StatQueueLockKey(tnt, args.ID))
	defer guardian.Guardian.UnguardIDs(lkID)
	sq, err := sS.getStatQueue(ctx, tnt, args.ID)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	metrics := make(map[string]*utils.Decimal, len(sq.SQMetrics))
	for metricID, metric := range sq.SQMetrics {
		metrics[metricID] = metric.GetValue()
	}
	*reply = metrics
	return
}

// V1GetQueueIDs returns list of queueIDs registered for a tenant
func (sS *StatS) V1GetQueueIDs(ctx *context.Context, args *utils.TenantWithAPIOpts, qIDs *[]string) (err error) {
	tenant := args.Tenant
	if tenant == utils.EmptyString {
		tenant = sS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.StatQueuePrefix + tenant + utils.ConcatenatedKeySep
	db, _, err := sS.dm.DBConns().GetConn(utils.MetaStatQueues)
	if err != nil {
		return err
	}
	keys, err := db.GetKeysForPrefix(ctx, prfx, utils.EmptyString)
	if err != nil {
		return err
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[len(prfx):]
	}
	*qIDs = retIDs
	return
}

// V1ResetStatQueue resets the stat queue
func (sS *StatS) V1ResetStatQueue(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, rply *string) (err error) {
	if missing := utils.MissingStructFields(tntID, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := tntID.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cfg.GeneralCfg().DefaultTenant
	}
	// make sure statQueue is locked at process level
	lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout,
		utils.StatQueueLockKey(tnt, tntID.ID))
	defer guardian.Guardian.UnguardIDs(lkID)
	var sq *utils.StatQueue
	if sq, err = sS.dm.GetStatQueue(ctx, tnt, tntID.ID,
		true, true, utils.NonTransactional); err != nil {
		return
	}
	sq.SQItems = make([]utils.SQItem, 0)
	metrics := sq.SQMetrics
	sq.SQMetrics = make(map[string]utils.StatMetric)
	for id, m := range metrics {
		var metric utils.StatMetric
		if metric, err = utils.NewStatMetric(id,
			m.GetMinItems(), m.GetFilterIDs()); err != nil {
			return
		}
		sq.SQMetrics[id] = metric
	}
	if sS.cfg.StatSCfg().StoreInterval != 0 {
		if sS.cfg.StatSCfg().StoreInterval == -1 {
			sS.StoreStatQueue(ctx, sq)
		} else {
			sS.ssqMux.Lock()
			sS.storedStatQueues.Add(sq.TenantID())
			sS.ssqMux.Unlock()
		}
	}
	*rply = utils.OK
	return
}
