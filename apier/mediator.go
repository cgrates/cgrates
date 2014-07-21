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
	"fmt"
	"time"

	"github.com/cgrates/cgrates/mediator"
	"github.com/cgrates/cgrates/utils"
)

type MediatorV1 struct {
	Medi *mediator.Mediator
}

// Remotely start mediation with specific runid, runs asynchronously, it's status will be displayed in syslog
func (self *MediatorV1) RateCdrs(attrs utils.AttrRateCdrs, reply *string) error {
	if self.Medi == nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, "MediatorNotRunning")
	}
	var tStart, tEnd time.Time
	var err error
	if len(attrs.TimeStart) != 0 {
		if tStart, err = utils.ParseTimeDetectLayout(attrs.TimeStart); err != nil {
			return err
		}
	}
	if len(attrs.TimeEnd) != 0 {
		if tEnd, err = utils.ParseTimeDetectLayout(attrs.TimeEnd); err != nil {
			return err
		}
	}
	if err := self.Medi.RateCdrs(tStart, tEnd, attrs.RerateErrors, attrs.RerateRated, attrs.SendToStats); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = utils.OK
	return nil
}
