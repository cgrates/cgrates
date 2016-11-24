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
func (apier *ApierV1) DebitUsage(usageRecord engine.UsageRecord, reply *string) error {
	return apier.DebitUsageWithOptions(AttrDebitUsageWithOptions{
		Options:     AttrDebitUsageOptions{AllowNegative: true},
		UsageRecord: usageRecord,
	}, reply)
}

// AttrDebitUsageOptions represents options that
// are applied to the DebitUsage request
type AttrDebitUsageOptions struct {
	AllowNegative bool
}

// AttrDebitUsageWithOptions represents the DebitUsage request
type AttrDebitUsageWithOptions struct {
	Options     AttrDebitUsageOptions
	UsageRecord engine.UsageRecord
}

// DebitUsageWithOptions will debit the account based on the usage cost with
// additional options to control if the balance can go negative
func (apier *ApierV1) DebitUsageWithOptions(usageRecordWithOptions AttrDebitUsageWithOptions, reply *string) error {
	var usageRecord = &usageRecordWithOptions.UsageRecord
	var options = &usageRecordWithOptions.Options

	if missing := utils.MissingStructFields(usageRecord, []string{"Account", "Destination", "Usage"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	err := engine.LoadUserProfile(usageRecord, "")
	if err != nil {
		*reply = err.Error()
		return err
	}

	// Set values for optional parameters
	if usageRecord.ToR == "" {
		usageRecord.ToR = utils.VOICE
	}
	if usageRecord.RequestType == "" {
		usageRecord.RequestType = apier.Config.DefaultReqType
	}
	if usageRecord.Direction == "" {
		usageRecord.Direction = utils.OUT
	}
	if usageRecord.Tenant == "" {
		usageRecord.Tenant = apier.Config.DefaultTenant
	}
	if usageRecord.Category == "" {
		usageRecord.Category = apier.Config.DefaultCategory
	}
	if usageRecord.Subject == "" {
		usageRecord.Subject = usageRecord.Account
	}
	if usageRecord.AnswerTime == "" {
		usageRecord.AnswerTime = utils.META_NOW
	}

	// Get the call descriptor from the usage record
	cd, err := usageRecord.AsCallDescriptor(apier.Config.DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}

	// Apply options
	cd.AllowNegative = options.AllowNegative

	// Calculate the cost for usage and debit the account
	var cc engine.CallCost
	if err := apier.Responder.Debit(cd, &cc); err != nil {
		return utils.NewErrServerError(err)
	}

	*reply = OK
	return nil
}
