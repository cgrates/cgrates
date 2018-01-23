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
package config

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

var dfCgrJsonCfg *CgrJsonCfg

// Loads up the default configuration and  tests it's sections one by one
func TestDfNewdfCgrJsonCfgFromReader(t *testing.T) {
	var err error
	if dfCgrJsonCfg, err = NewCgrJsonCfgFromReader(strings.NewReader(CGRATES_CFG_JSON)); err != nil {
		t.Error(err)
	}
}

func TestDfGeneralJsonCfg(t *testing.T) {
	eCfg := &GeneralJsonCfg{
		Node_id:              utils.StringPointer(""),
		Logger:               utils.StringPointer(utils.MetaSysLog),
		Log_level:            utils.IntPointer(utils.LOGLEVEL_INFO),
		Http_skip_tls_verify: utils.BoolPointer(false),
		Rounding_decimals:    utils.IntPointer(5),
		Dbdata_encoding:      utils.StringPointer("msgpack"),
		Tpexport_dir:         utils.StringPointer("/var/spool/cgrates/tpe"),
		Poster_attempts:      utils.IntPointer(3),
		Failed_posts_dir:     utils.StringPointer("/var/spool/cgrates/failed_posts"),
		Default_request_type: utils.StringPointer(utils.META_RATED),
		Default_category:     utils.StringPointer("call"),
		Default_tenant:       utils.StringPointer("cgrates.org"),
		Default_timezone:     utils.StringPointer("Local"),
		Connect_attempts:     utils.IntPointer(3),
		Reconnects:           utils.IntPointer(-1),
		Connect_timeout:      utils.StringPointer("1s"),
		Reply_timeout:        utils.StringPointer("2s"),
		Response_cache_ttl:   utils.StringPointer("0s"),
		Internal_ttl:         utils.StringPointer("2m"),
		Locking_timeout:      utils.StringPointer("5s"),
	}
	if gCfg, err := dfCgrJsonCfg.GeneralJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, gCfg) {
		t.Error("Received: ", utils.ToIJSON(gCfg))
	}
}

func TestCacheJsonCfg(t *testing.T) {
	eCfg := &CacheJsonCfg{
		utils.CacheDestinations: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheReverseDestinations: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheRatingPlans: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheRatingProfiles: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheLCRRules: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheCDRStatS: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheActions: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheActionPlans: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheAccountActionPlans: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheActionTriggers: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheSharedGroups: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheAliases: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheReverseAliases: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheDerivedChargers: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheTimings: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheResourceProfiles: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheResources: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheEventResources: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer("1m"), Static_ttl: utils.BoolPointer(false)},
		utils.CacheStatQueueProfiles: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer("1m"), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheStatQueues: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer("1m"), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheThresholdProfiles: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheThresholds: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheFilters: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheSupplierProfiles: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheAttributeProfiles: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheResourceFilterIndexes: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheResourceFilterRevIndexes: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheStatFilterIndexes: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheStatFilterRevIndexes: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheThresholdFilterIndexes: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheThresholdFilterRevIndexes: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheSupplierFilterIndexes: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheSupplierFilterRevIndexes: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheAttributeFilterIndexes: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheAttributeFilterRevIndexes: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
	}

	if gCfg, err := dfCgrJsonCfg.CacheJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, gCfg) {
		t.Errorf("expected: %s\n, received: %s", utils.ToJSON(eCfg), utils.ToJSON(gCfg))
	}
}

