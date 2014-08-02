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
	"fmt"
	"net/rpc"
	"sync"

	"github.com/cgrates/cgrates/utils"
)

type StatsInterface interface {
	GetValues(string, *map[string]float64) error
	GetQueueIds(int, *[]string) error
	AppendCDR(*utils.StoredCdr, *int) error
	AddQueue(*CdrStats, *int) error
	ReloadQueues([]string, *int) error
	ResetQueues([]string, *int) error
}

type Stats struct {
	queues   map[string]*StatsQueue
	mux      sync.RWMutex
	ratingDb RatingStorage
}

func NewStats(ratingDb RatingStorage) *Stats {
	cdrStats := &Stats{ratingDb: ratingDb}
	if css, err := ratingDb.GetAllCdrStats(); err == nil {
		cdrStats.UpdateQueues(css, nil)
	} else {
		Logger.Err(fmt.Sprintf("Cannot load cdr stats: %v", err))
	}
	return cdrStats
}

func (s *Stats) GetQueueIds(in int, ids *[]string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	result := make([]string, 0)
	for id, _ := range s.queues {
		result = append(result, id)
	}
	*ids = result
	return nil
}

func (s *Stats) GetValues(sqID string, values *map[string]float64) error {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if sq, ok := s.queues[sqID]; ok {
		*values = sq.GetStats()
		return nil
	}
	return errors.New("Not Found")
}

func (s *Stats) AddQueue(cs *CdrStats, out *int) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.queues == nil {
		s.queues = make(map[string]*StatsQueue)
	}
	if sq, exists := s.queues[cs.Id]; exists {
		sq.UpdateConf(cs)
	} else {
		s.queues[cs.Id] = NewStatsQueue(cs)
	}
	return nil
}

func (s *Stats) ReloadQueues(ids []string, out *int) error {
	if len(ids) == 0 {
		if css, err := s.ratingDb.GetAllCdrStats(); err == nil {
			s.UpdateQueues(css, nil)
		} else {
			return fmt.Errorf("Cannot load cdr stats: %v", err)
		}
	}
	for _, id := range ids {
		if cs, err := s.ratingDb.GetCdrStats(id); err == nil {
			s.AddQueue(cs, nil)
		} else {
			return err
		}
	}
	return nil
}

func (s *Stats) ResetQueues(ids []string, out *int) error {
	if len(ids) == 0 {
		for _, sq := range s.queues {
			sq.cdrs = make([]*QCdr, 0)
			sq.metrics = make(map[string]Metric, len(sq.conf.Metrics))
			for _, m := range sq.conf.Metrics {
				if metric := CreateMetric(m); metric != nil {
					sq.metrics[m] = metric
				}
			}
		}
	} else {
		for _, id := range ids {
			sq, exists := s.queues[id]
			if !exists {
				Logger.Warning(fmt.Sprintf("Cannot reset queue id %v: Not Fund", id))
				continue
			}
			sq.cdrs = make([]*QCdr, 0)
			sq.metrics = make(map[string]Metric, len(sq.conf.Metrics))
			for _, m := range sq.conf.Metrics {
				if metric := CreateMetric(m); metric != nil {
					sq.metrics[m] = metric
				}
			}
		}
	}
	return nil
}

// change the xisting ones
// add new ones
// delete the ones missing from the new list
func (s *Stats) UpdateQueues(css []*CdrStats, out *int) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	oldQueues := s.queues
	s.queues = make(map[string]*StatsQueue, len(css))
	if def, exists := oldQueues[utils.META_DEFAULT]; exists {
		def.UpdateConf(def.conf) // for reset
		s.queues[utils.META_DEFAULT] = def
	}
	for _, cs := range css {
		var sq *StatsQueue
		var existing bool
		if oldQueues != nil {
			if sq, existing = oldQueues[cs.Id]; existing {
				sq.UpdateConf(cs)
			}
		}
		if sq == nil {
			sq = NewStatsQueue(cs)
		}
		s.queues[cs.Id] = sq
	}
	return nil
}

func (s *Stats) AppendCDR(cdr *utils.StoredCdr, out *int) error {
	s.mux.RLock()
	defer s.mux.RUnlock()
	for _, sq := range s.queues {
		sq.AppendCDR(cdr)
	}
	return nil
}

type ProxyStats struct {
	Client *rpc.Client
}

func NewProxyStats(addr string) (*ProxyStats, error) {
	client, err := rpc.Dial("tcp", addr)

	if err != nil {
		return nil, err
	}
	return &ProxyStats{Client: client}, nil
}

func (ps *ProxyStats) GetValues(sqID string, values *map[string]float64) error {
	return ps.Client.Call("Stats.GetValues", sqID, values)
}

func (ps *ProxyStats) AppendCDR(cdr *utils.StoredCdr, out *int) error {
	return ps.Client.Call("Stats.AppendCDR", cdr, out)
}

func (ps *ProxyStats) GetQueueIds(in int, ids *[]string) error {
	return ps.Client.Call("Stats.GetQueueIds", in, ids)
}

func (ps *ProxyStats) AddQueue(cs *CdrStats, out *int) error {
	return ps.Client.Call("Stats.AddQueue", cs, out)
}

func (ps *ProxyStats) ReloadQueues(ids []string, out *int) error {
	return ps.Client.Call("Stats.ReloadQueues", ids, out)
}

func (ps *ProxyStats) ResetQueues(ids []string, out *int) error {
	return ps.Client.Call("Stats.ReserQueues", ids, out)
}
