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

// GetAttributeProfile returns an Attribute Profile
func (apierV1 *ApierV1) GetAttributeProfile(arg utils.TenantID, reply *engine.AttributeProfile) error {
	if missing := utils.MissingStructFields(&arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if alsPrf, err := apierV1.DataManager.GetAttributeProfile(arg.Tenant, arg.ID, false, utils.NonTransactional); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *alsPrf
	}
	return nil
}

//SetAttributeProfile add/update a new Attribute Profile
func (apierV1 *ApierV1) SetAttributeProfile(alsPrf *engine.AttributeProfile, reply *string) error {
	if missing := utils.MissingStructFields(alsPrf, []string{"Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataManager.SetAttributeProfile(alsPrf, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

type ArgRemoveAttrProfile struct {
	Tenant   string
	ID       string
	Contexts []string
}

//RemoveAttributeProfile remove a specific Attribute Profile
func (apierV1 *ApierV1) RemoveAttributeProfile(arg *ArgRemoveAttrProfile, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{"Tenant", "ID", "Contexts"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataManager.RemoveAttributeProfile(arg.Tenant, arg.ID, arg.Contexts, utils.NonTransactional, true); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = utils.OK
	return nil
}

func NewAttributeSv1(attrS *engine.AttributeService) *AttributeSv1 {
	return &AttributeSv1{attrS: attrS}
}

// Exports RPC from RLs
type AttributeSv1 struct {
	attrS *engine.AttributeService
}

// Call implements rpcclient.RpcClientConnection interface for internal RPC
func (alSv1 *AttributeSv1) Call(serviceMethod string,
	args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(alSv1, serviceMethod, args, reply)
}

// GetAttributeForEvent  returns matching AttributeProfile for Event
func (alSv1 *AttributeSv1) GetAttributeForEvent(ev *utils.CGREvent,
	reply *engine.AttributeProfile) error {
	return alSv1.attrS.V1GetAttributeForEvent(ev, reply)
}

// ProcessEvent will replace event fields with the ones in maching AttributeProfile
func (alSv1 *AttributeSv1) ProcessEvent(ev *utils.CGREvent,
	reply *engine.AttrSProcessEventReply) error {
	return alSv1.attrS.V1ProcessEvent(ev, reply)
}
