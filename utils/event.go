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

package utils

import (
	"time"
)

type Event interface {
	GetName() string
	GetCgrId() string
	GetUUID() string
	GetSessionIds() []string // Returns identifiers needed to control a session (eg disconnect)
	GetDirection(string) string
	GetSubject(string) string
	GetAccount(string) string
	GetDestination(string) string
	GetCallDestNr(string) string
	GetCategory(string) string
	GetTenant(string) string
	GetReqType(string) string
	GetSetupTime(string) (time.Time, error)
	GetAnswerTime(string) (time.Time, error)
	GetEndTime() (time.Time, error)
	GetDuration(string) (time.Duration, error)
	GetOriginatorIP(string) string
	GetExtraFields() map[string]string
	MissingParameter() bool
	ParseEventValue(*RSRField) string
	PassesFieldFilter(*RSRField) (bool, string)
	AsStoredCdr() *StoredCdr
	String() string
	AsEvent(string) Event
}
