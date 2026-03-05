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

package main

import (
	"fmt"
	"os"
)

type replGet struct {
	Method string // e.g. "GetAccount"
	Prefix string // e.g. "utils.AccountPrefix"
	Meta   string // e.g. "utils.MetaAccounts"
	Reply  string // e.g. "utils.Account"
	Drv    string // e.g. "GetAccountDrv"
}

type replSet struct {
	Method    string // e.g. "SetAccount"
	Meta      string // e.g. "utils.MetaAccounts"
	ArgsType  string // e.g. "*utils.AccountWithAPIOpts"
	Inner     string // field path on args, e.g. "args.Account"
	Drv       string // e.g. "SetAccountDrv"
	Cache     string // e.g. "utils.CacheAccounts" or "" for no cache
	FilterIDs bool   // pass &args.FilterIDs to CallCache
}

type replRemove struct {
	Method string // e.g. "RemoveAccount"
	Meta   string // e.g. "utils.MetaAccounts"
	Drv    string // e.g. "RemoveAccountDrv"
	Cache  string // e.g. "utils.CacheAccounts" or "" for no cache
}

var gets = []replGet{
	{"GetAccount", "utils.AccountPrefix", "utils.MetaAccounts", "utils.Account", "GetAccountDrv"},
	{"GetStatQueue", "utils.StatQueuePrefix", "utils.MetaStatQueues", "engine.StatQueue", "GetStatQueueDrv"},
	{"GetFilter", "utils.FilterPrefix", "utils.MetaFilters", "engine.Filter", "GetFilterDrv"},
	{"GetThreshold", "utils.ThresholdPrefix", "utils.MetaThresholds", "engine.Threshold", "GetThresholdDrv"},
	{"GetThresholdProfile", "utils.ThresholdProfilePrefix", "utils.MetaThresholdProfiles", "engine.ThresholdProfile", "GetThresholdProfileDrv"},
	{"GetStatQueueProfile", "utils.StatQueueProfilePrefix", "utils.MetaStatQueueProfiles", "engine.StatQueueProfile", "GetStatQueueProfileDrv"},
	{"GetTrendProfile", "utils.TrendProfilePrefix", "utils.MetaTrendProfiles", "utils.TrendProfile", "GetTrendProfileDrv"},
	{"GetResource", "utils.ResourcesPrefix", "utils.MetaResources", "utils.Resource", "GetResourceDrv"},
	{"GetResourceProfile", "utils.ResourceProfilesPrefix", "utils.MetaResourceProfiles", "utils.ResourceProfile", "GetResourceProfileDrv"},
	{"GetIPAllocations", "utils.IPAllocationsPrefix", "utils.MetaIPAllocations", "utils.IPAllocations", "GetIPAllocationsDrv"},
	{"GetIPProfile", "utils.IPProfilesPrefix", "utils.MetaIPProfiles", "utils.IPProfile", "GetIPProfileDrv"},
	{"GetRankingProfile", "utils.RankingProfilePrefix", "utils.MetaRankingProfiles", "utils.RankingProfile", "GetRankingProfileDrv"},
	{"GetRouteProfile", "utils.RouteProfilePrefix", "utils.MetaRouteProfiles", "utils.RouteProfile", "GetRouteProfileDrv"},
	{"GetAttributeProfile", "utils.AttributeProfilePrefix", "utils.MetaAttributeProfiles", "utils.AttributeProfile", "GetAttributeProfileDrv"},
	{"GetChargerProfile", "utils.ChargerProfilePrefix", "utils.MetaChargerProfiles", "utils.ChargerProfile", "GetChargerProfileDrv"},
	{"GetRateProfile", "utils.RateProfilePrefix", "utils.MetaRateProfiles", "utils.RateProfile", "GetRateProfileDrv"},
	{"GetActionProfile", "utils.ActionProfilePrefix", "utils.MetaActionProfiles", "utils.ActionProfile", "GetActionProfileDrv"},
}

