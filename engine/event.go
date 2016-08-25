/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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

package engine

import (
	"github.com/cgrates/cgrates/utils"
	"time"
)

type Event interface {
	GetName() string
	GetCgrId(timezone string) string
	GetUUID() string
	GetSessionIds() []string // Returns identifiers needed to control a session (eg disconnect)
	GetDirection(string) string
	GetSubject(string) string
	GetAccount(string) string
	GetDestination(string) string
	GetCallDestNr(string) string
	GetOriginatorIP(string) string
	GetCategory(string) string
	GetTenant(string) string
	GetReqType(string) string
	GetSetupTime(string, string) (time.Time, error)
	GetAnswerTime(string, string) (time.Time, error)
	GetEndTime(string, string) (time.Time, error)
	GetDuration(string) (time.Duration, error)
	GetPdd(string) (time.Duration, error)
	GetSupplier(string) string
	GetDisconnectCause(string) string
	GetExtraFields() map[string]string
	MissingParameter(string) bool
	ParseEventValue(*utils.RSRField, string) string
	AsStoredCdr(timezone string) *CDR
	String() string
	AsEvent(string) Event
	ComputeLcr() bool
	AsMapStringIface() (map[string]interface{}, error)
}
