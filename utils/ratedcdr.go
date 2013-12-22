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

// Rated CDR as extracted from StorDb. Kinda standard of internal CDR
type RatedCDR struct {
	CgrId          string
	AccId          string
	CdrHost        string
	CdrSource      string
	ReqType        string
	Direction      string
	Tenant         string
	TOR            string
	Account        string
	Subject        string
	Destination    string
	AnswerTime     time.Time
	Duration       int64
	ExtraFields    map[string]string
	MediationRunId string
	Cost           float64
}
