//go:build benchmark
// +build benchmark

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package general_tests

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	jsonrpc2 "github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func BenchmarkRPCCalls(b *testing.B) {
	benchCfgPath := path.Join(*utils.DataDir, "conf", "samples", "sessions_internal")
	benchCfg, err := config.NewCGRConfigFromPath(benchCfgPath)
	if err != nil {
		b.Fatal(err)
	}
	benchDelay := 1000
	if err := engine.InitDataDB(benchCfg); err != nil {
		b.Fatal(err)
	}
	if err := engine.InitStorDb(benchCfg); err != nil {
		b.Fatal(err)
	}
	if err := os.RemoveAll("/tmp/TestBenchRPC"); err != nil {
		b.Error(err)
	}
	if err := os.MkdirAll("/tmp/TestBenchRPC", 0755); err != nil {
		b.Error(err)
	}
	if _, err := engine.StopStartEngine(benchCfgPath, benchDelay); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() {
		engine.KillEngine(benchDelay)
		if err := os.RemoveAll("/tmp/TestBenchRPC"); err != nil {
			b.Error(err)
		}
	})
	benchRPC, err := jsonrpc2.Dial("tcp", "127.0.0.1:2012")
	if err != nil {
		b.Fatal("could not connect to engine: ", err.Error())
	}

	writeFile := func(fileName, data string) error {
		csvFile, err := os.Create(path.Join("/tmp/TestBenchRPC", fileName))
		if err != nil {
			return err
		}
		defer csvFile.Close()
		_, err = csvFile.WriteString(data)
		if err != nil {
			return err

		}
		return csvFile.Sync()
	}

	// Create and populate AccountActions.csv
	if err := writeFile(utils.AccountActionsCsv, `
#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,AP_PACKAGE_10,,,
cgrates.org,1002,AP_PACKAGE_10,,,
`); err != nil {
		b.Fatal(err)
	}

	// Create and populate ActionPlans.csv
	if err := writeFile(utils.ActionPlansCsv, `
#Id,ActionsId,TimingId,Weight
AP_PACKAGE_10,ACT_TOPUP_RST_10,*asap,10
`); err != nil {
		b.Fatal(err)
	}

	// Create and populate Actions.csv
	if err := writeFile(utils.ActionsCsv, `
#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP_RST_10,*topup_reset,,,test,*monetary,,*any,,,*unlimited,,10,10,false,false,10
`); err != nil {
		b.Fatal(err)
	}

	// Create and populate Attributes.csv
	if err := writeFile(utils.AttributesCsv, `
#Tenant,ID,Context,FilterIDs,ActivationInterval,AttributeFilterIDs,Path,Type,Value,Blocker,Weight
cgrates.org,ATTR_RPC,,,,,*req.Password,*constant,CGRateS.org,false,0
`); err != nil {
		b.Fatal(err)
	}

	// Create and populate Chargers.csv
	if err := writeFile(utils.ChargersCsv, `
#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,ATTR_RPC,0
`); err != nil {
		b.Fatal(err)
	}

	// Create and populate Destinations.csv
	if err := writeFile(utils.DestinationsCsv, `
#Id,Prefix
DST_1001,1001
DST_1002,1002
`); err != nil {
		b.Fatal(err)
	}

	// Create and populate DestinationRates.csv
	if err := writeFile(utils.DestinationRatesCsv, `
#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_1001_20CNT,DST_1001,RT_20CNT,*up,4,0,
DR_1002_20CNT,DST_1002,RT_20CNT,*up,4,0,
`); err != nil {
		b.Fatal(err)
	}

	// Create and populate RatingPlans.csv
	if err := writeFile(utils.RatingPlansCsv, `
#Id,DestinationRatesId,TimingTag,Weight
RP_1001,DR_1002_20CNT,*any,10
RP_1002,DR_1001_20CNT,*any,10
`); err != nil {
		b.Fatal(err)
	}

	// Create and populate RatingProfiles.csv
	if err := writeFile(utils.RatingProfilesCsv, `
#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_1001,
cgrates.org,call,1002,2014-01-14T00:00:00Z,RP_1002,
`); err != nil {
		b.Fatal(err)
	}

	// Create and populate Rates.csv
	if err := writeFile(utils.RatesCsv, `
#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_20CNT,0.4,0.2,60s,60s,0s
RT_20CNT,0,0.1,60s,1s,60s
`); err != nil {
		b.Fatal(err)
	}

	// Create and populate Resources.csv
	if err := writeFile(utils.ResourcesCsv, `
#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],TTL[4],Limit[5],AllocationMessage[6],Blocker[7],Stored[8],Weight[9],ThresholdIDs[10]
cgrates.org,RES_RPC,,,1h,1,,false,false,10,
`); err != nil {
		b.Fatal(err)
	}

	// Create and populate Routes.csv
	if err := writeFile(utils.RoutesCsv, `
#Tenant,ID,FilterIDs,ActivationInterval,Sorting,SortingParameters,RouteID,RouteFilterIDs,RouteAccountIDs,RouteRatingPlanIDs,RouteResourceIDs,RouteStatIDs,RouteWeight,RouteBlocker,RouteParameters,Weight
cgrates.org,ROUTE_RPC,,,*weight,,,,,,,,,,,10
cgrates.org,ROUTE_RPC,,,,,route1,,,,,,20,,,
cgrates.org,ROUTE_RPC,,,,,route2,,,,,,10,,,
`); err != nil {
		b.Fatal(err)
	}

	var replyLoad string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: "/tmp/TestBenchRPC"}
	if err := benchRPC.Call(
		context.Background(),
		utils.APIerSv1LoadTariffPlanFromFolder, attrs, &replyLoad); err != nil {
		b.Error(err)
	} else if replyLoad != utils.OK {
		b.Error("unexpected reply returned", replyLoad)
	}

	time.Sleep(100 * time.Millisecond)

	b.Run("CoreSv1Status", func(b *testing.B) {
		b.Skip()
		var reply map[string]any
		args := &utils.TenantWithAPIOpts{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := benchRPC.Call(
				context.Background(),
				utils.CoreSv1Status, args, &reply)
			if err != nil {
				b.Error(err)
			}
		}
	})

	b.Run("ChargerSv1ProcessEvent", func(b *testing.B) {
		b.Skip()
		processedEv := []*engine.ChrgSProcessEventReply{
			{
				ChargerSProfile:    "DEFAULT",
				AttributeSProfiles: []string{"cgrates.org:ATTR_RPC"},
				AlteredFields:      []string{utils.MetaReqRunID, "*req.Password"},
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event1",
					Event: map[string]any{
						utils.AccountField: "1001",
						"Password":         "CGRateS.org",
						"RunID":            utils.MetaDefault,
					},
					APIOpts: map[string]any{
						utils.MetaSubsys:               utils.MetaChargers,
						utils.OptsAttributesProfileIDs: []any{"ATTR_RPC"},
					},
				},
			},
		}
		var result []*engine.ChrgSProcessEventReply
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := benchRPC.Call(
				context.Background(),
				utils.ChargerSv1ProcessEvent, &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event1",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
				}, &result); err != nil {
				b.Error(err)
			}
			b.StopTimer()
			if !reflect.DeepEqual(result, processedEv) {
				b.Errorf("expected: %T,\nreceived: %T", utils.ToJSON(processedEv), utils.ToJSON(result))
			}
			b.StartTimer()
		}
	})

	b.Run("SessionSv1AuthorizeEvent", func(b *testing.B) {
		expReply := sessions.V1AuthorizeReply{
			Attributes: &engine.AttrSProcessEventReply{
				MatchedProfiles: []string{"cgrates.org:ATTR_RPC"},
				AlteredFields:   []string{"*req.Password"},
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "benchmarkSession",
					Event: map[string]any{
						utils.AccountField: "1001",
						utils.AnswerTime:   "2023-02-28T09:00:00Z",
						utils.CGRID:        "a276c0f511361744c0b999dc6e85004e96e23595",
						utils.Category:     "call",
						utils.Destination:  "1002",
						utils.OriginID:     "session_rpc",
						"Password":         "CGRateS.org",
						utils.RequestType:  "*prepaid",
						utils.SetupTime:    "2023-02-28T08:59:50Z",
						utils.Subject:      "1001",
						utils.Tenant:       "cgrates.org",
						utils.ToR:          "*voice",
						utils.Usage:        60000000000.,
					},
					APIOpts: map[string]any{
						"*attrProfileIDs": nil,
						"*rsUnits":        1.,
						"*rsUsageID":      "session_rpc",
						"*subsys":         "*sessions",
					},
				},
			},
			ResourceAllocation: utils.StringPointer("RES_RPC"),
			MaxUsage:           utils.DurationPointer(60000000000),
			RouteProfiles: engine.SortedRoutesList{
				{
					ProfileID: "ROUTE_RPC",
					Sorting:   "*weight",
					Routes: []*engine.SortedRoute{
						{
							RouteID:         "route1",
							RouteParameters: "",
							SortingData: map[string]any{
								"Weight": 20.,
							},
						},
						{
							RouteID:         "route2",
							RouteParameters: "",
							SortingData: map[string]any{
								"Weight": 10.,
							},
						},
					},
				},
			},
		}
		argsInit := &sessions.V1AuthorizeArgs{
			GetMaxUsage:        true,
			AuthorizeResources: true,
			GetRoutes:          true,
			GetAttributes:      true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "benchmarkSession",
				Event: map[string]any{
					utils.OriginID:     "session_rpc",
					utils.Tenant:       "cgrates.org",
					utils.Category:     utils.Call,
					utils.ToR:          utils.MetaVoice,
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					utils.Subject:      "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2023, time.February, 28, 8, 59, 50, 0, time.UTC),
					utils.AnswerTime:   time.Date(2023, time.February, 28, 9, 0, 0, 0, time.UTC),
					utils.Usage:        time.Minute,
				},
			},
		}

		var rplyAuth sessions.V1AuthorizeReply
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := benchRPC.Call(
				context.Background(),
				utils.SessionSv1AuthorizeEvent,
				argsInit, &rplyAuth); err != nil {
				b.Error(err)
			}
			b.StopTimer()
			if !reflect.DeepEqual(rplyAuth, expReply) {
				b.Errorf("expected: %s,\nreceived: %s", utils.ToJSON(expReply), utils.ToJSON(rplyAuth))
			}
			b.StartTimer()
		}
	})
}

