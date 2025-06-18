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
	"fmt"
	"runtime"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

// IPProfile defines the configuration of the IP.
type IPProfile struct {
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval
	TTL                time.Duration
	Type               string
	AddressPool        string
	Allocation         string
	Stored             bool
	Weight             float64

	lkID string
}

// Clone creates a deep copy of IPProfile for thread-safe use.
func (ip *IPProfile) Clone() *IPProfile {
	if ip == nil {
		return nil
	}
	return &IPProfile{
		Tenant:             ip.Tenant,
		ID:                 ip.ID,
		FilterIDs:          slices.Clone(ip.FilterIDs),
		ActivationInterval: ip.ActivationInterval.Clone(),
		TTL:                ip.TTL,
		Type:               ip.Type,
		AddressPool:        ip.AddressPool,
		Allocation:         ip.Allocation,
		Stored:             ip.Stored,
		Weight:             ip.Weight,
	}
}

// CacheClone returns a clone of IPProfile used by ltcache CacheCloner.
func (ip *IPProfile) CacheClone() any {
	return ip.Clone()
}

// IPProfileWithAPIOpts wraps IPProfile with APIOpts.
type IPProfileWithAPIOpts struct {
	*IPProfile
	APIOpts map[string]any
}

// TenantID returns the concatenated tenant and ID.
func (ip *IPProfile) TenantID() string {
	return utils.ConcatenatedKey(ip.Tenant, ip.ID)
}

// ipProfileLockKey returns the ID used to lock an IPProfile with guardian.
func ipProfileLockKey(tnt, id string) string {
	return utils.ConcatenatedKey(utils.CacheIPProfiles, tnt, id)
}

// lock will lock the IPProfile using guardian and store the lock within lkID.
// If lkID is provided as an argument, it assumes the lock is already acquired.
func (ip *IPProfile) lock(lkID string) {
	if lkID == utils.EmptyString {
		lkID = guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			ipProfileLockKey(ip.Tenant, ip.ID))
	}
	ip.lkID = lkID
}

// unlock releases the lock on the IPProfile and clears the lock ID.
func (ip *IPProfile) unlock() {
	if ip.lkID == utils.EmptyString {
		return
	}
	tmp := ip.lkID
	ip.lkID = utils.EmptyString
	guardian.Guardian.UnguardIDs(tmp)
}

// IPUsage represents an usage counted.
type IPUsage struct {
	Tenant     string
	ID         string
	ExpiryTime time.Time
	Units      float64
}

// TenantID returns the concatenated key between tenant and ID
func (u *IPUsage) TenantID() string {
	return utils.ConcatenatedKey(u.Tenant, u.ID)
}

// isActive checks ExpiryTime at some time
func (u *IPUsage) isActive(atTime time.Time) bool {
	return u.ExpiryTime.IsZero() || u.ExpiryTime.Sub(atTime) > 0
}

// Clone duplicates ru
func (u *IPUsage) Clone() *IPUsage {
	if u == nil {
		return nil
	}
	clone := *u
	return &clone
}

// IP represents ...
type IP struct {
	Tenant string
	ID     string
	Usages map[string]*IPUsage
	TTLIdx []string
	lkID   string
	ttl    *time.Duration
	tUsage *float64
	dirty  *bool
	cfg    *IPProfile
}

// Clone clones *IP (lkID excluded)
func (ip *IP) Clone() *IP {
	if ip == nil {
		return nil
	}
	clone := &IP{
		Tenant: ip.Tenant,
		ID:     ip.ID,
		TTLIdx: slices.Clone(ip.TTLIdx),
		cfg:    ip.cfg.Clone(),
	}
	if ip.Usages != nil {
		clone.Usages = make(map[string]*IPUsage, len(ip.Usages))
		for key, usage := range ip.Usages {
			clone.Usages[key] = usage.Clone()
		}
	}
	if ip.ttl != nil {
		ttlCopy := *ip.ttl
		clone.ttl = &ttlCopy
	}
	if ip.tUsage != nil {
		tUsageCopy := *ip.tUsage
		clone.tUsage = &tUsageCopy
	}
	if ip.dirty != nil {
		dirtyCopy := *ip.dirty
		clone.dirty = &dirtyCopy
	}
	return clone
}

// CacheClone returns a clone of IP used by ltcache CacheCloner
func (ip *IP) CacheClone() any {
	return ip.Clone()
}