func TestDfListenJsonCfg(t *testing.T) {
	eCfg := &ListenJsonCfg{
		Rpc_json: utils.StringPointer("127.0.0.1:2012"),
		Rpc_gob:  utils.StringPointer("127.0.0.1:2013"),
		Http:     utils.StringPointer("127.0.0.1:2080")}
	if cfg, err := dfCgrJsonCfg.ListenJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestDfDataDbJsonCfg(t *testing.T) {
	eCfg := &DbJsonCfg{
		Db_type:           utils.StringPointer("redis"),
		Db_host:           utils.StringPointer("127.0.0.1"),
		Db_port:           utils.IntPointer(6379),
		Db_name:           utils.StringPointer("10"),
		Db_user:           utils.StringPointer("cgrates"),
		Db_password:       utils.StringPointer(""),
		Load_history_size: utils.IntPointer(10),
	}
	if cfg, err := dfCgrJsonCfg.DbJsonCfg(DATADB_JSN); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestDfStorDBJsonCfg(t *testing.T) {
	eCfg := &DbJsonCfg{
		Db_type:           utils.StringPointer("mysql"),
		Db_host:           utils.StringPointer("127.0.0.1"),
		Db_port:           utils.IntPointer(3306),
		Db_name:           utils.StringPointer("cgrates"),
		Db_user:           utils.StringPointer("cgrates"),
		Db_password:       utils.StringPointer(""),
		Max_open_conns:    utils.IntPointer(100),
		Max_idle_conns:    utils.IntPointer(10),
		Conn_max_lifetime: utils.IntPointer(0),
		Cdrs_indexes:      &[]string{},
	}
	if cfg, err := dfCgrJsonCfg.DbJsonCfg(STORDB_JSN); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestDfRalsJsonCfg(t *testing.T) {
	eCfg := &RalsJsonCfg{
		Enabled:                     utils.BoolPointer(false),
		Thresholds_conns:            &[]*HaPoolJsonCfg{},
		Cdrstats_conns:              &[]*HaPoolJsonCfg{},
		Stats_conns:                 &[]*HaPoolJsonCfg{},
		Pubsubs_conns:               &[]*HaPoolJsonCfg{},
		Attributes_conns:            &[]*HaPoolJsonCfg{},
		Users_conns:                 &[]*HaPoolJsonCfg{},
		Aliases_conns:               &[]*HaPoolJsonCfg{},
		Rp_subject_prefix_matching:  utils.BoolPointer(false),
		Lcr_subject_prefix_matching: utils.BoolPointer(false),
		Max_computed_usage: &map[string]string{
			utils.ANY:   "189h",
			utils.VOICE: "72h",
			utils.DATA:  "107374182400",
			utils.SMS:   "10000"},
	}
	if cfg, err := dfCgrJsonCfg.RalsJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Received: %+v", cfg)
	}
}

func TestDfSchedulerJsonCfg(t *testing.T) {
	eCfg := &SchedulerJsonCfg{Enabled: utils.BoolPointer(false)}
	if cfg, err := dfCgrJsonCfg.SchedulerJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestDfCdrsJsonCfg(t *testing.T) {
	eCfg := &CdrsJsonCfg{
		Enabled:         utils.BoolPointer(false),
		Extra_fields:    &[]string{},
		Store_cdrs:      utils.BoolPointer(true),
		Sm_cost_retries: utils.IntPointer(5),
		Rals_conns: &[]*HaPoolJsonCfg{
			&HaPoolJsonCfg{
				Address: utils.StringPointer("*internal"),
			}},
		Pubsubs_conns:      &[]*HaPoolJsonCfg{},
		Attributes_conns:   &[]*HaPoolJsonCfg{},
		Users_conns:        &[]*HaPoolJsonCfg{},
		Aliases_conns:      &[]*HaPoolJsonCfg{},
		Cdrstats_conns:     &[]*HaPoolJsonCfg{},
		Thresholds_conns:   &[]*HaPoolJsonCfg{},
		Stats_conns:        &[]*HaPoolJsonCfg{},
		Online_cdr_exports: &[]string{},
	}
	if cfg, err := dfCgrJsonCfg.CdrsJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Received: %+v", *cfg)
	}
}

func TestDfCdrStatsJsonCfg(t *testing.T) {
	eCfg := &CdrStatsJsonCfg{
		Enabled:       utils.BoolPointer(false),
		Save_Interval: utils.StringPointer("1m"),
	}
	if cfg, err := dfCgrJsonCfg.CdrStatsJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", *cfg)
	}
}

func TestDfCdreJsonCfgs(t *testing.T) {
	eFields := []*CdrFieldJsonCfg{}
	eContentFlds := []*CdrFieldJsonCfg{
		&CdrFieldJsonCfg{Tag: utils.StringPointer("CGRID"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.CGRID)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("RunID"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.MEDI_RUNID)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("TOR"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.TOR)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("OriginID"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.OriginID)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("RequestType"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.RequestType)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Tenant"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.Tenant)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Category"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.Category)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Account"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.Account)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Subject"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.Subject)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Destination"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.Destination)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("SetupTime"),
			Type:   utils.StringPointer(utils.META_COMPOSED),
			Value:  utils.StringPointer(utils.SetupTime),
			Layout: utils.StringPointer("2006-01-02T15:04:05Z07:00")},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("AnswerTime"),
			Type:   utils.StringPointer(utils.META_COMPOSED),
			Value:  utils.StringPointer(utils.AnswerTime),
			Layout: utils.StringPointer("2006-01-02T15:04:05Z07:00")},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Usage"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.Usage)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Cost"),
			Type:              utils.StringPointer(utils.META_COMPOSED),
			Value:             utils.StringPointer(utils.COST),
			Rounding_decimals: utils.IntPointer(4)},
	}
	eCfg := map[string]*CdreJsonCfg{
		utils.META_DEFAULT: &CdreJsonCfg{
			Export_format:         utils.StringPointer(utils.MetaFileCSV),
			Export_path:           utils.StringPointer("/var/spool/cgrates/cdre"),
			Cdr_filter:            utils.StringPointer(""),
			Synchronous:           utils.BoolPointer(false),
			Attempts:              utils.IntPointer(1),
			Field_separator:       utils.StringPointer(","),
			Usage_multiply_factor: &map[string]float64{utils.ANY: 1.0},
			Cost_multiply_factor:  utils.Float64Pointer(1.0),
			Header_fields:         &eFields,
			Content_fields:        &eContentFlds,
			Trailer_fields:        &eFields,
		},
	}
	if cfg, err := dfCgrJsonCfg.CdreJsonCfgs(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		expect, _ := json.Marshal(eCfg)
		received, _ := json.Marshal(cfg)
		t.Errorf("Expecting:\n%s\nReceived:\n%s", string(expect), string(received))
	}
}