/*
post update with *internal conns between subsystems
client is sending requests to the json port

$ go test -bench=.  -run=^$ -benchtime=3s -count=10
goos: linux
goarch: amd64
pkg: github.com/cgrates/cgrates/general_tests
cpu: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
BenchmarkRPCCalls/CoreSv1Status         	   18502	    188571 ns/op
BenchmarkRPCCalls/CoreSv1Status         	   19021	    192018 ns/op
BenchmarkRPCCalls/CoreSv1Status         	   19064	    189424 ns/op
BenchmarkRPCCalls/CoreSv1Status         	   18990	    188198 ns/op
BenchmarkRPCCalls/CoreSv1Status         	   18793	    188656 ns/op
BenchmarkRPCCalls/CoreSv1Status         	   18919	    197833 ns/op
BenchmarkRPCCalls/CoreSv1Status         	   18702	    191525 ns/op
BenchmarkRPCCalls/CoreSv1Status         	   18729	    191704 ns/op
BenchmarkRPCCalls/CoreSv1Status         	   18736	    194340 ns/op
BenchmarkRPCCalls/CoreSv1Status         	   18345	    190447 ns/op

BenchmarkRPCCalls/ChargerSv1ProcessEvent         	   14058	    253366 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	   14263	    252998 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	   12042	    254556 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	   14148	    252528 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	   14218	    253519 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	   14152	    253389 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	   14169	    252325 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	   14058	    251939 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	   14196	    252007 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	   14390	    254939 ns/op

BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	    3589	    984448 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	    3398	    975741 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	    3501	    981680 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	    3482	   1109417 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	    3289	    978660 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	    3542	    982441 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	    3139	    983919 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	    3480	    976529 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	    3332	   1021091 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	    3556	    981816 ns/op
*/

