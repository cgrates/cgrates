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

package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"path"
	"strconv"
	"time"

	"github.com/cgrates/rpcclient"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")

func main() {
	flag.Parse()
	var err error
	var rpc *rpcclient.RPCClient
	var cfgPath string
	var cfg *config.CGRConfig
	cfgPath = path.Join(*dataDir, "conf", "samples", "cdrsv1mysql")
	if cfg, err = config.NewCGRConfigFromPath(cfgPath); err != nil {
		log.Fatal("Got config error: ", err.Error())
	}
	rpc, err = rpcclient.NewRPCClient(utils.TCP, cfg.ListenCfg().RPCJSONListen, false, "", "", "", 1, 1,
		time.Second, 2*time.Second, rpcclient.JSONrpc, nil, false)
	if err != nil {
		log.Fatal("Could not connect to rater: ", err.Error())
	}
	var sumApier float64
	var sumRateS float64
	var tApier time.Duration
	var tRateS time.Duration
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	//
	var sls []string
	for i := 100; i < 200; i++ {
		sls = append(sls, strconv.Itoa(i))
	}

	for i := 0; i < 1000; i++ {
		destination := fmt.Sprintf("%+v%+v", sls[r1.Intn(100)], 1000000+rand.Intn(9999999-1000000))
		usage := fmt.Sprintf("%+vm", r1.Intn(250))
		attrs := v1.AttrGetCost{
			Category:    "call",
			Tenant:      "cgrates.org",
			Subject:     "*any",
			AnswerTime:  utils.MetaNow,
			Destination: destination,
			Usage:       usage,
		}
		var replyApier *engine.EventCost
		tApiInit := time.Now()
		if err := rpc.Call(utils.APIerSv1GetCost, &attrs, &replyApier); err != nil {
			fmt.Println(err)
			return
		}
		tApier += time.Now().Sub(tApiInit)
		sumApier += *replyApier.Cost

		argsRateS := &utils.ArgsCostForEvent{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]interface{}{
					utils.Category:    "call",
					utils.Tenant:      "cgrates.org",
					utils.Subject:     "*any",
					utils.AnswerTime:  utils.MetaNow,
					utils.Destination: destination,
				},
				Opts: map[string]interface{}{
					utils.OptsRatesUsage: usage,
				},
			},
		}

		var rplyRateS *engine.RateProfileCost
		tRateSInit := time.Now()
		if err := rpc.Call(utils.RateSv1CostForEvent, argsRateS, &rplyRateS); err != nil {
			fmt.Printf("Unexpected nil error received for RateSv1CostForEvent: %+v\n", err.Error())
			return
		}
		tRateS += time.Now().Sub(tRateSInit)
		sumRateS += rplyRateS.Cost
	}
	fmt.Println("Cost for apier get cost : ")
	fmt.Println(sumApier)
	fmt.Println("Cost for RateS")
	fmt.Println(sumRateS)
	fmt.Println("Time for apier get cost : ")
	fmt.Println(tApier)
	fmt.Println("Time for RateS")
	fmt.Println(tRateS)

}
