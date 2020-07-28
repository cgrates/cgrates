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
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/utils"
)

func NewEventExporterSv1(eeS *ees.EventExporterS) *EventExporterSv1 {
	return &EventExporterSv1{eeS: eeS}
}

type EventExporterSv1 struct {
	eeS *ees.EventExporterS
}

func (eSv1 *EventExporterSv1) Ping(ign *utils.CGREventWithOpts, reply *string) error {
	*reply = utils.Pong
	return nil
}

// ProcessEvent triggers exports on EEs side
func (eSv1 *EventExporterSv1) ProcessEvent(args *utils.CGREventWithIDs,
	reply *string) error {
	return eSv1.eeS.V1ProcessEvent(args, reply)
}
