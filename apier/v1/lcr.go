/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package v1

import (
	"fmt"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Computes the LCR for a specific request emulating a call
func (self *ApierV1) GetLcr(lcrReq engine.LcrRequest, lcrReply *engine.LcrReply) error {
	cd, err := lcrReq.AsCallDescriptor(self.Config.DefaultTimezone)
	if err != nil {
		return err
	}
	var lcrQried engine.LCRCost
	if err := self.Responder.GetLCR(&engine.AttrGetLcr{CallDescriptor: cd, LCRFilter: lcrReq.LCRFilter, Paginator: lcrReq.Paginator}, &lcrQried); err != nil {
		return utils.NewErrServerError(err)
	}
	if lcrQried.Entry == nil {
		return utils.ErrNotFound
	}
	lcrReply.DestinationId = lcrQried.Entry.DestinationId
	lcrReply.RPCategory = lcrQried.Entry.RPCategory
	lcrReply.Strategy = lcrQried.Entry.Strategy
	for _, qriedSuppl := range lcrQried.SupplierCosts {
		if qriedSuppl.Error != "" {
			utils.Logger.Err(fmt.Sprintf("LCR_ERROR: supplier <%s>, error <%s>", qriedSuppl.Supplier, qriedSuppl.Error))
			if !lcrReq.IgnoreErrors {
				return fmt.Errorf("%s:%s", utils.ErrServerError.Error(), "LCR_COMPUTE_ERRORS")
			}
			continue
		}
		if dtcs, err := utils.NewDTCSFromRPKey(qriedSuppl.Supplier); err != nil {
			return utils.NewErrServerError(err)
		} else {
			lcrReply.Suppliers = append(lcrReply.Suppliers, &engine.LcrSupplier{Supplier: dtcs.Subject, Cost: qriedSuppl.Cost, QOS: qriedSuppl.QOS})
		}
	}
	return nil
}

// Computes the LCR for a specific request emulating a call, returns a comma separated list of suppliers
func (self *ApierV1) GetLcrSuppliers(lcrReq engine.LcrRequest, suppliers *string) (err error) {
	cd, err := lcrReq.AsCallDescriptor(self.Config.DefaultTimezone)
	if err != nil {
		return err
	}
	var lcrQried engine.LCRCost
	if err := self.Responder.GetLCR(&engine.AttrGetLcr{CallDescriptor: cd, LCRFilter: lcrReq.LCRFilter, Paginator: lcrReq.Paginator}, &lcrQried); err != nil {
		return utils.NewErrServerError(err)
	}
	if lcrQried.HasErrors() {
		lcrQried.LogErrors()
		if !lcrReq.IgnoreErrors {
			return fmt.Errorf("%s:%s", utils.ErrServerError.Error(), "LCR_COMPUTE_ERRORS")
		}
	}
	if suppliersStr, err := lcrQried.SuppliersString(); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*suppliers = suppliersStr
	}
	return nil
}
