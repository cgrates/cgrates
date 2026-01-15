/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package main

import (
	"fmt"
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	brpc             *birpc.BirpcClient
	disconnectEvChan = make(chan utils.CGREvent, 1)
)

type smock struct{}

func (*smock) DisconnectSession(ctx *context.Context,
	args utils.CGREvent, reply *string) error {
	disconnectEvChan <- args
	*reply = utils.OK
	return nil
}

func callSessions(ctx *context.Context, authDur, initDur, updateDur, terminateDur, cdrDur *[]time.Duration,
	reqAuth, reqInit, reqUpdate, reqTerminate, reqCdr *uint64,
	digitMin, digitMax int64, totalUsage time.Duration) (err error) {

	if *digits <= 0 {
		return fmt.Errorf(`"digits" should be bigger than 0`)
	} else if int(math.Pow10(*digits))-1 < *cps {
		return fmt.Errorf(`"digits" should amount to be more than "cps"`)
	}
	var appendMu sync.Mutex
	acc := utils.RandomInteger(digitMin, digitMax)
	dest := utils.RandomInteger(digitMin, digitMax)

	event := &utils.CGREvent{
		Tenant: *tenant,
		ID:     "EventID1",
		Time:   utils.TimePointer(time.Now()),
		Event: map[string]any{
			utils.AccountField: acc,
			utils.Destination:  dest,
			utils.OriginHost:   utils.Local,
			utils.RequestType:  *requestType,
			utils.Source:       utils.CGRTester,
			utils.OriginID:     utils.GenUUID(),
		},
		APIOpts: map[string]any{},
	}

	srv, err := birpc.NewService(new(smock), utils.AgentV1, true)
	if err != nil {
		return err
	}
	brpc, err = utils.NewBiJSONrpcClient(tstCfg.ListenCfg().BiJSONListen, srv)
	if err != nil {
		return err
	}

	//
	// SessionSv1AuthorizeEvent
	//

	event.Event[utils.SetupTime] = time.Now()
	authArgs := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent:    event,
	}
	var authRply sessions.V1AuthorizeReply
	atomic.AddUint64(reqAuth, 1)
	authStartTime := time.Now()
	if err = brpc.Call(ctx, utils.SessionSv1AuthorizeEvent, authArgs, &authRply); err != nil {
		return
	}
	appendMu.Lock()
	*authDur = append(*authDur, time.Since(authStartTime))
	appendMu.Unlock()
	if *verbose {
		log.Printf("Account: <%v>, Destination: <%v>, SessionSv1AuthorizeEvent reply: <%v>", acc, dest, utils.ToJSON(authRply))
	}

	// Delay between authorize and initiation for a more realistic case
	time.Sleep(time.Duration(utils.RandomInteger(50, 100)) * time.Millisecond)

	//
	// SessionSv1InitiateSession
	//
	event.Event[utils.AnswerTime] = time.Now()
	initArgs := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent:    event,
	}

	var initRply sessions.V1InitSessionReply
	atomic.AddUint64(reqInit, 1)
	initStartTime := time.Now()
	if err = brpc.Call(ctx, utils.SessionSv1InitiateSession, initArgs, &initRply); err != nil {
		return
	}
	appendMu.Lock()
	*initDur = append(*initDur, time.Since(initStartTime))
	appendMu.Unlock()
	if *verbose {
		log.Printf("Account: <%v>, Destination: <%v>, SessionSv1InitiateSession reply: <%v>", acc, dest, utils.ToJSON(initRply))
	}

	//
	// SessionSv1UpdateSession
	//
	var currentUsage time.Duration
	for currentUsage = time.Duration(1 * time.Second); currentUsage < totalUsage; currentUsage += *updateInterval {

		time.Sleep(*updateInterval)

		event.Event[utils.Usage] = currentUsage.String()
		upArgs := &sessions.V1UpdateSessionArgs{
			UpdateSession: true,
			CGREvent:      event,
		}
		var upRply sessions.V1UpdateSessionReply
		atomic.AddUint64(reqUpdate, 1)
		updateStartTime := time.Now()
		if err = brpc.Call(ctx, utils.SessionSv1UpdateSession, upArgs, &upRply); err != nil {
			return
		}
		appendMu.Lock()
		*updateDur = append(*updateDur, time.Since(updateStartTime))
		appendMu.Unlock()
		if *verbose {
			log.Printf("Account: <%v>, Destination: <%v>, SessionSv1UpdateSession reply: <%v>", acc, dest, utils.ToJSON(upRply))
		}
	}

	// Delay between last update and termination for a more realistic case
	time.Sleep(totalUsage - currentUsage)

	//
	// SessionSv1TerminateSession
	//
	event.Event[utils.Usage] = totalUsage.String()

	tArgs := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent:         event,
	}
	var tRply string
	atomic.AddUint64(reqTerminate, 1)
	terminateStartTime := time.Now()
	if err = brpc.Call(ctx, utils.SessionSv1TerminateSession, tArgs, &tRply); err != nil {
		return
	}
	appendMu.Lock()
	*terminateDur = append(*terminateDur, time.Since(terminateStartTime))
	appendMu.Unlock()
	if *verbose {
		log.Printf("Account: <%v>, Destination: <%v>, SessionSv1TerminateSession reply: <%v>", acc, dest, utils.ToJSON(tRply))
	}

	// Delay between terminate and processCDR for a more realistic case
	time.Sleep(time.Duration(utils.RandomInteger(20, 40)) * time.Millisecond)

	//
	// SessionSv1ProcessCDR
	//
	procArgs := event
	var pRply string
	atomic.AddUint64(reqCdr, 1)
	cdrStartTime := time.Now()
	if err = brpc.Call(ctx, utils.SessionSv1ProcessCDR, procArgs, &pRply); err != nil {
		return
	}
	appendMu.Lock()
	*cdrDur = append(*cdrDur, time.Since(cdrStartTime))
	appendMu.Unlock()
	if *verbose {
		log.Printf("Account: <%v>, Destination: <%v>, SessionSv1ProcessCDR reply: <%v>", acc, dest, utils.ToJSON(pRply))
	}
	return
}
