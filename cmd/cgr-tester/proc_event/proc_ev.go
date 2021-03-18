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
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"strconv"
	"sync"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	dataDir   = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
	requests  = flag.Int("requests", 10000, "Number of requests")
	gorutines = flag.Int("goroutines", 5, "Number of simultaneous goroutines")
)

// How to run:
// 1) Start the engine with the following configuration cgr-engine -config_path=/usr/share/cgrates/conf/samples/hundred_rates
// 2) Load the data with cgr-loader cgr-loader -config_path=/usr/share/cgrates/conf/samples/hundred_rates -verbose -path=/usr/share/cgrates/tariffplans/hundredrates
// 3) Run the program with go run proc_ev.go -requests=10000 -goroutines=5

func main() {
	flag.Parse()
	var err error
	var rpc *rpc.Client
	var cfgPath string
	var cfg *config.CGRConfig
	cfgPath = path.Join(*dataDir, "conf", "samples", "cdrsv1mysql")
	if cfg, err = config.NewCGRConfigFromPath(cfgPath); err != nil {
		log.Fatal("Got config error: ", err.Error())
	}
	if rpc, err = jsonrpc.Dial(utils.TCP, cfg.ListenCfg().RPCJSONListen); err != nil {
		return
	}
	var sumApier float64
	var sumRateS float64
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	//
	var sls []string
	for i := 100; i < 200; i++ {
		sls = append(sls, strconv.Itoa(i))
	}

	var wgApier sync.WaitGroup
	var wgRateS sync.WaitGroup
	var apierTime time.Duration
	var rateSTime time.Duration
	for i := 0; i < *requests; i++ {
		wgApier.Add(1)
		wgRateS.Add(1)
		destination := fmt.Sprintf("%+v%+v", sls[r1.Intn(100)], 1000000+rand.Intn(9999999-1000000))
		usage := fmt.Sprintf("%+vm", r1.Intn(250))
		go func() {
			attrs := v1.AttrGetCost{
				Category:    "call",
				Tenant:      "cgrates.org",
				Subject:     "*any",
				AnswerTime:  utils.MetaNow,
				Destination: destination,
				Usage:       usage,
			}
			var replyApier *engine.EventCost
			tNow := time.Now()
			if err := rpc.Call(utils.APIerSv1GetCost, &attrs, &replyApier); err != nil {
				fmt.Println(err)
				return
			}
			apierTime += time.Now().Sub(tNow)
			sumApier += *replyApier.Cost
			wgApier.Done()
		}()

		go func() {
			argsRateS := &utils.ArgsCostForEvent{
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     utils.UUIDSha1Prefix(),
					Event: map[string]interface{}{
						utils.Category:      "call",
						utils.Tenant:        "cgrates.org",
						utils.Subject:       "*any",
						utils.AnswerTime:    utils.MetaNow,
						utils.Destination:   destination,
						"PrefixDestination": destination[:3],
					},
					Opts: map[string]interface{}{
						utils.OptsRatesUsage: usage,
					},
				},
			}

			var rplyRateS *utils.RateProfileCost
			tNow := time.Now()
			if err := rpc.Call(utils.RateSv1CostForEvent, argsRateS, &rplyRateS); err != nil {
				fmt.Printf("Unexpected nil error received for RateSv1CostForEvent: %+v\n", err.Error())
				return
			}
			rateSTime += time.Now().Sub(tNow)
			sumRateS += rplyRateS.Cost
			wgRateS.Done()
		}()
		if i%*gorutines == 0 {
			wgApier.Wait()
			wgRateS.Wait()
		}

	}
	wgApier.Wait()
	wgRateS.Wait()
	fmt.Println("Cost for apier get cost : ")
	fmt.Println(sumApier)
	fmt.Println("Cost for RateS")
	fmt.Println(sumRateS)
	fmt.Println("Average ApierTime")
	fmt.Println(apierTime / time.Duration(*requests))
	fmt.Println("Average RateSTime")
	fmt.Println(rateSTime / time.Duration(*requests))
	fmt.Println("Total ApierTime")
	fmt.Println(apierTime)
	fmt.Println("Total RateSTime")
	fmt.Println(rateSTime)

}
