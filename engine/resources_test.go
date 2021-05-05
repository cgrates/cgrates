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
	"bytes"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestResourcesRemoveExpiredUnitsResetTotalUsage(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	r := &Resource{
		TTLIdx: []string{"ResGroup1", "ResGroup2", "ResGroup3"},
		Usages: map[string]*ResourceUsage{
			"ResGroup2": {
				Tenant:     "cgrates.org",
				ID:         "RU_2",
				Units:      11,
				ExpiryTime: time.Date(2021, 5, 3, 13, 0, 0, 0, time.UTC),
			},
			"ResGroup3": {
				Tenant: "cgrates.org",
				ID:     "RU_3",
			},
		},
		tUsage: utils.Float64Pointer(10),
	}

	exp := &Resource{
		TTLIdx: []string{"ResGroup3"},
		Usages: map[string]*ResourceUsage{
			"ResGroup3": {
				Tenant: "cgrates.org",
				ID:     "RU_3",
			},
		},
	}

	explog := "CGRateS <> [WARNING] resetting total usage for resourceID: , usage smaller than 0: -1.000000\n"
	r.removeExpiredUnits()

	if !reflect.DeepEqual(r, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, r)
	}

	rcvlog := buf.String()[20:]
	if rcvlog != explog {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", explog, rcvlog)
	}

	utils.Logger.SetLogLevel(0)
}

func TestResourcesAvailable(t *testing.T) {
	r := ResourceWithConfig{
		Resource: &Resource{
			Usages: map[string]*ResourceUsage{
				"RU_1": {
					Units: 4,
				},
				"RU_2": {
					Units: 7,
				},
			},
		},
		Config: &ResourceProfile{
			Limit: 10,
		},
	}

	exp := -1.0
	rcv := r.Available()

	if rcv != exp {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestResourcesRecordUsageZeroTTL(t *testing.T) {
	r := &Resource{
		Usages: map[string]*ResourceUsage{
			"RU_1": {
				Tenant: "cgrates.org",
				ID:     "RU_1",
			},
		},
		ttl: utils.DurationPointer(0),
	}
	ru := &ResourceUsage{
		ID: "RU_2",
	}

	err := r.recordUsage(ru)

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
}

func TestResourcesRecordUsageGtZeroTTL(t *testing.T) {
	r := &Resource{
		Usages: map[string]*ResourceUsage{
			"RU_1": {
				Tenant: "cgrates.org",
				ID:     "RU_1",
			},
		},
		TTLIdx: []string{"RU_1"},
		ttl:    utils.DurationPointer(1 * time.Second),
	}
	ru := &ResourceUsage{
		Tenant: "cgrates.org",
		ID:     "RU_2",
	}

	exp := &Resource{
		Usages: map[string]*ResourceUsage{
			"RU_1": {
				Tenant: "cgrates.org",
				ID:     "RU_1",
			},
			"RU_2": {
				Tenant: "cgrates.org",
				ID:     "RU_2",
			},
		},
		TTLIdx: []string{"RU_1", "RU_2"},
		ttl:    utils.DurationPointer(1 * time.Second),
	}
	err := r.recordUsage(ru)
	exp.Usages[ru.ID].ExpiryTime = r.Usages[ru.ID].ExpiryTime

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if !reflect.DeepEqual(r, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(r))
	}
}

type mockWriter struct {
	WriteF func(p []byte) (n int, err error)
}

func (mW *mockWriter) Write(p []byte) (n int, err error) {
	if mW.WriteF != nil {
		return mW.WriteF(p)
	}
	return 0, nil
}

func TestResourcesRecordUsageClearErr(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer

	rs := Resources{
		{
			Usages: map[string]*ResourceUsage{
				"RU_1": {
					Tenant: "cgrates.org",
					ID:     "RU_1",
				},
				"RU_2": {
					Tenant: "cgrates.org",
					ID:     "RU_2",
				},
			},
			TTLIdx: []string{"RU_1", "RU_2"},
			ttl:    utils.DurationPointer(1 * time.Second),
		},
		{
			Usages: map[string]*ResourceUsage{
				"RU_3": {
					Tenant: "cgrates.org",
					ID:     "RU_3",
				},
				"RU_4": {
					Tenant: "cgrates.org",
					ID:     "RU_4",
				},
			},
			TTLIdx: []string{"RU_3"},
			ttl:    utils.DurationPointer(2 * time.Second),
		},
	}

	ru := &ResourceUsage{
		Tenant: "cgrates.org",
		ID:     "RU_4",
	}

	exp := Resources{
		{
			Usages: map[string]*ResourceUsage{
				"RU_1": {
					Tenant: "cgrates.org",
					ID:     "RU_1",
				},
				"RU_2": {
					Tenant: "cgrates.org",
					ID:     "RU_2",
				},
			},
			TTLIdx: []string{"RU_1", "RU_2", "RU_4"},
			ttl:    utils.DurationPointer(1 * time.Second),
		},
		{
			Usages: map[string]*ResourceUsage{
				"RU_3": {
					Tenant: "cgrates.org",
					ID:     "RU_3",
				},
				"RU_4": {
					Tenant: "cgrates.org",
					ID:     "RU_4",
				},
			},
			TTLIdx: []string{"RU_3"},
			ttl:    utils.DurationPointer(2 * time.Second),
		},
	}

	explog := []string{
		fmt.Sprintf("CGRateS <> [WARNING] <%s>cannot record usage, err: duplicate resource usage with id: %s:%s", utils.ResourceS, ru.Tenant, ru.ID),
		fmt.Sprintf("CGRateS <> [WARNING] <%s> cannot clear usage, err: cannot find usage record with id: %s", utils.ResourceS, ru.ID),
	}
	experr := fmt.Sprintf("duplicate resource usage with id: %s", "cgrates.org:"+ru.ID)

	defer log.SetOutput(os.Stderr)
	log.SetOutput(&mockWriter{
		WriteF: func(p []byte) (n int, err error) {
			delete(rs[0].Usages, "RU_4")
			return buf.Write(p)
		},
	})

	err := rs.recordUsage(ru)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if !reflect.DeepEqual(rs, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(rs))
	}

	rcv := strings.Split(buf.String(), "\n")
	for idx, exp := range explog {
		rcv[idx] = rcv[idx][20:]
		if rcv[idx] != exp {
			t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv[idx])
		}
	}

	utils.Logger.SetLogLevel(0)
}
