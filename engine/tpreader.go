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

package engine

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

type TpReader struct {
	tpid               string
	timezone           string
	dm                 *DataManager
	lr                 LoadReader
	timings            map[string]*utils.TPTiming
	resProfiles        map[utils.TenantID]*utils.TPResourceProfile
	sqProfiles         map[utils.TenantID]*utils.TPStatProfile
	thProfiles         map[utils.TenantID]*utils.TPThresholdProfile
	filters            map[utils.TenantID]*utils.TPFilterProfile
	routeProfiles      map[utils.TenantID]*utils.TPRouteProfile
	attributeProfiles  map[utils.TenantID]*utils.TPAttributeProfile
	chargerProfiles    map[utils.TenantID]*utils.TPChargerProfile
	dispatcherProfiles map[utils.TenantID]*utils.TPDispatcherProfile
	dispatcherHosts    map[utils.TenantID]*utils.TPDispatcherHost
	rateProfiles       map[utils.TenantID]*utils.TPRateProfile
	actionProfiles     map[utils.TenantID]*utils.TPActionProfile
	accounts           map[utils.TenantID]*utils.TPAccount
	resources          []*utils.TenantID // IDs of resources which need creation based on resourceProfiles
	statQueues         []*utils.TenantID // IDs of statQueues which need creation based on statQueueProfiles
	thresholds         []*utils.TenantID // IDs of thresholds which need creation based on thresholdProfiles
	acntActionPlans    map[string][]string
	cacheConns         []string
	//schedulerConns     []string
	isInternalDB bool // do not reload cache if we use internalDB
}

func NewTpReader(db DataDB, lr LoadReader, tpid, timezone string,
	cacheConns, schedulerConns []string, isInternalDB bool) (*TpReader, error) {

	tpr := &TpReader{
		tpid:       tpid,
		timezone:   timezone,
		dm:         NewDataManager(db, config.CgrConfig().CacheCfg(), connMgr), // ToDo: add ChacheCfg as parameter to the NewTpReader
		lr:         lr,
		cacheConns: cacheConns,
		//schedulerConns: schedulerConns,
		isInternalDB: isInternalDB,
	}
	tpr.Init()
	//add default timing tag (in case of no timings file)
	tpr.addDefaultTimings()

	return tpr, nil
}

func (tpr *TpReader) Init() {
	tpr.timings = make(map[string]*utils.TPTiming)
	tpr.resProfiles = make(map[utils.TenantID]*utils.TPResourceProfile)
	tpr.sqProfiles = make(map[utils.TenantID]*utils.TPStatProfile)
	tpr.thProfiles = make(map[utils.TenantID]*utils.TPThresholdProfile)
	tpr.routeProfiles = make(map[utils.TenantID]*utils.TPRouteProfile)
	tpr.attributeProfiles = make(map[utils.TenantID]*utils.TPAttributeProfile)
	tpr.chargerProfiles = make(map[utils.TenantID]*utils.TPChargerProfile)
	tpr.dispatcherProfiles = make(map[utils.TenantID]*utils.TPDispatcherProfile)
	tpr.dispatcherHosts = make(map[utils.TenantID]*utils.TPDispatcherHost)
	tpr.rateProfiles = make(map[utils.TenantID]*utils.TPRateProfile)
	tpr.actionProfiles = make(map[utils.TenantID]*utils.TPActionProfile)
	tpr.accounts = make(map[utils.TenantID]*utils.TPAccount)
	tpr.filters = make(map[utils.TenantID]*utils.TPFilterProfile)
	tpr.acntActionPlans = make(map[string][]string)
}

func (tpr *TpReader) LoadTimings() (err error) {
	tps, err := tpr.lr.GetTPTimings(tpr.tpid, "")
	if err != nil {
		return err
	}
	var tpTimings map[string]*utils.TPTiming
	if tpTimings, err = MapTPTimings(tps); err != nil {
		return
	}
	// add default timings
	tpr.addDefaultTimings()
	// add timings defined by user
	for timingID, timing := range tpTimings {
		tpr.timings[timingID] = timing
	}
	return err
}

func (tpr *TpReader) LoadResourceProfilesFiltered(tag string) (err error) {
	rls, err := tpr.lr.GetTPResources(tpr.tpid, "", tag)
	if err != nil {
		return err
	}
	mapRsPfls := make(map[utils.TenantID]*utils.TPResourceProfile)
	for _, rl := range rls {
		if err = verifyInlineFilterS(rl.FilterIDs); err != nil {
			return
		}
		tpr.resources = append(tpr.resources, &utils.TenantID{Tenant: rl.Tenant, ID: rl.ID})
		mapRsPfls[utils.TenantID{Tenant: rl.Tenant, ID: rl.ID}] = rl
	}
	tpr.resProfiles = mapRsPfls
	return nil
}

func (tpr *TpReader) LoadResourceProfiles() error {
	return tpr.LoadResourceProfilesFiltered("")
}

func (tpr *TpReader) LoadStatsFiltered(tag string) (err error) {
	tps, err := tpr.lr.GetTPStats(tpr.tpid, "", tag)
	if err != nil {
		return err
	}
	mapSTs := make(map[utils.TenantID]*utils.TPStatProfile)
	for _, st := range tps {
		if err = verifyInlineFilterS(st.FilterIDs); err != nil {
			return
		}
		mapSTs[utils.TenantID{Tenant: st.Tenant, ID: st.ID}] = st
		tpr.statQueues = append(tpr.statQueues, &utils.TenantID{Tenant: st.Tenant, ID: st.ID})
	}
	tpr.sqProfiles = mapSTs
	return nil
}

func (tpr *TpReader) LoadStats() error {
	return tpr.LoadStatsFiltered("")
}

