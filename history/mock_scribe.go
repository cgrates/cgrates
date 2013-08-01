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

package history

import (
	"bufio"
	"bytes"
	"encoding/json"
	"sync"
)

type MockScribe struct {
	sync.RWMutex
	records records
	Buf     bytes.Buffer
}

func NewMockScribe() (Scribe, error) {
	return &MockScribe{}, nil
}

func (s *MockScribe) Record(key string, obj interface{}) error {
	s.Lock()
	defer s.Unlock()
	s.records = s.records.SetOrAdd(key, obj)
	s.save()
	return nil
}

func (s *MockScribe) save() error {
	b := bufio.NewWriter(&s.Buf)
	e := json.NewEncoder(b)
	defer b.Flush()
	s.records.Sort()
	return e.Encode(s.records)
}
