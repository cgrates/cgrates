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
	"testing"

	"github.com/cgrates/cgrates/utils"
)

var dfCgrJsonCfg *CgrJsonCfg

// Loads up the default configuration and  tests it's sections one by one
func TestDfNewdfCgrJsonCfgFromReader(t *testing.T) {
	var err error
	if dfCgrJsonCfg, err = NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON)); err != nil {
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
		Dbdata_encoding:      utils.StringPointer("*msgpack"),
		Tpexport_dir:         utils.StringPointer("/var/spool/cgrates/tpe"),
		Poster_attempts:      utils.IntPointer(3),
		Failed_posts_dir:     utils.StringPointer("/var/spool/cgrates/failed_posts"),
		Default_request_type: utils.StringPointer(utils.META_RATED),
		Default_category:     utils.StringPointer("call"),
		Default_tenant:       utils.StringPointer("cgrates.org"),
		Default_caching:      utils.StringPointer(utils.MetaReload),
		Default_timezone:     utils.StringPointer("Local"),
		Connect_attempts:     utils.IntPointer(5),
		Reconnects:           utils.IntPointer(-1),
		Connect_timeout:      utils.StringPointer("1s"),
		Reply_timeout:        utils.StringPointer("2s"),
		Locking_timeout:      utils.StringPointer("0"),
		Digest_separator:     utils.StringPointer(","),
		Digest_equal:         utils.StringPointer(":"),
		Rsr_separator:        utils.StringPointer(";"),
		Max_parralel_conns:   utils.IntPointer(100),
	}
	if gCfg, err := dfCgrJsonCfg.GeneralJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, gCfg) {
		t.Errorf("expecting: %s, \nreceived: %s", utils.ToIJSON(eCfg), utils.ToIJSON(gCfg))
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
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false)},
		utils.CacheStatQueueProfiles: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheStatQueues: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
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
		utils.CacheChargerProfiles: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheDispatcherProfiles: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheDispatcherHosts: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheResourceFilterIndexes: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false)},
		utils.CacheStatFilterIndexes: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false)},
		utils.CacheThresholdFilterIndexes: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false)},
		utils.CacheSupplierFilterIndexes: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false)},
		utils.CacheAttributeFilterIndexes: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false)},
		utils.CacheChargerFilterIndexes: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false)},
		utils.CacheDispatcherFilterIndexes: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false)},
		utils.CacheDispatcherRoutes: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false)},
		utils.CacheDiameterMessages: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer("3h"), Static_ttl: utils.BoolPointer(false)},
		utils.CacheRPCResponses: &CacheParamJsonCfg{Limit: utils.IntPointer(0),
			Ttl: utils.StringPointer("2s"), Static_ttl: utils.BoolPointer(false)},
		utils.CacheClosedSessions: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer("10s"), Static_ttl: utils.BoolPointer(false)},
		utils.CacheCDRIDs: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer("10m"), Static_ttl: utils.BoolPointer(false)},
		utils.CacheLoadIDs: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false),
			Precache: utils.BoolPointer(false)},
		utils.CacheRPCConnections: &CacheParamJsonCfg{Limit: utils.IntPointer(-1),
			Ttl: utils.StringPointer(""), Static_ttl: utils.BoolPointer(false)},
	}

	if gCfg, err := dfCgrJsonCfg.CacheJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, gCfg) {
		t.Errorf("expected: %s\n, received: %s", utils.ToJSON(eCfg), utils.ToJSON(gCfg))
	}
}

func TestDfListenJsonCfg(t *testing.T) {
	eCfg := &ListenJsonCfg{
		Rpc_json:     utils.StringPointer("127.0.0.1:2012"),
		Rpc_gob:      utils.StringPointer("127.0.0.1:2013"),
		Http:         utils.StringPointer("127.0.0.1:2080"),
		Rpc_json_tls: utils.StringPointer("127.0.0.1:2022"),
		Rpc_gob_tls:  utils.StringPointer("127.0.0.1:2023"),
		Http_tls:     utils.StringPointer("127.0.0.1:2280"),
	}
	if cfg, err := dfCgrJsonCfg.ListenJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestDfDataDbJsonCfg(t *testing.T) {
	eCfg := &DbJsonCfg{
		Db_type:           utils.StringPointer("*redis"),
		Db_host:           utils.StringPointer("127.0.0.1"),
		Db_port:           utils.IntPointer(6379),
		Db_name:           utils.StringPointer("10"),
		Db_user:           utils.StringPointer("cgrates"),
		Db_password:       utils.StringPointer(""),
		Redis_sentinel:    utils.StringPointer(""),
		Query_timeout:     utils.StringPointer("10s"),
		Replication_conns: &[]string{},
		Remote_conns:      &[]string{},
		Items: &map[string]*ItemOptJson{
			utils.MetaAccounts: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaReverseDestinations: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaDestinations: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaRatingPlans: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaRatingProfiles: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaActions: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaActionPlans: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaAccountActionPlans: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaActionTriggers: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaSharedGroups: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaTimings: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaResourceProfile: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaStatQueues: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaResources: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaStatQueueProfiles: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaThresholds: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaThresholdProfiles: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaFilters: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaSupplierProfiles: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaAttributeProfiles: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaDispatcherHosts: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaChargerProfiles: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaDispatcherProfiles: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaFilterIndexes: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
			utils.MetaLoadIDs: &ItemOptJson{
				Replicate: utils.BoolPointer(false),
				Remote:    utils.BoolPointer(false),
				Ttl:       utils.StringPointer(utils.EmptyString)},
		},
	}
	if cfg, err := dfCgrJsonCfg.DbJsonCfg(DATADB_JSN); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("expecting: %s, \nreceived: %s", utils.ToIJSON(eCfg), utils.ToIJSON(cfg))
	}
}

