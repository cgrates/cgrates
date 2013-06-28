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

type CDR interface {
	GetCgrId() string
	GetAccId() string
	GetCdrHost() string
	GetDirection() string
	GetOrigId() string
	GetSubject() string
	GetAccount() string
	GetDestination() string
	GetTOR() string
	GetTenant() string
	GetReqType() string
	GetAnswerTime() (time.Time, error)
	GetDuration() int64
	GetFallbackSubj() string
	GetExtraFields() map[string]string //Stores extra CDR Fields
}

type GenericCdr map[string]string

func (gcdr GenericCdr) GetCgrId() string {
	return ""
}
func (gcdr GenericCdr) GetAccId() string {
	return ""
}
func (gcdr GenericCdr) GetCdrHost() string {
	return ""
}
func (gcdr GenericCdr) GetDirection() string {
	return ""
}
func (gcdr GenericCdr) GetOrigId() string {
	return ""
}
func (gcdr GenericCdr) GetSubject() string {
	return ""
}
func (gcdr GenericCdr) GetAccount() string {
	return ""
}
func (gcdr GenericCdr) GetDestination() string {
	return ""
}
func (gcdr GenericCdr) GetTOR() string {
	return ""
}
func (gcdr GenericCdr) GetTenant() string {
	return ""
}
func (gcdr GenericCdr) GetReqType() string {
	return ""
}
func (gcdr GenericCdr) GetAnswerTime() (time.Time, error) {
	return time.Now(), nil
}
func (gcdr GenericCdr) GetDuration() int64 {
	return 0.0
}
func (gcdr GenericCdr) GetFallbackSubj() string {
	return ""
}
func (gcdr GenericCdr) GetExtraFields() map[string]string {
	return nil
}
