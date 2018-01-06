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
	"github.com/cgrates/cgrates/cache"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// GetSupplierProfile returns a Supplier configuration
func (apierV1 *ApierV1) GetSupplierProfile(arg utils.TenantID, reply *engine.SupplierProfile) error {
	if missing := utils.MissingStructFields(&arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if spp, err := apierV1.DataManager.GetSupplierProfile(arg.Tenant, arg.ID, false, utils.NonTransactional); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *spp
	}
	return nil
}

//SetSupplierProfile add a new Supplier configuration
func (apierV1 *ApierV1) SetSupplierProfile(spp *engine.SupplierProfile, reply *string) error {
	if missing := utils.MissingStructFields(spp, []string{"Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataManager.SetSupplierProfile(spp, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	cache.RemKey(utils.SupplierProfilePrefix+utils.ConcatenatedKey(spp.Tenant, spp.ID), true, "") // ToDo: Remove here with autoreload
	*reply = utils.OK
	return nil
}

//RemSupplierProfile remove a specific Supplier configuration
func (apierV1 *ApierV1) RemSupplierProfile(arg utils.TenantID, reply *string) error {
	if missing := utils.MissingStructFields(&arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataManager.RemoveSupplierProfile(arg.Tenant, arg.ID, utils.NonTransactional, true); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = utils.OK
	return nil
}

func NewSupplierSv1(splS *engine.SupplierService) *SupplierSv1 {
	return &SupplierSv1{splS: splS}
}

// Exports RPC from RLs
type SupplierSv1 struct {
	splS *engine.SupplierService
}

// Call implements rpcclient.RpcClientConnection interface for internal RPC
func (splv1 *SupplierSv1) Call(serviceMethod string,
	args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(splv1, serviceMethod, args, reply)
}

// GetSuppliers returns sorted list of suppliers for Event
func (splv1 *SupplierSv1) GetSuppliers(args *utils.CGREvent,
	reply *engine.SortedSuppliers) error {
	return splv1.splS.V1GetSuppliers(args, reply)
}
