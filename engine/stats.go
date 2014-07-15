/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package engine

import (
	"errors"
	"sync"

	"github.com/cgrates/cgrates/utils"
)

type Stats struct {
	queues map[string]*StatsQueue
	mux    sync.RWMutex
}

func (s *Stats) AddQueue(sq *StatsQueue) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.queues[sq.conf.Id] = sq
}

func (s *Stats) GetValues(sqID string) (map[string]float64, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if sq, ok := s.queues[sqID]; ok {
		return sq.GetStats(), nil
	}
	return nil, errors.New("Not Found")
}

func (s *Stats) AppendCDR(cdr *utils.StoredCdr) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	for _, sq := range s.queues {
		sq.AppendCDR(cdr)
	}
}
