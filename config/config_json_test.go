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
		Http_skip_tls_verify: utils.BoolPointer(false),
		Rounding_decimals:    utils.IntPointer(10),
		Dbdata_encoding:      utils.StringPointer("msgpack"),
		Tpexport_dir:         utils.StringPointer("/var/log/cgrates/tpe"),
		Http_failed_dir:      utils.StringPointer("/var/log/cgrates/http_failed"),
		Default_reqtype:      utils.StringPointer(utils.META_RATED),
		Default_category:     utils.StringPointer("call"),
		Default_tenant:       utils.StringPointer("cgrates.org"),
		Default_subject:      utils.StringPointer("cgrates"),
		Default_timezone:     utils.StringPointer("Local"),
		Connect_attempts:     utils.IntPointer(3),
		Reconnects:           utils.IntPointer(-1),
		Response_cache_ttl:   utils.StringPointer("3s"),
		Internal_ttl:         utils.StringPointer("2m")}
	if gCfg, err := dfCgrJsonCfg.GeneralJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, gCfg) {
		t.Error("Received: ", gCfg)
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

func TestDfDbJsonCfg(t *testing.T) {
	eCfg := &DbJsonCfg{
		Db_type:   utils.StringPointer("redis"),
		Db_host:   utils.StringPointer("127.0.0.1"),
		Db_port:   utils.IntPointer(6379),
		Db_name:   utils.StringPointer("10"),
		Db_user:   utils.StringPointer(""),
		Db_passwd: utils.StringPointer(""),
	}
	if cfg, err := dfCgrJsonCfg.DbJsonCfg(TPDB_JSN); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
	eCfg = &DbJsonCfg{
		Db_type:           utils.StringPointer("redis"),
		Db_host:           utils.StringPointer("127.0.0.1"),
		Db_port:           utils.IntPointer(6379),
		Db_name:           utils.StringPointer("11"),
		Db_user:           utils.StringPointer(""),
		Db_passwd:         utils.StringPointer(""),
		Load_history_size: utils.IntPointer(10),
	}
	if cfg, err := dfCgrJsonCfg.DbJsonCfg(DATADB_JSN); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
	eCfg = &DbJsonCfg{
		Db_type:        utils.StringPointer("mysql"),
		Db_host:        utils.StringPointer("127.0.0.1"),
		Db_port:        utils.IntPointer(3306),
		Db_name:        utils.StringPointer("cgrates"),
		Db_user:        utils.StringPointer("cgrates"),
		Db_passwd:      utils.StringPointer("CGRateS.org"),
		Max_open_conns: utils.IntPointer(100),
		Max_idle_conns: utils.IntPointer(10),
	}
	if cfg, err := dfCgrJsonCfg.DbJsonCfg(STORDB_JSN); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestDfBalancerJsonCfg(t *testing.T) {
	eCfg := &BalancerJsonCfg{Enabled: utils.BoolPointer(false)}
	if cfg, err := dfCgrJsonCfg.BalancerJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestDfRaterJsonCfg(t *testing.T) {
	eCfg := &RaterJsonCfg{Enabled: utils.BoolPointer(false), Balancer: utils.StringPointer(""), Cdrstats: utils.StringPointer(""),
		Historys: utils.StringPointer(""), Pubsubs: utils.StringPointer(""), Users: utils.StringPointer(""), Aliases: utils.StringPointer("")}
	if cfg, err := dfCgrJsonCfg.RaterJsonCfg(); err != nil {
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
		Extra_fields:    utils.StringSlicePointer([]string{}),
		Store_cdrs:      utils.BoolPointer(true),
		Rater:           utils.StringPointer("internal"),
		Pubsubs:         utils.StringPointer(""),
		Users:           utils.StringPointer(""),
		Aliases:         utils.StringPointer(""),
		Cdrstats:        utils.StringPointer(""),
		Cdr_replication: &[]*CdrReplicationJsonCfg{},
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
		&CdrFieldJsonCfg{Tag: utils.StringPointer("CgrId"),
			Cdr_field_id: utils.StringPointer(utils.CGRID),
			Type:         utils.StringPointer("cdrfield"),
			Value:        utils.StringPointer(utils.CGRID)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("RunId"),
			Cdr_field_id: utils.StringPointer(utils.MEDI_RUNID),
			Type:         utils.StringPointer("cdrfield"),
			Value:        utils.StringPointer(utils.MEDI_RUNID)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Tor"),
			Cdr_field_id: utils.StringPointer(utils.TOR),
			Type:         utils.StringPointer("cdrfield"),
			Value:        utils.StringPointer(utils.TOR)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("AccId"),
			Cdr_field_id: utils.StringPointer(utils.ACCID),
			Type:         utils.StringPointer("cdrfield"),
			Value:        utils.StringPointer(utils.ACCID)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("ReqType"),
			Cdr_field_id: utils.StringPointer(utils.REQTYPE),
			Type:         utils.StringPointer("cdrfield"),
			Value:        utils.StringPointer(utils.REQTYPE)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Direction"),
			Cdr_field_id: utils.StringPointer(utils.DIRECTION),
			Type:         utils.StringPointer("cdrfield"),
			Value:        utils.StringPointer(utils.DIRECTION)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Tenant"),
			Cdr_field_id: utils.StringPointer(utils.TENANT),
			Type:         utils.StringPointer("cdrfield"),
			Value:        utils.StringPointer(utils.TENANT)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Category"),
			Cdr_field_id: utils.StringPointer(utils.CATEGORY),
			Type:         utils.StringPointer("cdrfield"),
			Value:        utils.StringPointer(utils.CATEGORY)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Account"),
			Cdr_field_id: utils.StringPointer(utils.ACCOUNT),
			Type:         utils.StringPointer("cdrfield"),
			Value:        utils.StringPointer(utils.ACCOUNT)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Subject"),
			Cdr_field_id: utils.StringPointer(utils.SUBJECT),
			Type:         utils.StringPointer("cdrfield"),
			Value:        utils.StringPointer(utils.SUBJECT)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Destination"),
			Cdr_field_id: utils.StringPointer(utils.DESTINATION),
			Type:         utils.StringPointer("cdrfield"),
			Value:        utils.StringPointer(utils.DESTINATION)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("SetupTime"),
			Cdr_field_id: utils.StringPointer(utils.SETUP_TIME),
			Type:         utils.StringPointer("cdrfield"),
			Value:        utils.StringPointer(utils.SETUP_TIME),
			Layout:       utils.StringPointer("2006-01-02T15:04:05Z07:00")},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("AnswerTime"),
			Cdr_field_id: utils.StringPointer(utils.ANSWER_TIME),
			Type:         utils.StringPointer("cdrfield"),
			Value:        utils.StringPointer(utils.ANSWER_TIME),
			Layout:       utils.StringPointer("2006-01-02T15:04:05Z07:00")},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Usage"),
			Cdr_field_id: utils.StringPointer(utils.USAGE),
			Type:         utils.StringPointer("cdrfield"),
			Value:        utils.StringPointer(utils.USAGE)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Cost"),
			Cdr_field_id: utils.StringPointer(utils.COST),
			Type:         utils.StringPointer("cdrfield"),
			Value:        utils.StringPointer(utils.COST)},
	}
	eCfg := map[string]*CdreJsonCfg{
		utils.META_DEFAULT: &CdreJsonCfg{
			Cdr_format:                    utils.StringPointer("csv"),
			Field_separator:               utils.StringPointer(","),
			Data_usage_multiply_factor:    utils.Float64Pointer(1.0),
			Sms_usage_multiply_factor:     utils.Float64Pointer(1.0),
			Generic_usage_multiply_factor: utils.Float64Pointer(1.0),
			Cost_multiply_factor:          utils.Float64Pointer(1.0),
			Cost_rounding_decimals:        utils.IntPointer(-1),
			Cost_shift_digits:             utils.IntPointer(0),
			Mask_destination_id:           utils.StringPointer("MASKED_DESTINATIONS"),
			Mask_length:                   utils.IntPointer(0),
			Export_dir:                    utils.StringPointer("/var/log/cgrates/cdre"),
			Header_fields:                 &eFields,
			Content_fields:                &eContentFlds,
			Trailer_fields:                &eFields,
		},
	}
	if cfg, err := dfCgrJsonCfg.CdreJsonCfgs(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestDfCdrcJsonCfg(t *testing.T) {
	eFields := []*CdrFieldJsonCfg{}
	cdrFields := []*CdrFieldJsonCfg{
		&CdrFieldJsonCfg{Tag: utils.StringPointer("tor"), Cdr_field_id: utils.StringPointer(utils.TOR), Type: utils.StringPointer(utils.CDRFIELD),
			Value: utils.StringPointer("2"), Mandatory: utils.BoolPointer(true)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("accid"), Cdr_field_id: utils.StringPointer(utils.ACCID), Type: utils.StringPointer(utils.CDRFIELD),
			Value: utils.StringPointer("3"), Mandatory: utils.BoolPointer(true)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("reqtype"), Cdr_field_id: utils.StringPointer(utils.REQTYPE), Type: utils.StringPointer(utils.CDRFIELD),
			Value: utils.StringPointer("4"), Mandatory: utils.BoolPointer(true)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("direction"), Cdr_field_id: utils.StringPointer(utils.DIRECTION), Type: utils.StringPointer(utils.CDRFIELD),
			Value: utils.StringPointer("5"), Mandatory: utils.BoolPointer(true)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("tenant"), Cdr_field_id: utils.StringPointer(utils.TENANT), Type: utils.StringPointer(utils.CDRFIELD),
			Value: utils.StringPointer("6"), Mandatory: utils.BoolPointer(true)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("category"), Cdr_field_id: utils.StringPointer(utils.CATEGORY), Type: utils.StringPointer(utils.CDRFIELD),
			Value: utils.StringPointer("7"), Mandatory: utils.BoolPointer(true)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("account"), Cdr_field_id: utils.StringPointer(utils.ACCOUNT), Type: utils.StringPointer(utils.CDRFIELD),
			Value: utils.StringPointer("8"), Mandatory: utils.BoolPointer(true)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("subject"), Cdr_field_id: utils.StringPointer(utils.SUBJECT), Type: utils.StringPointer(utils.CDRFIELD),
			Value: utils.StringPointer("9"), Mandatory: utils.BoolPointer(true)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("destination"), Cdr_field_id: utils.StringPointer(utils.DESTINATION), Type: utils.StringPointer(utils.CDRFIELD),
			Value: utils.StringPointer("10"), Mandatory: utils.BoolPointer(true)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("setup_time"), Cdr_field_id: utils.StringPointer(utils.SETUP_TIME), Type: utils.StringPointer(utils.CDRFIELD),
			Value: utils.StringPointer("11"), Mandatory: utils.BoolPointer(true)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("answer_time"), Cdr_field_id: utils.StringPointer(utils.ANSWER_TIME), Type: utils.StringPointer(utils.CDRFIELD),
			Value: utils.StringPointer("12"), Mandatory: utils.BoolPointer(true)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("usage"), Cdr_field_id: utils.StringPointer(utils.USAGE), Type: utils.StringPointer(utils.CDRFIELD),
			Value: utils.StringPointer("13"), Mandatory: utils.BoolPointer(true)},
	}
	eCfg := map[string]*CdrcJsonCfg{
		"*default": &CdrcJsonCfg{
			Enabled:                    utils.BoolPointer(false),
			Dry_run:                    utils.BoolPointer(false),
			Cdrs:                       utils.StringPointer("internal"),
			Cdr_format:                 utils.StringPointer("csv"),
			Field_separator:            utils.StringPointer(","),
			Timezone:                   utils.StringPointer(""),
			Run_delay:                  utils.IntPointer(0),
			Max_open_files:             utils.IntPointer(1024),
			Data_usage_multiply_factor: utils.Float64Pointer(1024.0),
			Cdr_in_dir:                 utils.StringPointer("/var/log/cgrates/cdrc/in"),
			Cdr_out_dir:                utils.StringPointer("/var/log/cgrates/cdrc/out"),
			Failed_calls_prefix:        utils.StringPointer("missed_calls"),
			Cdr_source_id:              utils.StringPointer("freeswitch_csv"),
			Cdr_filter:                 utils.StringPointer(""),
			Partial_record_cache:       utils.StringPointer("10s"),
			Header_fields:              &eFields,
			Content_fields:             &cdrFields,
			Trailer_fields:             &eFields,
		},
	}
	if cfg, err := dfCgrJsonCfg.CdrcJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg["*default"])
	}
}

func TestSmFsJsonCfg(t *testing.T) {
	eCfg := &SmFsJsonCfg{
		Enabled:                utils.BoolPointer(false),
		Rater:                  utils.StringPointer("internal"),
		Cdrs:                   utils.StringPointer("internal"),
		Create_cdr:             utils.BoolPointer(false),
		Extra_fields:           utils.StringSlicePointer([]string{}),
		Debit_interval:         utils.StringPointer("10s"),
		Min_call_duration:      utils.StringPointer("0s"),
		Max_call_duration:      utils.StringPointer("3h"),
		Min_dur_low_balance:    utils.StringPointer("5s"),
		Low_balance_ann_file:   utils.StringPointer(""),
		Empty_balance_context:  utils.StringPointer(""),
		Empty_balance_ann_file: utils.StringPointer(""),
		Subscribe_park:         utils.BoolPointer(true),
		Channel_sync_interval:  utils.StringPointer("5m"),
		Connections: &[]*FsConnJsonCfg{
			&FsConnJsonCfg{
				Server:     utils.StringPointer("127.0.0.1:8021"),
				Password:   utils.StringPointer("ClueCon"),
				Reconnects: utils.IntPointer(5),
			}},
	}
	if cfg, err := dfCgrJsonCfg.SmFsJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestSmKamJsonCfg(t *testing.T) {
	eCfg := &SmKamJsonCfg{
		Enabled:           utils.BoolPointer(false),
		Rater:             utils.StringPointer("internal"),
		Cdrs:              utils.StringPointer("internal"),
		Create_cdr:        utils.BoolPointer(false),
		Debit_interval:    utils.StringPointer("10s"),
		Min_call_duration: utils.StringPointer("0s"),
		Max_call_duration: utils.StringPointer("3h"),
		Connections: &[]*KamConnJsonCfg{
			&KamConnJsonCfg{
				Evapi_addr: utils.StringPointer("127.0.0.1:8448"),
				Reconnects: utils.IntPointer(5),
			},
		},
	}
	if cfg, err := dfCgrJsonCfg.SmKamJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestSmOsipsJsonCfg(t *testing.T) {
	eCfg := &SmOsipsJsonCfg{
		Enabled:                   utils.BoolPointer(false),
		Listen_udp:                utils.StringPointer("127.0.0.1:2020"),
		Rater:                     utils.StringPointer("internal"),
		Cdrs:                      utils.StringPointer("internal"),
		Create_cdr:                utils.BoolPointer(false),
		Debit_interval:            utils.StringPointer("10s"),
		Min_call_duration:         utils.StringPointer("0s"),
		Max_call_duration:         utils.StringPointer("3h"),
		Events_subscribe_interval: utils.StringPointer("60s"),
		Mi_addr:                   utils.StringPointer("127.0.0.1:8020"),
	}
	if cfg, err := dfCgrJsonCfg.SmOsipsJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestDfHistServJsonCfg(t *testing.T) {
	eCfg := &HistServJsonCfg{
		Enabled:       utils.BoolPointer(false),
		History_dir:   utils.StringPointer("/var/log/cgrates/history"),
		Save_interval: utils.StringPointer("1s"),
	}
	if cfg, err := dfCgrJsonCfg.HistServJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
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
		Indexes: utils.StringSlicePointer([]string{}),
	}
	if cfg, err := dfCgrJsonCfg.UserServJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestDfMailerJsonCfg(t *testing.T) {
	eCfg := &MailerJsonCfg{
		Server:       utils.StringPointer("localhost"),
		Auth_user:    utils.StringPointer("cgrates"),
		Auth_passwd:  utils.StringPointer("CGRateS.org"),
		From_address: utils.StringPointer("cgr-mailer@localhost.localdomain"),
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
		Timezone:                utils.StringPointer("Local"),
		Include_local_cost:      utils.BoolPointer(false),
		Return_file_code:        utils.StringPointer("0"),
		Response_group:          utils.StringPointer("03"),
		Response_type:           utils.StringPointer("D4"),
		Regulatory_code:         utils.StringPointer("03"),
		Client_tracking:         utils.StringPointer("CgrId"),
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
	eCfg := &GeneralJsonCfg{Default_reqtype: utils.StringPointer(utils.META_PSEUDOPREPAID)}
	if gCfg, err := cgrJsonCfg.GeneralJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, gCfg) {
		t.Error("Received: ", gCfg)
	}
	cdrFields := []*CdrFieldJsonCfg{
		&CdrFieldJsonCfg{Cdr_field_id: utils.StringPointer(utils.TOR), Value: utils.StringPointer("~7:s/^(voice|data|sms|generic)$/*$1/")},
		&CdrFieldJsonCfg{Cdr_field_id: utils.StringPointer(utils.ANSWER_TIME), Value: utils.StringPointer("1")},
		&CdrFieldJsonCfg{Cdr_field_id: utils.StringPointer(utils.USAGE), Value: utils.StringPointer(`~9:s/^(\d+)$/${1}s/`)},
	}
	eCfgCdrc := map[string]*CdrcJsonCfg{
		"CDRC-CSV1": &CdrcJsonCfg{
			Enabled:       utils.BoolPointer(true),
			Cdr_in_dir:    utils.StringPointer("/tmp/cgrates/cdrc1/in"),
			Cdr_out_dir:   utils.StringPointer("/tmp/cgrates/cdrc1/out"),
			Cdr_source_id: utils.StringPointer("csv1"),
		},
		"CDRC-CSV2": &CdrcJsonCfg{
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
		t.Error("Received: ", cfg["CDRC-CSV2"])
	}
	eCfgSmFs := &SmFsJsonCfg{
		Enabled: utils.BoolPointer(true),
		Connections: &[]*FsConnJsonCfg{
			&FsConnJsonCfg{
				Server:     utils.StringPointer("1.2.3.4:8021"),
				Password:   utils.StringPointer("ClueCon"),
				Reconnects: utils.IntPointer(5),
			},
			&FsConnJsonCfg{
				Server:     utils.StringPointer("2.3.4.5:8021"),
				Password:   utils.StringPointer("ClueCon"),
				Reconnects: utils.IntPointer(5),
			},
		},
	}
	if smFsCfg, err := cgrJsonCfg.SmFsJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfgSmFs, smFsCfg) {
		t.Error("Received: ", smFsCfg)
	}
}