func TestDfStorDBJsonCfg(t *testing.T) {
	eCfg := &DbJsonCfg{
		Db_type:               utils.StringPointer("*mysql"),
		Db_host:               utils.StringPointer("127.0.0.1"),
		Db_port:               utils.IntPointer(3306),
		Db_name:               utils.StringPointer("cgrates"),
		Db_user:               utils.StringPointer("cgrates"),
		Db_password:           utils.StringPointer(""),
		Max_open_conns:        utils.IntPointer(100),
		Max_idle_conns:        utils.IntPointer(10),
		Conn_max_lifetime:     utils.IntPointer(0),
		String_indexed_fields: &[]string{},
		Prefix_indexed_fields: &[]string{},
		Query_timeout:         utils.StringPointer("10s"),
		Sslmode:               utils.StringPointer(utils.PostgressSSLModeDisable),
		Items: &map[string]*ItemOptJson{
			utils.TBLTPTimings: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
			utils.TBLTPDestinations: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
			utils.TBLTPRates: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
			utils.TBLTPDestinationRates: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
			utils.TBLTPRatingPlans: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
			utils.TBLTPRateProfiles: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
			utils.TBLTPSharedGroups: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
			utils.TBLTPActions: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
			utils.TBLTPActionTriggers: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
			utils.TBLTPAccountActions: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
			utils.TBLTPResources: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
			utils.TBLTPStats: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
			utils.TBLTPThresholds: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
			utils.TBLTPFilters: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
			utils.SessionCostsTBL: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
			utils.TBLTPActionPlans: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
			utils.TBLTPSuppliers: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
			utils.TBLTPAttributes: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
			utils.TBLTPChargers: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
			utils.TBLTPDispatchers: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
			utils.TBLTPDispatcherHosts: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
			utils.CDRsTBL: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
			utils.TBLVersions: &ItemOptJson{
				Ttl: utils.StringPointer(utils.EmptyString)},
		},
	}
	if cfg, err := dfCgrJsonCfg.DbJsonCfg(STORDB_JSN); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expected : %+v,\n Received: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfRalsJsonCfg(t *testing.T) {
	eCfg := &RalsJsonCfg{
		Enabled:                    utils.BoolPointer(false),
		Thresholds_conns:           &[]string{},
		Stats_conns:                &[]string{},
		CacheS_conns:               &[]string{utils.MetaInternal},
		Rp_subject_prefix_matching: utils.BoolPointer(false),
		Remove_expired:             utils.BoolPointer(true),
		Max_computed_usage: &map[string]string{
			utils.ANY:   "189h",
			utils.VOICE: "72h",
			utils.DATA:  "107374182400",
			utils.SMS:   "10000"},
		Max_increments: utils.IntPointer(1000000),
		Balance_rating_subject: &map[string]string{
			utils.ANY:      "*zero1ns",
			utils.VOICE:    "*zero1s",
			utils.DATA:     "*zero1ns",
			utils.SMS:      "*zero1ns",
			utils.MONETARY: "*zero1ns",
			utils.GENERIC:  "*zero1ns",
		},
	}
	if cfg, err := dfCgrJsonCfg.RalsJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Received: %+v", cfg)
	}
}

func TestDfSchedulerJsonCfg(t *testing.T) {
	eCfg := &SchedulerJsonCfg{
		Enabled:    utils.BoolPointer(false),
		Cdrs_conns: &[]string{},
		Filters:    &[]string{},
	}
	if cfg, err := dfCgrJsonCfg.SchedulerJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestDfCdrsJsonCfg(t *testing.T) {
	eCfg := &CdrsJsonCfg{
		Enabled:              utils.BoolPointer(false),
		Extra_fields:         &[]string{},
		Store_cdrs:           utils.BoolPointer(true),
		Session_cost_retries: utils.IntPointer(5),
		Chargers_conns:       &[]string{},
		Rals_conns:           &[]string{},
		Attributes_conns:     &[]string{},
		Thresholds_conns:     &[]string{},
		Stats_conns:          &[]string{},
		Online_cdr_exports:   &[]string{},
	}
	if cfg, err := dfCgrJsonCfg.CdrsJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Received: %+v", *cfg)
	}
}

