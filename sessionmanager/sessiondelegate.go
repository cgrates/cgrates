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
	"log"
)

type SessionDelegate interface {
	OnHeartBeat(*Event)
	OnChannelAnswer(*Event, *Session)
	OnChannelHangupComplete(*Event, *Session)
	LoopAction()
}

type DirectSessionDelegate byte

func (dsd *DirectSessionDelegate) OnHeartBeat(ev *Event) {
	log.Print("direct hearbeat")
}

func (dsd *DirectSessionDelegate) OnChannelAnswer(ev *Event, s *Session) {
	log.Print("direct answer")
}

func (dsd *DirectSessionDelegate) OnChannelHangupComplete(ev *Event, s *Session) {
	log.Print("direct hangup")
}

func (dsd *DirectSessionDelegate) LoopAction() {
	log.Print("Direct debit")
}

// 
type RPCSessionDelegate byte

func (rsd *RPCSessionDelegate) OnHeartBeat(ev *Event) {
	log.Print("rpc hearbeat")
}

func (rsd *RPCSessionDelegate) OnChannelAnswer(ev *Event, s *Session) {
	log.Print("rpc answer")
}

func (rsd *RPCSessionDelegate) OnChannelHangupComplete(ev *Event, s *Session) {
	log.Print("rpc hangup")
}

func (rsd *RPCSessionDelegate) LoopAction() {
	log.Print("Rpc debit")
}
