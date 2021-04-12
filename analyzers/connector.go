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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
)

func (aS *AnalyzerService) NewAnalyzerConnector(sc birpc.ClientConnector, enc, from, to string) birpc.ClientConnector {
	return &AnalyzerConnector{
		conn: sc,
		aS:   aS,
		enc:  enc,
		from: from,
		to:   to,
	}
}

type AnalyzerConnector struct {
	conn birpc.ClientConnector

	aS   *AnalyzerService
	enc  string
	from string
	to   string
}

func (c *AnalyzerConnector) Call(ctx *context.Context, serviceMethod string, args, reply interface{}) (err error) {
	sTime := time.Now()
	err = c.conn.Call(ctx, serviceMethod, args, reply)
	go c.aS.logTrafic(0, serviceMethod, args, reply, err, c.enc, c.from, c.to, sTime, time.Now())
	return
}
