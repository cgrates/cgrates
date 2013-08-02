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
	"io"
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
	s.Buf.Reset()
	b := bufio.NewWriter(&s.Buf)
	defer b.Flush()
	if err := s.format(b); err != nil {
		return err
	}
	return nil
}

func (s *MockScribe) format(b io.Writer) error {
	s.records.Sort()
	b.Write([]byte("["))
	for i, r := range s.records {
		src, err := json.Marshal(r)
		if err != nil {
			return err
		}
		b.Write(src)
		if i < len(s.records)-1 {
			b.Write([]byte("\n"))
		}
	}
	b.Write([]byte("]"))
	return nil
}
