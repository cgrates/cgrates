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

// CDR as extracted from StorDb. Kinda standard of internal CDR
type RatedCDR struct {
	CgrId         string
        AccId         string
        CdrHost       string
	ReqType       string
        Direction     string
	Tenant        string
	TOR           string
	Account       string
        Subject       string
	Destination   string
        AnswerTime    time.Time
        Duration      int64
        ExtraFields   map[string]string
	Cost          float64
}

func(ratedCdr *RatedCDR) GetCgrId() string {
	return ratedCdr.CgrId
}

func(ratedCdr *RatedCDR) GetAccId() string {
	return ratedCdr.AccId
}

func(ratedCdr *RatedCDR) GetCdrHost() string {
	return ratedCdr.CdrHost
}

func(ratedCdr *RatedCDR) GetDirection() string {
	return ratedCdr.Direction
}

func(ratedCdr *RatedCDR) GetSubject() string {
	return ratedCdr.Subject
}

func(ratedCdr *RatedCDR) GetAccount() string {
	return ratedCdr.Account
}

func(ratedCdr *RatedCDR) GetDestination() string {
	return ratedCdr.Destination
}

func(ratedCdr *RatedCDR) GetTOR() string {
	return ratedCdr.TOR
}

func(ratedCdr *RatedCDR) GetTenant() string {
	return ratedCdr.Tenant
}

func(ratedCdr *RatedCDR) GetReqType() string {
	return ratedCdr.ReqType
}

func(ratedCdr *RatedCDR) GetAnswerTime() (time.Time, error) {
	return ratedCdr.AnswerTime, nil
}

func(ratedCdr *RatedCDR) GetDuration() int64 {
	return ratedCdr.Duration
}

func(ratedCdr *RatedCDR) GetExtraFields() map[string]string {
	return ratedCdr.ExtraFields
}