// ipLockKey returns the ID used to lock a ip with guardian
func ipLockKey(tnt, id string) string {
	return utils.ConcatenatedKey(utils.CacheIPs, tnt, id)
}

// lock will lock the ip using guardian and store the lock within r.lkID
// if lkID is passed as argument, the lock is considered as executed
func (ip *IP) lock(lkID string) {
	if lkID == utils.EmptyString {
		lkID = guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			ipLockKey(ip.Tenant, ip.ID))
	}
	ip.lkID = lkID
}

// unlock will unlock the ip and clear r.lkID
func (ip *IP) unlock() {
	if ip.lkID == utils.EmptyString {
		return
	}
	tmp := ip.lkID
	ip.lkID = utils.EmptyString
	guardian.Guardian.UnguardIDs(tmp)

}

type IPWithAPIOpts struct {
	*IP
	APIOpts map[string]any
}

// TenantID returns the unique ID in a multi-tenant environment
func (ip *IP) TenantID() string {
	return utils.ConcatenatedKey(ip.Tenant, ip.ID)
}

// removeExpiredUnits removes units which are expired from the ip
func (ip *IP) removeExpiredUnits() {
	var firstActive int
	for _, rID := range ip.TTLIdx {
		if r, has := ip.Usages[rID]; has && r.isActive(time.Now()) {
			break
		}
		firstActive++
	}
	if firstActive == 0 {
		return
	}
	for _, rID := range ip.TTLIdx[:firstActive] {
		ru, has := ip.Usages[rID]
		if !has {
			continue
		}
		delete(ip.Usages, rID)
		if ip.tUsage != nil { //  total usage was not yet calculated so we do not need to update it
			*ip.tUsage -= ru.Units
			if *ip.tUsage < 0 { // something went wrong
				utils.Logger.Warning(
					fmt.Sprintf("resetting total usage for ipID: %s, usage smaller than 0: %f", ip.ID, *ip.tUsage))
				ip.tUsage = nil
			}
		}
	}
	ip.TTLIdx = ip.TTLIdx[firstActive:]
	ip.tUsage = nil
}

// TotalUsage returns the sum of all usage units.
func (ip *IP) TotalUsage() float64 {
	if ip.tUsage == nil {
		var tu float64
		for _, ru := range ip.Usages {
			tu += ru.Units
		}
		ip.tUsage = &tu
	}
	if ip.tUsage == nil {
		return 0
	}
	return *ip.tUsage
}

// recordUsage records a new usage
func (ip *IP) recordUsage(ru *IPUsage) (err error) {
	if _, hasID := ip.Usages[ru.ID]; hasID {
		return fmt.Errorf("duplicate ip usage with id: %s", ru.TenantID())
	}
	if ip.ttl != nil && *ip.ttl != -1 {
		if *ip.ttl == 0 {
			return // no recording for ttl of 0
		}
		ru = ru.Clone() // don't influence the initial ru
		ru.ExpiryTime = time.Now().Add(*ip.ttl)
	}
	ip.Usages[ru.ID] = ru
	if ip.tUsage != nil {
		*ip.tUsage += ru.Units
	}
	if !ru.ExpiryTime.IsZero() {
		ip.TTLIdx = append(ip.TTLIdx, ru.ID)
	}
	return
}

// clearUsage clears the usage for an ID
func (ip *IP) clearUsage(ruID string) (err error) {
	ru, hasIt := ip.Usages[ruID]
	if !hasIt {
		return fmt.Errorf("cannot find usage record with id: %s", ruID)
	}
	if !ru.ExpiryTime.IsZero() {
		for i, ruIDIdx := range ip.TTLIdx {
			if ruIDIdx == ruID {
				ip.TTLIdx = slices.Delete(ip.TTLIdx, i, i+1)
				break
			}
		}
	}
	if ip.tUsage != nil {
		*ip.tUsage -= ru.Units
	}
	delete(ip.Usages, ruID)
	return
}

// IPs is an orderable list of IPs based on Weight
type IPs []*IP

// Sort sorts based on Weight
func (ips IPs) Sort() {
	sort.Slice(ips, func(i, j int) bool { return ips[i].cfg.Weight > ips[j].cfg.Weight })
}

// unlock will unlock ips part of this slice
func (ips IPs) unlock() {
	for _, ip := range ips {
		ip.unlock()
		if ip.cfg != nil {
			ip.cfg.unlock()
		}
	}
}

