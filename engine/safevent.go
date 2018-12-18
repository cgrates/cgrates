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
	"net"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func NewSafEvent(mp map[string]interface{}) *SafEvent {
	return &SafEvent{Me: NewMapEvent(mp)}
}

// SafEvent is a generic event which is safe to read/write from multiple goroutines
type SafEvent struct {
	sync.RWMutex
	Me MapEvent // need it exportable so we can pass it on network
}

func (se *SafEvent) Clone() (cln *SafEvent) {
	se.RLock()
	cln = &SafEvent{Me: se.Me.Clone()}
	se.RUnlock()
	return
}

// MapEvent offers access to MapEvent methods, avoiding locks
func (se *SafEvent) MapEvent() (mp MapEvent) {
	return se.Me
}

func (se *SafEvent) String() (out string) {
	se.RLock()
	out = se.Me.String()
	se.RUnlock()
	return
}

func (se *SafEvent) FieldAsInterface(fldPath []string) (out interface{}, err error) {
	se.RLock()
	out, err = se.Me.FieldAsInterface(fldPath)
	se.RUnlock()
	return
}

func (se *SafEvent) FieldAsString(fldPath []string) (out string, err error) {
	se.RLock()
	out, err = se.Me.FieldAsString(fldPath)
	se.RUnlock()
	return
}

func (se *SafEvent) AsNavigableMap(fctemplate []*config.FCTemplate) (out *config.NavigableMap, err error) {
	se.RLock()
	out, err = se.Me.AsNavigableMap(fctemplate)
	se.RUnlock()
	return
}

func (se *SafEvent) RemoteHost() (out net.Addr) {
	se.RLock()
	out = se.Me.RemoteHost()
	se.RUnlock()
	return
}

func (se *SafEvent) HasField(fldName string) (has bool) {
	se.RLock()
	has = se.Me.HasField(fldName)
	se.RUnlock()
	return
}

func (se *SafEvent) Get(fldName string) (out interface{}, has bool) {
	se.RLock()
	out, has = se.Me[fldName]
	se.RUnlock()
	return
}

func (se *SafEvent) GetIgnoreErrors(fldName string) (out interface{}) {
	out, _ = se.Get(fldName)
	return
}

// Set will set a field's value
func (se *SafEvent) Set(fldName string, val interface{}) {
	se.Lock()
	se.Me[fldName] = val
	se.Unlock()
	return
}

// Remove will remove a field from map
func (se *SafEvent) Remove(fldName string) {
	se.Lock()
	delete(se.Me, fldName)
	se.Unlock()
	return
}

func (se *SafEvent) GetString(fldName string) (out string, err error) {
	se.RLock()
	out, err = se.Me.GetString(fldName)
	se.RUnlock()
	return
}

func (se *SafEvent) GetStringIgnoreErrors(fldName string) (out string) {
	out, _ = se.GetString(fldName)
	return
}

// GetDuration returns a field as Duration
func (se *SafEvent) GetDuration(fldName string) (d time.Duration, err error) {
	se.RLock()
	d, err = se.Me.GetDuration(fldName)
	se.RUnlock()
	return
}

// GetDurationPointer returns pointer towards duration, useful to detect presence of duration
func (se *SafEvent) GetDurationOrDefault(fldName string, dflt time.Duration) (d time.Duration, err error) {
	_, has := se.Get(fldName)
	if !has {
		return dflt, nil
	}
	return se.GetDuration(fldName)
}

// GetDuration returns a field as Duration, ignoring errors
func (se *SafEvent) GetDurationIgnoreErrors(fldName string) (d time.Duration) {
	d, _ = se.GetDuration(fldName)
	return
}

// GetDurationPointer returns pointer towards duration, useful to detect presence of duration
func (se *SafEvent) GetDurationPtr(fldName string) (d *time.Duration, err error) {
	fldIface, has := se.Get(fldName)
	if !has {
		return nil, utils.ErrNotFound
	}
	var dReal time.Duration
	if dReal, err = utils.IfaceAsDuration(fldIface); err != nil {
		return
	}
	return &dReal, nil
}

// GetDurationPointer returns pointer towards duration, useful to detect presence of duration
func (se *SafEvent) GetDurationPtrOrDefault(fldName string, dflt *time.Duration) (d *time.Duration, err error) {
	fldIface, has := se.Get(fldName)
	if !has {
		return dflt, nil
	}
	var dReal time.Duration
	if dReal, err = utils.IfaceAsDuration(fldIface); err != nil {
		return
	}
	return &dReal, nil
}

// GetTime returns a field as Time
func (se *SafEvent) GetTime(fldName string, tmz string) (t time.Time, err error) {
	se.RLock()
	t, err = se.Me.GetTime(fldName, tmz)
	se.RUnlock()
	return
}

// GetTimeIgnoreErrors returns a field as Time instance, ignoring errors
func (se *SafEvent) GetTimeIgnoreErrors(fldName string, tmz string) (t time.Time) {
	t, _ = se.GetTime(fldName, tmz)
	return
}

// GetSet will attempt to get a field value
// if field not present set it to the value received as parameter
func (se *SafEvent) GetSetString(fldName string, setVal string) (out string, err error) {
	se.Lock()
	defer se.Unlock()
	outIface, has := se.Me[fldName]
	if !has {
		se.Me[fldName] = setVal
		out = setVal
		return
	}
	// should be present, return it as string
	return utils.IfaceAsString(outIface)
}

// GetMapInterface returns the map stored internally without cloning it
func (se *SafEvent) GetMapInterface() (mp map[string]interface{}) {
	se.RLock()
	mp = se.Me
	se.RUnlock()
	return
}

// AsMapInterface returns the cloned map stored internally
func (se *SafEvent) AsMapInterface() (mp map[string]interface{}) {
	se.RLock()
	mp = se.Me.Clone()
	se.RUnlock()
	return
}

// AsMapString returns a map[string]string out of mp, ignoring specific fields if needed
// most used when needing to export extraFields
func (se *SafEvent) AsMapString(ignoredFlds utils.StringMap) (mp map[string]string, err error) {
	se.RLock()
	mp, err = se.Me.AsMapString(ignoredFlds)
	se.RUnlock()
	return
}

// AsMapString returns a map[string]string out of mp, ignoring specific fields if needed
// most used when needing to export extraFields
func (se *SafEvent) AsMapStringIgnoreErrors(ignoredFlds utils.StringMap) (mp map[string]string) {
	se.RLock()
	mp = se.Me.AsMapStringIgnoreErrors(ignoredFlds)
	se.RUnlock()
	return
}

// AsCDR exports the SafEvent as CDR
func (se *SafEvent) AsCDR(cfg *config.CGRConfig, tnt, tmz string) (cdr *CDR, err error) {
	se.RLock()
	cdr, err = se.Me.AsCDR(cfg, tnt, tmz)
	se.RUnlock()
	return
}
