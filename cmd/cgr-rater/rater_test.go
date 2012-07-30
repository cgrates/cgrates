/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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
	"testing"
)

func TestConfig(t *testing.T) {
	readConfig("/home/rif/Documents/prog/go/src/github.com/cgrates/cgrates/data/test.config")
	if redis_server != "test" ||
		redis_db != 1 ||
		redis_pass != "test" ||
		logging_db_type != "test" ||
		logging_db_host != "test" ||
		logging_db_port != "test" ||
		logging_db_db != "test" ||
		logging_db_user != "test" ||
		logging_db_password != "test" ||

		rater_enabled != true ||
		rater_balancer != "test" ||
		rater_listen != "test" ||
		rater_rpc_encoding != "test" ||

		balancer_enabled != true ||
		balancer_listen != "test" ||
		balancer_rpc_encoding != "test" ||

		scheduler_enabled != true ||

		sm_enabled != true ||
		sm_rater != "test" ||
		sm_freeswitch_server != "test" ||
		sm_freeswitch_pass != "test" ||
		sm_rpc_encoding != "test" ||

		mediator_enabled != true ||
		mediator_cdr_file != "test" ||
		mediator_result_file != "test" ||
		mediator_rater != "test" ||
		mediator_rpc_encoding != "test" ||
		mediator_skipdb != true ||
		stats_enabled != true ||
		stats_listen != "test" {
		t.Log(redis_server)
		t.Log(redis_db)
		t.Log(redis_pass)
		t.Log(logging_db_type)
		t.Log(logging_db_host)
		t.Log(logging_db_port)
		t.Log(logging_db_db)
		t.Log(logging_db_user)
		t.Log(logging_db_password)
		t.Log(rater_enabled)
		t.Log(rater_balancer)
		t.Log(rater_listen)
		t.Log(rater_rpc_encoding)
		t.Log(balancer_enabled)
		t.Log(balancer_listen)
		t.Log(balancer_rpc_encoding)
		t.Log(scheduler_enabled)
		t.Log(sm_enabled)
		t.Log(sm_rater)
		t.Log(sm_freeswitch_server)
		t.Log(sm_freeswitch_pass)
		t.Log(sm_rpc_encoding)
		t.Log(mediator_enabled)
		t.Log(mediator_cdr_file)
		t.Log(mediator_result_file)
		t.Log(mediator_rater)
		t.Log(stats_enabled)
		t.Log(stats_listen)

		t.Error("Config file read failed!")
	}
}

func TestRPCGet(t *testing.T) {
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
