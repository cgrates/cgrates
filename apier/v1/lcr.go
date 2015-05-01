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
	"errors"
	"fmt"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// A request for LCR
type LcrRequest struct {
	Direction   string
	Tenant      string
	Category    string
	Account     string
	Subject     string
	Destination string
	TimeStart   string
	Duration    string
}

// A LCR reply
type LcrReply struct {
	DestinationId string
	RPCategory    string
	Strategy      string
	Suppliers     []*LcrSupplier
}

// One supplier out of LCR reply
type LcrSupplier struct {
	Supplier string
	Cost     float64
}

// Computes the LCR for a specific request emulating a call
func (self *ApierV1) GetLcr(lcrReq LcrRequest, lcrReply *LcrReply) (err error) {
	if missing := utils.MissingStructFields(&lcrReq, []string{"Account", "Destination"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	// Handle defaults
	if len(lcrReq.Direction) == 0 {
		lcrReq.Direction = utils.OUT
	}
	if len(lcrReq.Tenant) == 0 {
		lcrReq.Tenant = self.Config.DefaultTenant
	}
	if len(lcrReq.Category) == 0 {
		lcrReq.Category = self.Config.DefaultCategory
	}
	if len(lcrReq.Subject) == 0 {
		lcrReq.Subject = lcrReq.Account
	}
	var timeStart time.Time
	if len(lcrReq.TimeStart) == 0 {
		timeStart = time.Now()
	} else if timeStart, err = utils.ParseTimeDetectLayout(lcrReq.TimeStart); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	var callDur time.Duration
	if len(lcrReq.Duration) == 0 {
		callDur = time.Duration(1) * time.Minute
	} else if callDur, err = utils.ParseDurationWithSecs(lcrReq.Duration); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	cd := &engine.CallDescriptor{
		Direction:   lcrReq.Direction,
		Tenant:      lcrReq.Tenant,
		Category:    lcrReq.Category,
		Account:     lcrReq.Account,
		Subject:     lcrReq.Subject,
		Destination: lcrReq.Destination,
		TimeStart:   timeStart,
		TimeEnd:     timeStart.Add(callDur),
	}
	var lcrQried engine.LCRCost
	if err := self.Responder.GetLCR(cd, &lcrQried); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	if lcrQried.Entry == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	}
	lcrReply.DestinationId = lcrQried.Entry.DestinationId
	lcrReply.RPCategory = lcrQried.Entry.RPCategory
	lcrReply.Strategy = lcrQried.Entry.Strategy
	for _, qriedSuppl := range lcrQried.SupplierCosts {
		if dtcs, err := utils.NewDTCSFromRPKey(qriedSuppl.Supplier); err != nil {
			return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
		} else {
			lcrReply.Suppliers = append(lcrReply.Suppliers, &LcrSupplier{Supplier: dtcs.Subject, Cost: qriedSuppl.Cost})
		}
	}
	return nil
}
