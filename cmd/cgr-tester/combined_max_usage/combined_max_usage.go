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
	"sync"
	"time"

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
// 1) Start the engine with the following configuration < cgr-engine -config_path=/usr/share/cgrates/conf/samples/accounts_mysql >
// 2) Load the data with < cgr-loader -config_path=/usr/share/cgrates/conf/samples/accounts_mysql -verbose -path=/usr/share/cgrates/tariffplans/oldaccvsnew >
// 3) Run the program with  < go run combined_max_usage.go -requests=10000 -goroutines=5 >
// Additional Information
// In this scenario we compare the old account system vs the new one. The balance contains the RateID in order to access it directly without the needed of matching.
// For this scenario the account 1002 is used

func main() {
	flag.Parse()
	var err error
	var rpc *rpc.Client
	var cfgPath string
	var cfg *config.CGRConfig
	cfgPath = path.Join(*dataDir, "conf", "samples", "accounts_mysql")
	if cfg, err = config.NewCGRConfigFromPath(cfgPath); err != nil {
		log.Fatal("Got config error: ", err.Error())
	}
	if rpc, err = jsonrpc.Dial(utils.TCP, cfg.ListenCfg().RPCJSONListen); err != nil {
		return
	}

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	var wgAccountS sync.WaitGroup
	var accountSTime time.Duration
	var sumAccountS float64

	var wgRALs sync.WaitGroup
	var ralsTime time.Duration
	var sumRALs float64

	for i := 0; i < *requests; i++ {
		wgAccountS.Add(1)
		wgRALs.Add(1)
		usage := fmt.Sprintf("%+vm", 1+r1.Intn(59))
		go func() {
			var eEc *utils.ExtEventCharges
			arg := &utils.ArgsAccountsForEvent{CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]interface{}{
					utils.AccountField: "1002",
					utils.ToR:          utils.MetaVoice,
					utils.Usage:        usage,
				}}}
			tNow := time.Now()
			if err := rpc.Call(utils.AccountSv1MaxAbstracts,
				arg, &eEc); err != nil {
				return
			}
			accountSTime += time.Now().Sub(tNow)
			sumAccountS += *eEc.Usage
			wgAccountS.Done()
		}()

		go func() {
			tStart := time.Date(2016, 3, 31, 0, 0, 0, 0, time.UTC)
			usageDur, _ := utils.ParseDurationWithNanosecs(usage)
			cd := &engine.CallDescriptorWithOpts{
				CallDescriptor: &engine.CallDescriptor{
					Category:    "call",
					Tenant:      "cgrates.org",
					Subject:     "1002",
					Account:     "1002",
					Destination: "1003",
					TimeStart:   tStart,
					TimeEnd:     tStart.Add(usageDur),
				},
			}
			var rply time.Duration
			tNow := time.Now()
			if err := rpc.Call(utils.ResponderGetMaxSessionTime, cd, &rply); err != nil {
				return
			}
			ralsTime += time.Now().Sub(tNow)
			sumRALs += rply.Seconds()
			wgRALs.Done()
		}()

		if i%*gorutines == 0 {
			wgAccountS.Wait()
			wgRALs.Wait()
		}

	}
	wgAccountS.Wait()
	wgRALs.Wait()

	fmt.Println("Sum AccountS MaxUsage")
	fmt.Println(sumAccountS)
	fmt.Println("Average AccountS MaxUsage")
	fmt.Println(accountSTime / time.Duration(*requests))
	fmt.Println("Total AccountS MaxUsage Time")
	fmt.Println(accountSTime)

	fmt.Println("Sum RALs GetMaxSessionTime")
	fmt.Println(sumRALs)
	fmt.Println("Average RALs GetMaxSessionTime")
	fmt.Println(ralsTime / time.Duration(*requests))
	fmt.Println("Total RALs GetMaxSessionTime Time")
	fmt.Println(ralsTime)

}
