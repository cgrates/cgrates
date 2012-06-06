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
	"log"
	"net/rpc/jsonrpc"
	"os"
	"time"
)

var (
	server  = flag.String("server", "127.0.0.1:2001", "server address host:port")
	tor     = flag.String("tor", "0", "Type of record")
	cstmid  = flag.String("cstmid", "vdf", "Customer identificator")
	subject = flag.String("subject", "rif", "The client who made the call")
	dest    = flag.String("dest", "0256", "Destination prefix")
	ts      = flag.String("ts", "2012-02-09T00:00:00Z", "Time start")
	te      = flag.String("te", "2012-02-09T00:10:00Z", "Time end")
	amount  = flag.Float64("amount", 100, "Amount for different operations")
)

func main() {
	flag.Parse()
	client, err := jsonrpc.Dial("tcp", *server)
	if err != nil {
		log.Fatal("Could not connect to server " + *server)
	}
	defer client.Close()

	timestart, err := time.Parse(time.RFC3339, *ts)
	if err != nil {
		log.Fatal("Time start format is invalid: ", err)
	}
	timeend, err := time.Parse(time.RFC3339, *te)
	if err != nil {
		log.Fatal("Time end format is invalid: ", err)
	}

	cd := &timespans.CallDescriptor{TOR: *tor,
		CstmId:            *cstmid,
		Subject:           *subject,
		DestinationPrefix: *dest,
		TimeStart:         timestart,
		TimeEnd:           timeend,
		Amount:            *amount,
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
	case "addvolumediscountseconds":
		var result float64
		if err = client.Call("Responder.AddVolumeDiscountSeconds", cd, &result); err == nil {
			fmt.Println(result)
		}
	case "resetvolumediscountseconds":
		var result float64
		if err = client.Call("Responder.ResetVolumeDiscountSeconds", cd, &result); err == nil {
			fmt.Println(result)
		}
	case "addrecievedcallseconds":
		var result float64
		if err = client.Call("Responder.AddRecievedCallSeconds", cd, &result); err == nil {
			fmt.Println(result)
		}
	case "resetuserbudget":
		var result float64
		if err = client.Call("Responder.ResetUserBudget", cd, &result); err == nil {
			fmt.Println(result)
		}
	case "status":
		var result string
		if err = client.Call("Responder.Status", cd, &result); err == nil {
			fmt.Println(result)
		}

	default:
		fmt.Println("List of commands:")
		fmt.Println("\tgetcost")
		fmt.Println("\tgetmaxsessiontime")
		fmt.Println("\tdebitbalance")
		fmt.Println("\tdebitsms")
		fmt.Println("\tdebitseconds")
		fmt.Println("\taddvolumediscountseconds")
		fmt.Println("\tresetvolumediscountseconds")
		fmt.Println("\taddrecievedcallseconds")
		fmt.Println("\tresetuserbudget")
		fmt.Println("\tstatus")
	}
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