func TestDfCdreJsonCfgs(t *testing.T) {
	eFields := []*FcTemplateJsonCfg{}
	eContentFlds := []*FcTemplateJsonCfg{
		{Tag: utils.StringPointer("CGRID"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.DynamicDataPrefix + utils.CGRID)},
		{Tag: utils.StringPointer("RunID"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.DynamicDataPrefix + utils.RunID)},
		{Tag: utils.StringPointer("TOR"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.DynamicDataPrefix + utils.ToR)},
		{Tag: utils.StringPointer("OriginID"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.DynamicDataPrefix + utils.OriginID)},
		{Tag: utils.StringPointer("RequestType"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.DynamicDataPrefix + utils.RequestType)},
		{Tag: utils.StringPointer("Tenant"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.DynamicDataPrefix + utils.Tenant)},
		{Tag: utils.StringPointer("Category"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.DynamicDataPrefix + utils.Category)},
		{Tag: utils.StringPointer("Account"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.DynamicDataPrefix + utils.Account)},
		{Tag: utils.StringPointer("Subject"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.DynamicDataPrefix + utils.Subject)},
		{Tag: utils.StringPointer("Destination"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.DynamicDataPrefix + utils.Destination)},
		{Tag: utils.StringPointer("SetupTime"),
			Type:   utils.StringPointer(utils.META_COMPOSED),
			Value:  utils.StringPointer(utils.DynamicDataPrefix + utils.SetupTime),
			Layout: utils.StringPointer("2006-01-02T15:04:05Z07:00")},
		{Tag: utils.StringPointer("AnswerTime"),
			Type:   utils.StringPointer(utils.META_COMPOSED),
			Value:  utils.StringPointer(utils.DynamicDataPrefix + utils.AnswerTime),
			Layout: utils.StringPointer("2006-01-02T15:04:05Z07:00")},
		{Tag: utils.StringPointer("Usage"),
			Type:  utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer(utils.DynamicDataPrefix + utils.Usage)},
		{Tag: utils.StringPointer("Cost"),
			Type:              utils.StringPointer(utils.META_COMPOSED),
			Value:             utils.StringPointer(utils.DynamicDataPrefix + utils.COST),
			Rounding_decimals: utils.IntPointer(4)},
	}
	eCfg := map[string]*CdreJsonCfg{
		utils.MetaDefault: {
			Export_format:      utils.StringPointer(utils.MetaFileCSV),
			Export_path:        utils.StringPointer("/var/spool/cgrates/cdre"),
			Synchronous:        utils.BoolPointer(false),
			Attempts:           utils.IntPointer(1),
			Tenant:             utils.StringPointer(""),
			Attributes_context: utils.StringPointer(""),
			Field_separator:    utils.StringPointer(","),
			Header_fields:      &eFields,
			Content_fields:     &eContentFlds,
			Trailer_fields:     &eFields,
			Filters:            &[]string{},
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

func TestSmgJsonCfg(t *testing.T) {
	eCfg := &SessionSJsonCfg{
		Enabled:               utils.BoolPointer(false),
		Listen_bijson:         utils.StringPointer("127.0.0.1:2014"),
		Chargers_conns:        &[]string{},
		Rals_conns:            &[]string{},
		Cdrs_conns:            &[]string{},
		Resources_conns:       &[]string{},
		Thresholds_conns:      &[]string{},
		Stats_conns:           &[]string{},
		Suppliers_conns:       &[]string{},
		Attributes_conns:      &[]string{},
		Replication_conns:     &[]string{},
		Debit_interval:        utils.StringPointer("0s"),
		Store_session_costs:   utils.BoolPointer(false),
		Min_call_duration:     utils.StringPointer("0s"),
		Max_call_duration:     utils.StringPointer("3h"),
		Session_ttl:           utils.StringPointer("0s"),
		Session_indexes:       &[]string{},
		Client_protocol:       utils.Float64Pointer(1.0),
		Channel_sync_interval: utils.StringPointer("0"),
		Terminate_attempts:    utils.IntPointer(5),
		Alterable_fields:      &[]string{},
	}
	if cfg, err := dfCgrJsonCfg.SessionSJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestFsAgentJsonCfg(t *testing.T) {
	eCfg := &FreeswitchAgentJsonCfg{
		Enabled:                utils.BoolPointer(false),
		Sessions_conns:         &[]string{utils.MetaInternal},
		Subscribe_park:         utils.BoolPointer(true),
		Create_cdr:             utils.BoolPointer(false),
		Extra_fields:           &[]string{},
		Empty_balance_context:  utils.StringPointer(""),
		Empty_balance_ann_file: utils.StringPointer(""),
		Max_wait_connection:    utils.StringPointer("2s"),
		Event_socket_conns: &[]*FsConnJsonCfg{
			{
				Address:    utils.StringPointer("127.0.0.1:8021"),
				Password:   utils.StringPointer("ClueCon"),
				Reconnects: utils.IntPointer(5),
				Alias:      utils.StringPointer(""),
			}},
	}
	if cfg, err := dfCgrJsonCfg.FreeswitchAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestKamAgentJsonCfg(t *testing.T) {
	eCfg := &KamAgentJsonCfg{
		Enabled:        utils.BoolPointer(false),
		Sessions_conns: &[]string{utils.MetaInternal},
		Create_cdr:     utils.BoolPointer(false),
		Evapi_conns: &[]*KamConnJsonCfg{
			{
				Address:    utils.StringPointer("127.0.0.1:8448"),
				Reconnects: utils.IntPointer(5),
			},
		},
	}
	if cfg, err := dfCgrJsonCfg.KamAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expecting: %s, received: %s: ",
			utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestAsteriskAgentJsonCfg(t *testing.T) {
	eCfg := &AsteriskAgentJsonCfg{
		Enabled:        utils.BoolPointer(false),
		Sessions_conns: &[]string{utils.MetaInternal},
		Create_cdr:     utils.BoolPointer(false),
		Asterisk_conns: &[]*AstConnJsonCfg{
			{
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
		Enabled:              utils.BoolPointer(false),
		Listen:               utils.StringPointer("127.0.0.1:3868"),
		Listen_net:           utils.StringPointer(utils.TCP),
		Dictionaries_path:    utils.StringPointer("/usr/share/cgrates/diameter/dict/"),
		Sessions_conns:       &[]string{utils.MetaInternal},
		Origin_host:          utils.StringPointer("CGR-DA"),
		Origin_realm:         utils.StringPointer("cgrates.org"),
		Vendor_id:            utils.IntPointer(0),
		Product_name:         utils.StringPointer("CGRateS"),
		Concurrent_requests:  utils.IntPointer(-1),
		Synced_conn_requests: utils.BoolPointer(false),
		Asr_template:         utils.StringPointer(""),
		Templates: map[string][]*FcTemplateJsonCfg{
			utils.MetaErr: {
				{Tag: utils.StringPointer("SessionId"),
					Path:      utils.StringPointer("*rep.Session-Id"),
					Type:      utils.StringPointer(utils.META_COMPOSED),
					Value:     utils.StringPointer("~*req.Session-Id"),
					Mandatory: utils.BoolPointer(true)},
				{Tag: utils.StringPointer("OriginHost"),
					Path:      utils.StringPointer("*rep.Origin-Host"),
					Type:      utils.StringPointer(utils.META_COMPOSED),
					Value:     utils.StringPointer("~*vars.OriginHost"),
					Mandatory: utils.BoolPointer(true)},
				{Tag: utils.StringPointer("OriginRealm"),
					Path:      utils.StringPointer("*rep.Origin-Realm"),
					Type:      utils.StringPointer(utils.META_COMPOSED),
					Value:     utils.StringPointer("~*vars.OriginRealm"),
					Mandatory: utils.BoolPointer(true)},
			},
			utils.MetaCCA: {
				{Tag: utils.StringPointer("SessionId"),
					Path:      utils.StringPointer("*rep.Session-Id"),
					Type:      utils.StringPointer(utils.META_COMPOSED),
					Value:     utils.StringPointer("~*req.Session-Id"),
					Mandatory: utils.BoolPointer(true)},
				{Tag: utils.StringPointer("ResultCode"),
					Path:  utils.StringPointer("*rep.Result-Code"),
					Type:  utils.StringPointer(utils.META_CONSTANT),
					Value: utils.StringPointer("2001")},
				{Tag: utils.StringPointer("OriginHost"),
					Path:      utils.StringPointer("*rep.Origin-Host"),
					Type:      utils.StringPointer(utils.META_COMPOSED),
					Value:     utils.StringPointer("~*vars.OriginHost"),
					Mandatory: utils.BoolPointer(true)},
				{Tag: utils.StringPointer("OriginRealm"),
					Path:      utils.StringPointer("*rep.Origin-Realm"),
					Type:      utils.StringPointer(utils.META_COMPOSED),
					Value:     utils.StringPointer("~*vars.OriginRealm"),
					Mandatory: utils.BoolPointer(true)},
				{Tag: utils.StringPointer("AuthApplicationId"),
					Path:      utils.StringPointer("*rep.Auth-Application-Id"),
					Type:      utils.StringPointer(utils.META_COMPOSED),
					Value:     utils.StringPointer("~*vars.*appid"),
					Mandatory: utils.BoolPointer(true)},
				{Tag: utils.StringPointer("CCRequestType"),
					Path:      utils.StringPointer("*rep.CC-Request-Type"),
					Type:      utils.StringPointer(utils.META_COMPOSED),
					Value:     utils.StringPointer("~*req.CC-Request-Type"),
					Mandatory: utils.BoolPointer(true)},
				{Tag: utils.StringPointer("CCRequestNumber"),
					Path:      utils.StringPointer("*rep.CC-Request-Number"),
					Type:      utils.StringPointer(utils.META_COMPOSED),
					Value:     utils.StringPointer("~*req.CC-Request-Number"),
					Mandatory: utils.BoolPointer(true)},
			},
			utils.MetaASR: {
				{Tag: utils.StringPointer("SessionId"),
					Path:      utils.StringPointer("*diamreq.Session-Id"),
					Type:      utils.StringPointer(utils.MetaVariable),
					Value:     utils.StringPointer("~*req.Session-Id"),
					Mandatory: utils.BoolPointer(true)},
				{Tag: utils.StringPointer("OriginHost"),
					Path:      utils.StringPointer("*diamreq.Origin-Host"),
					Type:      utils.StringPointer(utils.MetaVariable),
					Value:     utils.StringPointer("~*req.Destination-Host"),
					Mandatory: utils.BoolPointer(true)},
				{Tag: utils.StringPointer("OriginRealm"),
					Path:      utils.StringPointer("*diamreq.Origin-Realm"),
					Type:      utils.StringPointer(utils.MetaVariable),
					Value:     utils.StringPointer("~*req.Destination-Realm"),
					Mandatory: utils.BoolPointer(true)},
				{Tag: utils.StringPointer("DestinationRealm"),
					Path:      utils.StringPointer("*diamreq.Destination-Realm"),
					Type:      utils.StringPointer(utils.MetaVariable),
					Value:     utils.StringPointer("~*req.Origin-Realm"),
					Mandatory: utils.BoolPointer(true)},
				{Tag: utils.StringPointer("DestinationHost"),
					Path:      utils.StringPointer("*diamreq.Destination-Host"),
					Type:      utils.StringPointer(utils.MetaVariable),
					Value:     utils.StringPointer("~*req.Origin-Host"),
					Mandatory: utils.BoolPointer(true)},
				{Tag: utils.StringPointer("AuthApplicationId"),
					Path:      utils.StringPointer("*diamreq.Auth-Application-Id"),
					Type:      utils.StringPointer(utils.MetaVariable),
					Value:     utils.StringPointer("~*vars.*appid"),
					Mandatory: utils.BoolPointer(true)},
				{Tag: utils.StringPointer("UserName"),
					Path:      utils.StringPointer("*diamreq.User-Name"),
					Type:      utils.StringPointer(utils.MetaVariable),
					Value:     utils.StringPointer("~*req.User-Name"),
					Mandatory: utils.BoolPointer(true)},
				{Tag: utils.StringPointer("OriginStateID"),
					Path:  utils.StringPointer("*diamreq.Origin-State-Id"),
					Type:  utils.StringPointer(utils.META_CONSTANT),
					Value: utils.StringPointer("1")},
			},
		},
		Request_processors: &[]*ReqProcessorJsnCfg{},
	}
	if cfg, err := dfCgrJsonCfg.DiameterAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("expecting: %s, \n\nreceived: %s", utils.ToIJSON(eCfg), utils.ToIJSON(cfg))
	}
}

func TestRadiusAgentJsonCfg(t *testing.T) {
	eCfg := &RadiusAgentJsonCfg{
		Enabled:     utils.BoolPointer(false),
		Listen_net:  utils.StringPointer("udp"),
		Listen_auth: utils.StringPointer("127.0.0.1:1812"),
		Listen_acct: utils.StringPointer("127.0.0.1:1813"),
		Client_secrets: utils.MapStringStringPointer(map[string]string{
			utils.MetaDefault: "CGRateS.org",
		}),
		Client_dictionaries: utils.MapStringStringPointer(map[string]string{
			utils.MetaDefault: "/usr/share/cgrates/radius/dict/",
		}),
		Sessions_conns:     &[]string{utils.MetaInternal},
		Request_processors: &[]*ReqProcessorJsnCfg{},
	}
	if cfg, err := dfCgrJsonCfg.RadiusAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		rcv := *cfg.Request_processors
		t.Errorf("Received: %+v", rcv)
	}
}

func TestHttpAgentJsonCfg(t *testing.T) {
	eCfg := &[]*HttpAgentJsonCfg{}
	if cfg, err := dfCgrJsonCfg.HttpAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDNSAgentJsonCfg(t *testing.T) {
	eCfg := &DNSAgentJsonCfg{
		Enabled:            utils.BoolPointer(false),
		Listen_net:         utils.StringPointer("udp"),
		Listen:             utils.StringPointer("127.0.0.1:2053"),
		Sessions_conns:     &[]string{utils.ConcatenatedKey(utils.MetaInternal)},
		Timezone:           utils.StringPointer(""),
		Request_processors: &[]*ReqProcessorJsnCfg{},
	}
	if cfg, err := dfCgrJsonCfg.DNSAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfAttributeServJsonCfg(t *testing.T) {
	eCfg := &AttributeSJsonCfg{
		Enabled:               utils.BoolPointer(false),
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: nil,
		Prefix_indexed_fields: &[]string{},
		Process_runs:          utils.IntPointer(1),
		Nested_fields:         utils.BoolPointer(false),
	}
	if cfg, err := dfCgrJsonCfg.AttributeServJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", utils.ToJSON(cfg))
	}
}

func TestDfChargerServJsonCfg(t *testing.T) {
	eCfg := &ChargerSJsonCfg{
		Enabled:               utils.BoolPointer(false),
		Indexed_selects:       utils.BoolPointer(true),
		Attributes_conns:      &[]string{},
		String_indexed_fields: nil,
		Prefix_indexed_fields: &[]string{},
		Nested_fields:         utils.BoolPointer(false),
	}
	if cfg, err := dfCgrJsonCfg.ChargerServJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", utils.ToJSON(cfg))
	}
}

func TestDfFilterSJsonCfg(t *testing.T) {
	eCfg := &FilterSJsonCfg{
		Stats_conns:     &[]string{},
		Resources_conns: &[]string{},
		Rals_conns:      &[]string{},
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
		Indexed_selects:       utils.BoolPointer(true),
		Thresholds_conns:      &[]string{},
		Store_interval:        utils.StringPointer(""),
		String_indexed_fields: nil,
		Prefix_indexed_fields: &[]string{},
		Nested_fields:         utils.BoolPointer(false),
	}
	if cfg, err := dfCgrJsonCfg.ResourceSJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfStatServiceJsonCfg(t *testing.T) {
	eCfg := &StatServJsonCfg{
		Enabled:                  utils.BoolPointer(false),
		Indexed_selects:          utils.BoolPointer(true),
		Store_interval:           utils.StringPointer(""),
		Store_uncompressed_limit: utils.IntPointer(0),
		Thresholds_conns:         &[]string{},
		String_indexed_fields:    nil,
		Prefix_indexed_fields:    &[]string{},
		Nested_fields:            utils.BoolPointer(false),
	}
	if cfg, err := dfCgrJsonCfg.StatSJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", utils.ToJSON(cfg))
	}
}

func TestDfThresholdSJsonCfg(t *testing.T) {
	eCfg := &ThresholdSJsonCfg{
		Enabled:               utils.BoolPointer(false),
		Indexed_selects:       utils.BoolPointer(true),
		Store_interval:        utils.StringPointer(""),
		String_indexed_fields: nil,
		Prefix_indexed_fields: &[]string{},
		Nested_fields:         utils.BoolPointer(false),
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
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: nil,
		Prefix_indexed_fields: &[]string{},
		Attributes_conns:      &[]string{},
		Resources_conns:       &[]string{},
		Stats_conns:           &[]string{},
		Default_ratio:         utils.IntPointer(1),
		Nested_fields:         utils.BoolPointer(false),
	}
	if cfg, err := dfCgrJsonCfg.SupplierSJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfLoaderJsonCfg(t *testing.T) {
	eCfg := []*LoaderJsonCfg{
		{
			ID:              utils.StringPointer(utils.MetaDefault),
			Enabled:         utils.BoolPointer(false),
			Tenant:          utils.StringPointer(""),
			Dry_run:         utils.BoolPointer(false),
			Run_delay:       utils.IntPointer(0),
			Lock_filename:   utils.StringPointer(".cgr.lck"),
			Caches_conns:    &[]string{utils.MetaInternal},
			Field_separator: utils.StringPointer(","),
			Tp_in_dir:       utils.StringPointer("/var/spool/cgrates/loader/in"),
			Tp_out_dir:      utils.StringPointer("/var/spool/cgrates/loader/out"),
			Data: &[]*LoaderJsonDataType{
				{
					Type:      utils.StringPointer(utils.MetaAttributes),
					File_name: utils.StringPointer(utils.AttributesCsv),
					Fields: &[]*FcTemplateJsonCfg{
						{Tag: utils.StringPointer("TenantID"),
							Field_id:  utils.StringPointer(utils.Tenant),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~0"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("ProfileID"),
							Field_id:  utils.StringPointer(utils.ID),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~1"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("Contexts"),
							Field_id: utils.StringPointer(utils.Contexts),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~2")},
						{Tag: utils.StringPointer("FilterIDs"),
							Field_id: utils.StringPointer(utils.FilterIDs),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~3")},
						{Tag: utils.StringPointer("ActivationInterval"),
							Field_id: utils.StringPointer("ActivationInterval"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~4")},
						{Tag: utils.StringPointer("AttributeFilterIDs"),
							Field_id: utils.StringPointer("AttributeFilterIDs"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~5")},
						{Tag: utils.StringPointer("Path"),
							Field_id: utils.StringPointer(utils.Path),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~6")},
						{Tag: utils.StringPointer("Type"),
							Field_id: utils.StringPointer("Type"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~7")},
						{Tag: utils.StringPointer("Value"),
							Field_id: utils.StringPointer("Value"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~8")},
						{Tag: utils.StringPointer("Blocker"),
							Field_id: utils.StringPointer("Blocker"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~9")},
						{Tag: utils.StringPointer("Weight"),
							Field_id: utils.StringPointer(utils.Weight),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~10")},
					},
				},
				{
					Type:      utils.StringPointer(utils.MetaFilters),
					File_name: utils.StringPointer(utils.FiltersCsv),
					Fields: &[]*FcTemplateJsonCfg{
						{Tag: utils.StringPointer(utils.Tenant),
							Field_id:  utils.StringPointer(utils.Tenant),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~0"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer(utils.ID),
							Field_id:  utils.StringPointer(utils.ID),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~1"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("Type"),
							Field_id: utils.StringPointer("Type"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~2")},
						{Tag: utils.StringPointer("Element"),
							Field_id: utils.StringPointer("Element"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~3")},
						{Tag: utils.StringPointer("Values"),
							Field_id: utils.StringPointer("Values"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~4")},
						{Tag: utils.StringPointer("ActivationInterval"),
							Field_id: utils.StringPointer("ActivationInterval"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~5")},
					},
				},
				{
					Type:      utils.StringPointer(utils.MetaResources),
					File_name: utils.StringPointer(utils.ResourcesCsv),
					Fields: &[]*FcTemplateJsonCfg{
						{Tag: utils.StringPointer(utils.Tenant),
							Field_id:  utils.StringPointer(utils.Tenant),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~0"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer(utils.ID),
							Field_id:  utils.StringPointer(utils.ID),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~1"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("FilterIDs"),
							Field_id: utils.StringPointer("FilterIDs"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~2")},
						{Tag: utils.StringPointer("ActivationInterval"),
							Field_id: utils.StringPointer("ActivationInterval"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~3")},
						{Tag: utils.StringPointer("TTL"),
							Field_id: utils.StringPointer("UsageTTL"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~4")},
						{Tag: utils.StringPointer("Limit"),
							Field_id: utils.StringPointer("Limit"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~5")},
						{Tag: utils.StringPointer("AllocationMessage"),
							Field_id: utils.StringPointer("AllocationMessage"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~6")},
						{Tag: utils.StringPointer("Blocker"),
							Field_id: utils.StringPointer("Blocker"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~7")},
						{Tag: utils.StringPointer("Stored"),
							Field_id: utils.StringPointer("Stored"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~8")},
						{Tag: utils.StringPointer("Weight"),
							Field_id: utils.StringPointer("Weight"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~9")},
						{Tag: utils.StringPointer("ThresholdIDs"),
							Field_id: utils.StringPointer("ThresholdIDs"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~10")},
					},
				},
				{
					Type:      utils.StringPointer(utils.MetaStats),
					File_name: utils.StringPointer(utils.StatsCsv),
					Fields: &[]*FcTemplateJsonCfg{
						{Tag: utils.StringPointer(utils.Tenant),
							Field_id:  utils.StringPointer(utils.Tenant),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~0"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer(utils.ID),
							Field_id:  utils.StringPointer(utils.ID),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~1"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("FilterIDs"),
							Field_id: utils.StringPointer("FilterIDs"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~2")},
						{Tag: utils.StringPointer("ActivationInterval"),
							Field_id: utils.StringPointer("ActivationInterval"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~3")},
						{Tag: utils.StringPointer("QueueLength"),
							Field_id: utils.StringPointer("QueueLength"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~4")},
						{Tag: utils.StringPointer("TTL"),
							Field_id: utils.StringPointer("TTL"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~5")},
						{Tag: utils.StringPointer("MinItems"),
							Field_id: utils.StringPointer("MinItems"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~6")},
						{Tag: utils.StringPointer("MetricIDs"),
							Field_id: utils.StringPointer("MetricIDs"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~7")},
						{Tag: utils.StringPointer("MetricFilterIDs"),
							Field_id: utils.StringPointer("MetricFilterIDs"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~8")},
						{Tag: utils.StringPointer("Blocker"),
							Field_id: utils.StringPointer("Blocker"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~9")},
						{Tag: utils.StringPointer("Stored"),
							Field_id: utils.StringPointer("Stored"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~10")},
						{Tag: utils.StringPointer("Weight"),
							Field_id: utils.StringPointer("Weight"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~11")},

						{Tag: utils.StringPointer("ThresholdIDs"),
							Field_id: utils.StringPointer("ThresholdIDs"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~12")},
					},
				},
				{
					Type:      utils.StringPointer(utils.MetaThresholds),
					File_name: utils.StringPointer(utils.ThresholdsCsv),
					Fields: &[]*FcTemplateJsonCfg{
						{Tag: utils.StringPointer(utils.Tenant),
							Field_id:  utils.StringPointer(utils.Tenant),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~0"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer(utils.ID),
							Field_id:  utils.StringPointer(utils.ID),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~1"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("FilterIDs"),
							Field_id: utils.StringPointer("FilterIDs"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~2")},
						{Tag: utils.StringPointer("ActivationInterval"),
							Field_id: utils.StringPointer("ActivationInterval"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~3")},
						{Tag: utils.StringPointer("MaxHits"),
							Field_id: utils.StringPointer("MaxHits"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~4")},
						{Tag: utils.StringPointer("MinHits"),
							Field_id: utils.StringPointer("MinHits"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~5")},
						{Tag: utils.StringPointer("MinSleep"),
							Field_id: utils.StringPointer("MinSleep"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~6")},
						{Tag: utils.StringPointer("Blocker"),
							Field_id: utils.StringPointer("Blocker"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~7")},
						{Tag: utils.StringPointer("Weight"),
							Field_id: utils.StringPointer("Weight"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~8")},
						{Tag: utils.StringPointer("ActionIDs"),
							Field_id: utils.StringPointer("ActionIDs"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~9")},
						{Tag: utils.StringPointer("Async"),
							Field_id: utils.StringPointer("Async"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~10")},
					},
				},
				{
					Type:      utils.StringPointer(utils.MetaSuppliers),
					File_name: utils.StringPointer(utils.SuppliersCsv),
					Fields: &[]*FcTemplateJsonCfg{
						{Tag: utils.StringPointer(utils.Tenant),
							Field_id:  utils.StringPointer(utils.Tenant),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~0"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer(utils.ID),
							Field_id:  utils.StringPointer(utils.ID),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~1"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("FilterIDs"),
							Field_id: utils.StringPointer("FilterIDs"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~2")},
						{Tag: utils.StringPointer("ActivationInterval"),
							Field_id: utils.StringPointer("ActivationInterval"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~3")},
						{Tag: utils.StringPointer("Sorting"),
							Field_id: utils.StringPointer("Sorting"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~4")},
						{Tag: utils.StringPointer("SortingParamameters"),
							Field_id: utils.StringPointer("SortingParamameters"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~5")},
						{Tag: utils.StringPointer("SupplierID"),
							Field_id: utils.StringPointer("SupplierID"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~6")},
						{Tag: utils.StringPointer("SupplierFilterIDs"),
							Field_id: utils.StringPointer("SupplierFilterIDs"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~7")},
						{Tag: utils.StringPointer("SupplierAccountIDs"),
							Field_id: utils.StringPointer("SupplierAccountIDs"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~8")},
						{Tag: utils.StringPointer("SupplierRatingPlanIDs"),
							Field_id: utils.StringPointer("SupplierRatingPlanIDs"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~9")},
						{Tag: utils.StringPointer("SupplierResourceIDs"),
							Field_id: utils.StringPointer("SupplierResourceIDs"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~10")},
						{Tag: utils.StringPointer("SupplierStatIDs"),
							Field_id: utils.StringPointer("SupplierStatIDs"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~11")},
						{Tag: utils.StringPointer("SupplierWeight"),
							Field_id: utils.StringPointer("SupplierWeight"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~12")},
						{Tag: utils.StringPointer("SupplierBlocker"),
							Field_id: utils.StringPointer("SupplierBlocker"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~13")},
						{Tag: utils.StringPointer("SupplierParameters"),
							Field_id: utils.StringPointer("SupplierParameters"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~14")},
						{Tag: utils.StringPointer("Weight"),
							Field_id: utils.StringPointer("Weight"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~15")},
					},
				},
				{
					Type:      utils.StringPointer(utils.MetaChargers),
					File_name: utils.StringPointer(utils.ChargersCsv),
					Fields: &[]*FcTemplateJsonCfg{
						{Tag: utils.StringPointer(utils.Tenant),
							Field_id:  utils.StringPointer(utils.Tenant),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~0"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer(utils.ID),
							Field_id:  utils.StringPointer(utils.ID),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~1"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("FilterIDs"),
							Field_id: utils.StringPointer("FilterIDs"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~2")},
						{Tag: utils.StringPointer("ActivationInterval"),
							Field_id: utils.StringPointer("ActivationInterval"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~3")},
						{Tag: utils.StringPointer("RunID"),
							Field_id: utils.StringPointer("RunID"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~4")},
						{Tag: utils.StringPointer("AttributeIDs"),
							Field_id: utils.StringPointer("AttributeIDs"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~5")},
						{Tag: utils.StringPointer("Weight"),
							Field_id: utils.StringPointer("Weight"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~6")},
					},
				},
				{
					Type:      utils.StringPointer(utils.MetaDispatchers),
					File_name: utils.StringPointer(utils.DispatcherProfilesCsv),
					Fields: &[]*FcTemplateJsonCfg{
						{Tag: utils.StringPointer(utils.Tenant),
							Field_id:  utils.StringPointer(utils.Tenant),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~0"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer(utils.ID),
							Field_id:  utils.StringPointer(utils.ID),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~1"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("Contexts"),
							Field_id: utils.StringPointer("Contexts"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~2")},
						{Tag: utils.StringPointer("FilterIDs"),
							Field_id: utils.StringPointer("FilterIDs"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~3")},
						{Tag: utils.StringPointer("ActivationInterval"),
							Field_id: utils.StringPointer("ActivationInterval"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~4")},
						{Tag: utils.StringPointer("Strategy"),
							Field_id: utils.StringPointer("Strategy"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~5")},
						{Tag: utils.StringPointer("StrategyParameters"),
							Field_id: utils.StringPointer("StrategyParameters"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~6")},
						{Tag: utils.StringPointer("ConnID"),
							Field_id: utils.StringPointer("ConnID"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~7")},
						{Tag: utils.StringPointer("ConnFilterIDs"),
							Field_id: utils.StringPointer("ConnFilterIDs"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~8")},
						{Tag: utils.StringPointer("ConnWeight"),
							Field_id: utils.StringPointer("ConnWeight"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~9")},
						{Tag: utils.StringPointer("ConnBlocker"),
							Field_id: utils.StringPointer("ConnBlocker"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~10")},
						{Tag: utils.StringPointer("ConnParameters"),
							Field_id: utils.StringPointer("ConnParameters"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~11")},
						{Tag: utils.StringPointer("Weight"),
							Field_id: utils.StringPointer("Weight"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~12")},
					},
				},
				{
					Type:      utils.StringPointer(utils.MetaDispatcherHosts),
					File_name: utils.StringPointer(utils.DispatcherHostsCsv),
					Fields: &[]*FcTemplateJsonCfg{
						{Tag: utils.StringPointer(utils.Tenant),
							Field_id:  utils.StringPointer(utils.Tenant),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~0"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer(utils.ID),
							Field_id:  utils.StringPointer(utils.ID),
							Type:      utils.StringPointer(utils.MetaVariable),
							Value:     utils.StringPointer("~1"),
							Mandatory: utils.BoolPointer(true)},
						{Tag: utils.StringPointer("Address"),
							Field_id: utils.StringPointer("Address"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~2")},
						{Tag: utils.StringPointer("Transport"),
							Field_id: utils.StringPointer("Transport"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~3")},
						{Tag: utils.StringPointer("TLS"),
							Field_id: utils.StringPointer("TLS"),
							Type:     utils.StringPointer(utils.MetaVariable),
							Value:    utils.StringPointer("~4")},
					},
				},
			},
		},
	}
	if cfg, err := dfCgrJsonCfg.LoaderJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expecting: %s, received: %s ",
			utils.ToIJSON(eCfg), utils.ToIJSON(cfg))
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
		Client_tracking:         utils.StringPointer(utils.DynamicDataPrefix + utils.CGRID),
		Customer_number:         utils.StringPointer("~Subject"),
		Orig_number:             utils.StringPointer("~Subject"),
		Term_number:             utils.StringPointer("~Destination"),
		Bill_to_number:          utils.StringPointer(""),
		Zipcode:                 utils.StringPointer(""),
		Plus4:                   utils.StringPointer(""),
		P2PZipcode:              utils.StringPointer(""),
		P2PPlus4:                utils.StringPointer(""),
		Units:                   utils.StringPointer("1"),
		Unit_type:               utils.StringPointer("00"),
		Tax_included:            utils.StringPointer("0"),
		Tax_situs_rule:          utils.StringPointer("04"),
		Trans_type_code:         utils.StringPointer("010101"),
		Sales_type_code:         utils.StringPointer("R"),
		Tax_exemption_code_list: utils.StringPointer(""),
	}
	if cfg, err := dfCgrJsonCfg.SureTaxJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

// func TestNewCgrJsonCfgFromFile(t *testing.T) {
// 	cgrJsonCfg, err := NewCgrJsonCfgFromFile("cfg_data.json")
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	eCfg := &GeneralJsonCfg{Default_request_type: utils.StringPointer(utils.META_PSEUDOPREPAID)}
// 	if gCfg, err := cgrJsonCfg.GeneralJsonCfg(); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(eCfg, gCfg) {
// 		t.Errorf("Expecting: %+v, received: %+v", eCfg, gCfg)
// 	}
// 	cdrFields := []*FcTemplateJsonCfg{
// 		{Field_id: utils.StringPointer(utils.ToR), Value: utils.StringPointer("~7:s/^(voice|data|sms|mms|generic)$/*$1/")},
// 		{Field_id: utils.StringPointer(utils.AnswerTime), Value: utils.StringPointer("~1")},
// 		{Field_id: utils.StringPointer(utils.Usage), Value: utils.StringPointer(`~9:s/^(\d+)$/${1}s/`)},
// 	}
// 	eCfgCdrc := []*CdrcJsonCfg{
// 		{
// 			Id:            utils.StringPointer("CDRC-CSV1"),
// 			Enabled:       utils.BoolPointer(true),
// 			Cdr_in_path:   utils.StringPointer("/tmp/cgrates/cdrc1/in"),
// 			Cdr_out_path:  utils.StringPointer("/tmp/cgrates/cdrc1/out"),
// 			Cdr_source_id: utils.StringPointer("csv1"),
// 		},
// 		{
// 			Id:             utils.StringPointer("CDRC-CSV2"),
// 			Enabled:        utils.BoolPointer(true),
// 			Run_delay:      utils.IntPointer(1),
// 			Cdr_in_path:    utils.StringPointer("/tmp/cgrates/cdrc2/in"),
// 			Cdr_out_path:   utils.StringPointer("/tmp/cgrates/cdrc2/out"),
// 			Cdr_source_id:  utils.StringPointer("csv2"),
// 			Content_fields: &cdrFields,
// 		},
// 	}
// 	if cfg, err := cgrJsonCfg.CdrcJsonCfg(); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(eCfgCdrc, cfg) {
// 		t.Errorf("Expecting: %+v \n received: %+v", utils.ToIJSON(eCfgCdrc), utils.ToIJSON(cfg))
// 	}
// 	eCfgSmFs := &FreeswitchAgentJsonCfg{
// 		Enabled: utils.BoolPointer(true),
// 		Event_socket_conns: &[]*FsConnJsonCfg{
// 			{
// 				Address:    utils.StringPointer("1.2.3.4:8021"),
// 				Password:   utils.StringPointer("ClueCon"),
// 				Reconnects: utils.IntPointer(5),
// 			},
// 			{
// 				Address:    utils.StringPointer("2.3.4.5:8021"),
// 				Password:   utils.StringPointer("ClueCon"),
// 				Reconnects: utils.IntPointer(5),
// 			},
// 		},
// 	}
// 	if smFsCfg, err := cgrJsonCfg.FreeswitchAgentJsonCfg(); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(eCfgSmFs, smFsCfg) {
// 		t.Error("Received: ", smFsCfg)
// 	}
// }

func TestDfHttpJsonCfg(t *testing.T) {
	eCfg := &HTTPJsonCfg{
		Json_rpc_url:        utils.StringPointer("/jsonrpc"),
		Ws_url:              utils.StringPointer("/ws"),
		Freeswitch_cdrs_url: utils.StringPointer("/freeswitch_json"),
		Http_Cdrs:           utils.StringPointer("/cdr_http"),
		Use_basic_auth:      utils.BoolPointer(false),
		Auth_users:          utils.MapStringStringPointer(map[string]string{}),
	}
	if cfg, err := dfCgrJsonCfg.HttpJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestDfDispatcherSJsonCfg(t *testing.T) {
	eCfg := &DispatcherSJsonCfg{
		Enabled:               utils.BoolPointer(false),
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: nil,
		Prefix_indexed_fields: &[]string{},
		Attributes_conns:      &[]string{},
		Nested_fields:         utils.BoolPointer(false),
	}
	if cfg, err := dfCgrJsonCfg.DispatcherSJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfLoaderCfg(t *testing.T) {
	eCfg := &LoaderCfgJson{
		Tpid:            utils.StringPointer(""),
		Data_path:       utils.StringPointer("./"),
		Disable_reverse: utils.BoolPointer(false),
		Field_separator: utils.StringPointer(","),
		Caches_conns:    &[]string{utils.MetaLocalHost},
		Scheduler_conns: &[]string{utils.MetaLocalHost},
	}
	if cfg, err := dfCgrJsonCfg.LoaderCfgJson(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfMigratorCfg(t *testing.T) {
	eCfg := &MigratorCfgJson{
		Out_dataDB_type:     utils.StringPointer("redis"),
		Out_dataDB_host:     utils.StringPointer("127.0.0.1"),
		Out_dataDB_port:     utils.StringPointer("6379"),
		Out_dataDB_name:     utils.StringPointer("10"),
		Out_dataDB_user:     utils.StringPointer("cgrates"),
		Out_dataDB_password: utils.StringPointer(""),
		Out_dataDB_encoding: utils.StringPointer("msgpack"),
		Out_storDB_type:     utils.StringPointer("mysql"),
		Out_storDB_host:     utils.StringPointer("127.0.0.1"),
		Out_storDB_port:     utils.StringPointer("3306"),
		Out_storDB_name:     utils.StringPointer("cgrates"),
		Out_storDB_user:     utils.StringPointer("cgrates"),
		Out_storDB_password: utils.StringPointer(""),
		Users_filters:       &[]string{},
	}
	if cfg, err := dfCgrJsonCfg.MigratorCfgJson(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfTlsCfg(t *testing.T) {
	eCfg := &TlsJsonCfg{
		Server_certificate: utils.StringPointer(""),
		Server_key:         utils.StringPointer(""),
		Ca_certificate:     utils.StringPointer(""),
		Client_certificate: utils.StringPointer(""),
		Client_key:         utils.StringPointer(""),
		Server_name:        utils.StringPointer(""),
		Server_policy:      utils.IntPointer(4),
	}
	if cfg, err := dfCgrJsonCfg.TlsCfgJson(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfAnalyzerCfg(t *testing.T) {
	eCfg := &AnalyzerSJsonCfg{
		Enabled: utils.BoolPointer(false),
	}
	if cfg, err := dfCgrJsonCfg.AnalyzerCfgJson(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfApierCfg(t *testing.T) {
	eCfg := &ApierJsonCfg{
		Caches_conns:     &[]string{utils.MetaInternal},
		Scheduler_conns:  &[]string{},
		Attributes_conns: &[]string{},
	}
	if cfg, err := dfCgrJsonCfg.ApierCfgJson(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}

func TestDfEventReaderCfg(t *testing.T) {
	cdrFields := []*FcTemplateJsonCfg{
		{Tag: utils.StringPointer("TOR"), Field_id: utils.StringPointer(utils.ToR), Type: utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer("~*req.2"), Mandatory: utils.BoolPointer(true)},
		{Tag: utils.StringPointer("OriginID"), Field_id: utils.StringPointer(utils.OriginID), Type: utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer("~*req.3"), Mandatory: utils.BoolPointer(true)},
		{Tag: utils.StringPointer("RequestType"), Field_id: utils.StringPointer(utils.RequestType), Type: utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer("~*req.4"), Mandatory: utils.BoolPointer(true)},
		{Tag: utils.StringPointer("Tenant"), Field_id: utils.StringPointer(utils.Tenant), Type: utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer("~*req.6"), Mandatory: utils.BoolPointer(true)},
		{Tag: utils.StringPointer("Category"), Field_id: utils.StringPointer(utils.Category), Type: utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer("~*req.7"), Mandatory: utils.BoolPointer(true)},
		{Tag: utils.StringPointer("Account"), Field_id: utils.StringPointer(utils.Account), Type: utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer("~*req.8"), Mandatory: utils.BoolPointer(true)},
		{Tag: utils.StringPointer("Subject"), Field_id: utils.StringPointer(utils.Subject), Type: utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer("~*req.9"), Mandatory: utils.BoolPointer(true)},
		{Tag: utils.StringPointer("Destination"), Field_id: utils.StringPointer(utils.Destination), Type: utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer("~*req.10"), Mandatory: utils.BoolPointer(true)},
		{Tag: utils.StringPointer("SetupTime"), Field_id: utils.StringPointer(utils.SetupTime), Type: utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer("~*req.11"), Mandatory: utils.BoolPointer(true)},
		{Tag: utils.StringPointer("AnswerTime"), Field_id: utils.StringPointer(utils.AnswerTime), Type: utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer("~*req.12"), Mandatory: utils.BoolPointer(true)},
		{Tag: utils.StringPointer("Usage"), Field_id: utils.StringPointer(utils.Usage), Type: utils.StringPointer(utils.META_COMPOSED),
			Value: utils.StringPointer("~*req.13"), Mandatory: utils.BoolPointer(true)},
	}
	eCfg := &ERsJsonCfg{
		Enabled:        utils.BoolPointer(false),
		Sessions_conns: &[]string{utils.MetaInternal},
		Readers: &[]*EventReaderJsonCfg{
			&EventReaderJsonCfg{
				Id:                  utils.StringPointer(utils.MetaDefault),
				Type:                utils.StringPointer(utils.MetaFileCSV),
				Field_separator:     utils.StringPointer(","),
				Run_delay:           utils.IntPointer(0),
				Concurrent_requests: utils.IntPointer(1024),
				Source_path:         utils.StringPointer("/var/spool/cgrates/cdrc/in"),
				Processed_path:      utils.StringPointer("/var/spool/cgrates/cdrc/out"),
				Xml_root_path:       utils.StringPointer(utils.EmptyString),
				Tenant:              utils.StringPointer(utils.EmptyString),
				Timezone:            utils.StringPointer(utils.EmptyString),
				Filters:             &[]string{},
				Flags:               &[]string{},
				Header_fields:       &[]*FcTemplateJsonCfg{},
				Content_fields:      &cdrFields,
				Trailer_fields:      &[]*FcTemplateJsonCfg{},
			},
		},
	}
	if cfg, err := dfCgrJsonCfg.ERsJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Errorf("Expected: %+v, \nreceived: %+v", utils.ToJSON(eCfg), utils.ToJSON(cfg))
	}
}
