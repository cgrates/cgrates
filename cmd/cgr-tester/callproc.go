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
	"log"
	"time"

	"github.com/cenkalti/rpc2"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	brpc             *rpc2.Client
	disconnectEvChan = make(chan *utils.AttrDisconnectSession, 1)
	cpsCounter       int
)

func handleDisconnectSession(clnt *rpc2.Client,
	args *utils.AttrDisconnectSession, reply *string) error {
	disconnectEvChan <- args
	*reply = utils.OK
	return nil
}

func callSession() (err error) {

	var currentUsage time.Duration

	st := time.Now().Add(2 * time.Second)
	at := time.Now().Add(10 * time.Second)

	event := &utils.CGREvent{
		Tenant: *tenant,
		ID:     "TheEventID100000",
		Time:   utils.TimePointer(time.Now()),
		Event: map[string]any{
			utils.AccountField: *subject,
			utils.Destination:  *destination,
			utils.OriginHost:   "local",
			utils.RequestType:  utils.MetaRated,
			utils.SetupTime:    st,
			utils.Source:       "cgr_tester",
			utils.OriginID:     utils.GenUUID(),
		},
		APIOpts: map[string]any{},
	}

	cpsCounter += 1
	log.Printf("current call number: %+v", cpsCounter)

	if *updateInterval > *maxUsage {
		return fmt.Errorf(`"update_interval" should be smaller than "max_usage"`)
	} else if *maxUsage <= *minUsage {
		return fmt.Errorf(`"min_usage" should be smaller than "max_usage"`)
	}

	tstCfg.SessionSCfg().DebitInterval = 0

	clntHandlers := map[string]any{utils.SessionSv1DisconnectSession: handleDisconnectSession}
	brpc, err = utils.NewBiJSONrpcClient(tstCfg.SessionSCfg().ListenBijson, clntHandlers)
	if err != nil {
		return
	}

	//
	// SessionSv1AuthorizeEvent
	//
	authArgs := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent:    event,
	}
	var authRply sessions.V1AuthorizeReply
	if err = brpc.Call(utils.SessionSv1AuthorizeEvent, authArgs, &authRply); err != nil {
		return
	}
	// log.Printf("auth: %+v", utils.ToJSON(authRply))

	//
	// SessionSv1InitiateSession
	//
	event.Event[utils.AnswerTime] = at
	event.Event[utils.RequestType] = utils.MetaRated

	initArgs := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent:    event,
	}

	var initRply sessions.V1InitSessionReply
	if err = brpc.Call(utils.SessionSv1InitiateSession, initArgs, &initRply); err != nil {
		return
	}
	// log.Printf("init: %+v.", utils.ToJSON(initRply))

	//
	// SessionSv1UpdateSession
	//
	totalUsage := time.Duration(utils.RandomInteger(int64(*minUsage), int64(*maxUsage)))
	log.Println("randUsage", totalUsage)
	for currentUsage < totalUsage {
		// log.Println("currentUsage", currentUsage)
		currentUsage += *updateInterval
		if currentUsage >= totalUsage {
			break
		}
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
		// log.Printf("update: %+v.", utils.ToJSON(upRply))
	}

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
	// log.Printf("terminate: %+v.", utils.ToJSON(tRply))

	//
	// SessionSv1ProcessCDR
	//
	procArgs := event
	var pRply string
	if err = brpc.Call(utils.SessionSv1ProcessCDR, procArgs, &pRply); err != nil {
		return
	}
	// log.Printf("process: %+v.", utils.ToJSON(pRply))

	return
}
