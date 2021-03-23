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

package v2

import (
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type AttributeWithAPIOpts struct {
	*engine.APIAttributeProfile
	APIOpts map[string]interface{}
}

//SetAttributeProfile add/update a new Attribute Profile
func (APIerSv2 *APIerSv2) SetAttributeProfile(arg *AttributeWithAPIOpts, reply *string) error {
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
	if err := APIerSv2.APIerSv1.CallCache(utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), alsPrf.Tenant, utils.CacheAttributeProfiles,
		alsPrf.TenantID(), &alsPrf.FilterIDs, alsPrf.Contexts, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
