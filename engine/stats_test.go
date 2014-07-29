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

package engine

import (
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestStatsInit(t *testing.T) {
	sq := NewStatsQueue(&CdrStats{Metrics: []string{ASR, ACC}})
	if len(sq.metrics) != 2 {
		t.Error("Expected 2 metrics got ", len(sq.metrics))
	}
}

func TestStatsValue(t *testing.T) {
	sq := NewStatsQueue(&CdrStats{Metrics: []string{ASR, ACD, ACC}})
	cdr := &utils.StoredCdr{
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
		s[ACC] != 2 {
		t.Errorf("Error getting stats: %+v", s)
	}
}

func TestStatsSimplifyCDR(t *testing.T) {
	cdr := &utils.StoredCdr{
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
	qcdr := sq.simplifyCDR(cdr)
	if cdr.SetupTime != qcdr.SetupTime ||
		cdr.AnswerTime != qcdr.AnswerTime ||
		cdr.Usage != qcdr.Usage ||
		cdr.Cost != qcdr.Cost {
		t.Error("Failed to simplify cdr: %+v", qcdr)
	}
}

func TestAcceptCDR(t *testing.T) {
	sq := NewStatsQueue(nil)
	cdr := &utils.StoredCdr{
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
	if sq.acceptCDR(cdr) != true {
		t.Error("Should have accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{TOR: []string{"test"}}
	if sq.acceptCDR(cdr) == true {
		t.Error("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{CdrHost: []string{"test"}}
	if sq.acceptCDR(cdr) == true {
		t.Error("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{CdrSource: []string{"test"}}
	if sq.acceptCDR(cdr) == true {
		t.Error("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Direction: []string{"test"}}
	if sq.acceptCDR(cdr) == true {
		t.Error("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: []string{"test"}}
	if sq.acceptCDR(cdr) == true {
		t.Error("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Category: []string{"test"}}
	if sq.acceptCDR(cdr) == true {
		t.Error("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Account: []string{"test"}}
	if sq.acceptCDR(cdr) == true {
		t.Error("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Subject: []string{"test"}}
	if sq.acceptCDR(cdr) == true {
		t.Error("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{RatedAccount: []string{"test"}}
	if sq.acceptCDR(cdr) == true {
		t.Error("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{RatedSubject: []string{"test"}}
	if sq.acceptCDR(cdr) == true {
		t.Error("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{DestinationPrefix: []string{"test"}}
	if sq.acceptCDR(cdr) == true {
		t.Error("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{DestinationPrefix: []string{"test", "123"}}
	if sq.acceptCDR(cdr) != true {
		t.Error("Should have accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{SetupInterval: []time.Time{time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC)}}
	if sq.acceptCDR(cdr) == true {
		t.Error("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{SetupInterval: []time.Time{time.Date(2014, 7, 3, 13, 42, 0, 0, time.UTC), time.Date(2014, 7, 3, 13, 43, 0, 0, time.UTC)}}
	if sq.acceptCDR(cdr) == true {
		t.Error("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{SetupInterval: []time.Time{time.Date(2014, 7, 3, 13, 42, 0, 0, time.UTC)}}
	if sq.acceptCDR(cdr) != true {
		t.Error("Should have accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{SetupInterval: []time.Time{time.Date(2014, 7, 3, 13, 42, 0, 0, time.UTC), time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC)}}
	if sq.acceptCDR(cdr) != true {
		t.Error("Should have accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{UsageInterval: []time.Duration{11 * time.Second}}
	if sq.acceptCDR(cdr) == true {
		t.Error("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{UsageInterval: []time.Duration{1 * time.Second, 10 * time.Second}}
	if sq.acceptCDR(cdr) == true {
		t.Error("Should have NOT accepted thif CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{UsageInterval: []time.Duration{10 * time.Second, 11 * time.Second}}
	if sq.acceptCDR(cdr) != true {
		t.Error("Should have accepted thif CDR: %+v", cdr)
	}
}
