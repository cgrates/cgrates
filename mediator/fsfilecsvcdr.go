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

package mediator

import (
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"strconv"
	"time"
)

type FScsvCDR struct {
	rowData []string // The original row extracted form csv file
	accIdIdx,
	subjectIdx,
	reqtypeIdx,
	directionIdx,
	tenantIdx,
	torIdx,
	accountIdx,
	destinationIdx,
	answerTimeIdx,
	durationIdx int // Field indexes
	cgrCfg *config.CGRConfig // CGR Config instance
}

func NewFScsvCDR(cdrRow []string, accIdIdx, subjectIdx, reqtypeIdx, directionIdx, tenantIdx, torIdx,
	accountIdx, destinationIdx, answerTimeIdx, durationIdx int, cfg *config.CGRConfig) (*FScsvCDR, error) {
	fscdr := FScsvCDR{cdrRow, accIdIdx, subjectIdx, reqtypeIdx, directionIdx, tenantIdx,
		torIdx, accountIdx, destinationIdx, answerTimeIdx, durationIdx, cfg}
	return &fscdr, nil
}

func (self *FScsvCDR) GetCgrId() string {
	return utils.FSCgrId(self.rowData[self.accIdIdx])
}

func (self *FScsvCDR) GetAccId() string {
	return self.rowData[self.accIdIdx]
}

func (self *FScsvCDR) GetCdrHost() string {
	return utils.LOCALHOST // ToDo: Maybe extract dynamically the external IP address here
}

func (self *FScsvCDR) GetDirection() string {
	return "*out"
}

func (self *FScsvCDR) GetSubject() string {
	return self.rowData[self.subjectIdx]
}

func (self *FScsvCDR) GetAccount() string {
	return self.rowData[self.accountIdx]
}

func (self *FScsvCDR) GetDestination() string {
	return self.rowData[self.destinationIdx]
}

func (self *FScsvCDR) GetTOR() string {
	return self.rowData[self.torIdx]
}

func (self *FScsvCDR) GetTenant() string {
	return self.rowData[self.tenantIdx]
}

func (self *FScsvCDR) GetReqType() string {
	if self.reqtypeIdx == -1 {
		return self.cgrCfg.DefaultReqType
	}
	return self.rowData[self.reqtypeIdx]
}

func (self *FScsvCDR) GetAnswerTime() (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", self.rowData[self.answerTimeIdx])
}

func (self *FScsvCDR) GetDuration() int64 {
	dur, _ := strconv.ParseInt(self.rowData[self.durationIdx], 0, 64)
	return dur
}

func (self *FScsvCDR) GetExtraFields() map[string]string {
	return nil
}
