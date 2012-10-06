/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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

package sessionmanager

import (
	"bufio"
	"fmt"
	"github.com/cgrates/cgrates/fsock"
	"github.com/cgrates/cgrates/rater"
	"net"
	"strings"
	"time"
)

// The freeswitch session manager type holding a buffer for the network connection
// and the active sessions
type FSSessionManager struct {
	conn        net.Conn
	buf         *bufio.Reader
	sessions    []*Session
	connector   rater.Connector
	debitPeriod time.Duration
	loggerDB    rater.DataStorage
	delayFunc   func() int
}

func NewFSSessionManager(storage rater.DataStorage, connector rater.Connector, debitPeriod time.Duration) *FSSessionManager {
	return &FSSessionManager{loggerDB: storage}
}

// Connects to the freeswitch mod_event_socket server and starts
// listening for events in json format.
func (sm *FSSessionManager) Connect(address, pass string) error {
	sm.address = address
	sm.pass = pass
	if sm.conn != nil {
		// in case it is a reconnect
		sm.conn.Close()
	}
	conn, err := net.Dial("tcp", address)
	if err != nil {
		rater.Logger.Warning("Could not connect to freeswitch server!")
		return err
	}
	sm.conn = conn
	sm.buf = bufio.NewReaderSize(conn, 8192)
	fmt.Fprint(conn, fmt.Sprintf("auth %s\n\n", pass))
	fmt.Fprint(conn, "event json  HEARTBEAT CHANNEL_ANSWER CHANNEL_HANGUP_COMPLETE CHANNEL_PARK\n\n")
	fmt.Fprint(conn, "filter Call-Direction inbound\n\n")

	handlers := make(map[string]func(string))
	handlers["HEARTBEAT"] = func(s string) { fmt.Println(s) } // Example handler
	if officer.FS, err = officer.NewFSock(config.FS_SOCK, config.FS_PASWD, 3, handlers); err != nil {
		fmt.Println("FreeSWITCH error:", err)
		exitChan <- true
		return
	} else if officer.FS.Connected() {
		fmt.Println("Successfully connected to FreeSWITCH")
	}
	officer.FS.ReadEvents()
	exitChan <- true // If we have reached here something went wrong

	go func() {
		sm.delayFunc = fib()
		exitChan := make(chan bool)
		for {
			select {
			case <-exitChan:
				break
			default:
				sm.readNextEvent(exitChan)
			}
		}
	}()
	return nil
}

// Reads from freeswitch server buffer until it encounters a '}',
// than it creates an event object and calls the appropriate method
func (sm *FSSessionManager) readNextEvent(exitChan chan bool) (ev Event) {
	body, err := sm.buf.ReadString('}')
	if err != nil {
		rater.Logger.Warning("Could not read from freeswitch connection!")
		// wait until a sec
		time.Sleep(time.Duration(sm.delayFunc()) * time.Second)
		// try to reconnect
		err = sm.Connect(sm.address, sm.pass)
		if err == nil {
			rater.Logger.Info("Successfuly reconnected to freeswitch! ")
			exitChan <- true
		}
	}
	ev = new(FSEvent).New(body)
	switch ev.GetName() {
	case HEARTBEAT:
		sm.OnHeartBeat(ev)
	case PARK:
		sm.OnChannelPark(ev)
	case ANSWER:
		sm.OnChannelAnswer(ev)
	case HANGUP:
		sm.OnChannelHangupComplete(ev)
	}
	return
}

// Searches and return the session with the specifed uuid
func (sm *FSSessionManager) GetSession(uuid string) *Session {
	for _, s := range sm.sessions {
		if s.uuid == uuid {
			return s
		}
	}
	return nil
}

// Disconnects a session by sending hangup command to freeswitch
func (sm *FSSessionManager) DisconnectSession(s *Session, notify string) {
	fmt.Fprint(sm.conn, fmt.Sprintf("api uuid_setvar %s cgr_notify %s\n\n", s.uuid, notify))
	fmt.Fprint(sm.conn, fmt.Sprintf("SendMsg %s\ncall-command: hangup\nhangup-cause: MANAGER_REQUEST\n\n", s.uuid))
	s.Close()
}

