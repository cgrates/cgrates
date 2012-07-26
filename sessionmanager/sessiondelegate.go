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
	"github.com/cgrates/cgrates/timespans"
	"log"
	"net/rpc"
	"time"
)

const (
	DEBIT_PERIOD = 10 * time.Second
)

type Connector interface {
	GetCost(timespans.CallDescriptor, *timespans.CallCost) error
	Debit(timespans.CallDescriptor, *timespans.CallCost) error
	DebitCents(timespans.CallDescriptor, *float64) error
	DebitSeconds(timespans.CallDescriptor, *float64) error
	GetMaxSessionTime(timespans.CallDescriptor, *float64) error
}

type RPCClientConnector struct {
	Client *rpc.Client
}

func (rcc *RPCClientConnector) GetCost(cd timespans.CallDescriptor, cc *timespans.CallCost) error {
	return rcc.Client.Call("Responder.GetCost", cd, cc)
}

func (rcc *RPCClientConnector) Debit(cd timespans.CallDescriptor, cc *timespans.CallCost) error {
	return rcc.Client.Call("Responder.Debit", cd, cc)
}
func (rcc *RPCClientConnector) DebitCents(cd timespans.CallDescriptor, resp *float64) error {
	return rcc.Client.Call("Responder.DebitCents", cd, resp)
}
func (rcc *RPCClientConnector) DebitSeconds(cd timespans.CallDescriptor, resp *float64) error {
	return rcc.Client.Call("Responder.DebitSeconds", cd, resp)
}
func (rcc *RPCClientConnector) GetMaxSessionTime(cd timespans.CallDescriptor, resp *float64) error {
	return rcc.Client.Call("Responder.GetMaxSessionTime", cd, resp)
}

// Sample SessionDelegate calling the timespans methods through the RPC interface
type SessionDelegate struct {
	Connector Connector
}

func (rsd *SessionDelegate) OnHeartBeat(ev Event) {
	log.Print("freeswitch ♥")
}

func (rsd *SessionDelegate) OnChannelAnswer(ev Event, s *Session) {
	log.Print("freeswitch answer")
}

func (rsd *SessionDelegate) OnChannelHangupComplete(ev Event, s *Session) {
	lastCC := s.CallCosts[len(s.CallCosts)-1]
	// put credit back	
	start := time.Now()
	end := lastCC.Timespans[len(lastCC.Timespans)-1].TimeEnd
	refoundDuration := end.Sub(start).Seconds()
	cost := 0.0
	seconds := 0.0
	log.Printf("Refund duration: %v", refoundDuration)
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
		cd := &timespans.CallDescriptor{
			Direction:   lastCC.Direction,
			Tenant:      lastCC.Tenant,
			TOR:         lastCC.TOR,
			Subject:     lastCC.Subject,
			Account:     lastCC.Account,
			Destination: lastCC.Destination,
			Amount:      -cost,
		}
		var response float64
		err := rsd.Connector.DebitCents(*cd, &response)
		if err != nil {
			log.Printf("Debit cents failed: %v", err)
		}
	}
	if seconds > 0 {
		cd := &timespans.CallDescriptor{
			TOR:         lastCC.TOR,
			Tenant:      lastCC.Tenant,
			Subject:     lastCC.Subject,
			Account:     lastCC.Account,
			Destination: lastCC.Destination,
			Amount:      -seconds,
		}
		var response float64
		err := rsd.Connector.DebitSeconds(*cd, &response)
		if err != nil {
			log.Printf("Debit seconds failed: %v", err)
		}
	}
	lastCC.Cost -= cost
	log.Printf("Rambursed %v cents, %v seconds", cost, seconds)
}

func (rsd *SessionDelegate) LoopAction(s *Session, cd *timespans.CallDescriptor) {
	cc := &timespans.CallCost{}
	err := rsd.Connector.Debit(*cd, cc)
	if err != nil {
		log.Printf("Could not complete debit opperation: %v", err)
	}
	s.CallCosts = append(s.CallCosts, cc)
	log.Print(cc)
	cd.Amount = DEBIT_PERIOD.Seconds()
	var remainingSeconds float64
	err = rsd.Connector.GetMaxSessionTime(*cd, &remainingSeconds)
	if err != nil {
		log.Printf("Could not get max session time: %v", err)
	}
	if remainingSeconds == -1 && err == nil {
		log.Print("Postpaying client: happy talking!")
		return
	}
	if remainingSeconds == 0 || err != nil {
		log.Printf("No credit left: Disconnect %v", s)
		s.Disconnect()
		return
	}
	if remainingSeconds < DEBIT_PERIOD.Seconds() || err != nil {
		log.Printf("Not enough money for another debit period %v", s)
		s.Disconnect()
		return
	}
}
func (rsd *SessionDelegate) GetDebitPeriod() time.Duration {
	return DEBIT_PERIOD
}