// ids returns a map of ip IDs which is used for caching
func (ips IPs) ids() utils.StringSet {
	mp := make(utils.StringSet, len(ips))
	for _, ip := range ips {
		mp.Add(ip.ID)
	}
	return mp
}

// NewIPService  returns a new IPService
func NewIPService(dm *DataManager, cgrcfg *config.CGRConfig,
	filterS *FilterS, connMgr *ConnManager) *IPService {
	return &IPService{dm: dm,
		storedIPs:   make(utils.StringSet),
		cfg:         cgrcfg,
		fs:          filterS,
		loopStopped: make(chan struct{}),
		stopBackup:  make(chan struct{}),
		cm:          connMgr,
	}

}

// IPService is the service handling ips
type IPService struct {
	cfg          *config.CGRConfig
	cm           *ConnManager
	dm           *DataManager
	fs           *FilterS
	storedIPsMux sync.RWMutex    // protects storedIPs
	storedIPs    utils.StringSet // keep a record of ips which need saving, map[ipID]bool
	stopBackup   chan struct{}   // control storing process
	loopStopped  chan struct{}
}

// Reload stops the backupLoop and restarts it
func (rS *IPService) Reload() {
	close(rS.stopBackup)
	<-rS.loopStopped // wait until the loop is done
	rS.stopBackup = make(chan struct{})
	go rS.runBackup()
}

// StartLoop starts the gorutine with the backup loop
func (rS *IPService) StartLoop() {
	go rS.runBackup()
}

// Shutdown is called to shutdown the service
func (rS *IPService) Shutdown() {
	utils.Logger.Info("<IPs> service shutdown initialized")
	close(rS.stopBackup)
	rS.storeIPs()
	utils.Logger.Info("<IPs> service shutdown complete")
}

// backup will regularly store ips changed to dataDB
func (rS *IPService) runBackup() {
	storeInterval := rS.cfg.IPsCfg().StoreInterval
	if storeInterval <= 0 {
		rS.loopStopped <- struct{}{}
		return
	}
	for {
		rS.storeIPs()
		select {
		case <-rS.stopBackup:
			rS.loopStopped <- struct{}{}
			return
		case <-time.After(storeInterval):
		}
	}
}

// storeIPs represents one task of complete backup
func (rS *IPService) storeIPs() {
	var failedRIDs []string
	for { // don't stop until we store all dirty ips
		rS.storedIPsMux.Lock()
		rID := rS.storedIPs.GetOne()
		if rID != "" {
			rS.storedIPs.Remove(rID)
		}
		rS.storedIPsMux.Unlock()
		if rID == "" {
			break // no more keys, backup completed
		}
		rIf, ok := Cache.Get(utils.CacheIPs, rID)
		if !ok || rIf == nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> failed retrieving from cache ip with ID: %s", utils.IPs, rID))
			continue
		}
		r := rIf.(*IP)
		r.lock(utils.EmptyString)
		if err := rS.storeIP(r); err != nil {
			failedRIDs = append(failedRIDs, rID) // record failure so we can schedule it for next backup
		}
		r.unlock()
		// randomize the CPU load and give up thread control
		runtime.Gosched()
	}
	if len(failedRIDs) != 0 { // there were errors on save, schedule the keys for next backup
		rS.storedIPsMux.Lock()
		rS.storedIPs.AddSlice(failedRIDs)
		rS.storedIPsMux.Unlock()
	}
}

// StoreIP stores the ip in DB and corrects dirty flag
func (rS *IPService) storeIP(r *IP) (err error) {
	if r.dirty == nil || !*r.dirty {
		return
	}
	if err = rS.dm.SetIP(r); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<IPs> failed saving IP with ID: %s, error: %s",
				r.ID, err.Error()))
		return
	}
	//since we no longer handle cache in DataManager do here a manual caching
	if tntID := r.TenantID(); Cache.HasItem(utils.CacheIPs, tntID) { // only cache if previously there
		if err = Cache.Set(utils.CacheIPs, tntID, r, nil,
			true, utils.NonTransactional); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<IPs> failed caching IP with ID: %s, error: %s",
					tntID, err.Error()))
			return
		}
	}
	*r.dirty = false
	return
}