// Sends the transfer command to unpark the call to freeswitch
func (sm *FSSessionManager) UnparkCall(uuid, call_dest_nb, notify string) {
	fmt.Fprint(sm.conn, fmt.Sprintf("api uuid_setvar %s cgr_notify %s\n\n", uuid, notify))
	fmt.Fprint(sm.conn, fmt.Sprintf("api uuid_transfer %s %s\n\n", uuid, call_dest_nb))
}

func (sm *FSSessionManager) OnHeartBeat(ev Event) {
	rater.Logger.Info("freeswitch ♥")
}

func (sm *FSSessionManager) OnChannelPark(ev Event) {
	rater.Logger.Info("freeswitch park")
	startTime, err := ev.GetStartTime(PARK_TIME)
	if err != nil {
		rater.Logger.Err("Error parsing answer event start time, using time.Now!")
		startTime = time.Now()
	}
	// if there is no account configured leave the call alone
	if strings.TrimSpace(ev.GetReqType()) != REQTYPE_PREPAID {
		return
	}
	if ev.MissingParameter() {
		sm.UnparkCall(ev.GetUUID(), ev.GetCallDestNb(), MISSING_PARAMETER)
		rater.Logger.Err(fmt.Sprintf("Missing parameter for %s", ev.GetUUID()))
		return
	}
	cd := rater.CallDescriptor{
		Direction:   ev.GetDirection(),
		Tenant:      ev.GetTenant(),
		TOR:         ev.GetTOR(),
		Subject:     ev.GetSubject(),
		Account:     ev.GetAccount(),
		Destination: ev.GetDestination(),
		Amount:      sm.debitPeriod.Seconds(),
		TimeStart:   startTime}
	var remainingSeconds float64
	err = sm.connector.GetMaxSessionTime(cd, &remainingSeconds)
	if err != nil {
		rater.Logger.Err(fmt.Sprintf("Could not get max session time for %s: %v", ev.GetUUID(), err))
		sm.UnparkCall(ev.GetUUID(), ev.GetCallDestNb(), SYSTEM_ERROR)
		return
	}
	rater.Logger.Info(fmt.Sprintf("Remaining seconds: %v", remainingSeconds))
	if remainingSeconds == 0 {
		rater.Logger.Info(fmt.Sprintf("Not enough credit for trasferring the call %s for %s.", ev.GetUUID(), cd.GetKey()))
		sm.UnparkCall(ev.GetUUID(), ev.GetCallDestNb(), INSUFFICIENT_FUNDS)
		return
	}
	sm.UnparkCall(ev.GetUUID(), ev.GetCallDestNb(), AUTH_OK)
}

func (sm *FSSessionManager) OnChannelAnswer(ev Event) {
	rater.Logger.Info("freeswitch answer")
}

