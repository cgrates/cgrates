// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v2

import (
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type AttributeWithCache struct {
	*engine.ExternalAttributeProfile
	Cache *string
}

// SetAttributeProfile add/update a new Attribute Profile
func (APIerSv2 *APIerSv2) SetAttributeProfile(arg *AttributeWithCache, reply *string) error {
	if missing := utils.MissingStructFields(arg.ExternalAttributeProfile, []string{utils.Tenant, utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	alsPrf, err := arg.ExternalAttributeProfile.AsAttributeProfile()
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
	args := utils.ArgsGetCacheItem{
		CacheID: utils.CacheAttributeProfiles,
		ItemID:  alsPrf.TenantID(),
	}
	if err := APIerSv2.APIerSv1.CallCache(
		arg.Tenant,
		v1.GetCacheOpt(arg.Cache),
		args); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
