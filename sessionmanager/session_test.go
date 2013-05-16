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

package sessionmanager

import (
	"github.com/cgrates/cgrates/config"
	"testing"
	// "time"
)

var (
	newEventBody = `
"Event-Name":	"HEARTBEAT",
"Core-UUID":	"d5abc5b0-95c6-11e1-be05-43c90197c914",
"FreeSWITCH-Hostname":	"grace",
"FreeSWITCH-Switchname":	"grace",
"FreeSWITCH-IPv4":	"172.17.77.126",
"variable_sip_full_from":	"rif",
"variable_cgr_account":	"rif",
"variable_sip_full_to":	"0723045326",
"Caller-Dialplan":	"vdf",
"FreeSWITCH-IPv6":	"::1",
"Event-Date-Local":	"2012-05-04 14:38:23",
"Event-Date-GMT":	"Fri, 03 May 2012 11:38:23 GMT",
"Event-Date-Timestamp":	"1336131503218867",
"Event-Calling-File":	"switch_core.c",
"Event-Calling-Function":	"send_heartbeat",
"Event-Calling-Line-Number":	"68",
"Event-Sequence":	"4171",
"Event-Info":	"System Ready",
"Up-Time":	"0 years, 0 days, 2 hours, 43 minutes, 21 seconds, 349 milliseconds, 683 microseconds",
"Session-Count":	"0",
"Max-Sessions":	"1000",
"Session-Per-Sec":	"30",
"Session-Since-Startup":	"122",
"Idle-CPU":	"100.000000"
`
	conf_data = []byte(`
### Test data, not for production usage

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
cdr_in_dir = test # Freeswitch Master CSV CDR path.
cdr_out_dir = test
rater = test #address where to access rater. Can be internal, direct rater address or the address of a balancer
rpc_encoding = test # use JSON for RPC encoding
skipdb = true
pseudoprepaid = true

[scheduler]
enabled = true

[session_manager]
enabled = true
switch_type = test
rater = test #address where to access rater. Can be internal, direct rater address or the address of a balancer
debit_interval = 11
rpc_encoding = test # use JSON for RPC encoding

[freeswitch]
server = test # freeswitch address host:port
passwd = test # freeswitch address host:port
direction_index = test
tor_index         = test
tenant_index      = test
subject_index     = test
account_index     = test
destination_index = test
time_start_index  = test
duration_index    = test
uuid_index        = test
`)
)

/*func TestSessionDurationSingle(t *testing.T) {
	newEvent := new(FSEvent).New(newEventBody)
	sm := &FSSessionManager{}
	s := NewSession(newEvent, sm)
	defer s.Close()
	twoSeconds, _ := time.ParseDuration("2s")
	if d := s.getSessionDurationFrom(s.callDescriptor.TimeStart.Add(twoSeconds)); d.Seconds() < 2 || d.Seconds() > 3 {
		t.Errorf("Wrong session duration %v", d)
	}
}*/

func TestSessionNilSession(t *testing.T) {
	var errCfg error
	cfg, errCfg = config.NewCGRConfigBytes(conf_data) // Needed here to avoid nil on cfg variable
	if errCfg != nil {
		t.Errorf("Cannot get configuration %v", errCfg)
	}
	newEvent := new(FSEvent).New("")
	sm := &FSSessionManager{}
	s := NewSession(newEvent, sm)
	if s != nil {
		t.Error("no account and it still created session.")
	}
}