// storeMatchedIPs will store the list of ips based on the StoreInterval
func (s *IPService) storeMatchedIPs(matchedIPs IPs) (err error) {
	if s.cfg.IPsCfg().StoreInterval == 0 {
		return
	}
	if s.cfg.IPsCfg().StoreInterval > 0 {
		s.storedIPsMux.Lock()
		defer s.storedIPsMux.Unlock()
	}
	for _, ip := range matchedIPs {
		if ip.dirty != nil {
			*ip.dirty = true // mark it to be saved
			if s.cfg.IPsCfg().StoreInterval > 0 {
				s.storedIPs.Add(ip.TenantID())
				continue
			}
			if err = s.storeIP(ip); err != nil {
				return
			}
		}

	}
	return
}

// matchingIPsForEvent returns ordered list of matching ips which are active by the time of the call
func (s *IPService) matchingIPsForEvent(tnt string, ev *utils.CGREvent,
	evUUID string, usageTTL *time.Duration) (ips IPs, err error) {
	var ipIDs utils.StringSet
	evNm := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
	}
	if x, ok := Cache.Get(utils.CacheEventIPs, evUUID); ok { // The IPIDs were cached as utils.StringSet{"ipID":bool}
		if x == nil {
			return nil, utils.ErrNotFound
		}
		ipIDs = x.(utils.StringSet)
		defer func() { // make sure we uncache if we find errors
			if err != nil {
				if errCh := Cache.Remove(utils.CacheEventIPs, evUUID,
					cacheCommit(utils.NonTransactional), utils.NonTransactional); errCh != nil {
					err = errCh
				}
			}
		}()

	} else { // select the ipIDs out of dataDB
		ipIDs, err = MatchingItemIDsForEvent(evNm,
			s.cfg.IPsCfg().StringIndexedFields,
			s.cfg.IPsCfg().PrefixIndexedFields,
			s.cfg.IPsCfg().SuffixIndexedFields,
			s.cfg.IPsCfg().ExistsIndexedFields,
			s.dm, utils.CacheIPFilterIndexes, tnt,
			s.cfg.IPsCfg().IndexedSelects,
			s.cfg.IPsCfg().NestedFields,
		)
		if err != nil {
			if err == utils.ErrNotFound {
				if errCh := Cache.Set(utils.CacheEventIPs, evUUID, nil, nil, true, ""); errCh != nil { // cache negative match
					return nil, errCh
				}
			}
			return
		}
	}
	ips = make(IPs, 0, len(ipIDs))
	for id := range ipIDs {
		lkPrflID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			ipProfileLockKey(tnt, id))
		var profile *IPProfile
		if profile, err = s.dm.GetIPProfile(tnt, id,
			true, true, utils.NonTransactional); err != nil {
			guardian.Guardian.UnguardIDs(lkPrflID)
			if err == utils.ErrNotFound {
				continue
			}
			ips.unlock()
			return
		}
		profile.lock(lkPrflID)
		if profile.ActivationInterval != nil && ev.Time != nil &&
			!profile.ActivationInterval.IsActiveAtTime(*ev.Time) { // not active
			profile.unlock()
			continue
		}
		var pass bool
		if pass, err = s.fs.Pass(tnt, profile.FilterIDs,
			evNm); err != nil {
			profile.unlock()
			ips.unlock()
			return nil, err
		} else if !pass {
			profile.unlock()
			continue
		}
		lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
			config.CgrConfig().GeneralCfg().LockingTimeout,
			ipLockKey(profile.Tenant, profile.ID))
		var ip *IP
		if ip, err = s.dm.GetIP(profile.Tenant, profile.ID, true, true, ""); err != nil {
			guardian.Guardian.UnguardIDs(lkID)
			profile.unlock()
			ips.unlock()
			return nil, err
		}
		ip.lock(lkID) // pass the lock into ip so we have it as reference
		if profile.Stored && ip.dirty == nil {
			ip.dirty = utils.BoolPointer(false)
		}
		if usageTTL != nil {
			if *usageTTL != 0 {
				ip.ttl = usageTTL
			}
		} else if profile.TTL >= 0 {
			ip.ttl = utils.DurationPointer(profile.TTL)
		}
		ip.cfg = profile
		ips = append(ips, ip)
	}

	if len(ips) == 0 {
		return nil, utils.ErrNotFound
	}
	ips.Sort()
	if err = Cache.Set(utils.CacheEventIPs, evUUID, ips.ids(), nil, true, ""); err != nil {
		ips.unlock()
	}
	return
}

