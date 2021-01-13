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

// Call implements rpcclient.ClientConnector interface for internal RPC
func (rsv1 *RALsV1) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(rsv1, serviceMethod, args, reply)
}

// GetRatingPlansCost returns EventCosts matching RatingPlanIDs
func (rsv1 *RALsV1) GetRatingPlansCost(arg *utils.RatingPlanCostArg, reply *dispatchers.RatingPlanCost) error {
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

func (rsv1 *RALsV1) Ping(ign *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}
