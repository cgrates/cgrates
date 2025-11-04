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

package utils

import (
	"sync"
	"time"

	"github.com/cgrates/cron"
)

// ExporterMetrics stores export statistics with thread-safe access and
// cron-scheduled resets.
type ExporterMetrics struct {
	mu         sync.RWMutex
	MapStorage MapStorage
	cron       *cron.Cron
	loc        *time.Location
}

// NewExporterMetrics creates metrics with optional automatic reset.
// schedule is a cron expression for reset timing (empty to disable).
func NewExporterMetrics(schedule string, loc *time.Location) *ExporterMetrics {
	m := &ExporterMetrics{
		loc: loc,
	}
	m.Reset() // init MapStorage with default values

	if schedule != "" {
		m.cron = cron.New()
		m.cron.AddFunc(schedule, func() {
			m.Reset()
		})
		m.cron.Start()
	}
	return m
}

// Reset immediately clears all metrics and resets them to initial values.
func (m *ExporterMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MapStorage = MapStorage{
		NumberOfEvents:  int64(0),
		PositiveExports: StringSet{},
		NegativeExports: StringSet{},
		TimeNow:         time.Now().In(m.loc),
	}
}

// StopCron stops the automatic reset schedule if one is active.
func (m *ExporterMetrics) StopCron() {
	if m.cron == nil {
		return
	}
	m.cron.Stop()
	// ctx := m.cron.Stop()
	// <-ctx.Done() // wait for any running jobs to complete
}

// String returns the map as json string.
func (m *ExporterMetrics) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.MapStorage.String()
}

// FieldAsInterface returns the value from the path.
func (m *ExporterMetrics) FieldAsInterface(fldPath []string) (val any, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.MapStorage.FieldAsInterface(fldPath)
}

// FieldAsString returns the value from path as string.
func (m *ExporterMetrics) FieldAsString(fldPath []string) (str string, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.MapStorage.FieldAsString(fldPath)
}

// Set sets the value at the given path.
func (m *ExporterMetrics) Set(fldPath []string, val any) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.MapStorage.Set(fldPath, val)
}

// GetKeys returns all the keys from map.
func (m *ExporterMetrics) GetKeys(nested bool, nestedLimit int, prefix string) (keys []string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.MapStorage.GetKeys(nested, nestedLimit, prefix)
}

// Remove removes the item at path
func (m *ExporterMetrics) Remove(fldPath []string) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.MapStorage.Remove(fldPath)
}

func (m *ExporterMetrics) ClonedMapStorage() (msClone MapStorage) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.MapStorage.Clone()
}

// IncrementEvents increases the event counter (NumberOfEvents) by 1.
func (m *ExporterMetrics) IncrementEvents() {
	m.mu.Lock()
	defer m.mu.Unlock()
	count, _ := m.MapStorage[NumberOfEvents].(int64)
	m.MapStorage[NumberOfEvents] = count + 1
}

// Lock locks the ExporterMetrics mutex.
func (m *ExporterMetrics) Lock() {
	m.mu.Lock()
}

// Unlock unlocks the ExporterMetrics mutex.
func (m *ExporterMetrics) Unlock() {
	m.mu.Unlock()
}

// RLock locks the ExporterMetrics mutex for reading.
func (m *ExporterMetrics) RLock() {
	m.mu.RLock()
}

// RUnlock unlocks the read lock on the ExporterMetrics mutex.
func (m *ExporterMetrics) RUnlock() {
	m.mu.RUnlock()
}
