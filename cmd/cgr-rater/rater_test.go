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
	"code.google.com/p/goconf/conf"
	"net/rpc"
	"testing"
)

const (
	configText = `
[global]
datadb_type = test # 
datadb_host = test # The host to connect to. Values that start with / are for UNIX domain sockets.
datadb_port = test # The port to bind to.
datadb_name = test # The name of the database to connect to.
datadb_user =  test # The user to sign in as.
datadb_passwd =  test # The user's password.root
logdb_type = test # 
logdb_host = test # The host to connect to. Values that start with / are for UNIX domain sockets.
logdb_port = test # The port to bind to.
logdb_name = test # The name of the database to connect to.
logdb_user =  test # The user to sign in as.
logdb_passwd =  test # The user's password.root

[balancer]
enabled = true # Start balancer server
listen = test # Balancer listen interface
rpc_encoding = test # use JSON for RPC encoding	

[rater]
enabled = true
listen = test # listening address host:port, internal for internal communication only
balancer = test # if defined it will register to balancer as worker
rpc_encoding = test # use JSON for RPC encoding

[mediator]
enabled = true
cdr_path = test # Freeswitch Master CSV CDR file.
rater = test #address where to access rater. Can be internal, direct rater address or the address of a balancer
rpc_encoding = test # use JSON for RPC encoding
skipdb = true

[scheduler]
enabled = true

[session_manager]
enabled = true
switch_type = test
rater = test #address where to access rater. Can be internal, direct rater address or the address of a balancer
debit_period = 11
rpc_encoding = test # use JSON for RPC encoding

[stats]
enabled = true
listen = test # Web server address (for stat reports)
media_path = test

[freeswitch]
server = test # freeswitch address host:port
pass = test # freeswitch address host:port
direction_index = test
tor_index         = test
tenant_index      = test
subject_index     = test
account_index     = test
destination_index = test
time_start_index  = test
time_end_index    = test
`
)

func TestConfig(t *testing.T) {
	c, err := conf.ReadConfigBytes([]byte(configText))
	if err != nil {
		t.Log("Could not parse configuration!")
		t.FailNow()
	}
	readConfig(c)
	if data_db_type != "test" ||
		data_db_host != "test" ||
		data_db_port != "test" ||
		data_db_name != "test" ||
		data_db_user != "test" ||
		data_db_pass != "test" ||
		log_db_type != "test" ||
		log_db_host != "test" ||
		log_db_port != "test" ||
		log_db_name != "test" ||
		log_db_user != "test" ||
		log_db_pass != "test" ||

		rater_enabled != true ||
		rater_balancer != "test" ||
		rater_listen != "test" ||
		rater_rpc_encoding != "test" ||

		balancer_enabled != true ||
		balancer_listen != "test" ||
		balancer_rpc_encoding != "test" ||

		scheduler_enabled != true ||

		sm_enabled != true ||
		sm_switch_type != "test" ||
		sm_rater != "test" ||
		sm_debit_period != 11 ||
		sm_rpc_encoding != "test" ||

		mediator_enabled != true ||
		mediator_cdr_path != "test" ||
		mediator_rater != "test" ||
		mediator_rpc_encoding != "test" ||
		mediator_skipdb != true ||

		stats_enabled != true ||
		stats_listen != "test" ||
		stats_media_path != "test" ||
		freeswitch_server != "test" ||
		freeswitch_pass != "test" ||
		freeswitch_direction != "test" ||
		freeswitch_tor != "test" ||
		freeswitch_tenant != "test" ||
		freeswitch_subject != "test" ||
		freeswitch_account != "test" ||
		freeswitch_destination != "test" ||
		freeswitch_time_start != "test" ||
		freeswitch_time_end != "test" {
		t.Log(data_db_type)
		t.Log(data_db_host)
		t.Log(data_db_port)
		t.Log(data_db_name)
		t.Log(data_db_user)
		t.Log(data_db_pass)
		t.Log(log_db_type)
		t.Log(log_db_host)
		t.Log(log_db_port)
		t.Log(log_db_name)
		t.Log(log_db_user)
		t.Log(log_db_pass)
		t.Log(rater_enabled)
		t.Log(rater_balancer)
		t.Log(rater_listen)
		t.Log(rater_rpc_encoding)
		t.Log(balancer_enabled)
		t.Log(balancer_listen)
		t.Log(balancer_rpc_encoding)
		t.Log(scheduler_enabled)
		t.Log(sm_enabled)
		t.Log(sm_switch_type)
		t.Log(sm_rater)
		t.Log(sm_debit_period)
		t.Log(sm_rpc_encoding)
		t.Log(mediator_enabled)
		t.Log(mediator_cdr_path)
		t.Log(mediator_rater)
		t.Log(stats_enabled)
		t.Log(stats_listen)
		t.Log(stats_media_path)
		t.Log(freeswitch_server)
		t.Log(freeswitch_pass)
		t.Log(freeswitch_direction)
		t.Log(freeswitch_tor)
		t.Log(freeswitch_tenant)
		t.Log(freeswitch_subject)
		t.Log(freeswitch_account)
		t.Log(freeswitch_destination)
		t.Log(freeswitch_time_start)
		t.Log(freeswitch_time_end)
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

func TestVarReset(t *testing.T) {
	c, err := conf.ReadConfigBytes([]byte(configText))
	if err != nil {
		t.Log("Could not parse configuration!")
		t.FailNow()
	}
	myString := "default"
	myString, err = c.GetString("default", "non_existing")
	if err == nil {
		t.Error("Reding non exitsing variable did not issue error!")
	}
	if myString != "" {
		t.Error("Variable has not been reseted")
	}
	myBool := true
	myBool, err = c.GetBool("default", "non_existing")
	if err == nil {
		t.Error("Reding non exitsing variable did not issue error!")
	}
	if myBool {
		t.Error("Variable has not been reseted")
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
