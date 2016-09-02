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
	cdr := &CDR{
		SetupTime:  time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
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
	cdr := &CDR{
		ToR:         "tor",
		OriginID:    "accid",
		OriginHost:  "cdrhost",
		Source:      "cdrsource",
		RequestType: "reqtype",
		Direction:   "direction",
		Tenant:      "tenant",
		Category:    "category",
		Account:     "account",
		Subject:     "subject",
		Destination: "12345678",
		SetupTime:   time.Date(2014, 7, 3, 13, 43, 0, 0, time.UTC),
		Usage:       10 * time.Second,
		RunID:       "mri",
		Cost:        10,
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
	cdr := &CDR{
		ToR:             "tor",
		OriginID:        "accid",
		OriginHost:      "cdrhost",
		Source:          "cdrsource",
		RequestType:     "reqtype",
		Direction:       "direction",
		Tenant:          "tenant",
		Category:        "category",
		Account:         "account",
		Subject:         "subject",
		Destination:     "0723045326",
		SetupTime:       time.Date(2014, 7, 3, 13, 43, 0, 0, time.UTC),
		Usage:           10 * time.Second,
		PDD:             7 * time.Second,
		Supplier:        "supplier1",
		DisconnectCause: "normal",
		RunID:           "mri",
		Cost:            10,
	}
	sq.conf = &CdrStats{}
	if sq.conf.AcceptCdr(cdr) != true {
		t.Errorf("Should have accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{TOR: []string{"test"}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{CdrHost: []string{"test"}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{CdrSource: []string{"test"}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Direction: []string{"test"}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: []string{"test"}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Category: []string{"test"}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Account: []string{"test"}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Subject: []string{"test"}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Supplier: []string{"test"}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{DisconnectCause: []string{"test"}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{DestinationIds: []string{"test"}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{DestinationIds: []string{"NAT", "RET"}}
	if sq.conf.AcceptCdr(cdr) != true {
		t.Errorf("Should have accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{SetupInterval: []time.Time{time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC)}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{SetupInterval: []time.Time{time.Date(2014, 7, 3, 13, 42, 0, 0, time.UTC), time.Date(2014, 7, 3, 13, 43, 0, 0, time.UTC)}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{SetupInterval: []time.Time{time.Date(2014, 7, 3, 13, 42, 0, 0, time.UTC)}}
	if sq.conf.AcceptCdr(cdr) != true {
		t.Errorf("Should have accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{SetupInterval: []time.Time{time.Date(2014, 7, 3, 13, 42, 0, 0, time.UTC), time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC)}}
	if sq.conf.AcceptCdr(cdr) != true {
		t.Errorf("Should have accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{UsageInterval: []time.Duration{11 * time.Second}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{UsageInterval: []time.Duration{1 * time.Second, 10 * time.Second}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{PddInterval: []time.Duration{8 * time.Second}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{PddInterval: []time.Duration{3 * time.Second, 7 * time.Second}}
	if sq.conf.AcceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{PddInterval: []time.Duration{3 * time.Second, 8 * time.Second}}
	if sq.conf.AcceptCdr(cdr) != true {
		t.Errorf("Should have accepted this CDR: %+v", cdr)
	}
}

func TestStatsQueueIds(t *testing.T) {
	cdrStats := NewStats(ratingStorage, accountingStorage, 0)
	ids := []string{}
	if err := cdrStats.GetQueueIds(0, &ids); err != nil {
		t.Error("Errorf getting queue ids: ", err)
	}
	result := len(ids)
	expected := 5
	if result != expected {
		t.Errorf("Errorf loading stats queues. Expected %v was %v (%v)", expected, result, ids)
	}
}

func TestStatsAppendCdr(t *testing.T) {
	cdrStats := NewStats(ratingStorage, accountingStorage, 0)
	cdr := &CDR{
		Tenant:          "cgrates.org",
		Category:        "call",
		AnswerTime:      time.Now(),
		SetupTime:       time.Now(),
		Usage:           10 * time.Second,
		Cost:            10,
		Supplier:        "suppl1",
		DisconnectCause: "NORMAL_CLEARNING",
	}
	err := cdrStats.AppendCDR(cdr, nil)
	if err != nil {
		t.Error("Error appending cdr to stats: ", err)
	}
	t.Log(cdrStats.queues)
	if len(cdrStats.queues) != 5 ||
		len(cdrStats.queues["CDRST1"].Cdrs) != 0 ||
		len(cdrStats.queues["CDRST2"].Cdrs) != 1 {
		t.Error("Error appending cdr to queue: ", utils.ToIJSON(cdrStats.queues))
	}
}

func TestStatsGetValues(t *testing.T) {
	cdrStats := NewStats(ratingStorage, accountingStorage, 0)
	cdr := &CDR{
		Tenant:     "cgrates.org",
		Category:   "call",
		AnswerTime: time.Now(),
		SetupTime:  time.Now(),
		Usage:      10 * time.Second,
		Cost:       10,
	}
	cdrStats.AppendCDR(cdr, nil)
	cdr = &CDR{
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
	cdrStats := NewStats(ratingStorage, accountingStorage, 0)
	cdr := &CDR{
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
	expected := 5
	if result != expected {
		t.Errorf("Error loading stats queues. Expected %v was %v: %v", expected, result, ids)
	}
	valMap := make(map[string]float64)
	if err := cdrStats.GetValues("CDRST2", &valMap); err != nil {
		t.Error("Error getting metric values: ", err)
	}
	if len(valMap) != 2 || valMap["ACD"] != 10 || valMap["ASR"] != 100 {
		t.Error("Error on metric map: ", valMap)
	}
}

func TestStatsReloadQueuesWithDefault(t *testing.T) {
	cdrStats := NewStats(ratingStorage, accountingStorage, 0)
	cdrStats.AddQueue(&CdrStats{
		Id: utils.META_DEFAULT,
	}, nil)
	cdr := &CDR{
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
	expected := 6
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

func TestStatsReloadQueuesWithIds(t *testing.T) {
	cdrStats := NewStats(ratingStorage, accountingStorage, 0)
	cdr := &CDR{
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
	expected := 6
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

func TestStatsSaveQueues(t *testing.T) {
	cdrStats := NewStats(ratingStorage, accountingStorage, 0)
	cdr := &CDR{
		Tenant:     "cgrates.org",
		Category:   "call",
		AnswerTime: time.Now(),
		SetupTime:  time.Now(),
		Usage:      10 * time.Second,
		Cost:       10,
	}
	cdrStats.AppendCDR(cdr, nil)
	ids := []string{}
	cdrStats.GetQueueIds(0, &ids)
	if _, found := cdrStats.queueSavers["CDRST1"]; !found {
		t.Error("Error creating queue savers: ", cdrStats.queueSavers)
	}
}

func TestStatsResetQueues(t *testing.T) {
	cdrStats := NewStats(ratingStorage, accountingStorage, 0)
	cdr := &CDR{
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
	expected := 6
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
	cdrStats := NewStats(ratingStorage, accountingStorage, 0)
	cdr := &CDR{
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
	expected := 6
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

func TestStatsSaveRestoreQeue(t *testing.T) {
	sq := &StatsQueue{
		conf: &CdrStats{Id: "TTT"},
		Cdrs: []*QCdr{&QCdr{Cost: 9.0}},
	}
	if err := accountingStorage.SetCdrStatsQueue(sq); err != nil {
		t.Error("Error saving metric: ", err)
	}
	recovered, err := accountingStorage.GetCdrStatsQueue(sq.GetId())
	if err != nil {
		t.Error("Error loading metric: ", err)
	}
	if len(recovered.Cdrs) != 1 || recovered.Cdrs[0].Cost != sq.Cdrs[0].Cost {
		t.Errorf("Expecting %+v got: %+v", sq.Cdrs[0], recovered.Cdrs[0])
	}
}

func TestStatsPurgeTimeOne(t *testing.T) {
	sq := NewStatsQueue(&CdrStats{Metrics: []string{ASR, ACD, TCD, ACC, TCC}, TimeWindow: 30 * time.Minute})
	cdr := &CDR{
		SetupTime:  time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		Usage:      10 * time.Second,
		Cost:       1,
	}
	qcdr := sq.AppendCDR(cdr)
	qcdr.EventTime = qcdr.SetupTime
	s := sq.GetStats()
	if s[ASR] != -1 ||
		s[ACD] != -1 ||
		s[TCD] != -1 ||
		s[ACC] != -1 ||
		s[TCC] != -1 {
		t.Errorf("Error getting stats: %+v", s)
	}
}

func TestStatsPurgeTime(t *testing.T) {
	sq := NewStatsQueue(&CdrStats{Metrics: []string{ASR, ACD, TCD, ACC, TCC}, TimeWindow: 30 * time.Minute})
	cdr := &CDR{
		SetupTime:  time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		Usage:      10 * time.Second,
		Cost:       1,
	}
	qcdr := sq.AppendCDR(cdr)
	qcdr.EventTime = qcdr.SetupTime
	cdr.Cost = 2
	qcdr = sq.AppendCDR(cdr)
	qcdr.EventTime = qcdr.SetupTime
	cdr.Cost = 3
	qcdr = sq.AppendCDR(cdr)
	qcdr.EventTime = qcdr.SetupTime
	s := sq.GetStats()
	if s[ASR] != -1 ||
		s[ACD] != -1 ||
		s[TCD] != -1 ||
		s[ACC] != -1 ||
		s[TCC] != -1 {
		t.Errorf("Error getting stats: %+v", s)
	}
}

func TestStatsPurgeTimeFirst(t *testing.T) {
	sq := NewStatsQueue(&CdrStats{Metrics: []string{ASR, ACD, TCD, ACC, TCC}, TimeWindow: 30 * time.Minute})
	cdr := &CDR{
		SetupTime:  time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		Usage:      10 * time.Second,
		Cost:       1,
	}
	qcdr := sq.AppendCDR(cdr)
	cdr.Cost = 2
	cdr.SetupTime = time.Date(2024, 7, 14, 14, 25, 0, 0, time.UTC)
	cdr.AnswerTime = time.Date(2024, 7, 14, 14, 25, 0, 0, time.UTC)
	qcdr.EventTime = qcdr.SetupTime
	sq.AppendCDR(cdr)
	cdr.Cost = 3
	sq.AppendCDR(cdr)
	s := sq.GetStats()
	if s[ASR] != 100 ||
		s[ACD] != 10 ||
		s[TCD] != 20 ||
		s[ACC] != 2.5 ||
		s[TCC] != 5 {
		t.Errorf("Error getting stats: %+v", s)
	}
}

func TestStatsPurgeLength(t *testing.T) {
	sq := NewStatsQueue(&CdrStats{Metrics: []string{ASR, ACD, TCD, ACC, TCC}, QueueLength: 1})
	cdr := &CDR{
		SetupTime:  time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
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
		s[TCD] != 10 ||
		s[ACC] != 3 ||
		s[TCC] != 3 {
		t.Errorf("Error getting stats: %+v", s)
	}
}