func TestDfCdrcJsonCfg(t *testing.T) {
	eFields := []*CdrFieldJsonCfg{}
	cdrFields := []*CdrFieldJsonCfg{
		&CdrFieldJsonCfg{Tag: utils.StringPointer("TOR"), Field_id: utils.StringPointer(utils.TOR), Type: utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer("2"), Mandatory: utils.BoolPointer(true)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("OriginID"), Field_id: utils.StringPointer(utils.OriginID), Type: utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer("3"), Mandatory: utils.BoolPointer(true)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("RequestType"), Field_id: utils.StringPointer(utils.RequestType), Type: utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer("4"), Mandatory: utils.BoolPointer(true)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Tenant"), Field_id: utils.StringPointer(utils.Tenant), Type: utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer("6"), Mandatory: utils.BoolPointer(true)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Category"), Field_id: utils.StringPointer(utils.Category), Type: utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer("7"), Mandatory: utils.BoolPointer(true)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Account"), Field_id: utils.StringPointer(utils.Account), Type: utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer("8"), Mandatory: utils.BoolPointer(true)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Subject"), Field_id: utils.StringPointer(utils.Subject), Type: utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer("9"), Mandatory: utils.BoolPointer(true)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Destination"), Field_id: utils.StringPointer(utils.Destination), Type: utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer("10"), Mandatory: utils.BoolPointer(true)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("SetupTime"), Field_id: utils.StringPointer(utils.SetupTime), Type: utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer("11"), Mandatory: utils.BoolPointer(true)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("AnswerTime"), Field_id: utils.StringPointer(utils.AnswerTime), Type: utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer("12"), Mandatory: utils.BoolPointer(true)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Usage"), Field_id: utils.StringPointer(utils.Usage), Type: utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer("13"), Mandatory: utils.BoolPointer(true)},
	}
	cacheDumpFields := []*CdrFieldJsonCfg{
		&CdrFieldJsonCfg{Tag: utils.StringPointer("CGRID"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.CGRID)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("RunID"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.MEDI_RUNID)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("TOR"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.TOR)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("OriginID"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.OriginID)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("RequestType"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.RequestType)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Tenant"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.Tenant)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Category"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.Category)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Account"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.Account)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Subject"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.Subject)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Destination"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.Destination)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("SetupTime"),
			Type:   utils.StringPointer(utils.META_COMPOSED),
			Value:  utils.StringPointer(utils.SetupTime),
			Layout: utils.StringPointer("2006-01-02T15:04:05Z07:00")},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("AnswerTime"),
			Type:   utils.StringPointer(utils.META_COMPOSED),
			Value:  utils.StringPointer(utils.AnswerTime),
			Layout: utils.StringPointer("2006-01-02T15:04:05Z07:00")},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Usage"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.Usage)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Cost"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.COST)},
	}
	eCfg := []*CdrcJsonCfg{
		&CdrcJsonCfg{
			Id:      utils.StringPointer(utils.META_DEFAULT),
			Enabled: utils.BoolPointer(false),
			Dry_run: utils.BoolPointer(false),
			Cdrs_conns: &[]*HaPoolJsonCfg{&HaPoolJsonCfg{
				Address: utils.StringPointer(utils.MetaInternal),
			}},
			Cdr_format:                  utils.StringPointer("csv"),
			Field_separator:             utils.StringPointer(","),
			Timezone:                    utils.StringPointer(""),
			Run_delay:                   utils.IntPointer(0),
			Max_open_files:              utils.IntPointer(1024),
			Data_usage_multiply_factor:  utils.Float64Pointer(1024.0),
			Cdr_in_dir:                  utils.StringPointer("/var/spool/cgrates/cdrc/in"),
			Cdr_out_dir:                 utils.StringPointer("/var/spool/cgrates/cdrc/out"),
			Failed_calls_prefix:         utils.StringPointer("missed_calls"),
			Cdr_path:                    utils.StringPointer(""),
			Cdr_source_id:               utils.StringPointer("freeswitch_csv"),
			Cdr_filter:                  utils.StringPointer(""),
			Continue_on_success:         utils.BoolPointer(false),
			Partial_record_cache:        utils.StringPointer("10s"),
			Partial_cache_expiry_action: utils.StringPointer(utils.MetaDumpToFile),
			Header_fields:               &eFields,
			Content_fields:              &cdrFields,
			Trailer_fields:              &eFields,
			Cache_dump_fields:           &cacheDumpFields,
		},
	}
	if cfg, err := dfCgrJsonCfg.CdrcJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expecting: \n%s\n, received: \n%s\n: ", utils.ToIJSON(eCfg), utils.ToIJSON(cfg))
	}
}