/*
post update with *localhost conns between subsystems
client is sending requests to the json port

$ go test -bench=.  -run=^$ -benchtime=3s -count=10
goos: linux
goarch: amd64
pkg: github.com/cgrates/cgrates/general_tests
cpu: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
BenchmarkRPCCalls/CoreSv1Status         	   16543	    235314 ns/op
BenchmarkRPCCalls/CoreSv1Status         	   17408	    222892 ns/op
BenchmarkRPCCalls/CoreSv1Status         	   15898	    210535 ns/op
BenchmarkRPCCalls/CoreSv1Status         	   17500	    224985 ns/op
BenchmarkRPCCalls/CoreSv1Status         	   15091	    260805 ns/op
BenchmarkRPCCalls/CoreSv1Status         	   15942	    201977 ns/op
BenchmarkRPCCalls/CoreSv1Status         	   15996	    210657 ns/op
BenchmarkRPCCalls/CoreSv1Status         	   16407	    221494 ns/op
BenchmarkRPCCalls/CoreSv1Status         	   17919	    214834 ns/op
BenchmarkRPCCalls/CoreSv1Status         	   18087	    331801 ns/op

BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    6456	    617457 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    8203	    416304 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    9064	    406322 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    7372	    418886 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    8120	    477174 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    8268	    549279 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    7045	    499183 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    8192	    430688 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    7153	    430848 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    7917	    429561 ns/op

BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	    1552	   2369496 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	    1380	   2550935 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	    1345	   2872729 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	    1238	   2432697 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	    1568	   2297426 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	    1480	   2225039 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	    1515	   2222200 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	    1563	   2264733 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	    1560	   2235554 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	    1538	   2251135 ns/op
*/

