// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestNewAMQPee(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	em := &utils.ExporterMetrics{
		MapStorage: utils.MapStorage{
			utils.NumberOfEvents:  int64(0),
			utils.PositiveExports: utils.StringSet{},
			utils.NegativeExports: 5,
		},
	}
	cfg.EEsCfg().ExporterCfg(utils.MetaDefault).ConcurrentRequests = 2
	rcv := NewAMQPee(cfg.EEsCfg().ExporterCfg(utils.MetaDefault), em)
	exp := &AMQPee{
		cfg:  cfg.EEsCfg().ExporterCfg(utils.MetaDefault),
		em:   em,
		reqs: newConcReq(cfg.EEsCfg().ExporterCfg(utils.MetaDefault).ConcurrentRequests),
	}
	rcv.reqs = nil
	exp.reqs = nil
	exp.parseOpts(cfg.EEsCfg().ExporterCfg(utils.MetaDefault).Opts)
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}
}
