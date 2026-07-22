// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// DebitUsage will debit the balance for the usage cost, allowing the
// account to go negative if the cost calculated is greater than the balance
func (apierSv1 *APIerSv1) DebitUsage(ctx *context.Context, usageRecord *engine.UsageRecordWithAPIOpts, reply *string) error {
	return apierSv1.DebitUsageWithOptions(ctx,
		&AttrDebitUsageWithOptions{
			UsageRecord:          usageRecord,
			AllowNegativeAccount: true,
		}, reply)
}

// AttrDebitUsageWithOptions represents the DebitUsage request
type AttrDebitUsageWithOptions struct {
	UsageRecord          *engine.UsageRecordWithAPIOpts
	AllowNegativeAccount bool // allow account to go negative during debit
}

// DebitUsageWithOptions will debit the account based on the usage cost with
// additional options to control if the balance can go negative
func (apierSv1 *APIerSv1) DebitUsageWithOptions(ctx *context.Context, args *AttrDebitUsageWithOptions, reply *string) error {
	if apierSv1.Responder == nil {
		return utils.NewErrNotConnected(utils.RALService)
	}
	usageRecord := args.UsageRecord.UsageRecord
	if missing := utils.MissingStructFields(usageRecord, []string{"Account", "Destination", "Usage"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	// Set values for optional parameters
	if usageRecord.ToR == "" {
		usageRecord.ToR = utils.MetaVoice
	}
	if usageRecord.RequestType == "" {
		usageRecord.RequestType = apierSv1.Config.GeneralCfg().DefaultReqType
	}
	if usageRecord.Tenant == "" {
		usageRecord.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if usageRecord.Category == "" {
		usageRecord.Category = apierSv1.Config.GeneralCfg().DefaultCategory
	}
	if usageRecord.Subject == "" {
		usageRecord.Subject = usageRecord.Account
	}
	if usageRecord.AnswerTime == "" {
		usageRecord.AnswerTime = utils.MetaNow
	}

	// Get the call descriptor from the usage record
	cd, err := usageRecord.AsCallDescriptor(apierSv1.Config.GeneralCfg().DefaultTimezone,
		!args.AllowNegativeAccount)
	if err != nil {
		return utils.NewErrServerError(err)
	}

	// Calculate the cost for usage and debit the account
	var cc engine.CallCost
	if err := apierSv1.Responder.Debit(
		context.Background(),
		&engine.CallDescriptorWithAPIOpts{
			CallDescriptor: cd,
			APIOpts:        args.UsageRecord.APIOpts,
		}, &cc); err != nil {
		return utils.NewErrServerError(err)
	}

	*reply = utils.OK
	return nil
}
