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
package config

import (
	"github.com/cgrates/cgrates/utils"
)

type CDRReplicationCfg struct {
	Transport     string
	Address       string
	Synchronous   bool
	Attempts      int             // Number of attempts if not success
	CdrFilter     utils.RSRFields // Only replicate if the filters here are matching
	ContentFields []*CfgCdrField
}

func (rplCfg CDRReplicationCfg) FallbackFileName() string {
	fileSuffix := ".txt"
	switch rplCfg.Transport {
	case utils.MetaHTTPjsonCDR, utils.MetaHTTPjsonMap, utils.MetaAMQPjsonCDR, utils.MetaAMQPjsonMap:
		fileSuffix = ".json"
	case utils.META_HTTP_POST:
		fileSuffix = ".form"
	}
	ffn := &utils.FallbackFileName{Module: "cdr", Transport: rplCfg.Transport, Address: rplCfg.Address,
		RequestID: utils.GenUUID(), FileSuffix: fileSuffix}
	return ffn.AsString()
}
