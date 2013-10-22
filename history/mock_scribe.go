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
	sync.Mutex
	destinations   records
	ratingPlans    records
	ratingProfiles records
	DestBuf        bytes.Buffer
	RplBuf         bytes.Buffer
	RpBuf          bytes.Buffer
}

func NewMockScribe() (*MockScribe, error) {
	return &MockScribe{}, nil
}

func (s *MockScribe) Record(rec *Record, out *int) error {
	switch {
	case strings.HasPrefix(rec.Key, DESTINATION_PREFIX):
		s.destinations = s.destinations.SetOrAdd(&Record{rec.Key[len(DESTINATION_PREFIX):], rec.Object})
		s.save(DESTINATIONS_FILE)
	case strings.HasPrefix(rec.Key, RATING_PLAN_PREFIX):
		s.ratingPlans = s.ratingPlans.SetOrAdd(&Record{rec.Key[len(RATING_PLAN_PREFIX):], rec.Object})
		s.save(RATING_PLANS_FILE)
	case strings.HasPrefix(rec.Key, RATING_PROFILE_PREFIX):
		s.ratingProfiles = s.ratingProfiles.SetOrAdd(&Record{rec.Key[len(RATING_PROFILE_PREFIX):], rec.Object})
		s.save(RATING_PROFILES_FILE)
	}
	*out = 0
	return nil
}

func (s *MockScribe) save(filename string) error {
	s.Lock()
	defer s.Unlock()
	switch filename {
	case DESTINATIONS_FILE:
		s.DestBuf.Reset()
		b := bufio.NewWriter(&s.DestBuf)
		defer b.Flush()
		if err := s.format(b, s.destinations); err != nil {
			return err
		}
	case RATING_PLANS_FILE:
		s.RplBuf.Reset()
		b := bufio.NewWriter(&s.RplBuf)
		defer b.Flush()
		if err := s.format(b, s.ratingPlans); err != nil {
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
