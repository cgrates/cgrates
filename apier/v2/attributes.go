// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v2

import (
	"fmt"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type AttributeWithAPIOpts struct {
	*engine.APIAttributeProfile
	APIOpts map[string]any
}

// SetAttributeProfile add/update a new Attribute Profile
func (APIerSv2 *APIerSv2) SetAttributeProfile(ctx *context.Context, arg *AttributeWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg.APIAttributeProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = APIerSv2.Config.GeneralCfg().DefaultTenant
	}
	alsPrf, err := arg.APIAttributeProfile.AsAttributeProfile()
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := APIerSv2.DataManager.SetAttributeProfile(alsPrf, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheAttributeProfiles and store it in database
	if err := APIerSv2.DataManager.SetLoadIDs(
		map[string]int64{utils.CacheAttributeProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if APIerSv2.Config.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<V2SetAttributeProfile> Delaying cache call for %v", APIerSv2.Config.GeneralCfg().CachingDelay))
		time.Sleep(APIerSv2.Config.GeneralCfg().CachingDelay)
	}
	if err := APIerSv2.APIerSv1.CallCache(utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), alsPrf.Tenant, utils.CacheAttributeProfiles,
		alsPrf.TenantID(), utils.EmptyString, &alsPrf.FilterIDs, alsPrf.Contexts, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
