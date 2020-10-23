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

package analyzers

import (
	"time"

	"github.com/cgrates/rpcclient"
)

func (aS *AnalyzerService) NewAnalyzeConnector(sc rpcclient.ClientConnector, enc, from, to string) rpcclient.ClientConnector {
	return &AnalyzeConnector{
		conn: sc,
		aS:   aS,
		extrainfo: &extraInfo{
			enc:  enc,
			from: from,
			to:   to,
		},
	}
}

type AnalyzeConnector struct {
	conn rpcclient.ClientConnector

	aS        *AnalyzerService
	extrainfo *extraInfo
}

func (c *AnalyzeConnector) Call(serviceMethod string, args interface{}, reply interface{}) (err error) {
	sTime := time.Now()
	err = c.conn.Call(serviceMethod, args, reply)
	go c.aS.logTrafic(0, serviceMethod, args, reply, err, c.extrainfo, sTime, time.Now())
	return
}
