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

type AttrGetTPDestination struct {
	TPid            string
	DestinationId string
}

// Return destinations profile for a destination tag received as parameter
func (self *Apier) GetTPDestination(attrs AttrGetTPDestination, reply *rater.Destination) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DestinationId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if dst, err := self.StorDb.GetTPDestination(attrs.TPid, attrs.DestinationId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if dst == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = *dst
	}
	return nil
}

type AttrSetTPDestination struct {
	TPid            string
	DestinationId string
	Prefixes []string
}

func (self *Apier) SetTPDestination(attrs AttrSetTPDestination, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DestinationId", "Prefixes"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if err := self.StorDb.SetTPDestination(attrs.TPid, &rater.Destination{attrs.DestinationId, attrs.Prefixes}); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = "OK"
	return nil
}
