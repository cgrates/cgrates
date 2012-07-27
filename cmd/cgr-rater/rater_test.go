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

		rater_enabled != true ||
		rater_balancer != "test" ||
		rater_listen != "test" ||
		rater_json != true ||

		balancer_enabled != true ||
		balancer_listen_rater != "test" ||
		balancer_listen != "test" ||
		balancer_json != true ||

		scheduler_enabled != true ||

		sm_enabled != true ||
		sm_rater != "test" ||
		sm_freeswitch_server != "test" ||
		sm_freeswitch_pass != "test" ||
		sm_json != true ||

		mediator_enabled != true ||
		mediator_cdr_file != "test" ||
		mediator_result_file != "test" ||
		mediator_rater != "test" ||
		mediator_host != "test" ||
		mediator_port != "test" ||
		mediator_db != "test" ||
		mediator_user != "test" ||
		mediator_password != "test" ||
		mediator_json != true ||
		mediator_skipdb != true ||
		stats_enabled != true ||
		stats_listen != "test" {
		t.Log(redis_server)
		t.Log(redis_db)
		t.Log(rater_enabled)
		t.Log(rater_balancer)
		t.Log(rater_listen)
		t.Log(rater_json)
		t.Log(balancer_enabled)
		t.Log(balancer_listen_rater)
		t.Log(balancer_listen)
		t.Log(balancer_json)
		t.Log(scheduler_enabled)
		t.Log(sm_enabled)
		t.Log(sm_rater)
		t.Log(sm_freeswitch_server)
		t.Log(sm_freeswitch_pass)
		t.Log(sm_json)
		t.Log(mediator_enabled)
		t.Log(mediator_cdr_file)
		t.Log(mediator_result_file)
		t.Log(mediator_rater)
		t.Log(mediator_host)
		t.Log(mediator_port)
		t.Log(mediator_db)
		t.Log(mediator_user)
		t.Log(mediator_password)
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