func (sm *FSSessionManager) OnChannelHangupComplete(ev Event) {
	s := sm.GetSession(ev.GetUUID())
	if ev.GetReqType() == REQTYPE_POSTPAID {
		startTime, err := ev.GetStartTime(START_TIME)
		if err != nil {
			rater.Logger.Crit("Error parsing postpaid call start time from event")
			return
		}
		endTime, err := ev.GetEndTime()
		if err != nil {
			rater.Logger.Crit("Error parsing postpaid call start time from event")
			return
		}
		cd := rater.CallDescriptor{
			Direction:   ev.GetDirection(),
			Tenant:      ev.GetTenant(),
			TOR:         ev.GetTOR(),
			Subject:     ev.GetSubject(),
			Account:     ev.GetAccount(),
			Destination: ev.GetDestination(),
			TimeStart:   startTime,
			TimeEnd:     endTime,
		}
		cc := &rater.CallCost{}
		err = sm.connector.Debit(cd, cc)
		if err != nil {
			rater.Logger.Err(fmt.Sprintf("Error making the general debit for postpaid call: %v", ev.GetUUID()))
			return
		}
		s.CallCosts = append(s.CallCosts, cc)
		return
	}

	if s == nil || len(s.CallCosts) == 0 {
		return // why would we have 0 callcosts
	}
	lastCC := s.CallCosts[len(s.CallCosts)-1]
	// put credit back	
	start := time.Now()
	end := lastCC.Timespans[len(lastCC.Timespans)-1].TimeEnd
	refoundDuration := end.Sub(start).Seconds()
	cost := 0.0
	seconds := 0.0
	rater.Logger.Info(fmt.Sprintf("Refund duration: %v", refoundDuration))
	for i := len(lastCC.Timespans) - 1; i >= 0; i-- {
		ts := lastCC.Timespans[i]
		tsDuration := ts.GetDuration().Seconds()
		if refoundDuration <= tsDuration {
			// find procentage
			procentage := (refoundDuration * 100) / tsDuration
			tmpCost := (procentage * ts.Cost) / 100
			ts.Cost -= tmpCost
			cost += tmpCost
			if ts.MinuteInfo != nil {
				// DestinationPrefix and Price take from lastCC and above caclulus
				seconds += (procentage * ts.MinuteInfo.Quantity) / 100
			}
			// set the end time to now
			ts.TimeEnd = start
			break // do not go to other timespans
		} else {
			cost += ts.Cost
			if ts.MinuteInfo != nil {
				seconds += ts.MinuteInfo.Quantity
			}
			// remove the timestamp entirely
			lastCC.Timespans = lastCC.Timespans[:i]
			// continue to the next timespan with what is left to refound
			refoundDuration -= tsDuration
		}
	}
	if cost > 0 {
		cd := &rater.CallDescriptor{
			Direction:   lastCC.Direction,
			Tenant:      lastCC.Tenant,
			TOR:         lastCC.TOR,
			Subject:     lastCC.Subject,
			Account:     lastCC.Account,
			Destination: lastCC.Destination,
			Amount:      -cost,
		}
		var response float64
		err := sm.connector.DebitCents(*cd, &response)
		if err != nil {
			rater.Logger.Err(fmt.Sprintf("Debit cents failed: %v", err))
		}
	}
	if seconds > 0 {
		cd := &rater.CallDescriptor{
			Direction:   lastCC.Direction,
			TOR:         lastCC.TOR,
			Tenant:      lastCC.Tenant,
			Subject:     lastCC.Subject,
			Account:     lastCC.Account,
			Destination: lastCC.Destination,
			Amount:      -seconds,
		}
		var response float64
		err := sm.connector.DebitSeconds(*cd, &response)
		if err != nil {
			rater.Logger.Err(fmt.Sprintf("Debit seconds failed: %v", err))
		}
	}
	lastCC.Cost -= cost
	rater.Logger.Info(fmt.Sprintf("Rambursed %v cents, %v seconds", cost, seconds))
}

func (sm *FSSessionManager) LoopAction(s *Session, cd *rater.CallDescriptor) {
	cc := &rater.CallCost{}
	cd.Amount = sm.debitPeriod.Seconds()
	err := sm.connector.MaxDebit(*cd, cc)
	if err != nil {
		rater.Logger.Err(fmt.Sprintf("Could not complete debit opperation: %v", err))
		// disconnect session
		s.sessionManager.DisconnectSession(s, SYSTEM_ERROR)
	}
	nbts := len(cc.Timespans)
	remainingSeconds := 0.0
	rater.Logger.Debug(fmt.Sprintf("Result of MaxDebit call: %v", cc))
	if nbts > 0 {
		remainingSeconds = cc.Timespans[nbts-1].TimeEnd.Sub(cc.Timespans[0].TimeStart).Seconds()
	}
	if remainingSeconds == 0 || err != nil {
		rater.Logger.Info(fmt.Sprintf("No credit left: Disconnect %v", s))
		s.Disconnect()
		return
	}
	s.CallCosts = append(s.CallCosts, cc)
}
func (sm *FSSessionManager) GetDebitPeriod() time.Duration {
	return sm.debitPeriod
}
func (sm *FSSessionManager) GetDbLogger() rater.DataStorage {
	return sm.loggerDB
}

// successive Fibonacci numbers.
func fib() func() int {
	a, b := 0, 1
	return func() int {
		a, b = b, a+b
		return a
	}
}
