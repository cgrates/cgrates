/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package apier

import (
	"errors"
	"fmt"
	"github.com/cgrates/cgrates/rater"
	"github.com/cgrates/cgrates/utils"
)

type ApierTPDestination struct {
	TPid          string   // Tariff plan id
	DestinationId string   // Destination id
	Prefixes      []string // Prefixes attached to this destination
}

// Creates a new destination within a tariff plan
func (self *Apier) SetTPDestination(attrs ApierTPDestination, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DestinationId", "Prefixes"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if exists, err := self.StorDb.ExistsTPDestination(attrs.TPid, attrs.DestinationId); err != nil {
		return fmt.Errorf("%s:%v", utils.ERR_SERVER_ERROR, err.Error())
	} else if exists {
		return errors.New(utils.ERR_DUPLICATE)
	}
	if err := self.StorDb.SetTPDestination(attrs.TPid, &rater.Destination{attrs.DestinationId, attrs.Prefixes}); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = "OK"
	return nil
}

type AttrGetTPDestination struct {
	TPid          string // Tariff plan id
	DestinationId string // Destination id
}

// Queries a specific destination
func (self *Apier) GetTPDestination(attrs AttrGetTPDestination, reply *ApierTPDestination) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DestinationId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if dst, err := self.StorDb.GetTPDestination(attrs.TPid, attrs.DestinationId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if dst == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ApierTPDestination{attrs.TPid, dst.Id, dst.Prefixes}
	}
	return nil
}

type AttrGetTPDestinationIds struct {
	TPid string // Tariff plan id
}

// Queries destination identities on specific tariff plan.
func (self *Apier) GetTPDestinationIds(attrs AttrGetTPDestinationIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ids, err := self.StorDb.GetTPDestinationIds(attrs.TPid); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if ids == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ids
	}
	return nil
}