/*
pre update with *internal conns between subsystems
client is sending requests to the json port

$ go test -bench=.  -run=^$ -benchtime=1s -count=10
goos: linux
goarch: amd64
pkg: github.com/cgrates/cgrates/general_tests
cpu: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
BenchmarkRPCCalls/CoreSv1Status         	    6313	    165823 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    6236	    165079 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    6313	    167061 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    7208	    161504 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    7179	    167144 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    7272	    163666 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    7326	    168609 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    7354	    169409 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    6609	    166113 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    7012	    164124 ns/op

BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    1075	    988083 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    1215	   1024714 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    1197	    996550 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    1174	   1061729 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    1171	   1055582 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    1245	   1194874 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    1106	   1019410 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    1233	   1030683 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    1354	    999383 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    1216	    929263 ns/op

BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	     602	   2047117 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	     582	   2101084 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	     626	   2029212 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	     571	   1900435 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	     591	   1990043 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	     636	   2880445 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	     588	   2010745 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	     580	   2006579 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	     567	   1951949 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	     613	   2111439 ns/op
*/

/*
pre update with *localhost conns between subsystems
client is sending requests to the json port

$ go test -bench=.  -run=^$ -benchtime=1s -count=10
goos: linux
goarch: amd64
pkg: github.com/cgrates/cgrates/general_tests
cpu: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
BenchmarkRPCCalls/CoreSv1Status         	    7424	    160989 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    6706	    166310 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    6352	    169117 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    7365	    168071 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    6532	    166651 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    6648	    162875 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    7458	    161252 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    6786	    159809 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    6907	    159827 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    6901	    165061 ns/op

BenchmarkRPCCalls/ChargerSv1ProcessEvent         	     927	   1236949 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    1042	   1210147 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    1033	   1194677 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	     904	   1292240 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    1012	   1179045 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    1018	   1160659 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	     979	   1184598 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    1018	   1234070 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    1033	   1181330 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	     938	   1217141 ns/op

BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	     326	   3615195 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	     345	   3344256 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	     339	   3661695 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	     340	   3389555 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	     352	   3421626 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	     333	   3835950 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	     229	   4937094 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	     283	   3776658 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	     373	   4403626 ns/op
BenchmarkRPCCalls/SessionSv1AuthorizeEvent       	     319	   3789267 ns/op
*/

