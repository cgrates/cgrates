/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package tpes

import (
	"bytes"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestTPEnewTPStats(t *testing.T) {
	// dataDB := &engine.DataDBM
	// dm := &engine.NewDataManager()
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(&engine.DataDBMock{
		GetStatQueueProfileDrvF: func(ctx *context.Context, tnt string, id string) (*engine.StatQueueProfile, error) {
			stq := &engine.StatQueueProfile{
				Tenant: "cgrates.org",
				ID:     "SQ_2",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				QueueLength: 14,
				Metrics: []*engine.MetricWithFilters{
					{
						MetricID: utils.MetaASR,
					},
					{
						MetricID: utils.MetaTCD,
					},
					{
						MetricID: utils.MetaPDD,
					},
					{
						MetricID: utils.MetaTCC,
					},
					{
						MetricID: utils.MetaTCD,
					},
				},
				ThresholdIDs: []string{utils.MetaNone},
			}
			return stq, nil
		},
	}, cfg, connMng)
	exp := &TPStats{
		dm: dm,
	}
	rcv := newTPStats(dm)
	if rcv.dm != exp.dm {
		t.Errorf("Expected %v \nbut received %v", exp, rcv)
	}
}

func TestTPEExportStats(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	tpStq := TPStats{
		dm: dm,
	}
	stq := &engine.StatQueueProfile{
		Tenant: "cgrates.org",
		ID:     "SQ_2",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		QueueLength: 14,
		Metrics: []*engine.MetricWithFilters{
			{
				MetricID: utils.MetaASR,
			},
			{
				MetricID: utils.MetaTCD,
			},
			{
				MetricID: utils.MetaPDD,
			},
			{
				MetricID: utils.MetaTCC,
			},
			{
				MetricID: utils.MetaTCD,
			},
		},
		ThresholdIDs: []string{utils.MetaNone},
	}
	tpStq.dm.SetStatQueueProfile(context.Background(), stq, false)
	err := tpStq.exportItems(context.Background(), wrtr, "cgrates.org", []string{"SQ_2"})
	if err != nil {
		t.Errorf("Expected nil\n but received %v", err)
	}
}

func TestTPEExportItemsStatsNoDbConn(t *testing.T) {
	engine.Cache.Clear(nil)
	wrtr := new(bytes.Buffer)
	tpStq := TPStats{
		dm: nil,
	}
	stq := &engine.StatQueueProfile{
		Tenant: "cgrates.org",
		ID:     "SQ_2",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		QueueLength: 14,
		Metrics: []*engine.MetricWithFilters{
			{
				MetricID: utils.MetaASR,
			},
			{
				MetricID: utils.MetaTCD,
			},
			{
				MetricID: utils.MetaPDD,
			},
			{
				MetricID: utils.MetaTCC,
			},
			{
				MetricID: utils.MetaTCD,
			},
		},
		ThresholdIDs: []string{utils.MetaNone},
	}
	tpStq.dm.SetStatQueueProfile(context.Background(), stq, false)
	err := tpStq.exportItems(context.Background(), wrtr, "cgrates.org", []string{"SQ_2"})
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected %v\n but received %v", utils.ErrNoDatabaseConn, err)
	}
}

func TestTPEExportItemsStatsIDNotFound(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	tpStq := TPStats{
		dm: dm,
	}
	stq := &engine.StatQueueProfile{
		Tenant: "cgrates.org",
		ID:     "SQ_2",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		QueueLength: 14,
		Metrics: []*engine.MetricWithFilters{
			{
				MetricID: utils.MetaASR,
			},
			{
				MetricID: utils.MetaTCD,
			},
			{
				MetricID: utils.MetaPDD,
			},
			{
				MetricID: utils.MetaTCC,
			},
			{
				MetricID: utils.MetaTCD,
			},
		},
		ThresholdIDs: []string{utils.MetaNone},
	}
	tpStq.dm.SetStatQueueProfile(context.Background(), stq, false)
	err := tpStq.exportItems(context.Background(), wrtr, "cgrates.org", []string{"SQ_3"})
	errExpect := "<NOT_FOUND> cannot find StatQueueProfile with id: <SQ_3>"
	if err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}
