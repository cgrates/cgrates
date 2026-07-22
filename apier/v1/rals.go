// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewRALsV1() *RALsV1 {
	return &RALsV1{}
}

// Exports RPC from RALs
type RALsV1 struct {
}

// GetRatingPlansCost returns EventCosts matching RatingPlanIDs
func (rsv1 *RALsV1) GetRatingPlansCost(ctx *context.Context, arg *utils.RatingPlanCostArg, reply *dispatchers.RatingPlanCost) error {
	if missing := utils.MissingStructFields(arg, []string{utils.RatingPlanIDs,
		utils.Destination, utils.SetupTime, utils.Usage}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	//parse SetupTime and Usage
	setupTime, err := utils.ParseTimeDetectLayout(arg.SetupTime,
		config.CgrConfig().GeneralCfg().DefaultTimezone)
	if err != nil {
		return err
	}
	usage, err := utils.ParseDurationWithNanosecs(arg.Usage)
	if err != nil {
		return err
	}
	tenant := utils.UUIDSha1Prefix()
	category := utils.MetaRatingPlanCost
	subject := utils.UUIDSha1Prefix()
	cd := &engine.CallDescriptor{
		Category:      category,
		Tenant:        tenant,
		Subject:       subject,
		Destination:   arg.Destination,
		TimeStart:     setupTime,
		TimeEnd:       setupTime.Add(usage),
		DurationIndex: usage,
	}
	for _, rp := range arg.RatingPlanIDs { // loop through RatingPlans until we find one without errors
		rPrfl := &engine.RatingProfile{
			Id: utils.ConcatenatedKey(utils.MetaOut,
				tenant, category, subject),
			RatingPlanActivations: engine.RatingPlanActivations{
				&engine.RatingPlanActivation{
					ActivationTime: setupTime,
					RatingPlanId:   rp,
				},
			},
		}
		// force cache set so it can be picked by calldescriptor for cost calculation
		if err := engine.Cache.Set(utils.CacheRatingProfilesTmp, rPrfl.Id, rPrfl, nil,
			true, utils.NonTransactional); err != nil {
			return err
		}
		cc, err := cd.GetCost()
		if err := engine.Cache.Remove(utils.CacheRatingProfilesTmp, rPrfl.Id, // Remove here so we don't overload memory
			true, utils.NonTransactional); err != nil {
			return err
		}
		if err != nil {
			// in case we have UnauthorizedDestination
			// or NotFound try next RatingPlan
			if err != utils.ErrUnauthorizedDestination &&
				err != utils.ErrNotFound {
				return err
			}
			continue
		}
		ec := engine.NewEventCostFromCallCost(cc, utils.EmptyString, utils.EmptyString)
		ec.Compute()
		*reply = dispatchers.RatingPlanCost{
			EventCost:    ec,
			RatingPlanID: rp,
		}
		break
	}
	return nil
}
