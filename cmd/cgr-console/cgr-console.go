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
	"flag"
	"fmt"
	"github.com/cgrates/cgrates/timespans"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"time"
)

var (
	server    = flag.String("server", "127.0.0.1:2001", "server address host:port")
	tor       = flag.String("tor", "0", "Type of record")
	direction = flag.String("direction", "OUT", "Call direction")
	tenant    = flag.String("tenant", "vdf", "Tenant identificator")
	subject   = flag.String("subject", "rif", "The client who made the call")
	account   = flag.String("account", "rif", "The the user balance to be used")
	dest      = flag.String("dest", "041", "Call destination")
	start     = flag.String("start", "2012-02-09T00:00:00Z", "Time start")
	end       = flag.String("end", "2012-02-09T00:10:00Z", "Time end")
	amount    = flag.Float64("amount", 100, "Amount for different operations")
	json      = flag.Bool("json", false, "Use JSON for RPC encoding.")
)

func main() {
	flag.Parse()
	var client *rpc.Client
	var err error
	if *json {
		client, err = jsonrpc.Dial("tcp", *server)
	} else {
		client, err = rpc.Dial("tcp", *server)
	}
	if err != nil {
		timespans.Logger.Crit(fmt.Sprintf("Could not connect to server " + *server))
		os.Exit(1)
	}
	defer client.Close()

	timestart, err := time.Parse(time.RFC3339, *start)
	if err != nil {
		timespans.Logger.Crit(fmt.Sprintf("Time start format is invalid: ", err))
		os.Exit(2)
	}
	timeend, err := time.Parse(time.RFC3339, *end)
	if err != nil {
		timespans.Logger.Crit(fmt.Sprintf("Time end format is invalid: ", err))
		os.Exit(3)
	}

	cd := &timespans.CallDescriptor{
		Direction:   *direction,
		TOR:         *tor,
		Tenant:      *tenant,
		Subject:     *subject,
		Account:     *account,
		Destination: *dest,
		TimeStart:   timestart,
		TimeEnd:     timeend,
		Amount:      *amount,
	}

	switch flag.Arg(0) {
	case "getcost":
		result := timespans.CallCost{}
		if err = client.Call("Responder.GetCost", cd, &result); err == nil {
			fmt.Println(result)
		}
	case "debit":
		result := timespans.CallCost{}
		if err = client.Call("Responder.Debit", cd, &result); err == nil {
			fmt.Println(result)
		}
	case "maxdebit":
		result := timespans.CallCost{}
		if err = client.Call("Responder.MaxDebit", cd, &result); err == nil {
			fmt.Println(result)
		}
	case "getmaxsessiontime":
		var result float64
		if err = client.Call("Responder.GetMaxSessionTime", cd, &result); err == nil {
			fmt.Println(result)
		}
	case "debitbalance":
		var result float64
		if err = client.Call("Responder.DebitBalance", cd, &result); err == nil {
			fmt.Println(result)
		}
	case "debitsms":
		var result float64
		if err = client.Call("Responder.DebitSMS", cd, &result); err == nil {
			fmt.Println(result)
		}
	case "debitseconds":
		var result float64
		if err = client.Call("Responder.DebitSeconds", cd, &result); err == nil {
			fmt.Println(result)
		}
	case "resetuserbudget":
		var result float64
		if err = client.Call("Responder.ResetUserBudget", cd, &result); err == nil {
			fmt.Println(result)
		}
	case "status":
		var result string
		if err = client.Call("Responder.Status", "", &result); err == nil {
			fmt.Println(result)
		}
	case "shutdown":
		var result string
		if err = client.Call("Responder.Shutdown", "", &result); err == nil {
			fmt.Println(result)
		}
	default:
		fmt.Println("List of commands:")
		fmt.Println("\tgetcost")
		fmt.Println("\tgetmaxsessiontime")
		fmt.Println("\tdebitbalance")
		fmt.Println("\tdebitsms")
		fmt.Println("\tdebitseconds")
		fmt.Println("\tresetuserbudget")
		fmt.Println("\tstatus")
		fmt.Println("\tshutdown")
		flag.PrintDefaults()
	}
	if err != nil {
		timespans.Logger.Crit(err.Error())
		os.Exit(1)
	}
}
