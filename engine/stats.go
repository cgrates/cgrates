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
package engine

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type StatsInterface interface {
	GetValues(string, *map[string]float64) error
	GetQueueIds(int, *[]string) error
	GetQueue(string, *StatsQueue) error
	GetQueueTriggers(string, *ActionTriggers) error
	AppendCDR(*CDR, *int) error
	AddQueue(*CdrStats, *int) error
	RemoveQueue(string, *int) error
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
	ratingDb     RatingStorage
	accountingDb AccountingStorage
}

func newQueueSaver(saveInterval time.Duration, sq *StatsQueue, rdb RatingStorage, adb AccountingStorage) *queueSaver {
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
				sq.Save(rdb, adb)
			case <-svr.stopper:
				break
			}
		}
	}(saveInterval, sq, adb)
	return svr
}

func (svr *queueSaver) stop() {
	svr.sq.Save(svr.ratingDb, svr.accountingDb)
	svr.ticker.Stop()
	svr.stopper <- true
}

func NewStats(ratingDb RatingStorage, accountingDb AccountingStorage, saveInterval time.Duration) *Stats {
	cdrStats := &Stats{ratingDb: ratingDb, accountingDb: accountingDb, defaultSaveInterval: saveInterval}
	if css, err := ratingDb.GetAllCdrStats(); err == nil {
		cdrStats.UpdateQueues(css, nil)
	} else {
		utils.Logger.Err(fmt.Sprintf("Cannot load cdr stats: %v", err))
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

func (s *Stats) GetQueue(id string, sq *StatsQueue) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	q, found := s.queues[id]
	if !found {
		return utils.ErrNotFound
	}
	*sq = *q
	return nil
}

func (s *Stats) GetQueueTriggers(id string, ats *ActionTriggers) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	q, found := s.queues[id]
	if !found {
		return utils.ErrNotFound
	}
	if q.conf.Triggers != nil {
		*ats = q.conf.Triggers
	} else {
		*ats = ActionTriggers{}
	}
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
	var sq *StatsQueue
	var exists bool
	if sq, exists = s.queues[cs.Id]; exists {
		sq.UpdateConf(cs)
	} else {
		sq = NewStatsQueue(cs)
		s.queues[cs.Id] = sq
	}
	// save the conf
	if err := s.ratingDb.SetCdrStats(cs); err != nil {
		return err
	}
	if _, exists = s.queueSavers[sq.GetId()]; !exists {
		s.setupQueueSaver(sq)
	}
	return nil
}

func (s *Stats) RemoveQueue(qID string, out *int) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.queues == nil {
		s.queues = make(map[string]*StatsQueue)
	}
	if s.queueSavers == nil {
		s.queueSavers = make(map[string]*queueSaver)
	}

	delete(s.queues, qID)
	delete(s.queueSavers, qID)

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
				utils.Logger.Warning(fmt.Sprintf("Cannot reset queue id %v: Not Fund", id))
				continue
			}
			sq.Cdrs = make([]*QCdr, 0)
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
		s.queueSavers[sq.GetId()] = newQueueSaver(si, sq, s.ratingDb, s.accountingDb)
	}
}

func (s *Stats) AppendCDR(cdr *CDR, out *int) error {
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

func (s *Stats) Call(serviceMethod string, args interface{}, reply interface{}) error {
	parts := strings.Split(serviceMethod, ".")
	if len(parts) != 2 {
		return utils.ErrNotImplemented
	}
	// get method
	method := reflect.ValueOf(s).MethodByName(parts[1])
	if !method.IsValid() {
		return utils.ErrNotImplemented
	}

	// construct the params
	params := []reflect.Value{reflect.ValueOf(args), reflect.ValueOf(reply)}

	ret := method.Call(params)
	if len(ret) != 1 {
		return utils.ErrServerError
	}
	if ret[0].Interface() == nil {
		return nil
	}
	err, ok := ret[0].Interface().(error)
	if !ok {
		return utils.ErrServerError
	}
	return err
}