// V1GetIPsForEvent returns active ip configs matching the event
func (s *IPService) V1GetIPsForEvent(ctx *context.Context, args *utils.CGREvent, reply *IPs) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	usageID := utils.GetStringOpts(args, s.cfg.IPsCfg().Opts.UsageID, utils.OptsIPsUsageID)
	if usageID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.UsageID)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.IPsV1GetIPsForEvent, utils.ConcatenatedKey(tnt, args.ID))
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)
		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*IPs)
			}
			return cachedResp.Error
		}
		defer Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	var ttl *time.Duration
	if ttl, err = utils.GetDurationPointerOpts(args, s.cfg.IPsCfg().Opts.TTL,
		utils.OptsIPsTTL); err != nil {
		return
	}
	var ips IPs
	if ips, err = s.matchingIPsForEvent(tnt, args, usageID, ttl); err != nil {
		return err
	}
	defer ips.unlock()
	*reply = ips
	return
}

// V1AuthorizeIPs queries service to find if an Usage is allowed.
func (s *IPService) V1AuthorizeIPs(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	usageID := utils.GetStringOpts(args, s.cfg.IPsCfg().Opts.UsageID, utils.OptsIPsUsageID)
	if usageID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.UsageID)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.IPsV1AuthorizeIPs, utils.ConcatenatedKey(tnt, args.ID))
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)
		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*string)
			}
			return cachedResp.Error
		}
		defer Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	var ttl *time.Duration
	if ttl, err = utils.GetDurationPointerOpts(args, s.cfg.IPsCfg().Opts.TTL,
		utils.OptsIPsTTL); err != nil {
		return
	}
	var ips IPs
	if ips, err = s.matchingIPsForEvent(tnt, args, usageID, ttl); err != nil {
		return err
	}
	defer ips.unlock()

	if _, err = utils.GetFloat64Opts(args, s.cfg.IPsCfg().Opts.Units,
		utils.OptsIPsUnits); err != nil {
		return
	}

	/*
		authorize logic
		...
	*/

	*reply = utils.OK
	return
}

// V1AllocateIPs is called when a ip requires allocation.
func (s *IPService) V1AllocateIPs(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	usageID := utils.GetStringOpts(args, s.cfg.IPsCfg().Opts.UsageID, utils.OptsIPsUsageID)
	if usageID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.UsageID)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.IPsV1AllocateIPs, utils.ConcatenatedKey(tnt, args.ID))
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)
		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*string)
			}
			return cachedResp.Error
		}
		defer Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	var ttl *time.Duration
	if ttl, err = utils.GetDurationPointerOpts(args, s.cfg.IPsCfg().Opts.TTL,
		utils.OptsIPsTTL); err != nil {
		return
	}
	var ips IPs
	if ips, err = s.matchingIPsForEvent(tnt, args, usageID,
		ttl); err != nil {
		return err
	}
	defer ips.unlock()

	if _, err = utils.GetFloat64Opts(args, s.cfg.IPsCfg().Opts.Units,
		utils.OptsIPsUnits); err != nil {
		return
	}

	/*
		allocate logic
		...
	*/

	// index it for storing
	if err = s.storeMatchedIPs(ips); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// V1ReleaseIPs is called when we need to clear an allocation.
func (s *IPService) V1ReleaseIPs(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	usageID := utils.GetStringOpts(args, s.cfg.IPsCfg().Opts.UsageID, utils.OptsIPsUsageID)
	if usageID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.UsageID)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.IPsV1ReleaseIPs, utils.ConcatenatedKey(tnt, args.ID))
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)
		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*string)
			}
			return cachedResp.Error
		}
		defer Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	var ttl *time.Duration
	if ttl, err = utils.GetDurationPointerOpts(args, s.cfg.IPsCfg().Opts.TTL,
		utils.OptsIPsTTL); err != nil {
		return
	}
	var ips IPs
	if ips, err = s.matchingIPsForEvent(tnt, args, usageID,
		ttl); err != nil {
		return
	}
	defer ips.unlock()

	/*
		release logic
		...
	*/

	if err = s.storeMatchedIPs(ips); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// V1GetIP returns retrieves an IP from database.
func (s *IPService) V1GetIP(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *IP) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout,
		ipLockKey(tnt, arg.ID))
	defer guardian.Guardian.UnguardIDs(lkID)

	ip, err := s.dm.GetIP(tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		return err
	}
	*reply = *ip
	return nil
}