var sets = []replSet{
	{"SetAccount", "utils.MetaAccounts", "*utils.AccountWithAPIOpts", "args.Account", "SetAccountDrv", "", false},
	{"SetThresholdProfile", "utils.MetaThresholdProfiles", "*engine.ThresholdProfileWithAPIOpts", "args.ThresholdProfile", "SetThresholdProfileDrv", "utils.CacheThresholdProfiles", true},
	{"SetThreshold", "utils.MetaThresholds", "*engine.ThresholdWithAPIOpts", "args.Threshold", "SetThresholdDrv", "utils.CacheThresholds", false},
	{"SetTrendProfile", "utils.MetaTrendProfiles", "*utils.TrendProfileWithAPIOpts", "args.TrendProfile", "SetTrendProfileDrv", "utils.CacheTrendProfiles", false},
	{"SetTrend", "utils.MetaTrends", "*utils.TrendWithAPIOpts", "args.Trend", "SetTrendDrv", "utils.CacheTrends", false},
	{"SetStatQueueProfile", "utils.MetaStatQueueProfiles", "*engine.StatQueueProfileWithAPIOpts", "args.StatQueueProfile", "SetStatQueueProfileDrv", "utils.CacheStatQueueProfiles", true},
	{"SetFilter", "utils.MetaFilters", "*engine.FilterWithAPIOpts", "args.Filter", "SetFilterDrv", "utils.CacheFilters", false},
	{"SetResourceProfile", "utils.MetaResourceProfiles", "*utils.ResourceProfileWithAPIOpts", "args.ResourceProfile", "SetResourceProfileDrv", "utils.CacheResourceProfiles", true},
	{"SetResource", "utils.MetaResources", "*utils.ResourceWithAPIOpts", "args.Resource", "SetResourceDrv", "utils.CacheResources", false},
	{"SetIPProfile", "utils.MetaIPProfiles", "*utils.IPProfileWithAPIOpts", "args.IPProfile", "SetIPProfileDrv", "utils.CacheIPProfiles", true},
	{"SetIPAllocations", "utils.MetaIPAllocations", "*utils.IPAllocationsWithAPIOpts", "args.IPAllocations", "SetIPAllocationsDrv", "utils.CacheIPAllocations", false},
	{"SetRankingProfile", "utils.MetaRankingProfiles", "*utils.RankingProfileWithAPIOpts", "args.RankingProfile", "SetRankingProfileDrv", "utils.CacheRankingProfiles", false},
	{"SetRanking", "utils.MetaRankings", "*utils.RankingWithAPIOpts", "args.Ranking", "SetRankingDrv", "utils.CacheRankings", false},
	{"SetRouteProfile", "utils.MetaRouteProfiles", "*utils.RouteProfileWithAPIOpts", "args.RouteProfile", "SetRouteProfileDrv", "utils.CacheRouteProfiles", true},
	{"SetAttributeProfile", "utils.MetaAttributeProfiles", "*utils.AttributeProfileWithAPIOpts", "args.AttributeProfile", "SetAttributeProfileDrv", "utils.CacheAttributeProfiles", true},
	{"SetChargerProfile", "utils.MetaChargerProfiles", "*utils.ChargerProfileWithAPIOpts", "args.ChargerProfile", "SetChargerProfileDrv", "utils.CacheChargerProfiles", true},
	{"SetActionProfile", "utils.MetaActionProfiles", "*utils.ActionProfileWithAPIOpts", "args.ActionProfile", "SetActionProfileDrv", "utils.CacheActionProfiles", true},
}

