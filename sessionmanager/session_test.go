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
	"testing"
	"time"
)

func TestSessionDurationSingle(t *testing.T) {
	s := &Session{}
	s.AddCallToSession("", time.Now())
	twoSeconds, _ := time.ParseDuration("2s")
	if d := s.GetSessionDurationFrom(time.Now().Add(twoSeconds)); d.Seconds() < 2 || d.Seconds() > 3 {
		t.Errorf("Wrong duration %v", d)
	}
}

func TestSessionDurationMultiple(t *testing.T) {
	s := &Session{}
	s.AddCallToSession("", time.Now())
	s.AddCallToSession("", time.Now())
	s.AddCallToSession("", time.Now())
	twoSeconds, _ := time.ParseDuration("2s")
	if d := s.GetSessionDurationFrom(time.Now().Add(twoSeconds)); d.Seconds() < 6 || d.Seconds() > 7 {
		t.Errorf("Wrong duration %v", d)
	}
}

func TestSessionCostSingle(t *testing.T) {
	s := &Session{customer: "vdf", subject: "rif"}
	s.AddCallToSession("0723", time.Now())
	twoSeconds, _ := time.ParseDuration("2s")
	if ccs, err := s.GetSessionCostFrom(time.Now().Add(twoSeconds)); err != nil {
		t.Errorf("Get cost returned error %v", err)
	} else {
		log.Print(ccs)
		for i, cc := range ccs {
			log.Printf("here: %d. %v", i, cc.Cost)
		}
	}
}

func TestSessionCostMultiple(t *testing.T) {
	s := &Session{customer: "vdf", subject: "rif"}
	s.AddCallToSession("0723", time.Now())
	s.AddCallToSession("0723", time.Now())
	s.AddCallToSession("0723", time.Now())
	twoSeconds, _ := time.ParseDuration("2s")
	if ccs, err := s.GetSessionCostFrom(time.Now().Add(twoSeconds)); err != nil {
		t.Errorf("Get cost returned error %v", err)
	} else {
		log.Print(ccs)
		for i, cc := range ccs {
			log.Printf("here: %d. %v", i, cc.Cost)
		}
	}
}
