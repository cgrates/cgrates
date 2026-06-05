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

package thresholds

import (
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

// V1ProcessEvent implements ThresholdService method for processing an Event
func (s *ThresholdS) V1ProcessEvent(ctx *context.Context, args *utils.CGREvent, reply *[]string) error {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	tnt := args.Tenant
	if tnt == "" {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}
	ids, err := s.processEvent(ctx, tnt, args)
	if err != nil {
		return err
	}
	*reply = ids
	return nil
}

// V1GetThresholdsForEvent queries thresholds matching an Event
func (s *ThresholdS) V1GetThresholdsForEvent(ctx *context.Context, args *utils.CGREvent, reply *[]*utils.Threshold) error {
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
	matched, unlock, err := s.matchingThresholdsForEvent(ctx, tnt, args)
	if err != nil {
		return err
	}
	defer unlock()
	out := make([]*utils.Threshold, 0, len(matched))
	for _, mt := range matched {
		out = append(out, mt.threshold)
	}
	*reply = out
	return nil
}

// V1GetThresholdIDs returns list of thresholdIDs configured for a tenant
func (s *ThresholdS) V1GetThresholdIDs(ctx *context.Context, args *utils.TenantWithAPIOpts, tIDs *[]string) error {
	tenant := args.Tenant
	if tenant == utils.EmptyString {
		tenant = s.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.ThresholdPrefix + tenant + utils.ConcatenatedKeySep
	db, _, err := s.dm.DBConns().GetConn(utils.MetaThresholds)
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
	*tIDs = retIDs
	return nil
}

// V1GetThreshold retrieves a Threshold
func (s *ThresholdS) V1GetThreshold(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, th *utils.Threshold) error {
	tnt := tntID.Tenant
	if tnt == "" {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}
	// make sure threshold is locked at process level
	lockID := guardian.Guardian.GuardIDs("",
		s.cfg.GeneralCfg().LockingTimeout,
		utils.ThresholdLockKey(tnt, tntID.ID))
	defer guardian.Guardian.UnguardIDs(lockID)
	thd, err := s.dm.GetThreshold(ctx, tnt, tntID.ID, true, true, "")
	if err != nil {
		return err
	}
	*th = *thd
	return nil
}

// V1ResetThreshold resets the threshold hits
func (s *ThresholdS) V1ResetThreshold(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, rply *string) error {
	tnt := tntID.Tenant
	if tnt == "" {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}
	// make sure threshold is locked at process level
	lockID := guardian.Guardian.GuardIDs("",
		s.cfg.GeneralCfg().LockingTimeout,
		utils.ThresholdLockKey(tnt, tntID.ID))
	defer guardian.Guardian.UnguardIDs(lockID)
	thd, err := s.dm.GetThreshold(ctx, tnt, tntID.ID, true, true, "")
	if err != nil {
		return err
	}
	if thd.Hits != 0 {
		thd.Hits = 0
		thd.Snooze = time.Time{}
		if s.cfg.ThresholdSCfg().StoreInterval == -1 {
			if err := s.StoreThreshold(ctx, thd); err != nil {
				return err
			}
		} else {
			s.storedMu.Lock()
			s.storedThresholds.Add(thd.TenantID())
			s.storedMu.Unlock()
		}
	}
	*rply = utils.OK
	return nil
}
