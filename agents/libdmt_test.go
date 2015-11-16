/*
Real-time Charging System for Telecom & ISP environments
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

package agents

import (
	"testing"
	"time"
)

func TestDisectUsageForCCR(t *testing.T) {
	if reqType, reqNr, ccTime := disectUsageForCCR(time.Duration(0)*time.Second, time.Duration(300)*time.Second, false); reqType != 1 || reqNr != 0 || ccTime != 300 {
		t.Error(reqType, reqNr, ccTime)
	}
	if reqType, reqNr, ccTime := disectUsageForCCR(time.Duration(35)*time.Second, time.Duration(300)*time.Second, false); reqType != 2 || reqNr != 0 || ccTime != 300 {
		t.Error(reqType, reqNr, ccTime)
	}
	if reqType, reqNr, ccTime := disectUsageForCCR(time.Duration(935)*time.Second, time.Duration(300)*time.Second, false); reqType != 2 || reqNr != 3 || ccTime != 300 {
		t.Error(reqType, reqNr, ccTime)
	}
	if reqType, reqNr, ccTime := disectUsageForCCR(time.Duration(35)*time.Second, time.Duration(300)*time.Second, true); reqType != 3 || reqNr != 1 || ccTime != 35 {
		t.Error(reqType, reqNr, ccTime)
	}
	if reqType, reqNr, ccTime := disectUsageForCCR(time.Duration(935)*time.Second, time.Duration(300)*time.Second, true); reqType != 3 || reqNr != 4 || ccTime != 35 {
		t.Error(reqType, reqNr, ccTime)
	}

}

func TestGetUsageFromCCR(t *testing.T) {
	if usage := getUsageFromCCR(1, 0, 300, time.Duration(300)*time.Second); usage != time.Duration(300)*time.Second {
		t.Error(usage)
	}
	if usage := getUsageFromCCR(2, 0, 300, time.Duration(300)*time.Second); usage != time.Duration(300)*time.Second {
		t.Error(usage)
	}
	if usage := getUsageFromCCR(2, 3, 300, time.Duration(300)*time.Second); usage != time.Duration(1200)*time.Second {
		t.Error(usage)
	}
	if usage := getUsageFromCCR(3, 4, 35, time.Duration(300)*time.Second); usage != time.Duration(935)*time.Second {
		t.Error(usage)
	}
	if usage := getUsageFromCCR(3, 1, 35, time.Duration(300)*time.Second); usage != time.Duration(35)*time.Second {
		t.Error(usage)
	}
}
