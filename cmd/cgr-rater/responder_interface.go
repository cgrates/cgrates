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

package main

import (
	"github.com/cgrates/cgrates/timespans"
)

type ResponderInterface interface {
	GetCost(timespans.CallDescriptor, *timespans.CallCost) error
	Debit(timespans.CallDescriptor, *timespans.CallCost) error
	DebitCents(timespans.CallDescriptor, *float64) error
	DebitSMS(timespans.CallDescriptor, *float64) error
	DebitSeconds(timespans.CallDescriptor, *float64) error
	GetMaxSessionTime(timespans.CallDescriptor, *float64) error
	AddRecievedCallSeconds(timespans.CallDescriptor, *float64) error
	Status(timespans.CallDescriptor, *string) error
	Shutdown(string, *string) error
}

type Responder struct {
	resp ResponderInterface
}

func (r *Responder) GetCost(cd timespans.CallDescriptor, reply *timespans.CallCost) error {
	return r.resp.GetCost(cd, reply)
}

func (r *Responder) Debit(cd timespans.CallDescriptor, reply *timespans.CallCost) error {
	return r.resp.Debit(cd, reply)
}

func (r *Responder) DebitCents(cd timespans.CallDescriptor, reply *float64) error {
	return r.resp.DebitCents(cd, reply)
}

func (r *Responder) DebitSMS(cd timespans.CallDescriptor, reply *float64) error {
	return r.resp.DebitSMS(cd, reply)
}

func (r *Responder) DebitSeconds(cd timespans.CallDescriptor, reply *float64) error {
	return r.resp.DebitSeconds(cd, reply)
}

func (r *Responder) GetMaxSessionTime(cd timespans.CallDescriptor, reply *float64) error {
	return r.resp.GetMaxSessionTime(cd, reply)
}

func (r *Responder) AddRecievedCallSeconds(cd timespans.CallDescriptor, reply *float64) error {
	return r.resp.AddRecievedCallSeconds(cd, reply)
}

func (r *Responder) Status(cd timespans.CallDescriptor, reply *string) error {
	return r.resp.Status(cd, reply)
}

func (r *Responder) Shutdown(args string, reply *string) error {
	return r.resp.Shutdown(args, reply)
}
