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

type ExternalAliasProfile struct {
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // Activation interval
	Aliases            []*ExternalAliasEntry
	Weight             float64
}

type ExternalAliasEntry struct {
	FieldName string
	Initial   string
	Alias     string
}

// GetAliasProfile returns an Alias Profile
func (apierV1 *ApierV1) GetAliasProfile(arg utils.TenantID, reply *engine.AliasProfile) error {
	if missing := utils.MissingStructFields(&arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if alsPrf, err := apierV1.DataManager.GetAliasProfile(arg.Tenant, arg.ID, false, utils.NonTransactional); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *alsPrf
	}
	return nil
}

//SetAliasProfile add a new Alias Profile
func (apierV1 *ApierV1) SetAliasProfile(alsPrf *ExternalAliasProfile, reply *string) error {
	if missing := utils.MissingStructFields(alsPrf, []string{"Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	alsPrfEngine := &engine.AliasProfile{
		Tenant:             alsPrf.Tenant,
		ID:                 alsPrf.ID,
		Weight:             alsPrf.Weight,
		FilterIDs:          alsPrf.FilterIDs,
		ActivationInterval: alsPrf.ActivationInterval,
	}
	alsMap := make(map[string]map[string]string)
	for _, als := range alsPrf.Aliases {
		alsMap[als.FieldName] = make(map[string]string)
		alsMap[als.FieldName][als.Initial] = als.Alias
	}
	alsPrfEngine.Aliases = alsMap
	if err := apierV1.DataManager.SetAliasProfile(alsPrfEngine); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

//RemAliasProfile remove a specific Alias Profile
func (apierV1 *ApierV1) RemAliasProfile(arg utils.TenantID, reply *string) error {
	if missing := utils.MissingStructFields(&arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataManager.RemoveAliasProfile(arg.Tenant, arg.ID, utils.NonTransactional); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = utils.OK
	return nil
}