var removes = []replRemove{
	{"RemoveAccount", "utils.MetaAccounts", "RemoveAccountDrv", ""},
	{"RemoveThreshold", "utils.MetaThresholds", "RemoveThresholdDrv", "utils.CacheThresholds"},
	{"RemoveThresholdProfile", "utils.MetaThresholdProfiles", "RemThresholdProfileDrv", "utils.CacheThresholdProfiles"},
	{"RemoveTrend", "utils.MetaTrends", "RemoveTrendDrv", "utils.CacheTrends"},
	{"RemoveTrendProfile", "utils.MetaTrendProfiles", "RemTrendProfileDrv", "utils.CacheTrendProfiles"},
	{"RemoveStatQueue", "utils.MetaStatQueues", "RemStatQueueDrv", "utils.CacheStatQueues"},
	{"RemoveStatQueueProfile", "utils.MetaStatQueueProfiles", "RemStatQueueProfileDrv", "utils.CacheStatQueueProfiles"},
	{"RemoveFilter", "utils.MetaFilters", "RemoveFilterDrv", "utils.CacheFilters"},
	{"RemoveResource", "utils.MetaResources", "RemoveResourceDrv", "utils.CacheResources"},
	{"RemoveResourceProfile", "utils.MetaResourceProfiles", "RemoveResourceProfileDrv", "utils.CacheResourceProfiles"},
	{"RemoveIPAllocations", "utils.MetaIPAllocations", "RemoveIPAllocationsDrv", "utils.CacheIPAllocations"},
	{"RemoveIPProfile", "utils.MetaIPProfiles", "RemoveIPProfileDrv", "utils.CacheIPProfiles"},
	{"RemoveRankingProfile", "utils.MetaRankingProfiles", "RemRankingProfileDrv", "utils.CacheRankingProfiles"},
	{"RemoveRanking", "utils.MetaRankings", "RemoveRankingDrv", "utils.CacheRankings"},
	{"RemoveRouteProfile", "utils.MetaRouteProfiles", "RemoveRouteProfileDrv", "utils.CacheRouteProfiles"},
	{"RemoveAttributeProfile", "utils.MetaAttributeProfiles", "RemoveAttributeProfileDrv", "utils.CacheAttributeProfiles"},
	{"RemoveChargerProfile", "utils.MetaChargerProfiles", "RemoveChargerProfileDrv", "utils.CacheChargerProfiles"},
	{"RemoveActionProfile", "utils.MetaActionProfiles", "RemoveActionProfileDrv", "utils.CacheActionProfiles"},
}

