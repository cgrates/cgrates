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

package utils

import (
	"time"
)

var PrimaryCdrFields []string = []string{ACCID, CDRHOST, CDRSOURCE, REQTYPE, DIRECTION, TENANT, TOR, ACCOUNT, SUBJECT, DESTINATION, ANSWER_TIME, DURATION}

type CDR interface {
	GetCgrId() string
	GetAccId() string
	GetCdrHost() string
	GetCdrSource() string
	GetDirection() string
	GetSubject() string
	GetAccount() string
	GetDestination() string
	GetTOR() string
	GetTenant() string
	GetReqType() string
	GetAnswerTime() (time.Time, error)
	GetDuration() int64
	GetExtraFields() map[string]string //Stores extra CDR Fields
}
