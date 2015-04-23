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
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestStatsQueueInit(t *testing.T) {
	sq := NewStatsQueue(&CdrStats{Metrics: []string{ASR, ACC}})
	if len(sq.metrics) != 2 {
		t.Error("Expected 2 metrics got ", len(sq.metrics))
	}
}

func TestStatsValue(t *testing.T) {
	sq := NewStatsQueue(&CdrStats{Metrics: []string{ASR, ACD, TCD, ACC, TCC}})
	cdr := &StoredCdr{
		AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		Usage:      10 * time.Second,
		Cost:       1,
	}
	sq.AppendCDR(cdr)
	cdr.Cost = 2
	sq.AppendCDR(cdr)
	cdr.Cost = 3
	sq.AppendCDR(cdr)
	s := sq.GetStats()
	if s[ASR] != 100 ||
		s[ACD] != 10 ||
		s[TCD] != 30 ||
		s[ACC] != 2 ||
		s[TCC] != 6 {
		t.Errorf("Error getting stats: %+v", s)
	}
}

func TestStatsSimplifyCDR(t *testing.T) {
	cdr := &StoredCdr{
		TOR:            "tor",
		AccId:          "accid",
		CdrHost:        "cdrhost",
		CdrSource:      "cdrsource",
		ReqType:        "reqtype",
		Direction:      "direction",
		Tenant:         "tenant",
		Category:       "category",
		Account:        "account",
		Subject:        "subject",
		Destination:    "12345678",
		SetupTime:      time.Date(2014, 7, 3, 13, 43, 0, 0, time.UTC),
		Usage:          10 * time.Second,
		MediationRunId: "mri",
		Cost:           10,
	}
	sq := &StatsQueue{}
	qcdr := sq.simplifyCdr(cdr)
	if cdr.SetupTime != qcdr.SetupTime ||
		cdr.AnswerTime != qcdr.AnswerTime ||
		cdr.Usage != qcdr.Usage ||
		cdr.Cost != qcdr.Cost {
		t.Errorf("Failed to simplify cdr: %+v", qcdr)
	}
}

