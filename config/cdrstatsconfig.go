/*
Real-time Charging System for Telecom & ISP environments
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

package config

import (
	"code.google.com/p/goconf/conf"
	"github.com/cgrates/cgrates/utils"
	"strconv"
	"time"
)

// Parse the configuration file for CDRStatConfigs
func ParseCfgDefaultCDRStatsConfig(c *conf.ConfigFile) (*CdrStatsConfig, error) {
	var err error
	csCfg := new(CdrStatsConfig)
	csCfg.Id = utils.DEFAULT_RUNID
	if hasOpt := c.HasOption("cdrstats", "queue_length"); hasOpt {
		csCfg.QueueLength, _ = c.GetInt("cdrstats", "queue_length")
	}
	if hasOpt := c.HasOption("cdrstats", "time_window"); hasOpt {
		durStr, _ := c.GetString("cdrstats", "time_window")
		if csCfg.TimeWindow, err = utils.ParseDurationWithSecs(durStr); err != nil {
			return nil, err
		}
	}
	if hasOpt := c.HasOption("cdrstats", "metrics"); hasOpt {
		metricsStr, _ := c.GetString("cdrstats", "metrics")
		if csCfg.Metrics, err = ConfigSlice(metricsStr); err != nil {
			return nil, err
		}
	}
	if hasOpt := c.HasOption("cdrstats", "setup_interval"); hasOpt {
		setupIntervalStr, _ := c.GetString("cdrstats", "setup_interval")
		if len(setupIntervalStr) != 0 { // If we parse empty, will get empty time, we prefer nil
			if setupIntervalSlc, err := ConfigSlice(setupIntervalStr); err != nil {
				return nil, err
			} else {
				for _, setupTimeStr := range setupIntervalSlc {
					if setupTime, err := utils.ParseTimeDetectLayout(setupTimeStr); err != nil {
						return nil, err
					} else {
						csCfg.SetupInterval = append(csCfg.SetupInterval, setupTime)
					}
				}
			}
		}
	}
	if hasOpt := c.HasOption("cdrstats", "tors"); hasOpt {
		torsStr, _ := c.GetString("cdrstats", "tors")
		if csCfg.TOR, err = ConfigSlice(torsStr); err != nil {
			return nil, err
		}
	}
	if hasOpt := c.HasOption("cdrstats", "cdr_hosts"); hasOpt {
		valsStr, _ := c.GetString("cdrstats", "cdr_hosts")
		if csCfg.CdrHost, err = ConfigSlice(valsStr); err != nil {
			return nil, err
		}
	}
	if hasOpt := c.HasOption("cdrstats", "cdr_sources"); hasOpt {
		valsStr, _ := c.GetString("cdrstats", "cdr_sources")
		if csCfg.CdrSource, err = ConfigSlice(valsStr); err != nil {
			return nil, err
		}
	}
	if hasOpt := c.HasOption("cdrstats", "req_types"); hasOpt {
		valsStr, _ := c.GetString("cdrstats", "req_types")
		if csCfg.ReqType, err = ConfigSlice(valsStr); err != nil {
			return nil, err
		}
	}
	if hasOpt := c.HasOption("cdrstats", "directions"); hasOpt {
		valsStr, _ := c.GetString("cdrstats", "directions")
		if csCfg.Direction, err = ConfigSlice(valsStr); err != nil {
			return nil, err
		}
	}
	if hasOpt := c.HasOption("cdrstats", "tenants"); hasOpt {
		valsStr, _ := c.GetString("cdrstats", "tenants")
		if csCfg.Tenant, err = ConfigSlice(valsStr); err != nil {
			return nil, err
		}
	}
	if hasOpt := c.HasOption("cdrstats", "categories"); hasOpt {
		valsStr, _ := c.GetString("cdrstats", "categories")
		if csCfg.Category, err = ConfigSlice(valsStr); err != nil {
			return nil, err
		}
	}
	if hasOpt := c.HasOption("cdrstats", "accounts"); hasOpt {
		valsStr, _ := c.GetString("cdrstats", "accounts")
		if csCfg.Account, err = ConfigSlice(valsStr); err != nil {
			return nil, err
		}
	}
	if hasOpt := c.HasOption("cdrstats", "subjects"); hasOpt {
		valsStr, _ := c.GetString("cdrstats", "subjects")
		if csCfg.Subject, err = ConfigSlice(valsStr); err != nil {
			return nil, err
		}
	}
	if hasOpt := c.HasOption("cdrstats", "destination_prefixes"); hasOpt {
		valsStr, _ := c.GetString("cdrstats", "destination_prefixes")
		if csCfg.DestinationPrefix, err = ConfigSlice(valsStr); err != nil {
			return nil, err
		}
	}
	if hasOpt := c.HasOption("cdrstats", "usage_interval"); hasOpt {
		usageIntervalStr, _ := c.GetString("cdrstats", "usage_interval")
		if usageIntervalSlc, err := ConfigSlice(usageIntervalStr); err != nil {
			return nil, err
		} else {
			for _, usageDurStr := range usageIntervalSlc {
				if usageDur, err := utils.ParseDurationWithSecs(usageDurStr); err != nil {
					return nil, err
				} else {
					csCfg.UsageInterval = append(csCfg.UsageInterval, usageDur)
				}
			}
		}
	}
	if hasOpt := c.HasOption("cdrstats", "mediation_run_ids"); hasOpt {
		valsStr, _ := c.GetString("cdrstats", "mediation_run_ids")
		if csCfg.MediationRunIds, err = ConfigSlice(valsStr); err != nil {
			return nil, err
		}
	}
	if hasOpt := c.HasOption("cdrstats", "rated_accounts"); hasOpt {
		valsStr, _ := c.GetString("cdrstats", "rated_accounts")
		if csCfg.RatedAccount, err = ConfigSlice(valsStr); err != nil {
			return nil, err
		}
	}
	if hasOpt := c.HasOption("cdrstats", "rated_subjects"); hasOpt {
		valsStr, _ := c.GetString("cdrstats", "rated_subjects")
		if csCfg.RatedSubject, err = ConfigSlice(valsStr); err != nil {
			return nil, err
		}
	}
	if hasOpt := c.HasOption("cdrstats", "cost_intervals"); hasOpt {
		valsStr, _ := c.GetString("cdrstats", "cost_intervals")
		if costSlc, err := ConfigSlice(valsStr); err != nil {
			return nil, err
		} else {
			for _, costStr := range costSlc {
				if cost, err := strconv.ParseFloat(costStr, 64); err != nil {
					return nil, err
				} else {
					csCfg.CostInterval = append(csCfg.CostInterval, cost)
				}
			}
		}
	}

	return csCfg, nil
}

type CdrStatsConfig struct {
	Id                string        // Config id, unique per config instance
	QueueLength       int           // Number of items in the stats buffer
	TimeWindow        time.Duration // Will only keep the CDRs who's call setup time is not older than time.Now()-TimeWindow
	Metrics           []string      // ASR, ACD, ACC
	SetupInterval     []time.Time   // 2 or less items (>= start interval,< stop_interval)
	TOR               []string
	CdrHost           []string
	CdrSource         []string
	ReqType           []string
	Direction         []string
	Tenant            []string
	Category          []string
	Account           []string
	Subject           []string
	DestinationPrefix []string
	UsageInterval     []time.Duration // 2 or less items (>= Usage, <Usage)
	MediationRunIds   []string
	RatedAccount      []string
	RatedSubject      []string
	CostInterval      []float64 // 2 or less items, (>=Cost, <Cost)
}
