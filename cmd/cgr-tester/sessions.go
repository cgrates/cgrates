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

package main

import (
	"fmt"
	"time"

	"github.com/cenkalti/rpc2"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	brpc             *rpc2.Client
	disconnectEvChan = make(chan *utils.AttrDisconnectSession, 1)
)

func handleDisconnectSession(clnt *rpc2.Client,
	args *utils.AttrDisconnectSession, reply *string) error {
	disconnectEvChan <- args
	*reply = utils.OK
	return nil
}

func callSession() (err error) {

	event := &utils.CGREvent{
		Tenant: *tenant,
		ID:     "TheEventID100000",
		Time:   utils.TimePointer(time.Now()),
		Event: map[string]any{
			utils.AccountField: *subject,
			utils.Destination:  *destination,
			utils.OriginHost:   utils.Local,
			utils.RequestType:  *requestType,
			utils.Source:       utils.CGRTester,
			utils.OriginID:     utils.GenUUID(),
		},
		APIOpts: map[string]any{},
	}

	if *updateInterval > *maxUsage {
		return fmt.Errorf(`"update_interval" should be smaller than "max_usage"`)
	} else if *maxUsage <= *minUsage {
		return fmt.Errorf(`"min_usage" should be smaller than "max_usage"`)
	}

	clntHandlers := map[string]any{utils.SessionSv1DisconnectSession: handleDisconnectSession}
	brpc, err = utils.NewBiJSONrpcClient(tstCfg.SessionSCfg().ListenBijson, clntHandlers)
	if err != nil {
		return
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
	if err = brpc.Call(utils.SessionSv1AuthorizeEvent, authArgs, &authRply); err != nil {
		return
	}

	// Delay between authorize and initiation for a more realistic case
	time.Sleep(time.Duration(utils.RandomInteger(1, 5)) * time.Second)

	//
	// SessionSv1InitiateSession
	//
	event.Event[utils.AnswerTime] = time.Now()

	initArgs := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent:    event,
	}

	var initRply sessions.V1InitSessionReply
	if err = brpc.Call(utils.SessionSv1InitiateSession, initArgs, &initRply); err != nil {
		return
	}

	//
	// SessionSv1UpdateSession
	//
	totalUsage := time.Duration(utils.RandomInteger(int64(*minUsage), int64(*maxUsage)))
	for currentUsage := time.Duration(0); currentUsage < totalUsage; currentUsage += *updateInterval {

		event.Event[utils.Usage] = currentUsage.String()
		upArgs := &sessions.V1UpdateSessionArgs{
			GetAttributes: true,
			UpdateSession: true,
			CGREvent:      event,
		}
		var upRply sessions.V1UpdateSessionReply
		if err = brpc.Call(utils.SessionSv1UpdateSession, upArgs, &upRply); err != nil {
			return
		}
	}

	// Delay between last update and termination for a more realistic case
	time.Sleep(time.Duration(utils.RandomInteger(10, 20)) * time.Millisecond)

	//
	// SessionSv1TerminateSession
	//
	event.Event[utils.Usage] = totalUsage.String()

	tArgs := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent:         event,
	}
	var tRply string
	if err = brpc.Call(utils.SessionSv1TerminateSession, tArgs, &tRply); err != nil {
		return
	}

	// Delay between terminate and processCDR for a more realistic case
	time.Sleep(time.Duration(utils.RandomInteger(1, 3)) * time.Millisecond)

	//
	// SessionSv1ProcessCDR
	//
	procArgs := event
	var pRply string
	if err = brpc.Call(utils.SessionSv1ProcessCDR, procArgs, &pRply); err != nil {
		return
	}

	return
}
