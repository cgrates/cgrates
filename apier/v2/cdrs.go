// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v2

import (
	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Retrieves CDRs based on the filters
func (apier *APIerSv2) GetCDRs(ctx *context.Context, attrs *utils.RPCCDRsFilter, reply *[]*engine.ExternalCDR) error {
	cdrsFltr, err := attrs.AsCDRsFilter(apier.Config.GeneralCfg().DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if cdrs, _, err := apier.CdrDb.GetCDRs(cdrsFltr, false); err != nil {
		if err.Error() != utils.NotFoundCaps {
			err = utils.NewErrServerError(err)
		}
		return err
	} else if len(cdrs) == 0 {
		*reply = make([]*engine.ExternalCDR, 0)
	} else {
		for _, cdr := range cdrs {
			*reply = append(*reply, cdr.AsExternalCDR())
		}
	}
	return nil
}

func (apier *APIerSv2) CountCDRs(ctx *context.Context, attrs *utils.RPCCDRsFilter, reply *int64) error {
	cdrsFltr, err := attrs.AsCDRsFilter(apier.Config.GeneralCfg().DefaultTimezone)
	if err != nil {
		if err.Error() != utils.NotFoundCaps {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	cdrsFltr.Count = true
	if _, count, err := apier.CdrDb.GetCDRs(cdrsFltr, false); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = count
	}
	return nil
}

// Receive CDRs via RPC methods, not included with APIer because it has way less dependencies and can be standalone
type CDRsV2 struct {
	v1.CDRsV1
}

func (cdrSv2 *CDRsV2) StoreSessionCost(ctx *context.Context, args *engine.ArgsV2CDRSStoreSMCost, reply *string) error {
	return cdrSv2.CDRs.V2StoreSessionCost(ctx, args, reply)
}

// ProcessEvent will process an Event based on the flags attached
func (cdrSv2 *CDRsV2) ProcessEvent(ctx *context.Context, arg *engine.ArgV1ProcessEvent, evs *[]*utils.EventWithFlags) error {
	return cdrSv2.CDRs.V2ProcessEvent(ctx, arg, evs)
}
