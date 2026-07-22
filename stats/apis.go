// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package stats

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// V1ProcessEvent implements StatV1 method for processing an Event
func (s *StatS) V1ProcessEvent(ctx *context.Context, args *utils.CGREvent, reply *[]string) error {
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
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}
	ids, err := s.processEvent(ctx, tnt, args)
	if err != nil {
		return err
	}
	*reply = ids
	return nil
}

// V1GetStatQueuesForEvent implements StatV1 method for processing an Event
func (s *StatS) V1GetStatQueuesForEvent(ctx *context.Context, args *utils.CGREvent, reply *[]string) error {
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
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}
	sQs, unlock, err := s.matchingStatQueuesForEvent(ctx, tnt, args)
	if err != nil {
		return err
	}
	defer unlock()
	*reply = getStatQueueIDs(sQs)
	return nil
}

// V1GetStatQueue returns a StatQueue object
func (s *StatS) V1GetStatQueue(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *utils.StatQueue) error {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == "" {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}
	// make sure statQueue is locked at process level
	unlock := s.dm.Lock(utils.StatQueueLockKey(tnt, args.ID))
	defer unlock()
	sq, err := s.getStatQueue(ctx, tnt, args.ID)
	if err != nil {
		return err
	}
	*reply = *sq.Clone() // clone so the reply is marshaled safely after the lock is released
	return nil
}

// V1GetQueueStringMetrics returns the metrics of a Queue as string values
func (s *StatS) V1GetQueueStringMetrics(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *map[string]string) error {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == "" {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}
	// make sure statQueue is locked at process level
	unlock := s.dm.Lock(utils.StatQueueLockKey(tnt, args.ID))
	defer unlock()
	sq, err := s.getStatQueue(ctx, tnt, args.ID)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	var rnd int
	if rnd, err = engine.GetIntOpts(ctx, tnt, engine.MapEvent{utils.Tenant: tnt, "*opts": map[string]any{}}, nil, s.filters,
		s.cfg.StatSCfg().Opts.RoundingDecimals,
		utils.OptsRoundingDecimals); err != nil {
		return err
	}
	metrics := make(map[string]string, len(sq.SQMetrics))
	for metricID, metric := range sq.SQMetrics {
		metrics[metricID] = metric.GetStringValue(rnd)
	}
	*reply = metrics
	return nil
}

// V1GetQueueFloatMetrics returns the metrics as float64 values
func (s *StatS) V1GetQueueFloatMetrics(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *map[string]float64) error {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == "" {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}
	// make sure statQueue is locked at process level
	unlock := s.dm.Lock(utils.StatQueueLockKey(tnt, args.ID))
	defer unlock()
	sq, err := s.getStatQueue(ctx, tnt, args.ID)
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
	return nil
}

// V1GetQueueDecimalMetrics returns the metrics as decimal values
func (s *StatS) V1GetQueueDecimalMetrics(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *map[string]*utils.Decimal) error {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == "" {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}
	// make sure statQueue is locked at process level
	unlock := s.dm.Lock(utils.StatQueueLockKey(tnt, args.ID))
	defer unlock()
	sq, err := s.getStatQueue(ctx, tnt, args.ID)
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
	return nil
}

// V1GetQueueIDs returns list of queueIDs registered for a tenant
func (s *StatS) V1GetQueueIDs(ctx *context.Context, args *utils.TenantWithAPIOpts, qIDs *[]string) error {
	tenant := args.Tenant
	if tenant == utils.EmptyString {
		tenant = s.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.StatQueuePrefix + tenant + utils.ConcatenatedKeySep
	db, _, err := s.dm.DBConns().GetConn(utils.MetaStatQueues)
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
	return nil
}

// V1ResetStatQueue resets the stat queue
func (s *StatS) V1ResetStatQueue(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, rply *string) error {
	if missing := utils.MissingStructFields(tntID, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := tntID.Tenant
	if tnt == "" {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}
	// make sure statQueue is locked at process level
	unlock := s.dm.Lock(utils.StatQueueLockKey(tnt, tntID.ID))
	defer unlock()
	sq, err := s.dm.GetStatQueue(ctx, tnt, tntID.ID,
		true, true, utils.NonTransactional)
	if err != nil {
		return err
	}
	sq.SQItems = make([]utils.SQItem, 0)
	metrics := sq.SQMetrics
	sq.SQMetrics = make(map[string]utils.StatMetric)
	for id, m := range metrics {
		metric, err := utils.NewStatMetric(id,
			m.GetMinItems(), m.GetFilterIDs())
		if err != nil {
			return err
		}
		sq.SQMetrics[id] = metric
	}
	if s.cfg.StatSCfg().StoreInterval != 0 {
		if s.cfg.StatSCfg().StoreInterval == -1 {
			s.storeStatQueue(ctx, sq)
		} else {
			s.storedMu.Lock()
			s.storedStatQueues.Add(sq.TenantID())
			s.storedMu.Unlock()
		}
	}
	*rply = utils.OK
	return nil
}
