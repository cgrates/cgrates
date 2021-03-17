/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
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

package v1

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// DebitUsage will debit the balance for the usage cost, allowing the
// account to go negative if the cost calculated is greater than the balance
func (apierSv1 *APIerSv1) DebitUsage(usageRecord *engine.UsageRecordWithOpts, reply *string) error {
	return apierSv1.DebitUsageWithOptions(&AttrDebitUsageWithOptions{
		UsageRecord:          usageRecord,
		AllowNegativeAccount: true,
	}, reply)
}

// AttrDebitUsageWithOptions represents the DebitUsage request
type AttrDebitUsageWithOptions struct {
	UsageRecord          *engine.UsageRecordWithOpts
	AllowNegativeAccount bool // allow account to go negative during debit
}

// DebitUsageWithOptions will debit the account based on the usage cost with
// additional options to control if the balance can go negative
func (apierSv1 *APIerSv1) DebitUsageWithOptions(args *AttrDebitUsageWithOptions, reply *string) error {
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
	if err := apierSv1.Responder.Debit(&engine.CallDescriptorWithAPIOpts{
		CallDescriptor: cd,
		APIOpts:        args.UsageRecord.Opts,
	}, &cc); err != nil {
		return utils.NewErrServerError(err)
	}

	*reply = utils.OK
	return nil
}
