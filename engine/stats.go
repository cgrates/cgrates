/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"fmt"
	"sync"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type StatsInterface interface {
	GetValues(string, *map[string]float64) error
	GetQueueIds(int, *[]string) error
	AppendCDR(*StoredCdr, *int) error
	AddQueue(*CdrStats, *int) error
	ReloadQueues([]string, *int) error
	ResetQueues([]string, *int) error
	Stop(int, *int) error
}

type Stats struct {
	queues              map[string]*StatsQueue
	queueSavers         map[string]*queueSaver
	mux                 sync.RWMutex
	ratingDb            RatingStorage
	accountingDb        AccountingStorage
	defaultSaveInterval time.Duration
}

type queueSaver struct {
	ticker       *time.Ticker
	stopper      chan bool
	save         func(*queueSaver)
	sq           *StatsQueue
	accountingDb AccountingStorage
}

func newQueueSaver(saveInterval time.Duration, sq *StatsQueue, adb AccountingStorage) *queueSaver {
	svr := &queueSaver{
		ticker:       time.NewTicker(saveInterval),
		stopper:      make(chan bool),
		sq:           sq,
		accountingDb: adb,
	}
	go func(saveInterval time.Duration, sq *StatsQueue, adb AccountingStorage) {
		for {
			select {
			case <-svr.ticker.C:
				sq.Save(adb)
			case <-svr.stopper:
				break
			}
		}
	}(saveInterval, sq, adb)
	return svr
}

func (svr *queueSaver) stop() {
	svr.ticker.Stop()
	svr.stopper <- true
	svr.sq.Save(svr.accountingDb)
}

func NewStats(ratingDb RatingStorage, accountingDb AccountingStorage, saveInterval time.Duration) *Stats {
	cdrStats := &Stats{ratingDb: ratingDb, accountingDb: accountingDb, defaultSaveInterval: saveInterval}
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
	return utils.ErrNotFound
}

func (s *Stats) AddQueue(cs *CdrStats, out *int) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.queues == nil {
		s.queues = make(map[string]*StatsQueue)
	}
	if s.queueSavers == nil {
		s.queueSavers = make(map[string]*queueSaver)
	}
	if sq, exists := s.queues[cs.Id]; exists {
		sq.UpdateConf(cs)
	} else {
		s.queues[cs.Id] = NewStatsQueue(cs)
		s.setupQueueSaver(sq)
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
			sq.Cdrs = make([]*QCdr, 0)
			sq.Metrics = make(map[string]Metric, len(sq.conf.Metrics))
			for _, m := range sq.conf.Metrics {
				if metric := CreateMetric(m); metric != nil {
					sq.Metrics[m] = metric
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
			sq.Cdrs = make([]*QCdr, 0)
			sq.Metrics = make(map[string]Metric, len(sq.conf.Metrics))
			for _, m := range sq.conf.Metrics {
				if metric := CreateMetric(m); metric != nil {
					sq.Metrics[m] = metric
				}
			}
		}
	}
	return nil
}

// change the existing ones
// add new ones
// delete the ones missing from the new list
func (s *Stats) UpdateQueues(css []*CdrStats, out *int) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	oldQueues := s.queues
	oldSavers := s.queueSavers
	s.queues = make(map[string]*StatsQueue, len(css))
	s.queueSavers = make(map[string]*queueSaver, len(css))
	for _, cs := range css {
		var sq *StatsQueue
		var existing bool
		if oldQueues != nil {
			if sq, existing = oldQueues[cs.Id]; existing {
				sq.UpdateConf(cs)
				s.queueSavers[cs.Id] = oldSavers[cs.Id]
				delete(oldSavers, cs.Id)
			}
		}
		if sq == nil {
			sq = NewStatsQueue(cs)
			// load queue from storage if exists
			if saved, err := s.accountingDb.GetCdrStatsQueue(sq.GetId()); err == nil {
				sq.Load(saved)
			} else {
				Logger.Info(err.Error())
			}
			s.setupQueueSaver(sq)
		}
		s.queues[cs.Id] = sq
	}
	// stop obsolete savers
	for _, saver := range oldSavers {
		saver.stop()
	}
	return nil
}

func (s *Stats) setupQueueSaver(sq *StatsQueue) {
	if sq == nil {
		return
	}
	// setup queue saver
	if s.queueSavers == nil {
		s.queueSavers = make(map[string]*queueSaver)
	}
	var si time.Duration
	if sq.conf != nil {
		si = sq.conf.SaveInterval
	}
	if si == 0 {
		si = s.defaultSaveInterval
	}
	if si > 0 {
		s.queueSavers[sq.GetId()] = newQueueSaver(si, sq, s.accountingDb)
	}
}

func (s *Stats) AppendCDR(cdr *StoredCdr, out *int) error {
	s.mux.RLock()
	defer s.mux.RUnlock()
	for _, sq := range s.queues {
		sq.AppendCDR(cdr)
	}
	return nil
}

func (s *Stats) Stop(int, *int) error {
	s.mux.RLock()
	defer s.mux.RUnlock()
	for _, saver := range s.queueSavers {
		saver.stop()
	}
	return nil
}

type ProxyStats struct {
	Client *rpcclient.RpcClient
}

func NewProxyStats(addr string, reconnects int) (*ProxyStats, error) {
	client, err := rpcclient.NewRpcClient("tcp", addr, reconnects, utils.GOB)
	if err != nil {
		return nil, err
	}
	return &ProxyStats{Client: client}, nil
}

func (ps *ProxyStats) GetValues(sqID string, values *map[string]float64) error {
	return ps.Client.Call("Stats.GetValues", sqID, values)
}

func (ps *ProxyStats) AppendCDR(cdr *StoredCdr, out *int) error {
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
	return ps.Client.Call("Stats.ResetQueues", ids, out)
}

func (ps *ProxyStats) Stop(i int, r *int) error {
	return ps.Client.Call("Stats.Stop", 0, i)
}
