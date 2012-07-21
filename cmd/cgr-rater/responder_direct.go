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
	"fmt"
	"github.com/cgrates/cgrates/timespans"
	"os"
	"runtime"
)

type DirectResponder struct {
	sg timespans.StorageGetter
}

/*
RPC method providing the rating information from the storage.
*/
func (s *DirectResponder) GetCost(cd timespans.CallDescriptor, reply *timespans.CallCost) (err error) {
	r, e := timespans.AccLock.GuardGetCost(cd.GetUserBalanceKey(), func() (*timespans.CallCost, error) {
		return (&cd).GetCost()
	})
	*reply, err = *r, e
	return err
}

func (s *DirectResponder) Debit(cd timespans.CallDescriptor, reply *timespans.CallCost) (err error) {
	r, e := timespans.AccLock.GuardGetCost(cd.GetUserBalanceKey(), func() (*timespans.CallCost, error) {
		return (&cd).Debit()
	})
	*reply, err = *r, e
	return err
}

func (s *DirectResponder) DebitCents(cd timespans.CallDescriptor, reply *float64) (err error) {
	r, e := timespans.AccLock.Guard(cd.GetUserBalanceKey(), func() (float64, error) {
		return (&cd).DebitCents()
	})
	*reply, err = r, e
	return err
}

func (s *DirectResponder) DebitSMS(cd timespans.CallDescriptor, reply *float64) (err error) {
	r, e := timespans.AccLock.Guard(cd.GetUserBalanceKey(), func() (float64, error) {
		return (&cd).DebitSMS()
	})
	*reply, err = r, e
	return err
}

func (s *DirectResponder) DebitSeconds(cd timespans.CallDescriptor, reply *float64) (err error) {
	r, e := timespans.AccLock.Guard(cd.GetUserBalanceKey(), func() (float64, error) {
		return 0, (&cd).DebitSeconds()
	})
	*reply, err = r, e
	return err
}

func (s *DirectResponder) GetMaxSessionTime(cd timespans.CallDescriptor, reply *float64) (err error) {
	r, e := timespans.AccLock.Guard(cd.GetUserBalanceKey(), func() (float64, error) {
		return (&cd).GetMaxSessionTime()
	})
	*reply, err = r, e
	return err
}

func (s *DirectResponder) AddRecievedCallSeconds(cd timespans.CallDescriptor, reply *float64) (err error) {
	r, e := timespans.AccLock.Guard(cd.GetUserBalanceKey(), func() (float64, error) {
		return 0, (&cd).AddRecievedCallSeconds()
	})
	*reply, err = r, e
	return err
}

func (r *DirectResponder) Status(arg timespans.CallDescriptor, replay *string) (err error) {
	memstats := new(runtime.MemStats)
	runtime.ReadMemStats(memstats)
	*replay = fmt.Sprintf("memstats before GC: %dKb footprint: %dKb", memstats.HeapAlloc/1024, memstats.Sys/1024)
	return
}

/*
RPC method that triggers rater shutdown in case of balancer exit.
*/
func (s *DirectResponder) Shutdown(args string, reply *string) (err error) {
	s.sg.Close()
	defer os.Exit(0)
	*reply = "Done!"
	return nil
}
