// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// DebitUsage will debit the balance for the usage cost, allowing the
// account to go negative if the cost calculated is greater than the balance
func (apier *APIerSv1) DebitUsage(usageRecord engine.UsageRecordWithArgDispatcher, reply *string) error {
	return apier.DebitUsageWithOptions(AttrDebitUsageWithOptions{
		UsageRecord:          &usageRecord,
		AllowNegativeAccount: true,
	}, reply)
}

// AttrDebitUsageWithOptions represents the DebitUsage request
type AttrDebitUsageWithOptions struct {
	UsageRecord          *engine.UsageRecordWithArgDispatcher
	AllowNegativeAccount bool // allow account to go negative during debit
}

// DebitUsageWithOptions will debit the account based on the usage cost with
// additional options to control if the balance can go negative
func (apier *APIerSv1) DebitUsageWithOptions(args AttrDebitUsageWithOptions, reply *string) error {
	usageRecord := args.UsageRecord.UsageRecord
	if missing := utils.MissingStructFields(usageRecord, []string{"Account", "Destination", "Usage"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	// Set values for optional parameters
	if usageRecord.ToR == "" {
		usageRecord.ToR = utils.VOICE
	}
	if usageRecord.RequestType == "" {
		usageRecord.RequestType = apier.Config.GeneralCfg().DefaultReqType
	}
	if usageRecord.Tenant == "" {
		usageRecord.Tenant = apier.Config.GeneralCfg().DefaultTenant
	}
	if usageRecord.Category == "" {
		usageRecord.Category = apier.Config.GeneralCfg().DefaultCategory
	}
	if usageRecord.Subject == "" {
		usageRecord.Subject = usageRecord.Account
	}
	if usageRecord.AnswerTime == "" {
		usageRecord.AnswerTime = utils.META_NOW
	}

	// Get the call descriptor from the usage record
	cd, err := usageRecord.AsCallDescriptor(apier.Config.GeneralCfg().DefaultTimezone,
		!args.AllowNegativeAccount)
	if err != nil {
		return utils.NewErrServerError(err)
	}

	// Calculate the cost for usage and debit the account
	var cc engine.CallCost
	if err := apier.Responder.Debit(&engine.CallDescriptorWithArgDispatcher{CallDescriptor: cd,
		ArgDispatcher: args.UsageRecord.ArgDispatcher}, &cc); err != nil {
		return utils.NewErrServerError(err)
	}

	*reply = utils.OK
	return nil
}
