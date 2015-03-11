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

package history

import (
	"bufio"
	"bytes"
	"sync"
)

type MockScribe struct {
	mu     sync.Mutex
	BufMap map[string]*bytes.Buffer
}

func NewMockScribe() (*MockScribe, error) {
	return &MockScribe{BufMap: map[string]*bytes.Buffer{
		DESTINATIONS_FN:    bytes.NewBuffer(nil),
		RATING_PLANS_FN:    bytes.NewBuffer(nil),
		RATING_PROFILES_FN: bytes.NewBuffer(nil),
	}}, nil
}

func (s *MockScribe) Record(rec Record, out *int) error {
	s.mu.Lock()
	fn := rec.Filename
	recordsMap[fn] = recordsMap[fn].SetOrAdd(&rec)
	s.mu.Unlock()
	s.save(fn)
	return nil
}

func (s *MockScribe) save(filename string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	records := recordsMap[filename]
	s.BufMap[filename].Reset()
	b := bufio.NewWriter(s.BufMap[filename])
	defer b.Flush()
	if err := format(b, records); err != nil {
		return err
	}
	return nil
}

func (s *MockScribe) GetBuffer(fn string) *bytes.Buffer {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.BufMap[fn]
}
