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
	"github.com/cgrates/cgrates/utils"
	"reflect"
	"testing"
)

var cgrJsonCfg CgrJsonCfg

func TestNewCgrJsonCfgFromFile(t *testing.T) {
	var err error
	if cgrJsonCfg, err = NewCgrJsonCfgFromFile("cgrates_cfg_defaults.json"); err != nil {
		t.Error(err.Error())
	}
}

func TestGeneralJsonCfg(t *testing.T) {
	eCfg := &GeneralJsonCfg{
		Http_skip_tls_veify: utils.BoolPointer(false),
		Rounding_decimals:   utils.IntPointer(10),
		Dbdata_encoding:     utils.StringPointer("msgpack"),
		Tpexport_dir:        utils.StringPointer("/var/log/cgrates/tpe"),
		Default_reqtype:     utils.StringPointer("rated"),
		Default_category:    utils.StringPointer("call"),
		Default_tenant:      utils.StringPointer("cgrates.org"),
		Default_subject:     utils.StringPointer("cgrates")}
	if gCfg, err := cgrJsonCfg.GeneralJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, gCfg) {
		t.Error("Received: ", gCfg)
	}
}

func TestListenJsonCfg(t *testing.T) {
	eCfg := &ListenJsonCfg{
		Rpc_json: utils.StringPointer("127.0.0.1:2012"),
		Rpc_gob:  utils.StringPointer("127.0.0.1:2013"),
		Http:     utils.StringPointer("127.0.0.1:2080")}
	if cfg, err := cgrJsonCfg.ListenJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestDbJsonCfg(t *testing.T) {
	eCfg := &DbJsonCfg{
		Db_type:   utils.StringPointer("redis"),
		Db_host:   utils.StringPointer("127.0.0.1"),
		Db_port:   utils.IntPointer(6379),
		Db_name:   utils.StringPointer("10"),
		Db_user:   utils.StringPointer(""),
		Db_passwd: utils.StringPointer(""),
	}
	if cfg, err := cgrJsonCfg.DbJsonCfg(RATINGDB_JSN); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
	eCfg = &DbJsonCfg{
		Db_type:   utils.StringPointer("redis"),
		Db_host:   utils.StringPointer("127.0.0.1"),
		Db_port:   utils.IntPointer(6379),
		Db_name:   utils.StringPointer("11"),
		Db_user:   utils.StringPointer(""),
		Db_passwd: utils.StringPointer(""),
	}
	if cfg, err := cgrJsonCfg.DbJsonCfg(ACCOUNTINGDB_JSN); err != nil {
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
		Max_open_conns: utils.IntPointer(0),
		Max_idle_conns: utils.IntPointer(-1),
	}
	if cfg, err := cgrJsonCfg.DbJsonCfg(STORDB_JSN); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestBalancerJsonCfg(t *testing.T) {
	eCfg := &BalancerJsonCfg{Enabled: utils.BoolPointer(false)}
	if cfg, err := cgrJsonCfg.BalancerJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestRaterJsonCfg(t *testing.T) {
	eCfg := &RaterJsonCfg{Enabled: utils.BoolPointer(false), Balancer: utils.StringPointer("")}
	if cfg, err := cgrJsonCfg.RaterJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestSchedulerJsonCfg(t *testing.T) {
	eCfg := &SchedulerJsonCfg{Enabled: utils.BoolPointer(false)}
	if cfg, err := cgrJsonCfg.SchedulerJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestCdrsJsonCfg(t *testing.T) {
	eCfg := &CdrsJsonCfg{
		Enabled:       utils.BoolPointer(false),
		Extra_fields:  utils.StringSlicePointer([]string{}),
		Mediator:      utils.StringPointer(""),
		Cdrstats:      utils.StringPointer(""),
		Store_disable: utils.BoolPointer(false),
	}
	if cfg, err := cgrJsonCfg.CdrsJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestMediatorJsonCfg(t *testing.T) {
	eCfg := &MediatorJsonCfg{
		Enabled:       utils.BoolPointer(false),
		Reconnects:    utils.IntPointer(3),
		Rater:         utils.StringPointer("internal"),
		Cdrstats:      utils.StringPointer(""),
		Store_disable: utils.BoolPointer(false),
	}
	if cfg, err := cgrJsonCfg.MediatorJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestCdrStatsJsonCfg(t *testing.T) {
	eCfg := &CdrStatsJsonCfg{
		Enabled:              utils.BoolPointer(false),
		Queue_length:         utils.IntPointer(50),
		Time_window:          utils.StringPointer("1h"),
		Metrics:              utils.StringSlicePointer([]string{"ASR", "ACD", "ACC"}),
		Setup_interval:       utils.StringSlicePointer([]string{}),
		Tors:                 utils.StringSlicePointer([]string{}),
		Cdr_hosts:            utils.StringSlicePointer([]string{}),
		Cdr_sources:          utils.StringSlicePointer([]string{}),
		Req_types:            utils.StringSlicePointer([]string{}),
		Directions:           utils.StringSlicePointer([]string{}),
		Tenants:              utils.StringSlicePointer([]string{}),
		Categories:           utils.StringSlicePointer([]string{}),
		Accounts:             utils.StringSlicePointer([]string{}),
		Subjects:             utils.StringSlicePointer([]string{}),
		Destination_prefixes: utils.StringSlicePointer([]string{}),
		Usage_interval:       utils.StringSlicePointer([]string{}),
		Mediation_run_ids:    utils.StringSlicePointer([]string{}),
		Rated_accounts:       utils.StringSlicePointer([]string{}),
		Rated_subjects:       utils.StringSlicePointer([]string{}),
		Cost_intervals:       utils.StringSlicePointer([]string{}),
	}
	if cfg, err := cgrJsonCfg.CdrStatsJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestCdreJsonCfgs(t *testing.T) {
	eFields := []*CdrFieldJsonCfg{}
	eContentFlds := []*CdrFieldJsonCfg{
		&CdrFieldJsonCfg{Tag: utils.StringPointer("CgrId"),
			Type:      utils.StringPointer("cdrfield"),
			Value:     utils.StringPointer("cgrid"),
			Width:     utils.IntPointer(40),
			Mandatory: utils.BoolPointer(true)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("RunId"),
			Type:  utils.StringPointer("cdrfield"),
			Value: utils.StringPointer("mediation_runid"),
			Width: utils.IntPointer(20)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Tor"),
			Type:  utils.StringPointer("cdrfield"),
			Value: utils.StringPointer("tor"),
			Width: utils.IntPointer(6)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("AccId"),
			Type:  utils.StringPointer("cdrfield"),
			Value: utils.StringPointer("accid"),
			Width: utils.IntPointer(36)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("ReqType"),
			Type:  utils.StringPointer("cdrfield"),
			Value: utils.StringPointer("reqtype"),
			Width: utils.IntPointer(13)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Direction"),
			Type:  utils.StringPointer("cdrfield"),
			Value: utils.StringPointer("direction"),
			Width: utils.IntPointer(4)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Tenant"),
			Type:  utils.StringPointer("cdrfield"),
			Value: utils.StringPointer("tenant"),
			Width: utils.IntPointer(24)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Category"),
			Type:  utils.StringPointer("cdrfield"),
			Value: utils.StringPointer("category"),
			Width: utils.IntPointer(10)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Account"),
			Type:  utils.StringPointer("cdrfield"),
			Value: utils.StringPointer("account"),
			Width: utils.IntPointer(24)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Subject"),
			Type:  utils.StringPointer("cdrfield"),
			Value: utils.StringPointer("subject"),
			Width: utils.IntPointer(24)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Destination"),
			Type:  utils.StringPointer("cdrfield"),
			Value: utils.StringPointer("destination"),
			Width: utils.IntPointer(24)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("SetupTime"),
			Type:   utils.StringPointer("cdrfield"),
			Value:  utils.StringPointer("setup_time"),
			Width:  utils.IntPointer(30),
			Layout: utils.StringPointer("2006-01-02T15:04:05Z07:00")},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("AnswerTime"),
			Type:   utils.StringPointer("cdrfield"),
			Value:  utils.StringPointer("answer_time"),
			Width:  utils.IntPointer(30),
			Layout: utils.StringPointer("2006-01-02T15:04:05Z07:00")},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Usage"),
			Type:  utils.StringPointer("cdrfield"),
			Value: utils.StringPointer("usage"),
			Width: utils.IntPointer(30)},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("Cost"),
			Type:  utils.StringPointer("cdrfield"),
			Value: utils.StringPointer("cost"),
			Width: utils.IntPointer(24)},
	}
	eCfg := map[string]*CdreJsonCfg{
		"CDRE-FW1": &CdreJsonCfg{
			Cdr_format:                 utils.StringPointer("csv"),
			Data_usage_multiply_factor: utils.Float64Pointer(1.0),
			Cost_multiply_factor:       utils.Float64Pointer(1.0),
			Cost_rounding_decimals:     utils.IntPointer(-1),
			Cost_shift_digits:          utils.IntPointer(0),
			Mask_destination_id:        utils.StringPointer("MASKED_DESTINATIONS"),
			Mask_length:                utils.IntPointer(0),
			Export_dir:                 utils.StringPointer("/var/log/cgrates/cdre"),
			Header_fields:              &eFields,
			Content_fields:             &eContentFlds,
			Trailer_fields:             &eFields,
		},
	}
	if cfg, err := cgrJsonCfg.CdreJsonCfgs(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestCdrcJsonCfg(t *testing.T) {
	cdrFields := []*CdrFieldJsonCfg{
		&CdrFieldJsonCfg{Tag: utils.StringPointer("accid"), Value: utils.StringPointer("0;13")},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("reqtype"), Value: utils.StringPointer("1")},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("direction"), Value: utils.StringPointer("2")},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("tenant"), Value: utils.StringPointer("3")},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("category"), Value: utils.StringPointer("4")},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("account"), Value: utils.StringPointer("5")},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("subject"), Value: utils.StringPointer("6")},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("destination"), Value: utils.StringPointer("7")},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("setup_time"), Value: utils.StringPointer("8")},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("answer_time"), Value: utils.StringPointer("9")},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("usage"), Value: utils.StringPointer("10")},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("extr1"), Value: utils.StringPointer("11")},
		&CdrFieldJsonCfg{Tag: utils.StringPointer("extr2"), Value: utils.StringPointer("12")},
	}
	eCfg := map[string]*CdrcJsonCfg{
		"instance1": &CdrcJsonCfg{
			Enabled:                    utils.BoolPointer(false),
			Cdrs_address:               utils.StringPointer("internal"),
			Cdr_format:                 utils.StringPointer("csv"),
			Field_separator:            utils.StringPointer(","),
			Run_delay:                  utils.IntPointer(0),
			Data_usage_multiply_factor: utils.Float64Pointer(1024.0),
			Cdr_in_dir:                 utils.StringPointer("/var/log/cgrates/cdrc/in"),
			Cdr_out_dir:                utils.StringPointer("/var/log/cgrates/cdrc/out"),
			Cdr_source_id:              utils.StringPointer("freeswitch_csv"),
			Cdr_fields:                 &cdrFields,
		},
	}
	if cfg, err := cgrJsonCfg.CdrcJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestSessionManagerJsonCfg(t *testing.T) {
	eCfg := &SessionManagerJsonCfg{
		Enabled:           utils.BoolPointer(false),
		Switch_type:       utils.StringPointer("freeswitch"),
		Rater:             utils.StringPointer("internal"),
		Cdrs:              utils.StringPointer(""),
		Reconnects:        utils.IntPointer(3),
		Debit_interval:    utils.IntPointer(10),
		Min_call_duration: utils.StringPointer("0s"),
		Max_call_duration: utils.StringPointer("3h"),
	}
	if cfg, err := cgrJsonCfg.SessionManagerJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestFSJsonCfg(t *testing.T) {
	eCfg := &FSJsonCfg{
		Server:                 utils.StringPointer("127.0.0.1:8021"),
		Password:               utils.StringPointer("ClueCon"),
		Reconnects:             utils.IntPointer(5),
		Min_dur_low_balance:    utils.StringPointer("5s"),
		Low_balance_ann_file:   utils.StringPointer(""),
		Empty_balance_context:  utils.StringPointer(""),
		Empty_balance_ann_file: utils.StringPointer(""),
		Cdr_extra_fields:       utils.StringSlicePointer([]string{}),
	}
	if cfg, err := cgrJsonCfg.FSJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestKamailioJsonCfg(t *testing.T) {
	eCfg := &KamailioJsonCfg{
		Evapi_addr: utils.StringPointer("127.0.0.1:8448"),
		Reconnects: utils.IntPointer(3),
	}
	if cfg, err := cgrJsonCfg.KamailioJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestOsipsJsonCfg(t *testing.T) {
	eCfg := &OsipsJsonCfg{
		Listen_udp:                utils.StringPointer("127.0.0.1:2020"),
		Mi_addr:                   utils.StringPointer("127.0.0.1:8020"),
		Events_subscribe_interval: utils.StringPointer("60s"),
		Reconnects:                utils.IntPointer(3),
	}
	if cfg, err := cgrJsonCfg.OsipsJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestHistServJsonCfg(t *testing.T) {
	eCfg := &HistServJsonCfg{
		Enabled:       utils.BoolPointer(false),
		History_dir:   utils.StringPointer("/var/log/cgrates/history"),
		Save_interval: utils.StringPointer("1s"),
	}
	if cfg, err := cgrJsonCfg.HistServJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestHistAgentJsonCfg(t *testing.T) {
	eCfg := &HistAgentJsonCfg{
		Enabled: utils.BoolPointer(false),
		Server:  utils.StringPointer("internal"),
	}
	if cfg, err := cgrJsonCfg.HistAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}

func TestMailerJsonCfg(t *testing.T) {
	eCfg := &MailerJsonCfg{
		Server:       utils.StringPointer("localhost"),
		Auth_user:    utils.StringPointer("cgrates"),
		Auth_passwd:  utils.StringPointer("CGRateS.org"),
		From_address: utils.StringPointer("cgr-mailer@localhost.localdomain"),
	}
	if cfg, err := cgrJsonCfg.MailerJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCfg, cfg) {
		t.Error("Received: ", cfg)
	}
}
