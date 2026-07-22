// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

func (c *AnalyzerConnector) Call(ctx *context.Context, serviceMethod string, args, reply any) (err error) {
	sTime := time.Now()
	err = c.conn.Call(ctx, serviceMethod, args, reply)
	go c.aS.logTrafic(0, serviceMethod, args, reply, err, c.enc, c.from, c.to, sTime, time.Now())
	return
}
