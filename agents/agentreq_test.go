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

package agents

import (
	"fmt"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestAgReqAsNavigableMap(t *testing.T) {
	data, _ := engine.NewMapStorage()
	dm := engine.NewDataManager(data)
	cfg, _ := config.NewDefaultCGRConfig()
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := newAgentRequest(nil, nil,
		"cgrates.org", filterS)
	agReq.CGRReply.Set([]string{utils.CGRID}, utils.Sha1("dsafdsaf",
		time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), false)
	agReq.CGRReply.Set([]string{utils.Account}, "1001", false)
	agReq.CGRReply.Set([]string{utils.Destination}, "1002", false)
	agReq.CGRReply.Set([]string{utils.AnswerTime},
		time.Date(2013, 12, 30, 15, 0, 1, 0, time.UTC), false)
	agReq.CGRReply.Set([]string{utils.RequestType}, utils.META_PREPAID, false)
	agReq.CGRReply.Set([]string{utils.Usage}, time.Duration(2*time.Minute), false)
	agReq.CGRReply.Set([]string{utils.Cost}, 1.2, false)
	fmt.Printf("CGRReply: %+v\n", agReq.CGRReply.AsMapStringInterface())
	fmt.Printf("agReq: %+v\n", agReq)
}