func (tpr *TpReader) LoadThresholdsFiltered(tag string) (err error) {
	tps, err := tpr.lr.GetTPThresholds(tpr.tpid, "", tag)
	if err != nil {
		return err
	}
	mapTHs := make(map[utils.TenantID]*utils.TPThresholdProfile)
	for _, th := range tps {
		if err = verifyInlineFilterS(th.FilterIDs); err != nil {
			return
		}
		mapTHs[utils.TenantID{Tenant: th.Tenant, ID: th.ID}] = th
	}
	tpr.thProfiles = mapTHs
	for tntID := range mapTHs {
		tpr.thresholds = append(tpr.thresholds, &utils.TenantID{Tenant: tntID.Tenant, ID: tntID.ID})

	}
	return nil
}

func (tpr *TpReader) LoadThresholds() error {
	return tpr.LoadThresholdsFiltered("")
}

func (tpr *TpReader) LoadFiltersFiltered(tag string) error {
	tps, err := tpr.lr.GetTPFilters(tpr.tpid, "", tag)
	if err != nil {
		return err
	}
	mapTHs := make(map[utils.TenantID]*utils.TPFilterProfile)
	for _, th := range tps {
		mapTHs[utils.TenantID{Tenant: th.Tenant, ID: th.ID}] = th
	}
	tpr.filters = mapTHs
	return nil
}

func (tpr *TpReader) LoadFilters() error {
	return tpr.LoadFiltersFiltered("")
}

func (tpr *TpReader) LoadRouteProfilesFiltered(tag string) (err error) {
	rls, err := tpr.lr.GetTPRoutes(tpr.tpid, "", tag)
	if err != nil {
		return err
	}
	mapRsPfls := make(map[utils.TenantID]*utils.TPRouteProfile)
	for _, rl := range rls {
		if err = verifyInlineFilterS(rl.FilterIDs); err != nil {
			return
		}
		mapRsPfls[utils.TenantID{Tenant: rl.Tenant, ID: rl.ID}] = rl
	}
	tpr.routeProfiles = mapRsPfls
	return nil
}

func (tpr *TpReader) LoadRouteProfiles() error {
	return tpr.LoadRouteProfilesFiltered("")
}

func (tpr *TpReader) LoadAttributeProfilesFiltered(tag string) (err error) {
	attrs, err := tpr.lr.GetTPAttributes(tpr.tpid, "", tag)
	if err != nil {
		return err
	}
	mapAttrPfls := make(map[utils.TenantID]*utils.TPAttributeProfile)
	for _, attr := range attrs {
		if err = verifyInlineFilterS(attr.FilterIDs); err != nil {
			return
		}
		for _, at := range attr.Attributes {
			if at.Path == utils.EmptyString { // we do not suppot empty Path in Attributes
				err = fmt.Errorf("empty path in AttributeProfile <%s>", utils.ConcatenatedKey(attr.Tenant, attr.ID))
				return
			}
		}
		mapAttrPfls[utils.TenantID{Tenant: attr.Tenant, ID: attr.ID}] = attr
	}
	tpr.attributeProfiles = mapAttrPfls
	return nil
}

func (tpr *TpReader) LoadAttributeProfiles() error {
	return tpr.LoadAttributeProfilesFiltered("")
}

func (tpr *TpReader) LoadChargerProfilesFiltered(tag string) (err error) {
	rls, err := tpr.lr.GetTPChargers(tpr.tpid, "", tag)
	if err != nil {
		return err
	}
	mapChargerProfile := make(map[utils.TenantID]*utils.TPChargerProfile)
	for _, rl := range rls {
		if err = verifyInlineFilterS(rl.FilterIDs); err != nil {
			return
		}
		mapChargerProfile[utils.TenantID{Tenant: rl.Tenant, ID: rl.ID}] = rl
	}
	tpr.chargerProfiles = mapChargerProfile
	return nil
}

func (tpr *TpReader) LoadChargerProfiles() error {
	return tpr.LoadChargerProfilesFiltered("")
}

func (tpr *TpReader) LoadDispatcherProfilesFiltered(tag string) (err error) {
	rls, err := tpr.lr.GetTPDispatcherProfiles(tpr.tpid, "", tag)
	if err != nil {
		return err
	}
	mapDispatcherProfile := make(map[utils.TenantID]*utils.TPDispatcherProfile)
	for _, rl := range rls {
		if err = verifyInlineFilterS(rl.FilterIDs); err != nil {
			return
		}
		mapDispatcherProfile[utils.TenantID{Tenant: rl.Tenant, ID: rl.ID}] = rl
	}
	tpr.dispatcherProfiles = mapDispatcherProfile
	return nil
}

func (tpr *TpReader) LoadDispatcherProfiles() error {
	return tpr.LoadDispatcherProfilesFiltered("")
}

func (tpr *TpReader) LoadDispatcherHostsFiltered(tag string) (err error) {
	rls, err := tpr.lr.GetTPDispatcherHosts(tpr.tpid, "", tag)
	if err != nil {
		return err
	}
	mapDispatcherHost := make(map[utils.TenantID]*utils.TPDispatcherHost)
	for _, rl := range rls {
		mapDispatcherHost[utils.TenantID{Tenant: rl.Tenant, ID: rl.ID}] = rl
	}
	tpr.dispatcherHosts = mapDispatcherHost
	return nil
}

func (tpr *TpReader) LoadRateProfiles() error {
	return tpr.LoadRateProfilesFiltered("")
}

func (tpr *TpReader) LoadRateProfilesFiltered(tag string) (err error) {
	rls, err := tpr.lr.GetTPRateProfiles(tpr.tpid, "", tag)
	if err != nil {
		return err
	}
	mapRateProfiles := make(map[utils.TenantID]*utils.TPRateProfile)
	for _, rl := range rls {
		if err = verifyInlineFilterS(rl.FilterIDs); err != nil {
			return
		}
		mapRateProfiles[utils.TenantID{Tenant: rl.Tenant, ID: rl.ID}] = rl
	}
	tpr.rateProfiles = mapRateProfiles
	return nil
}

func (tpr *TpReader) LoadActionProfiles() error {
	return tpr.LoadActionProfilesFiltered("")
}

