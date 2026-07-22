// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"strconv"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Returns MaxUsage (for calls in seconds), -1 for no limit
func (apierSv1 *APIerSv1) GetMaxUsage(ctx *context.Context, usageRecord *engine.UsageRecordWithAPIOpts, maxUsage *time.Duration) error {
	if apierSv1.Responder == nil {
		return utils.NewErrNotConnected(utils.RALService)
	}
	if usageRecord.ToR == utils.EmptyString {
		usageRecord.ToR = utils.MetaVoice
	}
	if usageRecord.RequestType == utils.EmptyString {
		usageRecord.RequestType = apierSv1.Config.GeneralCfg().DefaultReqType
	}
	if usageRecord.Tenant == utils.EmptyString {
		usageRecord.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if usageRecord.Category == utils.EmptyString {
		usageRecord.Category = apierSv1.Config.GeneralCfg().DefaultCategory
	}
	if usageRecord.Subject == utils.EmptyString {
		usageRecord.Subject = usageRecord.Account
	}
	if usageRecord.SetupTime == utils.EmptyString {
		usageRecord.SetupTime = utils.MetaNow
	}
	if usageRecord.Usage == utils.EmptyString {
		usageRecord.Usage = strconv.FormatFloat(
			float64(apierSv1.Config.SessionSCfg().GetDefaultUsage(usageRecord.ToR)), 'f', -1, 64)
	}
	cd, err := usageRecord.AsCallDescriptor(apierSv1.Config.GeneralCfg().DefaultTimezone, false)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	var maxDur time.Duration
	if err := apierSv1.Responder.GetMaxSessionTime(ctx,
		&engine.CallDescriptorWithAPIOpts{
			CallDescriptor: cd,
			APIOpts:        usageRecord.APIOpts,
		}, &maxDur); err != nil {
		return err
	}
	if maxDur == time.Duration(-1) {
		*maxUsage = -1
		return nil
	}
	*maxUsage = maxDur
	return nil
}