func main() {
	f, err := os.Create("replicator_gen.go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	p := func(format string, args ...any) {
		fmt.Fprintf(f, format, args...)
	}

	p("/*\n")
	p("Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments\n")
	p("Copyright (C) ITsysCOM GmbH\n")
	p("\n")
	p("This program is free software: you can redistribute it and/or modify\n")
	p("it under the terms of the GNU Affero General Public License as published by\n")
	p("the Free Software Foundation, either version 3 of the License, or\n")
	p("(at your option) any later version.\n")
	p("\n")
	p("This program is distributed in the hope that it will be useful,\n")
	p("but WITHOUT ANY WARRANTY; without even the implied warranty of\n")
	p("MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the\n")
	p("GNU Affero General Public License for more details.\n")
	p("\n")
	p("You should have received a copy of the GNU Affero General Public License\n")
	p("along with this program.  If not, see <https://www.gnu.org/licenses/>\n")
	p("*/\n\n")
	p("// Code generated by generate_replicator; DO NOT EDIT.\n\n")
	p("package apis\n\n")
	p("import (\n")
	p("\t\"fmt\"\n")
	p("\t\"time\"\n\n")
	p("\t\"github.com/cgrates/birpc/context\"\n")
	p("\t\"github.com/cgrates/cgrates/engine\"\n")
	p("\t\"github.com/cgrates/cgrates/utils\"\n")
	p(")\n\n")

	// Struct definition.
	p("// ReplicatorSv1 exports DataDB methods as RPC endpoints for replication.\n")
	p("type ReplicatorSv1 struct {\n")
	p("\tping\n")
	p("\tdm    *engine.DataManager\n")
	p("\tadmin *AdminSv1\n")
	p("}\n\n")

	// Constructor.
	p("// NewReplicatorSv1 creates a new ReplicatorSv1.\n")
	p("func NewReplicatorSv1(dm *engine.DataManager, admin *AdminSv1) *ReplicatorSv1 {\n")
	p("\treturn &ReplicatorSv1{dm: dm, admin: admin}\n")
	p("}\n")

	// Get methods.
	for _, g := range gets {
		p("\nfunc (r *ReplicatorSv1) %s(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *%s) error {\n", g.Method, g.Reply)
		p("\tengine.UpdateReplicationFilters(%s, args.TenantID.TenantID(), utils.IfaceAsString(args.APIOpts[utils.RemoteHostOpt]))\n", g.Prefix)
		p("\tdataDB, _, err := r.dm.DBConns().GetConn(%s)\n", g.Meta)
		p("\tif err != nil {\n")
		p("\t\treturn err\n")
		p("\t}\n")
		p("\trcv, err := dataDB.%s(ctx, args.Tenant, args.ID)\n", g.Drv)
		p("\tif err != nil {\n")
		p("\t\treturn err\n")
		p("\t}\n")
		p("\t*reply = *rcv\n")
		p("\treturn nil\n")
		p("}\n")
	}

	// Set methods.
	for _, s := range sets {
		p("\nfunc (r *ReplicatorSv1) %s(ctx *context.Context, args %s, reply *string) error {\n", s.Method, s.ArgsType)
		p("\tdataDB, _, err := r.dm.DBConns().GetConn(%s)\n", s.Meta)
		p("\tif err != nil {\n")
		p("\t\treturn err\n")
		p("\t}\n")
		p("\tif err := dataDB.%s(ctx, %s); err != nil {\n", s.Drv, s.Inner)
		p("\t\treturn err\n")
		p("\t}\n")
		if s.Cache != "" {
			p("\tif r.admin.cfg.GeneralCfg().CachingDelay != 0 {\n")
			p("\t\tutils.Logger.Info(fmt.Sprintf(\"<ReplicatorSv1.%s> Delaying cache call for %%v\", r.admin.cfg.GeneralCfg().CachingDelay))\n", s.Method)
			p("\t\ttime.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)\n")
			p("\t}\n")
			filterArg := "nil"
			if s.FilterIDs {
				filterArg = "&args.FilterIDs"
			}
			p("\tif err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),\n")
			p("\t\targs.Tenant, %s, args.TenantID(), \"\", %s, args.APIOpts); err != nil {\n", s.Cache, filterArg)
			p("\t\treturn err\n")
			p("\t}\n")
		}
		p("\t*reply = utils.OK\n")
		p("\treturn nil\n")
		p("}\n")
	}

	// Remove methods.
	for _, rm := range removes {
		p("\nfunc (r *ReplicatorSv1) %s(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {\n", rm.Method)
		p("\tdataDB, _, err := r.dm.DBConns().GetConn(%s)\n", rm.Meta)
		p("\tif err != nil {\n")
		p("\t\treturn err\n")
		p("\t}\n")
		p("\tif err := dataDB.%s(ctx, args.Tenant, args.ID); err != nil {\n", rm.Drv)
		p("\t\treturn err\n")
		p("\t}\n")
		if rm.Cache != "" {
			p("\tif r.admin.cfg.GeneralCfg().CachingDelay != 0 {\n")
			p("\t\tutils.Logger.Info(fmt.Sprintf(\"<ReplicatorSv1.%s> Delaying cache call for %%v\", r.admin.cfg.GeneralCfg().CachingDelay))\n", rm.Method)
			p("\t\ttime.Sleep(r.admin.cfg.GeneralCfg().CachingDelay)\n")
			p("\t}\n")
			p("\tif err := r.admin.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]),\n")
			p("\t\targs.Tenant, %s, args.TenantID.TenantID(), \"\", nil, args.APIOpts); err != nil {\n", rm.Cache)
			p("\t\treturn err\n")
			p("\t}\n")
		}
		p("\t*reply = utils.OK\n")
		p("\treturn nil\n")
		p("}\n")
	}
}