func (tpr *TpReader) LoadActionProfilesFiltered(tag string) (err error) {
	aps, err := tpr.lr.GetTPActionProfiles(tpr.tpid, "", tag)
	if err != nil {
		return err
	}
	mapActionProfiles := make(map[utils.TenantID]*utils.TPActionProfile)
	for _, ap := range aps {
		if err = verifyInlineFilterS(ap.FilterIDs); err != nil {
			return
		}
		mapActionProfiles[utils.TenantID{Tenant: ap.Tenant, ID: ap.ID}] = ap
	}
	tpr.actionProfiles = mapActionProfiles
	return nil
}

func (tpr *TpReader) LoadAccounts() error {
	return tpr.LoadAccountsFiltered("")
}

func (tpr *TpReader) LoadAccountsFiltered(tag string) (err error) {
	aps, err := tpr.lr.GetTPAccounts(tpr.tpid, "", tag)
	if err != nil {
		return err
	}
	mapAccounts := make(map[utils.TenantID]*utils.TPAccount)
	for _, ap := range aps {
		if err = verifyInlineFilterS(ap.FilterIDs); err != nil {
			return
		}
		mapAccounts[utils.TenantID{Tenant: ap.Tenant, ID: ap.ID}] = ap
	}
	tpr.accounts = mapAccounts
	return nil
}

func (tpr *TpReader) LoadDispatcherHosts() error {
	return tpr.LoadDispatcherHostsFiltered("")
}