func TestAcceptCdr(t *testing.T) {
	sq := NewStatsQueue(nil)
	cdr := &StoredCdr{
		TOR:            "tor",
		AccId:          "accid",
		CdrHost:        "cdrhost",
		CdrSource:      "cdrsource",
		ReqType:        "reqtype",
		Direction:      "direction",
		Tenant:         "tenant",
		Category:       "category",
		Account:        "account",
		Subject:        "subject",
		Destination:    "12345678",
		SetupTime:      time.Date(2014, 7, 3, 13, 43, 0, 0, time.UTC),
		Usage:          10 * time.Second,
		MediationRunId: "mri",
		Cost:           10,
	}
	sq.conf = &CdrStats{}
	if sq.conf.AcceptCdr(cdr) != true {
		t.Errorf("Should have accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{TOR: []string{"test"}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{CdrHost: []string{"test"}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{CdrSource: []string{"test"}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Direction: []string{"test"}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: []string{"test"}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Category: []string{"test"}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Account: []string{"test"}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Subject: []string{"test"}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{RatedAccount: []string{"test"}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{RatedSubject: []string{"test"}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{DestinationPrefix: []string{"test"}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{DestinationPrefix: []string{"test", "123"}}
	if sq.conf.AcceptCdr(cdr) != true {
		t.Errorf("Should have accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{SetupInterval: []time.Time{time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC)}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{SetupInterval: []time.Time{time.Date(2014, 7, 3, 13, 42, 0, 0, time.UTC), time.Date(2014, 7, 3, 13, 43, 0, 0, time.UTC)}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{SetupInterval: []time.Time{time.Date(2014, 7, 3, 13, 42, 0, 0, time.UTC)}}
	if sq.conf.AcceptCdr(cdr) != true {
		t.Errorf("Should have accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{SetupInterval: []time.Time{time.Date(2014, 7, 3, 13, 42, 0, 0, time.UTC), time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC)}}
	if sq.conf.AcceptCdr(cdr) != true {
		t.Errorf("Should have accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{UsageInterval: []time.Duration{11 * time.Second}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{UsageInterval: []time.Duration{1 * time.Second, 10 * time.Second}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{UsageInterval: []time.Duration{10 * time.Second, 11 * time.Second}}
	if sq.conf.AcceptCdr(cdr) != true {
		t.Errorf("Should have accepted thif CDR: %+v", cdr)
	}
}

func TestStatsQueueIds(t *testing.T) {
	cdrStats := NewStats(dataStorage)
	ids := []string{}
	if err := cdrStats.GetQueueIds(0, &ids); err != nil {
		t.Error("Errorf getting queue ids: ", err)
	}
	result := len(ids)
	expected := 2
	if result != expected {
		t.Errorf("Errorf loading stats queues. Expected %v was %v", expected, result)
	}
}

func TestStatsAppendCdr(t *testing.T) {
	cdrStats := NewStats(dataStorage)
	cdr := &StoredCdr{
		Tenant:     "cgrates.org",
		Category:   "call",
		AnswerTime: time.Now(),
		SetupTime:  time.Now(),
		Usage:      10 * time.Second,
		Cost:       10,
	}
	err := cdrStats.AppendCDR(cdr, nil)
	if err != nil {
		t.Error("Error appending cdr to stats: ", err)
	}
	if len(cdrStats.queues["CDRST1"].cdrs) != 0 ||
		len(cdrStats.queues["CDRST2"].cdrs) != 1 {
		t.Error("Error appending cdr to queue: ", len(cdrStats.queues["CDRST2"].cdrs))
	}
}

func TestStatsGetValues(t *testing.T) {
	cdrStats := NewStats(dataStorage)
	cdr := &StoredCdr{
		Tenant:     "cgrates.org",
		Category:   "call",
		AnswerTime: time.Now(),
		SetupTime:  time.Now(),
		Usage:      10 * time.Second,
		Cost:       10,
	}
	cdrStats.AppendCDR(cdr, nil)
	cdr = &StoredCdr{
		Tenant:     "cgrates.org",
		Category:   "call",
		AnswerTime: time.Now(),
		SetupTime:  time.Now(),
		Usage:      2 * time.Second,
		Cost:       4,
	}
	cdrStats.AppendCDR(cdr, nil)
	valMap := make(map[string]float64)
	if err := cdrStats.GetValues("CDRST2", &valMap); err != nil {
		t.Error("Error getting metric values: ", err)
	}
	if len(valMap) != 2 || valMap["ACD"] != 6 || valMap["ASR"] != 100 {
		t.Error("Error on metric map: ", valMap)
	}
}

func TestStatsReloadQueues(t *testing.T) {
	cdrStats := NewStats(dataStorage)
	cdr := &StoredCdr{
		Tenant:     "cgrates.org",
		Category:   "call",
		AnswerTime: time.Now(),
		SetupTime:  time.Now(),
		Usage:      10 * time.Second,
		Cost:       10,
	}
	cdrStats.AppendCDR(cdr, nil)
	if err := cdrStats.ReloadQueues(nil, nil); err != nil {
		t.Error("Error reloading queues: ", err)
	}
	ids := []string{}
	if err := cdrStats.GetQueueIds(0, &ids); err != nil {
		t.Error("Error getting queue ids: ", err)
	}
	result := len(ids)
	expected := 2
	if result != expected {
		t.Errorf("Error loading stats queues. Expected %v was %v", expected, result)
	}
	valMap := make(map[string]float64)
	if err := cdrStats.GetValues("CDRST2", &valMap); err != nil {
		t.Error("Error getting metric values: ", err)
	}
	if len(valMap) != 2 || valMap["ACD"] != STATS_NA || valMap["ASR"] != STATS_NA {
		t.Error("Error on metric map: ", valMap)
	}
}

func TestStatsReloadQueuesWithDefault(t *testing.T) {
	cdrStats := NewStats(dataStorage)
	cdrStats.AddQueue(&CdrStats{
		Id: utils.META_DEFAULT,
	}, nil)
	cdr := &StoredCdr{
		Tenant:     "cgrates.org",
		Category:   "call",
		AnswerTime: time.Now(),
		SetupTime:  time.Now(),
		Usage:      10 * time.Second,
		Cost:       10,
	}
	cdrStats.AppendCDR(cdr, nil)

	if err := cdrStats.ReloadQueues(nil, nil); err != nil {
		t.Error("Error reloading queues: ", err)
	}
	ids := []string{}
	if err := cdrStats.GetQueueIds(0, &ids); err != nil {
		t.Error("Error getting queue ids: ", err)
	}
	result := len(ids)
	expected := 3
	if result != expected {
		t.Errorf("Error loading stats queues. Expected %v was %v", expected, result)
	}
	valMap := make(map[string]float64)
	if err := cdrStats.GetValues("CDRST2", &valMap); err != nil {
		t.Error("Error getting metric values: ", err)
	}
	if len(valMap) != 2 || valMap["ACD"] != STATS_NA || valMap["ASR"] != STATS_NA {
		t.Error("Error on metric map: ", valMap)
	}
}

func TestStatsReloadQueuesWithIds(t *testing.T) {
	cdrStats := NewStats(dataStorage)
	cdr := &StoredCdr{
		Tenant:     "cgrates.org",
		Category:   "call",
		AnswerTime: time.Now(),
		SetupTime:  time.Now(),
		Usage:      10 * time.Second,
		Cost:       10,
	}
	cdrStats.AppendCDR(cdr, nil)
	if err := cdrStats.ReloadQueues([]string{"CDRST1"}, nil); err != nil {
		t.Error("Error reloading queues: ", err)
	}
	ids := []string{}
	if err := cdrStats.GetQueueIds(0, &ids); err != nil {
		t.Error("Error getting queue ids: ", err)
	}
	result := len(ids)
	expected := 2
	if result != expected {
		t.Errorf("Error loading stats queues. Expected %v was %v", expected, result)
	}
	valMap := make(map[string]float64)
	if err := cdrStats.GetValues("CDRST2", &valMap); err != nil {
		t.Error("Error getting metric values: ", err)
	}
	if len(valMap) != 2 || valMap["ACD"] != 10 || valMap["ASR"] != 100 {
		t.Error("Error on metric map: ", valMap)
	}
}

func TestStatsResetQueues(t *testing.T) {
	cdrStats := NewStats(dataStorage)
	cdr := &StoredCdr{
		Tenant:     "cgrates.org",
		Category:   "call",
		AnswerTime: time.Now(),
		SetupTime:  time.Now(),
		Usage:      10 * time.Second,
		Cost:       10,
	}
	cdrStats.AppendCDR(cdr, nil)
	if err := cdrStats.ResetQueues(nil, nil); err != nil {
		t.Error("Error reloading queues: ", err)
	}
	ids := []string{}
	if err := cdrStats.GetQueueIds(0, &ids); err != nil {
		t.Error("Error getting queue ids: ", err)
	}
	result := len(ids)
	expected := 2
	if result != expected {
		t.Errorf("Error loading stats queues. Expected %v was %v", expected, result)
	}
	valMap := make(map[string]float64)
	if err := cdrStats.GetValues("CDRST2", &valMap); err != nil {
		t.Error("Error getting metric values: ", err)
	}
	if len(valMap) != 2 || valMap["ACD"] != STATS_NA || valMap["ASR"] != STATS_NA {
		t.Error("Error on metric map: ", valMap)
	}
}

func TestStatsResetQueuesWithIds(t *testing.T) {
	cdrStats := NewStats(dataStorage)
	cdr := &StoredCdr{
		Tenant:     "cgrates.org",
		Category:   "call",
		AnswerTime: time.Now(),
		SetupTime:  time.Now(),
		Usage:      10 * time.Second,
		Cost:       10,
	}
	cdrStats.AppendCDR(cdr, nil)
	if err := cdrStats.ResetQueues([]string{"CDRST1"}, nil); err != nil {
		t.Error("Error reloading queues: ", err)
	}
	ids := []string{}
	if err := cdrStats.GetQueueIds(0, &ids); err != nil {
		t.Error("Error getting queue ids: ", err)
	}
	result := len(ids)
	expected := 2
	if result != expected {
		t.Errorf("Error loading stats queues. Expected %v was %v", expected, result)
	}
	valMap := make(map[string]float64)
	if err := cdrStats.GetValues("CDRST2", &valMap); err != nil {
		t.Error("Error getting metric values: ", err)
	}
	if len(valMap) != 2 || valMap["ACD"] != 10 || valMap["ASR"] != 100 {
		t.Error("Error on metric map: ", valMap)
	}
}
