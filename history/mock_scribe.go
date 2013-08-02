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
	"strings"
	"sync"
)

type MockScribe struct {
	sync.RWMutex
	destinations   records
	ratingProfiles records
	DestBuf        bytes.Buffer
	RpBuf          bytes.Buffer
}

func NewMockScribe() (Scribe, error) {
	return &MockScribe{}, nil
}

func (s *MockScribe) Record(key string, obj interface{}) error {
	s.Lock()
	defer s.Unlock()
	switch {
	case strings.HasPrefix(key, DESTINATION_PREFIX):
		s.destinations = s.destinations.SetOrAdd(key[len(DESTINATION_PREFIX):], obj)
		s.save(DESTINATIONS_FILE)
	case strings.HasPrefix(key, RATING_PROFILE_PREFIX):
		s.ratingProfiles = s.ratingProfiles.SetOrAdd(key[len(DESTINATION_PREFIX):], obj)
		s.save(RATING_PROFILES_FILE)
	}
	return nil
}

func (s *MockScribe) save(filename string) error {
	switch filename {
	case DESTINATIONS_FILE:
		s.DestBuf.Reset()
		b := bufio.NewWriter(&s.DestBuf)
		defer b.Flush()
		if err := s.format(b, s.destinations); err != nil {
			return err
		}
	case RATING_PROFILES_FILE:
		s.RpBuf.Reset()
		b := bufio.NewWriter(&s.RpBuf)
		defer b.Flush()
		if err := s.format(b, s.ratingProfiles); err != nil {
			return err
		}
	}

	return nil
}

func (s *MockScribe) format(b io.Writer, recs records) error {
	recs.Sort()
	b.Write([]byte("["))
	for i, r := range recs {
		src, err := json.Marshal(r)
		if err != nil {
			return err
		}
		b.Write(src)
		if i < len(recs)-1 {
			b.Write([]byte("\n"))
		}
	}
	b.Write([]byte("]"))
	return nil
}