func (tpr *TpReader) LoadAll() (err error) {
	if err = tpr.LoadTimings(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadFilters(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadResourceProfiles(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadStats(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadThresholds(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadRouteProfiles(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadAttributeProfiles(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadChargerProfiles(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadDispatcherProfiles(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadDispatcherHosts(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadRateProfiles(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadActionProfiles(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadAccounts(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	return nil
}

func (tpr *TpReader) WriteToDatabase(verbose, disableReverse bool) (err error) {
	if tpr.dm.dataDB == nil {
		return errors.New("no database connection")
	}
	//generate a loadID
	loadID := time.Now().UnixNano()
	loadIDs := make(map[string]int64)

	if verbose {
		log.Print("Filters:")
	}
	for _, tpTH := range tpr.filters {
		var th *Filter
		if th, err = APItoFilter(tpTH, tpr.timezone); err != nil {
			return
		}
		if err = tpr.dm.SetFilter(context.TODO(), th, true); err != nil {
			return
		}
		if verbose {
			log.Print("\t", th.TenantID())
		}
	}
	if len(tpr.filters) != 0 {
		loadIDs[utils.CacheFilters] = loadID
	}
	if verbose {
		log.Print("ResourceProfiles:")
	}
	for _, tpRsp := range tpr.resProfiles {
		var rsp *ResourceProfile
		if rsp, err = APItoResource(tpRsp, tpr.timezone); err != nil {
			return
		}
		if err = tpr.dm.SetResourceProfile(rsp, true); err != nil {
			return
		}
		if verbose {
			log.Print("\t", rsp.TenantID())
		}
	}
	if len(tpr.resProfiles) != 0 {
		loadIDs[utils.CacheResourceProfiles] = loadID
	}
	if verbose {
		log.Print("Resources:")
	}
	for _, rTid := range tpr.resources {
		var ttl *time.Duration
		if tpr.resProfiles[*rTid].UsageTTL != utils.EmptyString {
			ttl = new(time.Duration)
			if *ttl, err = utils.ParseDurationWithNanosecs(tpr.resProfiles[*rTid].UsageTTL); err != nil {
				return
			}
			if *ttl <= 0 {
				ttl = nil
			}
		}
		var limit float64
		if tpr.resProfiles[*rTid].Limit != utils.EmptyString {
			if limit, err = strconv.ParseFloat(tpr.resProfiles[*rTid].Limit, 64); err != nil {
				return
			}
		}
		// for non stored we do not save the resource
		if err = tpr.dm.SetResource(
			&Resource{
				Tenant: rTid.Tenant,
				ID:     rTid.ID,
				Usages: make(map[string]*ResourceUsage),
			}, ttl, limit, !tpr.resProfiles[*rTid].Stored); err != nil {
			return
		}
		if verbose {
			log.Print("\t", rTid.TenantID())
		}
	}
	if len(tpr.resources) != 0 {
		loadIDs[utils.CacheResources] = loadID
	}
	if verbose {
		log.Print("StatQueueProfiles:")
	}
	for _, tpST := range tpr.sqProfiles {
		var st *StatQueueProfile
		if st, err = APItoStats(tpST, tpr.timezone); err != nil {
			return
		}
		if err = tpr.dm.SetStatQueueProfile(st, true); err != nil {
			return
		}
		if verbose {
			log.Print("\t", st.TenantID())
		}
	}
	if len(tpr.sqProfiles) != 0 {
		loadIDs[utils.CacheStatQueueProfiles] = loadID
	}
	if verbose {
		log.Print("StatQueues:")
	}
	for _, sqTntID := range tpr.statQueues {
		var ttl *time.Duration
		if tpr.sqProfiles[*sqTntID].TTL != utils.EmptyString {
			ttl = new(time.Duration)
			if *ttl, err = utils.ParseDurationWithNanosecs(tpr.sqProfiles[*sqTntID].TTL); err != nil {
				return
			}
			if *ttl <= 0 {
				ttl = nil
			}
		}
		metrics := make([]*MetricWithFilters, len(tpr.sqProfiles[*sqTntID].Metrics))
		for i, metric := range tpr.sqProfiles[*sqTntID].Metrics {
			metrics[i] = &MetricWithFilters{
				MetricID:  metric.MetricID,
				FilterIDs: metric.FilterIDs,
			}
		}
		sq := &StatQueue{
			Tenant: sqTntID.Tenant,
			ID:     sqTntID.ID,
		}
		if !tpr.sqProfiles[*sqTntID].Stored { //for not stored queues create the metrics
			if sq, err = NewStatQueue(sqTntID.Tenant, sqTntID.ID, metrics,
				tpr.sqProfiles[*sqTntID].MinItems); err != nil {
				return
			}
		}
		// for non stored we do not save the metrics
		if err = tpr.dm.SetStatQueue(sq, metrics,
			tpr.sqProfiles[*sqTntID].MinItems,
			ttl, tpr.sqProfiles[*sqTntID].QueueLength,
			!tpr.sqProfiles[*sqTntID].Stored); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", sqTntID.TenantID())
		}
	}
	if len(tpr.statQueues) != 0 {
		loadIDs[utils.CacheStatQueues] = loadID
	}
	if verbose {
		log.Print("ThresholdProfiles:")
	}
	for _, tpTH := range tpr.thProfiles {
		var th *ThresholdProfile
		if th, err = APItoThresholdProfile(tpTH, tpr.timezone); err != nil {
			return
		}
		if err = tpr.dm.SetThresholdProfile(th, true); err != nil {
			return
		}
		if verbose {
			log.Print("\t", th.TenantID())
		}
	}
	if len(tpr.thProfiles) != 0 {
		loadIDs[utils.CacheThresholdProfiles] = loadID
	}
	if verbose {
		log.Print("Thresholds:")
	}
	for _, thd := range tpr.thresholds {
		var minSleep time.Duration
		if tpr.thProfiles[*thd].MinSleep != utils.EmptyString {
			if minSleep, err = utils.ParseDurationWithNanosecs(tpr.thProfiles[*thd].MinSleep); err != nil {
				return
			}
		}
		if err = tpr.dm.SetThreshold(&Threshold{Tenant: thd.Tenant, ID: thd.ID}, minSleep, false); err != nil {
			return
		}
		if verbose {
			log.Print("\t", thd.TenantID())
		}
	}
	if len(tpr.thresholds) != 0 {
		loadIDs[utils.CacheThresholds] = loadID
	}
	if verbose {
		log.Print("RouteProfiles:")
	}
	for _, tpTH := range tpr.routeProfiles {
		var th *RouteProfile
		if th, err = APItoRouteProfile(tpTH, tpr.timezone); err != nil {
			return
		}
		if err = tpr.dm.SetRouteProfile(th, true); err != nil {
			return
		}
		if verbose {
			log.Print("\t", th.TenantID())
		}
	}
	if len(tpr.routeProfiles) != 0 {
		loadIDs[utils.CacheRouteProfiles] = loadID
	}
	if verbose {
		log.Print("AttributeProfiles:")
	}
	for _, tpTH := range tpr.attributeProfiles {
		var th *AttributeProfile
		if th, err = APItoAttributeProfile(tpTH, tpr.timezone); err != nil {
			return
		}
		if err = tpr.dm.SetAttributeProfile(context.TODO(), th, true); err != nil {
			return
		}
		if verbose {
			log.Print("\t", th.TenantID())
		}
	}
	if len(tpr.attributeProfiles) != 0 {
		loadIDs[utils.CacheAttributeProfiles] = loadID
	}
	if verbose {
		log.Print("ChargerProfiles:")
	}
	for _, tpTH := range tpr.chargerProfiles {
		var th *ChargerProfile
		if th, err = APItoChargerProfile(tpTH, tpr.timezone); err != nil {
			return
		}
		if err = tpr.dm.SetChargerProfile(th, true); err != nil {
			return
		}
		if verbose {
			log.Print("\t", th.TenantID())
		}
	}
	if len(tpr.chargerProfiles) != 0 {
		loadIDs[utils.CacheChargerProfiles] = loadID
	}
	if verbose {
		log.Print("DispatcherProfiles:")
	}
	for _, tpTH := range tpr.dispatcherProfiles {
		var th *DispatcherProfile
		if th, err = APItoDispatcherProfile(tpTH, tpr.timezone); err != nil {
			return
		}
		if err = tpr.dm.SetDispatcherProfile(th, true); err != nil {
			return
		}
		if verbose {
			log.Print("\t", th.TenantID())
		}
	}
	if len(tpr.dispatcherProfiles) != 0 {
		loadIDs[utils.CacheDispatcherProfiles] = loadID
	}
	if verbose {
		log.Print("DispatcherHosts:")
	}
	for _, tpTH := range tpr.dispatcherHosts {
		th := APItoDispatcherHost(tpTH)
		if err = tpr.dm.SetDispatcherHost(th); err != nil {
			return
		}
		if verbose {
			log.Print("\t", th.TenantID())
		}
	}
	if len(tpr.dispatcherHosts) != 0 {
		loadIDs[utils.CacheDispatcherHosts] = loadID
	}

	if verbose {
		log.Print("RateProfiles:")
	}
	for _, tpTH := range tpr.rateProfiles {
		var th *utils.RateProfile
		if th, err = APItoRateProfile(tpTH, tpr.timezone); err != nil {
			return
		}
		if err = tpr.dm.SetRateProfile(context.Background(), th, true); err != nil {
			return
		}
		if verbose {
			log.Print("\t", th.TenantID())
		}
	}
	if len(tpr.rateProfiles) != 0 {
		loadIDs[utils.CacheRateProfiles] = loadID
	}

	if verbose {
		log.Print("ActionProfiles:")
	}
	for _, tpAP := range tpr.actionProfiles {
		var ap *ActionProfile
		if ap, err = APItoActionProfile(tpAP, tpr.timezone); err != nil {
			return
		}
		if err = tpr.dm.SetActionProfile(ap, true); err != nil {
			return
		}
		if verbose {
			log.Print("\t", ap.TenantID())
		}
	}
	if len(tpr.actionProfiles) != 0 {
		loadIDs[utils.CacheActionProfiles] = loadID
	}

	if verbose {
		log.Print("Accounts:")
	}
	for _, tpAP := range tpr.accounts {
		var ap *utils.Account
		if ap, err = APItoAccount(tpAP, tpr.timezone); err != nil {
			return
		}
		if err = tpr.dm.SetAccount(ap, true); err != nil {
			return
		}
		if verbose {
			log.Print("\t", ap.TenantID())
		}
	}
	if len(tpr.accounts) != 0 {
		loadIDs[utils.CacheAccounts] = loadID
	}

	if verbose {
		log.Print("Timings:")
	}
	for _, t := range tpr.timings {
		if err = tpr.dm.SetTiming(context.Background(), t); err != nil {
			return
		}
		if verbose {
			log.Print("\t", t.ID)
		}
	}
	if len(tpr.timings) != 0 {
		loadIDs[utils.CacheTimings] = loadID
	}

	return tpr.dm.SetLoadIDs(context.TODO(), loadIDs)
}

func (tpr *TpReader) ShowStatistics() {
	// resource profiles
	log.Print("ResourceProfiles: ", len(tpr.resProfiles))
	// stats
	log.Print("Stats: ", len(tpr.sqProfiles))
	// thresholds
	log.Print("Thresholds: ", len(tpr.thProfiles))
	// filters
	log.Print("Filters: ", len(tpr.filters))
	// Route profiles
	log.Print("RouteProfiles: ", len(tpr.routeProfiles))
	// Attribute profiles
	log.Print("AttributeProfiles: ", len(tpr.attributeProfiles))
	// Charger profiles
	log.Print("ChargerProfiles: ", len(tpr.chargerProfiles))
	// Dispatcher profiles
	log.Print("DispatcherProfiles: ", len(tpr.dispatcherProfiles))
	// Dispatcher Hosts
	log.Print("DispatcherHosts: ", len(tpr.dispatcherHosts))
	// Rate profiles
	log.Print("RateProfiles: ", len(tpr.rateProfiles))
	// Action profiles
	log.Print("ActionProfiles: ", len(tpr.actionProfiles))
}

// GetLoadedIds returns the identities loaded for a specific category, useful for cache reloads
func (tpr *TpReader) GetLoadedIds(categ string) ([]string, error) {
	switch categ {
	case utils.ResourcesPrefix:
		keys := make([]string, len(tpr.resources))
		for i, k := range tpr.resources {
			keys[i] = k.TenantID()
		}
		return keys, nil
	case utils.StatQueuePrefix:
		keys := make([]string, len(tpr.statQueues))
		for i, k := range tpr.statQueues {
			keys[i] = k.TenantID()
		}
		return keys, nil
	case utils.ThresholdPrefix:
		keys := make([]string, len(tpr.thresholds))
		for i, k := range tpr.thresholds {
			keys[i] = k.TenantID()
		}
		return keys, nil
	case utils.TimingsPrefix:
		keys := make([]string, len(tpr.timings))
		i := 0
		for k := range tpr.timings {
			keys[i] = k
			i++
		}
		return keys, nil
	case utils.ResourceProfilesPrefix:
		keys := make([]string, len(tpr.resProfiles))
		i := 0
		for k := range tpr.resProfiles {
			keys[i] = k.TenantID()
			i++
		}
		return keys, nil

	case utils.StatQueueProfilePrefix:
		keys := make([]string, len(tpr.sqProfiles))
		i := 0
		for k := range tpr.sqProfiles {
			keys[i] = k.TenantID()
			i++
		}
		return keys, nil
	case utils.ThresholdProfilePrefix:
		keys := make([]string, len(tpr.thProfiles))
		i := 0
		for k := range tpr.thProfiles {
			keys[i] = k.TenantID()
			i++
		}
		return keys, nil
	case utils.FilterPrefix:
		keys := make([]string, len(tpr.filters))
		i := 0
		for k := range tpr.filters {
			keys[i] = k.TenantID()
			i++
		}
		return keys, nil
	case utils.RouteProfilePrefix:
		keys := make([]string, len(tpr.routeProfiles))
		i := 0
		for k := range tpr.routeProfiles {
			keys[i] = k.TenantID()
			i++
		}
		return keys, nil
	case utils.AttributeProfilePrefix:
		keys := make([]string, len(tpr.attributeProfiles))
		i := 0
		for k := range tpr.attributeProfiles {
			keys[i] = k.TenantID()
			i++
		}
		return keys, nil
	case utils.ChargerProfilePrefix:
		keys := make([]string, len(tpr.chargerProfiles))
		i := 0
		for k := range tpr.chargerProfiles {
			keys[i] = k.TenantID()
			i++
		}
		return keys, nil
	case utils.DispatcherProfilePrefix:
		keys := make([]string, len(tpr.dispatcherProfiles))
		i := 0
		for k := range tpr.dispatcherProfiles {
			keys[i] = k.TenantID()
			i++
		}
		return keys, nil

	case utils.DispatcherHostPrefix:
		keys := make([]string, len(tpr.dispatcherHosts))
		i := 0
		for k := range tpr.dispatcherHosts {
			keys[i] = k.TenantID()
			i++
		}
		return keys, nil

	case utils.RateProfilePrefix:
		keys := make([]string, len(tpr.rateProfiles))
		i := 0
		for k := range tpr.rateProfiles {
			keys[i] = k.TenantID()
			i++
		}
		return keys, nil
	case utils.ActionProfilePrefix:
		keys := make([]string, len(tpr.actionProfiles))
		i := 0
		for k := range tpr.actionProfiles {
			keys[i] = k.TenantID()
			i++
		}
		return keys, nil
	}
	return nil, errors.New("Unsupported load category")
}

func (tpr *TpReader) RemoveFromDatabase(verbose, disableReverse bool) (err error) {
	loadID := time.Now().UnixNano()
	loadIDs := make(map[string]int64)
	if verbose {
		log.Print("ResourceProfiles:")
	}
	for _, tpRsp := range tpr.resProfiles {
		if err = tpr.dm.RemoveResourceProfile(tpRsp.Tenant, tpRsp.ID, utils.NonTransactional, true); err != nil {
			return
		}
		if verbose {
			log.Print("\t", utils.ConcatenatedKey(tpRsp.Tenant, tpRsp.ID))
		}
	}
	if verbose {
		log.Print("Resources:")
	}
	for _, rTid := range tpr.resources {
		if err = tpr.dm.RemoveResource(rTid.Tenant, rTid.ID, utils.NonTransactional); err != nil {
			return
		}
		if verbose {
			log.Print("\t", rTid.TenantID())
		}
	}
	if verbose {
		log.Print("StatQueueProfiles:")
	}
	for _, tpST := range tpr.sqProfiles {
		if err = tpr.dm.RemoveStatQueueProfile(tpST.Tenant, tpST.ID, utils.NonTransactional, true); err != nil {
			return
		}
		if verbose {
			log.Print("\t", utils.ConcatenatedKey(tpST.Tenant, tpST.ID))
		}
	}
	if verbose {
		log.Print("StatQueues:")
	}
	for _, sqTntID := range tpr.statQueues {
		if err = tpr.dm.RemoveStatQueue(sqTntID.Tenant, sqTntID.ID, utils.NonTransactional); err != nil {
			return
		}
		if verbose {
			log.Print("\t", sqTntID.TenantID())
		}
	}
	if verbose {
		log.Print("ThresholdProfiles:")
	}
	for _, tpTH := range tpr.thProfiles {
		if err = tpr.dm.RemoveThresholdProfile(tpTH.Tenant, tpTH.ID, utils.NonTransactional, true); err != nil {
			return
		}
		if verbose {
			log.Print("\t", utils.ConcatenatedKey(tpTH.Tenant, tpTH.ID))
		}
	}
	if verbose {
		log.Print("Thresholds:")
	}
	for _, thd := range tpr.thresholds {
		if err = tpr.dm.RemoveThreshold(thd.Tenant, thd.ID, utils.NonTransactional); err != nil {
			return
		}
		if verbose {
			log.Print("\t", thd.TenantID())
		}
	}

	if verbose {
		log.Print("RouteProfiles:")
	}
	for _, tpSpl := range tpr.routeProfiles {
		if err = tpr.dm.RemoveRouteProfile(tpSpl.Tenant, tpSpl.ID, utils.NonTransactional, true); err != nil {
			return
		}
		if verbose {
			log.Print("\t", utils.ConcatenatedKey(tpSpl.Tenant, tpSpl.ID))
		}
	}

	if verbose {
		log.Print("AttributeProfiles:")
	}
	for _, tpAttr := range tpr.attributeProfiles {
		if err = tpr.dm.RemoveAttributeProfile(context.TODO(), tpAttr.Tenant, tpAttr.ID,
			utils.NonTransactional, true); err != nil {
			return
		}
		if verbose {
			log.Print("\t", utils.ConcatenatedKey(tpAttr.Tenant, tpAttr.ID))
		}
	}

	if verbose {
		log.Print("ChargerProfiles:")
	}
	for _, tpChr := range tpr.chargerProfiles {
		if err = tpr.dm.RemoveChargerProfile(tpChr.Tenant, tpChr.ID,
			utils.NonTransactional, true); err != nil {
			return
		}
		if verbose {
			log.Print("\t", utils.ConcatenatedKey(tpChr.Tenant, tpChr.ID))
		}
	}

	if verbose {
		log.Print("DispatcherProfiles:")
	}
	for _, tpDsp := range tpr.dispatcherProfiles {
		if err = tpr.dm.RemoveDispatcherProfile(tpDsp.Tenant, tpDsp.ID,
			utils.NonTransactional, true); err != nil {
			return
		}
		if verbose {
			log.Print("\t", utils.ConcatenatedKey(tpDsp.Tenant, tpDsp.ID))
		}
	}
	if verbose {
		log.Print("DispatcherHosts:")
	}
	for _, tpDsh := range tpr.dispatcherHosts {
		if err = tpr.dm.RemoveDispatcherHost(tpDsh.Tenant, tpDsh.ID,
			utils.NonTransactional); err != nil {
			return
		}
		if verbose {
			log.Print("\t", utils.ConcatenatedKey(tpDsh.Tenant, tpDsh.ID))
		}
	}

	if verbose {
		log.Print("RateProfiles:")
	}
	for _, tpRp := range tpr.rateProfiles {
		if err = tpr.dm.RemoveRateProfile(context.TODO(), tpRp.Tenant, tpRp.ID,
			utils.NonTransactional, true); err != nil {
			return
		}
		if verbose {
			log.Print("\t", utils.ConcatenatedKey(tpRp.Tenant, tpRp.ID))
		}
	}

	if verbose {
		log.Print("ActionProfiles:")
	}
	for _, tpAp := range tpr.actionProfiles {
		if err = tpr.dm.RemoveActionProfile(tpAp.Tenant, tpAp.ID,
			utils.NonTransactional, true); err != nil {
			return
		}
		if verbose {
			log.Print("\t", utils.ConcatenatedKey(tpAp.Tenant, tpAp.ID))
		}
	}

	if verbose {
		log.Print("Accounts:")
	}
	for _, tpAp := range tpr.accounts {
		if err = tpr.dm.RemoveAccount(tpAp.Tenant, tpAp.ID,
			utils.NonTransactional, true); err != nil {
			return
		}
		if verbose {
			log.Print("\t", utils.ConcatenatedKey(tpAp.Tenant, tpAp.ID))
		}
	}

	if verbose {
		log.Print("Timings:")
	}
	for _, t := range tpr.timings {
		if err = tpr.dm.RemoveTiming(t.ID, utils.NonTransactional); err != nil {
			return
		}
		if verbose {
			log.Print("\t", t.ID)
		}
	}
	//We remove the filters at the end because of indexes
	if verbose {
		log.Print("Filters:")
	}
	for _, tpFltr := range tpr.filters {
		if err = tpr.dm.RemoveFilter(context.TODO(), tpFltr.Tenant, tpFltr.ID,
			utils.NonTransactional, true); err != nil {
			return
		}
		if verbose {
			log.Print("\t", utils.ConcatenatedKey(tpFltr.Tenant, tpFltr.ID))
		}
	}
	if len(tpr.filters) != 0 {
		loadIDs[utils.CacheFilters] = loadID
	}
	if len(tpr.resProfiles) != 0 {
		loadIDs[utils.CacheResourceProfiles] = loadID
	}
	if len(tpr.resources) != 0 {
		loadIDs[utils.CacheResources] = loadID
	}
	if len(tpr.sqProfiles) != 0 {
		loadIDs[utils.CacheStatQueueProfiles] = loadID
	}
	if len(tpr.statQueues) != 0 {
		loadIDs[utils.CacheStatQueues] = loadID
	}
	if len(tpr.thProfiles) != 0 {
		loadIDs[utils.CacheThresholdProfiles] = loadID
	}
	if len(tpr.thresholds) != 0 {
		loadIDs[utils.CacheThresholds] = loadID
	}
	if len(tpr.routeProfiles) != 0 {
		loadIDs[utils.CacheRouteProfiles] = loadID
	}
	if len(tpr.attributeProfiles) != 0 {
		loadIDs[utils.CacheAttributeProfiles] = loadID
	}
	if len(tpr.chargerProfiles) != 0 {
		loadIDs[utils.CacheChargerProfiles] = loadID
	}
	if len(tpr.dispatcherProfiles) != 0 {
		loadIDs[utils.CacheDispatcherProfiles] = loadID
	}
	if len(tpr.dispatcherHosts) != 0 {
		loadIDs[utils.CacheDispatcherHosts] = loadID
	}
	if len(tpr.rateProfiles) != 0 {
		loadIDs[utils.CacheRateProfiles] = loadID
	}
	if len(tpr.actionProfiles) != 0 {
		loadIDs[utils.CacheActionProfiles] = loadID
	}
	if len(tpr.accounts) != 0 {
		loadIDs[utils.CacheAccounts] = loadID
	}
	if len(tpr.timings) != 0 {
		loadIDs[utils.CacheTimings] = loadID
	}
	return tpr.dm.SetLoadIDs(context.TODO(), loadIDs)
}

func (tpr *TpReader) ReloadCache(ctx *context.Context, caching string, verbose bool, opts map[string]interface{}) (err error) {
	if tpr.isInternalDB {
		return
	}
	if len(tpr.cacheConns) == 0 {
		log.Print("Disabled automatic reload")
		return
	}
	// take IDs for each type
	tmgIds, _ := tpr.GetLoadedIds(utils.TimingsPrefix)
	rspIDs, _ := tpr.GetLoadedIds(utils.ResourceProfilesPrefix)
	resIDs, _ := tpr.GetLoadedIds(utils.ResourcesPrefix)
	stqIDs, _ := tpr.GetLoadedIds(utils.StatQueuePrefix)
	stqpIDs, _ := tpr.GetLoadedIds(utils.StatQueueProfilePrefix)
	trsIDs, _ := tpr.GetLoadedIds(utils.ThresholdPrefix)
	trspfIDs, _ := tpr.GetLoadedIds(utils.ThresholdProfilePrefix)
	flrIDs, _ := tpr.GetLoadedIds(utils.FilterPrefix)
	routeIDs, _ := tpr.GetLoadedIds(utils.RouteProfilePrefix)
	apfIDs, _ := tpr.GetLoadedIds(utils.AttributeProfilePrefix)
	chargerIDs, _ := tpr.GetLoadedIds(utils.ChargerProfilePrefix)
	dppIDs, _ := tpr.GetLoadedIds(utils.DispatcherProfilePrefix)
	dphIDs, _ := tpr.GetLoadedIds(utils.DispatcherHostPrefix)
	ratePrfIDs, _ := tpr.GetLoadedIds(utils.RateProfilePrefix)
	actionPrfIDs, _ := tpr.GetLoadedIds(utils.ActionProfilePrefix)
	accountPrfIDs, _ := tpr.GetLoadedIds(utils.AccountPrefix)

	//compose Reload Cache argument
	cacheArgs := map[string][]string{
		utils.TimingIDs:            tmgIds,
		utils.ResourceProfileIDs:   rspIDs,
		utils.ResourceIDs:          resIDs,
		utils.StatsQueueIDs:        stqIDs,
		utils.StatsQueueProfileIDs: stqpIDs,
		utils.ThresholdIDs:         trsIDs,
		utils.ThresholdProfileIDs:  trspfIDs,
		utils.FilterIDs:            flrIDs,
		utils.RouteProfileIDs:      routeIDs,
		utils.AttributeProfileIDs:  apfIDs,
		utils.ChargerProfileIDs:    chargerIDs,
		utils.DispatcherProfileIDs: dppIDs,
		utils.DispatcherHostIDs:    dphIDs,
		utils.RateProfileIDs:       ratePrfIDs,
		utils.ActionProfileIDs:     actionPrfIDs,
	}

	// verify if we need to clear indexes
	var cacheIDs []string
	if len(apfIDs) != 0 {
		cacheIDs = append(cacheIDs, utils.CacheAttributeFilterIndexes)
	}
	if len(routeIDs) != 0 {
		cacheIDs = append(cacheIDs, utils.CacheRouteFilterIndexes)
	}
	if len(trspfIDs) != 0 {
		cacheIDs = append(cacheIDs, utils.CacheThresholdFilterIndexes)
	}
	if len(stqpIDs) != 0 {
		cacheIDs = append(cacheIDs, utils.CacheStatFilterIndexes)
	}
	if len(rspIDs) != 0 {
		cacheIDs = append(cacheIDs, utils.CacheResourceFilterIndexes)
	}
	if len(chargerIDs) != 0 {
		cacheIDs = append(cacheIDs, utils.CacheChargerFilterIndexes)
	}
	if len(dppIDs) != 0 {
		cacheIDs = append(cacheIDs, utils.CacheDispatcherFilterIndexes)
	}
	if len(ratePrfIDs) != 0 {
		cacheIDs = append(cacheIDs, utils.CacheRateProfilesFilterIndexes)
		cacheIDs = append(cacheIDs, utils.CacheRateFilterIndexes)
	}
	if len(actionPrfIDs) != 0 {
		cacheIDs = append(cacheIDs, utils.CacheActionProfilesFilterIndexes)
	}
	if len(accountPrfIDs) != 0 {
		cacheIDs = append(cacheIDs, utils.CacheAccountsFilterIndexes)
	}
	if len(flrIDs) != 0 {
		cacheIDs = append(cacheIDs, utils.CacheReverseFilterIndexes)
	}

	if err = CallCache(connMgr, ctx, tpr.cacheConns, caching, cacheArgs, cacheIDs, opts, verbose); err != nil {
		return
	}
	//get loadIDs for all types
	var loadIDs map[string]int64
	if loadIDs, err = tpr.dm.GetItemLoadIDs(ctx, utils.EmptyString, false); err != nil {
		return
	}
	cacheLoadIDs := populateCacheLoadIDs(loadIDs, cacheArgs)
	for key, val := range cacheLoadIDs {
		if err = Cache.Set(ctx, utils.CacheLoadIDs, key, val, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional); err != nil {
			return
		}
	}
	return
}

// CallCache call the cache reload after data load
func CallCache(connMgr *ConnManager, ctx *context.Context, cacheConns []string, caching string, args map[string][]string, cacheIDs []string, opts map[string]interface{}, verbose bool) (err error) {
	for k, v := range args {
		if len(v) == 0 {
			delete(args, k)
		}
	}
	var method, reply string
	var cacheArgs interface{} = utils.AttrReloadCacheWithAPIOpts{
		APIOpts:   opts,
		ArgsCache: args,
	}
	switch caching {
	case utils.MetaNone:
		return
	case utils.MetaReload:
		method = utils.CacheSv1ReloadCache
	case utils.MetaLoad:
		method = utils.CacheSv1LoadCache
	case utils.MetaRemove:
		method = utils.CacheSv1RemoveItems
	case utils.MetaClear:
		method = utils.CacheSv1Clear
		cacheArgs = &utils.AttrCacheIDsWithAPIOpts{APIOpts: opts}
	}
	if verbose {
		log.Print("Reloading cache")
	}

	if err = connMgr.Call(ctx, cacheConns, method, cacheArgs, &reply); err != nil {
		return
	}

	if len(cacheIDs) != 0 {
		if verbose {
			log.Print("Clearing indexes")
		}
		if err = connMgr.Call(ctx, cacheConns, utils.CacheSv1Clear, &utils.AttrCacheIDsWithAPIOpts{
			APIOpts:  opts,
			CacheIDs: cacheIDs,
		}, &reply); err != nil {
			if verbose {
				log.Printf("WARNING: Got error on cache clear: %s\n", err.Error())
			}
		}
	}
	return
}

func (tpr *TpReader) ReloadScheduler(verbose bool) (err error) { // ToDoNext: add reload to new actions
	// var reply string
	// aps, _ := tpr.GetLoadedIds(utils.ActionPlanPrefix)
	// // in case we have action plans reload the scheduler
	// if len(aps) == 0 {
	// 	return
	// }
	// if verbose {
	// 	log.Print("Reloading scheduler")
	// }
	// if err = connMgr.Call(tpr.schedulerConns, nil, utils.SchedulerSv1Reload,
	// 	new(utils.CGREvent), &reply); err != nil {
	// 	log.Printf("WARNING: Got error on scheduler reload: %s\n", err.Error())
	// }
	return
}

func (tpr *TpReader) addDefaultTimings() {
	tpr.timings[utils.MetaAny] = &utils.TPTiming{
		ID:        utils.MetaAny,
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{},
		StartTime: "00:00:00",
		EndTime:   "",
	}
	tpr.timings[utils.MetaASAP] = &utils.TPTiming{
		ID:        utils.MetaASAP,
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{},
		StartTime: utils.MetaASAP,
		EndTime:   "",
	}
	tpr.timings[utils.MetaEveryMinute] = &utils.TPTiming{
		ID:        utils.MetaEveryMinute,
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{},
		StartTime: utils.ConcatenatedKey(utils.Meta, utils.Meta, strconv.Itoa(time.Now().Second())),
		EndTime:   "",
	}
	tpr.timings[utils.MetaHourly] = &utils.TPTiming{
		ID:        utils.MetaHourly,
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{},
		StartTime: utils.ConcatenatedKey(utils.Meta, strconv.Itoa(time.Now().Minute()), strconv.Itoa(time.Now().Second())),
		EndTime:   "",
	}
	startTime := time.Now().Format("15:04:05")
	tpr.timings[utils.MetaDaily] = &utils.TPTiming{
		ID:        utils.MetaDaily,
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{},
		StartTime: startTime,
		EndTime:   "",
	}
	tpr.timings[utils.MetaWeekly] = &utils.TPTiming{
		ID:        utils.MetaWeekly,
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{time.Now().Weekday()},
		StartTime: startTime,
		EndTime:   "",
	}
	tpr.timings[utils.MetaMonthly] = &utils.TPTiming{
		ID:        utils.MetaMonthly,
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{time.Now().Day()},
		WeekDays:  utils.WeekDays{},
		StartTime: startTime,
		EndTime:   "",
	}
	tpr.timings[utils.MetaMonthlyEstimated] = &utils.TPTiming{
		ID:        utils.MetaMonthlyEstimated,
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{time.Now().Day()},
		WeekDays:  utils.WeekDays{},
		StartTime: startTime,
		EndTime:   "",
	}
	tpr.timings[utils.MetaMonthEnd] = &utils.TPTiming{
		ID:        utils.MetaMonthEnd,
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{-1},
		WeekDays:  utils.WeekDays{},
		StartTime: startTime,
		EndTime:   "",
	}
	tpr.timings[utils.MetaYearly] = &utils.TPTiming{
		ID:        utils.MetaYearly,
		Years:     utils.Years{},
		Months:    utils.Months{time.Now().Month()},
		MonthDays: utils.MonthDays{time.Now().Day()},
		WeekDays:  utils.WeekDays{},
		StartTime: startTime,
		EndTime:   "",
	}

}