/*
1.0 (commit #61a781675c4bfdbfcd0fde6abafa35fd61e51db2)
*internal conns between subsystems
client is sending requests to the json port

$ go test -bench=.  -run=^$ -benchtime=1s -count=10
goos: linux
goarch: amd64
pkg: github.com/cgrates/cgrates/general_tests
cpu: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
BenchmarkRPCCalls/CoreSv1Status         	    5997	    169817 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    7062	    168369 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    6366	    168987 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    6744	    170695 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    6456	    169385 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    6723	    168479 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    7118	    373589 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    5866	    175374 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    6260	    185668 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    5992	    202995 ns/op

BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    3858	    298691 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    4044	    292432 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    3966	    318905 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    3394	    297489 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    3862	    317578 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    3523	    289216 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    4228	    286180 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    4004	    281659 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    3752	    397958 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    4188	    281087 ns/op
*/

/*
1.0 (commit #61a781675c4bfdbfcd0fde6abafa35fd61e51db2)
*localhost conns between subsystems
client is sending requests to the json port

$ go test -bench=.  -run=^$ -benchtime=1s -count=10
goos: linux
goarch: amd64
pkg: github.com/cgrates/cgrates/general_tests
cpu: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
BenchmarkRPCCalls/CoreSv1Status         	    6046	    169929 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    6754	    174506 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    6930	    167632 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    6730	    175220 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    5874	    173871 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    6502	    209486 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    5608	    201035 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    6093	    176390 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    6624	    198762 ns/op
BenchmarkRPCCalls/CoreSv1Status         	    6460	    192686 ns/op

BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    2509	    506078 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    2379	    504408 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    2097	    506364 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    2421	    458854 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    1928	    538660 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    2215	    772484 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    2074	    521243 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    2152	    468308 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    2403	    476163 ns/op
BenchmarkRPCCalls/ChargerSv1ProcessEvent         	    2614	    468202 ns/op
*/

type ArgsSum struct {
	A int
	B int
}
type BiRPCObj struct{}

func (b *BiRPCObj) Add(ctx *context.Context, args *ArgsSum, reply *int) error {
	if args == nil {
		return errors.New("missing args")
	}
	*reply = args.A + args.B
	return nil
}

type StdRPCObj struct{}

func (b *StdRPCObj) Add(args *ArgsSum, reply *int) error {
	if args == nil {
		return errors.New("missing args")
	}
	*reply = args.A + args.B
	return nil
}

type server interface {
	Register(any) error
}

