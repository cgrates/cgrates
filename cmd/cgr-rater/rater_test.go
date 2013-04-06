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

package main

import (
	"net/rpc"
	"fmt"
	"testing"
)

func TestConfig(t *testing.T) {
	cfgPth := "test_data.txt"
	cfg, err = NewCGRConfig( &cfgPth )
	if err != nil {
		t.Log(fmt.Sprintf("Could not parse config: %s!", err))
		t.FailNow()
	}
	if cfg.data_db_type != "test" ||
		cfg.data_db_host != "test" ||
		cfg.data_db_port != "test" ||
		cfg.data_db_name != "test" ||
		cfg.data_db_user != "test" ||
		cfg.data_db_pass != "test" ||
		cfg.log_db_type != "test" ||
		cfg.log_db_host != "test" ||
		cfg.log_db_port != "test" ||
		cfg.log_db_name != "test" ||
		cfg.log_db_user != "test" ||
		cfg.log_db_pass != "test" ||
		cfg.rater_enabled != true ||
		cfg.rater_balancer != "test" ||
		cfg.rater_listen != "test" ||
		cfg.rater_rpc_encoding != "test" ||
		cfg.balancer_enabled != true ||
		cfg.balancer_listen != "test" ||
		cfg.balancer_rpc_encoding != "test" ||
		cfg.scheduler_enabled != true ||
		cfg.sm_enabled != true ||
		cfg.sm_switch_type != "test" ||
		cfg.sm_rater != "test" ||
		cfg.sm_debit_period != 11 ||
		cfg.sm_rpc_encoding != "test" ||
		cfg.mediator_enabled != true ||
		cfg.mediator_cdr_path != "test" ||
		cfg.mediator_cdr_out_path != "test" ||
		cfg.mediator_rater != "test" ||
		cfg.mediator_rpc_encoding != "test" ||
		cfg.mediator_skipdb != true ||
		cfg.mediator_pseudo_prepaid != true ||
		cfg.freeswitch_server != "test" ||
		cfg.freeswitch_pass != "test" ||
		cfg.freeswitch_direction != "test" ||
		cfg.freeswitch_tor != "test" ||
		cfg.freeswitch_tenant != "test" ||
		cfg.freeswitch_subject != "test" ||
		cfg.freeswitch_account != "test" ||
		cfg.freeswitch_destination != "test" ||
		cfg.freeswitch_time_start != "test" ||
		cfg.freeswitch_duration != "test" ||
		cfg.freeswitch_uuid != "test" {
		t.Log(cfg.data_db_type)
		t.Log(cfg.data_db_host)
		t.Log(cfg.data_db_port)
		t.Log(cfg.data_db_name)
		t.Log(cfg.data_db_user)
		t.Log(cfg.data_db_pass)
		t.Log(cfg.log_db_type)
		t.Log(cfg.log_db_host)
		t.Log(cfg.log_db_port)
		t.Log(cfg.log_db_name)
		t.Log(cfg.log_db_user)
		t.Log(cfg.log_db_pass)
		t.Log(cfg.rater_enabled)
		t.Log(cfg.rater_balancer)
		t.Log(cfg.rater_listen)
		t.Log(cfg.rater_rpc_encoding)
		t.Log(cfg.balancer_enabled)
		t.Log(cfg.balancer_listen)
		t.Log(cfg.balancer_rpc_encoding)
		t.Log(cfg.scheduler_enabled)
		t.Log(cfg.sm_enabled)
		t.Log(cfg.sm_switch_type)
		t.Log(cfg.sm_rater)
		t.Log(cfg.sm_debit_period)
		t.Log(cfg.sm_rpc_encoding)
		t.Log(cfg.mediator_enabled)
		t.Log(cfg.mediator_cdr_path)
		t.Log(cfg.mediator_cdr_out_path)
		t.Log(cfg.mediator_rater)
		t.Log(cfg.mediator_skipdb)
		t.Log(cfg.mediator_pseudo_prepaid)
		t.Log(cfg.freeswitch_server)
		t.Log(cfg.freeswitch_pass)
		t.Log(cfg.freeswitch_direction)
		t.Log(cfg.freeswitch_tor)
		t.Log(cfg.freeswitch_tenant)
		t.Log(cfg.freeswitch_subject)
		t.Log(cfg.freeswitch_account)
		t.Log(cfg.freeswitch_destination)
		t.Log(cfg.freeswitch_time_start)
		t.Log(cfg.freeswitch_duration)
		t.Log(cfg.freeswitch_uuid)
		t.Error("Config file read failed!")
	}
}

/*func TestRPCGet(t *testing.T) {
	client, err := rpc.DialHTTPPath("tcp", "localhost:2000", "/rpc")
	if err != nil {
		t.Error("Balancer server not started!")
		t.FailNow()
	}
	var reply string
	client.Call("Responder.Get", "test", &reply)
	const expect = "12223"
	if reply != expect {
		t.Errorf("replay == %v, want %v", reply, expect)
	}
}*/

func TestParamOverwrite(t *testing.T) {
	cfgPth := "test_data.txt"
	cfg, err = NewCGRConfig( &cfgPth )
	if err != nil {
		t.Log(fmt.Sprintf("Could not parse config: %s!", err))
		t.FailNow()
	}
	if cfg.freeswitch_reconnects !=  5 { // one default which is not overwritten in test data
		t.Errorf("freeswitch_reconnects set == %d, expect 5",cfg.freeswitch_reconnects)
	} else if cfg.scheduler_enabled != true { // one parameter which should be overwritten in test data
		t.Errorf("scheduler_enabled set == %d, expect true",cfg.freeswitch_reconnects)
	}
}

func BenchmarkRPCGet(b *testing.B) {
	b.StopTimer()
	client, _ := rpc.DialHTTPPath("tcp", "localhost:2000", "/rpc")
	b.StartTimer()
	var reply string
	for i := 0; i < b.N; i++ {
		client.Call("Responder.Get", "test", &reply)
	}
}
