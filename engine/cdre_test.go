// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"testing"

	"github.com/cgrates/cgrates/config"
)

func TestCDRENewCDRExporter(t *testing.T) {
	_, err := NewCDRExporter([]*CDR{}, &config.CdreCfg{}, "test", "test", "test", "test", false, 1, 'a', false, []string{}, &FilterS{})
	if err != nil {
		t.Error(err)
	}
}

func TestCDREmetaHandler(t *testing.T) {
	cdre := CDRExporter{}

	rcv, err := cdre.metaHandler("test", "test")

	if err != nil {
		if err.Error() != "Unsupported METATAG: test" {
			t.Fatal(err)
		}
	}

	if rcv != "" {
		t.Error(rcv)
	}
}

func TestCDREcomposeHeader(t *testing.T) {
	cdre := CDRExporter{
		exportTemplate: &config.CdreCfg{
			Fields: []*config.FCTemplate{
				{
					Type:  "*filler",
					Value: config.RSRParsers{{Rules: "test()"}},
				},
			},
		},
	}

	err := cdre.composeHeader()

	if err != nil {
		if err.Error() != "err" {
			t.Error(err)
		}
	}
}