func BenchmarkRPC(b *testing.B) {
	tests := []struct {
		lib    string
		codec  string
		method string
	}{
		{"birpc", "json", "BiRPCObj.Add"},
		{"rpc", "json", "StdRPCObj.Add"},
		{"birpc", "gob", "BiRPCObj.Add"},
		{"rpc", "gob", "StdRPCObj.Add"},
	}

	for _, tt := range tests {
		b.Run(fmt.Sprintf("%s %s", tt.lib, tt.codec), func(b *testing.B) {
			var err error
			errChan := make(chan error)
			stopChan := make(chan struct{})

			// Setting up and starting the server.
			var server server
			switch tt.lib {
			case "birpc":
				server = birpc.NewServer()
				var service *birpc.Service
				service, err = birpc.NewService(&BiRPCObj{}, "", false)
				if err != nil {
					b.Fatal(err)
				}
				err = server.Register(service)
			case "rpc":
				server = rpc.NewServer()
				err = server.Register(&StdRPCObj{})
			default:
				b.Fatal("unsupported lib")
			}
			if err != nil {
				b.Fatal(err)
			}

			go listenAndServe(server, tt.codec, "tcp", "127.0.0.1:2012", errChan, stopChan)

			select {
			case err := <-errChan:
				b.Fatal(err)
			case <-time.After(time.Second):
			}

			// Creating the client.
			var client any
			switch tt.lib + "." + tt.codec {
			case "birpc.json":
				client, err = jsonrpc2.Dial("tcp", "127.0.0.1:2012")
			case "birpc.gob":
				client, err = birpc.Dial("tcp", "127.0.0.1:2012")
			case "rpc.json":
				client, err = jsonrpc.Dial("tcp", "127.0.0.1:2012")
			case "rpc.gob":
				client, err = rpc.Dial("tcp", "127.0.0.1:2012")
			default:
				b.Fatal("unsupported codec")
			}
			if err != nil {
				b.Fatal(err)
			}

			args := &ArgsSum{A: 5, B: 3}
			runBenchmark(b, client, tt.method, args)

			close(stopChan)
		})
	}
}
func listenAndServe(server server, codec, network, address string, errChan chan error, stopChan chan struct{}) {
	l, err := net.Listen(network, address)
	if err != nil {
		errChan <- err
		return
	}
	defer l.Close()

	var serveFunc func(conn io.ReadWriteCloser)

	switch s := server.(type) {
	case *birpc.Server:
		if codec == "json" {
			serveFunc = func(conn io.ReadWriteCloser) { s.ServeCodec(jsonrpc2.NewServerCodec(conn)) }
		} else {
			serveFunc = s.ServeConn
		}
	case *rpc.Server:
		if codec == "json" {
			serveFunc = func(conn io.ReadWriteCloser) { s.ServeCodec(jsonrpc.NewServerCodec(conn)) }
		} else {
			serveFunc = s.ServeConn
		}
	}

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go serveFunc(conn)
		}
	}()
	<-stopChan
}

func runBenchmark(b *testing.B, client any, method string, args *ArgsSum) {
	var reply int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var err error
		switch c := client.(type) {
		case *birpc.Client:
			err = c.Call(context.Background(), method, args, &reply)
		case *rpc.Client:
			err = c.Call(method, args, &reply)
		}
		b.StopTimer()
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()
	}
}