func TestSmgJsonCfg(t *testing.T) {
	eCfg := &SessionSJsonCfg{
		Enabled:       utils.BoolPointer(false),
		Listen_bijson: utils.StringPointer("127.0.0.1:2014"),
		Rals_conns: &[]*HaPoolJsonCfg{
			&HaPoolJsonCfg{
				Address: utils.StringPointer(utils.MetaInternal),
			}},
		Cdrs_conns: &[]*HaPoolJsonCfg{
			&HaPoolJsonCfg{
				Address: utils.StringPointer(utils.MetaInternal),
			}},
		Resources_conns:           &[]*HaPoolJsonCfg{},
		Suppliers_conns:           &[]*HaPoolJsonCfg{},
		Attributes_conns:          &[]*HaPoolJsonCfg{},
		Session_replication_conns: &[]*HaPoolJsonCfg{},
		Debit_interval:            utils.StringPointer("0s"),
		Min_call_duration:         utils.StringPointer("0s"),
		Max_call_duration:         utils.StringPointer("3h"),
		Session_ttl:               utils.StringPointer("0s"),
		Session_indexes:           &[]string{},
		Client_protocol:           utils.Float64Pointer(1.0),
	}
	if cfg, err := dfCgrJsonCfg.SessionSJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestFsAgentJsonCfg(t *testing.T) {
	eCfg := &FreeswitchAgentJsonCfg{
		Enabled: utils.BoolPointer(false),
		Sessions_conns: &[]*HaPoolJsonCfg{
			&HaPoolJsonCfg{
				Address: utils.StringPointer(utils.MetaInternal),
			}},
		Subscribe_park:         utils.BoolPointer(true),
		Create_cdr:             utils.BoolPointer(false),
		Extra_fields:           &[]string{},
		Empty_balance_context:  utils.StringPointer(""),
		Empty_balance_ann_file: utils.StringPointer(""),
		Channel_sync_interval:  utils.StringPointer("5m"),
		Max_wait_connection:    utils.StringPointer("2s"),
		Event_socket_conns: &[]*FsConnJsonCfg{
			&FsConnJsonCfg{
				Address:    utils.StringPointer("127.0.0.1:8021"),
				Password:   utils.StringPointer("ClueCon"),
				Reconnects: utils.IntPointer(5),
			}},
	}
	if cfg, err := dfCgrJsonCfg.FreeswitchAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestSmKamJsonCfg(t *testing.T) {
	eCfg := &SmKamJsonCfg{
		Enabled: utils.BoolPointer(false),
		Rals_conns: &[]*HaPoolJsonCfg{
			&HaPoolJsonCfg{
				Address: utils.StringPointer(utils.MetaInternal),
			}},
		Cdrs_conns: &[]*HaPoolJsonCfg{
			&HaPoolJsonCfg{
				Address: utils.StringPointer(utils.MetaInternal),
			}},
		Resources_conns:   &[]*HaPoolJsonCfg{},
		Create_cdr:        utils.BoolPointer(false),
		Debit_interval:    utils.StringPointer("10s"),
		Min_call_duration: utils.StringPointer("0s"),
		Max_call_duration: utils.StringPointer("3h"),
		Evapi_conns: &[]*KamConnJsonCfg{
			&KamConnJsonCfg{
				Address:    utils.StringPointer("127.0.0.1:8448"),
				Reconnects: utils.IntPointer(5),
			},
		},
	}
	if cfg, err := dfCgrJsonCfg.SmKamJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expecting: %s, received: %s: ",
			utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestAsteriskAgentJsonCfg(t *testing.T) {
	eCfg := &AsteriskAgentJsonCfg{
		Enabled: utils.BoolPointer(false),
		Sessions_conns: &[]*HaPoolJsonCfg{
			&HaPoolJsonCfg{
				Address: utils.StringPointer(utils.MetaInternal),
			}},
		Create_cdr: utils.BoolPointer(false),
		Asterisk_conns: &[]*AstConnJsonCfg{
			&AstConnJsonCfg{
				Address:          utils.StringPointer("127.0.0.1:8088"),
				User:             utils.StringPointer("cgrates"),
				Password:         utils.StringPointer("CGRateS.org"),
				Connect_attempts: utils.IntPointer(3),
				Reconnects:       utils.IntPointer(5),
			},
		},
	}
	if cfg, err := dfCgrJsonCfg.AsteriskAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expecting: %s, received: %s ", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDiameterAgentJsonCfg(t *testing.T) {
	eCfg := &DiameterAgentJsonCfg{
		Enabled:          utils.BoolPointer(false),
		Listen:           utils.StringPointer("127.0.0.1:3868"),
		Dictionaries_dir: utils.StringPointer("/usr/share/cgrates/diameter/dict/"),
		Sessions_conns: &[]*HaPoolJsonCfg{
			&HaPoolJsonCfg{
				Address: utils.StringPointer(utils.MetaInternal),
			}},
		Pubsubs_conns:        &[]*HaPoolJsonCfg{},
		Create_cdr:           utils.BoolPointer(true),
		Cdr_requires_session: utils.BoolPointer(true),
		Debit_interval:       utils.StringPointer("5m"),
		Timezone:             utils.StringPointer(""),
		Origin_host:          utils.StringPointer("CGR-DA"),
		Origin_realm:         utils.StringPointer("cgrates.org"),
		Vendor_id:            utils.IntPointer(0),
		Product_name:         utils.StringPointer("CGRateS"),
		Request_processors:   &[]*DARequestProcessorJsnCfg{},
	}
	if cfg, err := dfCgrJsonCfg.DiameterAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		rcv := *cfg.Request_processors
		t.Errorf("Received: %+v", rcv[0].CCA_fields)
	}
}

func TestRadiusAgentJsonCfg(t *testing.T) {
	eCfg := &RadiusAgentJsonCfg{
		Enabled:     utils.BoolPointer(false),
		Listen_net:  utils.StringPointer("udp"),
		Listen_auth: utils.StringPointer("127.0.0.1:1812"),
		Listen_acct: utils.StringPointer("127.0.0.1:1813"),
		Client_secrets: utils.MapStringStringPointer(map[string]string{
			utils.META_DEFAULT: "CGRateS.org",
		}),
		Client_dictionaries: utils.MapStringStringPointer(map[string]string{
			utils.META_DEFAULT: "/usr/share/cgrates/radius/dict/",
		}),
		Sessions_conns: &[]*HaPoolJsonCfg{
			&HaPoolJsonCfg{
				Address: utils.StringPointer(utils.MetaInternal),
			}},
		Create_cdr:           utils.BoolPointer(true),
		Cdr_requires_session: utils.BoolPointer(false),
		Timezone:             utils.StringPointer(""),
		Request_processors:   &[]*RAReqProcessorJsnCfg{},
	}
	if cfg, err := dfCgrJsonCfg.RadiusAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		rcv := *cfg.Request_processors
		t.Errorf("Received: %+v", rcv)
	}
}

func TestDfPubSubServJsonCfg(t *testing.T) {
	eCfg := &PubSubServJsonCfg{
		Enabled: utils.BoolPointer(false),
	}
	if cfg, err := dfCgrJsonCfg.PubSubServJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestDfAliasesServJsonCfg(t *testing.T) {
	eCfg := &AliasesServJsonCfg{
		Enabled: utils.BoolPointer(false),
	}
	if cfg, err := dfCgrJsonCfg.AliasesServJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestDfUserServJsonCfg(t *testing.T) {
	eCfg := &UserServJsonCfg{
		Enabled: utils.BoolPointer(false),
		Indexes: &[]string{},
	}
	if cfg, err := dfCgrJsonCfg.UserServJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestDfAttributeServJsonCfg(t *testing.T) {
	eCfg := &AttributeSJsonCfg{
		Enabled:               utils.BoolPointer(false),
		String_indexed_fields: nil,
		Prefix_indexed_fields: &[]string{},
	}
	if cfg, err := dfCgrJsonCfg.AttributeServJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestDfFilterSJsonCfg(t *testing.T) {
	eCfg := &FilterSJsonCfg{
		Stats_conns: &[]*HaPoolJsonCfg{},
	}
	if cfg, err := dfCgrJsonCfg.FilterSJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfResourceLimiterSJsonCfg(t *testing.T) {
	eCfg := &ResourceSJsonCfg{
		Enabled:               utils.BoolPointer(false),
		Thresholds_conns:      &[]*HaPoolJsonCfg{},
		Store_interval:        utils.StringPointer(""),
		String_indexed_fields: nil,
		Prefix_indexed_fields: &[]string{},
	}
	if cfg, err := dfCgrJsonCfg.ResourceSJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfStatServiceJsonCfg(t *testing.T) {
	eCfg := &StatServJsonCfg{
		Enabled:               utils.BoolPointer(false),
		Store_interval:        utils.StringPointer(""),
		Thresholds_conns:      &[]*HaPoolJsonCfg{},
		String_indexed_fields: nil,
		Prefix_indexed_fields: &[]string{},
	}
	if cfg, err := dfCgrJsonCfg.StatSJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestDfThresholdSJsonCfg(t *testing.T) {
	eCfg := &ThresholdSJsonCfg{
		Enabled:               utils.BoolPointer(false),
		Store_interval:        utils.StringPointer(""),
		String_indexed_fields: nil,
		Prefix_indexed_fields: &[]string{},
	}
	if cfg, err := dfCgrJsonCfg.ThresholdSJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfSupplierSJsonCfg(t *testing.T) {
	eCfg := &SupplierSJsonCfg{
		Enabled:               utils.BoolPointer(false),
		String_indexed_fields: nil,
		Prefix_indexed_fields: &[]string{},
		Rals_conns: &[]*HaPoolJsonCfg{
			&HaPoolJsonCfg{
				Address: utils.StringPointer("*internal"),
			},
		},
		Resources_conns: &[]*HaPoolJsonCfg{},
		Stats_conns:     &[]*HaPoolJsonCfg{},
	}
	if cfg, err := dfCgrJsonCfg.SupplierSJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("expecting: %+v, received: %+v", eCfg, cfg)
	}
}

func TestDfMailerJsonCfg(t *testing.T) {
	eCfg := &MailerJsonCfg{
		Server:        utils.StringPointer("localhost"),
		Auth_user:     utils.StringPointer("cgrates"),
		Auth_password: utils.StringPointer("CGRateS.org"),
		From_address:  utils.StringPointer("cgr-mailer@localhost.localdomain"),
	}
	if cfg, err := dfCgrJsonCfg.MailerJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestDfSureTaxJsonCfg(t *testing.T) {
	eCfg := &SureTaxJsonCfg{
		Url:                     utils.StringPointer(""),
		Client_number:           utils.StringPointer(""),
		Validation_key:          utils.StringPointer(""),
		Business_unit:           utils.StringPointer(""),
		Timezone:                utils.StringPointer("Local"),
		Include_local_cost:      utils.BoolPointer(false),
		Return_file_code:        utils.StringPointer("0"),
		Response_group:          utils.StringPointer("03"),
		Response_type:           utils.StringPointer("D4"),
		Regulatory_code:         utils.StringPointer("03"),
		Client_tracking:         utils.StringPointer(utils.CGRID),
		Customer_number:         utils.StringPointer("Subject"),
		Orig_number:             utils.StringPointer("Subject"),
		Term_number:             utils.StringPointer("Destination"),
		Bill_to_number:          utils.StringPointer(""),
		Zipcode:                 utils.StringPointer(""),
		Plus4:                   utils.StringPointer(""),
		P2PZipcode:              utils.StringPointer(""),
		P2PPlus4:                utils.StringPointer(""),
		Units:                   utils.StringPointer("^1"),
		Unit_type:               utils.StringPointer("^00"),
		Tax_included:            utils.StringPointer("^0"),
		Tax_situs_rule:          utils.StringPointer("^04"),
		Trans_type_code:         utils.StringPointer("^010101"),
		Sales_type_code:         utils.StringPointer("^R"),
		Tax_exemption_code_list: utils.StringPointer(""),
	}
	if cfg, err := dfCgrJsonCfg.SureTaxJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestNewCgrJsonCfgFromFile(t *testing.T) {
	cgrJsonCfg, err := NewCgrJsonCfgFromFile("cfg_data.json")
	if err != nil {
		t.Error(err)
	}
	eCfg := &GeneralJsonCfg{Default_request_type: utils.StringPointer(utils.META_PSEUDOPREPAID)}
	if gCfg, err := cgrJsonCfg.GeneralJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, gCfg) {
		t.Errorf("Expecting: %+v, received: %+v", eCfg, gCfg)
	}
	cdrFields := []*CdrFieldJsonCfg{
		&CdrFieldJsonCfg{Field_id: utils.StringPointer(utils.TOR), Value: utils.StringPointer("~7:s/^(voice|data|sms|mms|generic)$/*$1/")},
		&CdrFieldJsonCfg{Field_id: utils.StringPointer(utils.AnswerTime), Value: utils.StringPointer("1")},
		&CdrFieldJsonCfg{Field_id: utils.StringPointer(utils.Usage), Value: utils.StringPointer(`~9:s/^(\d+)$/${1}s/`)},
	}
	eCfgCdrc := []*CdrcJsonCfg{
		&CdrcJsonCfg{
			Id:            utils.StringPointer("CDRC-CSV1"),
			Enabled:       utils.BoolPointer(true),
			Cdr_in_dir:    utils.StringPointer("/tmp/cgrates/cdrc1/in"),
			Cdr_out_dir:   utils.StringPointer("/tmp/cgrates/cdrc1/out"),
			Cdr_source_id: utils.StringPointer("csv1"),
		},
		&CdrcJsonCfg{
			Id:                         utils.StringPointer("CDRC-CSV2"),
			Enabled:                    utils.BoolPointer(true),
			Data_usage_multiply_factor: utils.Float64Pointer(0.000976563),
			Run_delay:                  utils.IntPointer(1),
			Cdr_in_dir:                 utils.StringPointer("/tmp/cgrates/cdrc2/in"),
			Cdr_out_dir:                utils.StringPointer("/tmp/cgrates/cdrc2/out"),
			Cdr_source_id:              utils.StringPointer("csv2"),
			Content_fields:             &cdrFields,
		},
	}
	if cfg, err := cgrJsonCfg.CdrcJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfgCdrc, cfg) {
		t.Errorf("Expecting:\n %+v\n received:\n %+v\n", utils.ToIJSON(eCfgCdrc), utils.ToIJSON(cfg))
	}
	eCfgSmFs := &FreeswitchAgentJsonCfg{
		Enabled: utils.BoolPointer(true),
		Event_socket_conns: &[]*FsConnJsonCfg{
			&FsConnJsonCfg{
				Address:    utils.StringPointer("1.2.3.4:8021"),
				Password:   utils.StringPointer("ClueCon"),
				Reconnects: utils.IntPointer(5),
			},
			&FsConnJsonCfg{
				Address:    utils.StringPointer("2.3.4.5:8021"),
				Password:   utils.StringPointer("ClueCon"),
				Reconnects: utils.IntPointer(5),
			},
		},
	}
	if smFsCfg, err := cgrJsonCfg.FreeswitchAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfgSmFs, smFsCfg) {
		t.Error("Received: ", smFsCfg)
	}
}

func TestDfHttpJsonCfg(t *testing.T) {
	eCfg := &HTTPJsonCfg{
		Json_rpc_url:   utils.StringPointer("/jsonrpc"),
		Ws_url:         utils.StringPointer("/ws"),
		Use_basic_auth: utils.BoolPointer(false),
		Auth_users:     utils.MapStringStringPointer(map[string]string{})}
	if cfg, err := dfCgrJsonCfg.HttpJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}
