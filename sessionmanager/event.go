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

package sessionmanager

import (
	"time"
)

type Event interface {
	New(string) Event
	GetName() string
	GetDirection() string
	GetOrigId() string
	GetSubject() string
	GetAccount() string
	GetDestination() string
	GetCallDestNr() string
	GetTOR() string
	GetUUID() string
	GetTenant() string
	GetReqType() string
	GetStartTime(string) (time.Time, error)
	GetEndTime() (time.Time, error)
	GetFallbackSubj() string
	MissingParameter() bool
}

// Returns first non empty string out of vals. Useful to extract defaults
func firstNonEmpty(vals ...string) string {
	for _, val := range vals {
		if len(val) != 0 {
			return val
		}
	}
	return ""
}