/*
$ go test -bench=.  -run=^$ -benchtime=1s -count=10 -benchmem
goos: linux
goarch: amd64
pkg: github.com/cgrates/cgrates/general_tests
cpu: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
BenchmarkRPC/birpc_json         	   18892	     60452 ns/op	    1096 B/op	      30 allocs/op
BenchmarkRPC/birpc_json         	   19707	     60350 ns/op	    1096 B/op	      30 allocs/op
BenchmarkRPC/birpc_json         	   19647	     61393 ns/op	    1096 B/op	      30 allocs/op
BenchmarkRPC/birpc_json         	   19410	     60748 ns/op	    1096 B/op	      30 allocs/op
BenchmarkRPC/birpc_json         	   19650	     60855 ns/op	    1096 B/op	      30 allocs/op
BenchmarkRPC/birpc_json         	   19716	     60592 ns/op	    1096 B/op	      30 allocs/op
BenchmarkRPC/birpc_json         	   19425	     61236 ns/op	    1096 B/op	      30 allocs/op
BenchmarkRPC/birpc_json         	   19503	     60812 ns/op	    1096 B/op	      30 allocs/op
BenchmarkRPC/birpc_json         	   18788	     60251 ns/op	    1096 B/op	      30 allocs/op
BenchmarkRPC/birpc_json         	   19716	     60292 ns/op	    1096 B/op	      30 allocs/op

BenchmarkRPC/rpc_json                 	   21067	     57189 ns/op	     960 B/op	      27 allocs/op
BenchmarkRPC/rpc_json                 	   20818	     57024 ns/op	     960 B/op	      27 allocs/op
BenchmarkRPC/rpc_json                 	   20786	     56784 ns/op	     960 B/op	      27 allocs/op
BenchmarkRPC/rpc_json                 	   20493	     57376 ns/op	     960 B/op	      27 allocs/op
BenchmarkRPC/rpc_json                 	   20658	     56967 ns/op	     960 B/op	      27 allocs/op
BenchmarkRPC/rpc_json                 	   20936	     56488 ns/op	     960 B/op	      27 allocs/op
BenchmarkRPC/rpc_json                 	   19545	     56708 ns/op	     960 B/op	      27 allocs/op
BenchmarkRPC/rpc_json                 	   10000	    129676 ns/op	     960 B/op	      27 allocs/op
BenchmarkRPC/rpc_json                 	   21284	     57897 ns/op	     960 B/op	      27 allocs/op
BenchmarkRPC/rpc_json                 	   20736	     56793 ns/op	     960 B/op	      27 allocs/op

BenchmarkRPC/birpc_gob           	   21524	     55731 ns/op	     576 B/op	      16 allocs/op
BenchmarkRPC/birpc_gob           	   21249	     55345 ns/op	     576 B/op	      16 allocs/op
BenchmarkRPC/birpc_gob           	   21350	     55174 ns/op	     576 B/op	      16 allocs/op
BenchmarkRPC/birpc_gob           	   20890	     55861 ns/op	     576 B/op	      16 allocs/op
BenchmarkRPC/birpc_gob           	   20028	     57341 ns/op	     576 B/op	      16 allocs/op
BenchmarkRPC/birpc_gob           	   21411	     73762 ns/op	     576 B/op	      16 allocs/op
BenchmarkRPC/birpc_gob           	   21535	     55137 ns/op	     576 B/op	      16 allocs/op
BenchmarkRPC/birpc_gob           	   21597	     55909 ns/op	     576 B/op	      16 allocs/op
BenchmarkRPC/birpc_gob           	   21319	     55335 ns/op	     576 B/op	      16 allocs/op
BenchmarkRPC/birpc_gob           	   21225	     55201 ns/op	     576 B/op	      16 allocs/op

BenchmarkRPC/rpc_gob                   	   24734	     48072 ns/op	     464 B/op	      14 allocs/op
BenchmarkRPC/rpc_gob                   	   24391	     47952 ns/op	     464 B/op	      14 allocs/op
BenchmarkRPC/rpc_gob                   	   24158	     48643 ns/op	     464 B/op	      14 allocs/op
BenchmarkRPC/rpc_gob                   	   24338	     47862 ns/op	     464 B/op	      14 allocs/op
BenchmarkRPC/rpc_gob                   	   24516	     47860 ns/op	     464 B/op	      14 allocs/op
BenchmarkRPC/rpc_gob                   	   24823	     47655 ns/op	     464 B/op	      14 allocs/op
BenchmarkRPC/rpc_gob                   	   25010	     47496 ns/op	     464 B/op	      14 allocs/op
BenchmarkRPC/rpc_gob                   	   25168	     47556 ns/op	     464 B/op	      14 allocs/op
BenchmarkRPC/rpc_gob                   	   25087	     49648 ns/op	     464 B/op	      14 allocs/op
BenchmarkRPC/rpc_gob                   	   23858	     49094 ns/op	     464 B/op	      14 allocs/op
*/
