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

package ips

import (
	"cmp"
	"fmt"
	"runtime"
	"slices"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

// ipProfile represents the user configuration for the ip
type ipProfile struct {
	IPProfile *utils.IPProfile
	lkID      string // holds the reference towards guardian lock key

}

// lock will lock the ipProfile using guardian and store the lock within r.lkID
// if lkID is passed as argument, the lock is considered as executed
func (ip *ipProfile) lock(lkID string) {
	if lkID == utils.EmptyString {
		lkID = guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			utils.IPProfileLockKey(ip.IPProfile.Tenant, ip.IPProfile.ID))
	}
	ip.lkID = lkID
}

// unlock will unlock the ipProfile and clear rp.lkID
func (ip *ipProfile) unlock() {
	if ip.lkID == utils.EmptyString {
		return
	}
	guardian.Guardian.UnguardIDs(ip.lkID)
	ip.lkID = utils.EmptyString
}

// isLocked returns the locks status of this ipProfile
func (ip *ipProfile) isLocked() bool {
	return ip.lkID != utils.EmptyString
}

// ip represents an ip in the system
// not thread safe, needs locking at process level
type ip struct {
	IP     *utils.IP
	lkID   string         // ID of the lock used when matching the ip
	ttl    *time.Duration // time to leave for this ip, picked up on each IP initialization out of config
	tUsage *float64       // sum of all usages
	dirty  *bool          // the usages were modified, needs save, *bool so we only save if enabled in config
	rPrf   *ipProfile     // for ordering purposes
}

// lock will lock the ip using guardian and store the lock within r.lkID
// if lkID is passed as argument, the lock is considered as executed
func (ip *ip) lock(lkID string) {
	if lkID == utils.EmptyString {
		lkID = guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			utils.IPLockKey(ip.IP.Tenant, ip.IP.ID))
	}
	ip.lkID = lkID
}

// unlock will unlock the ip and clear r.lkID
func (ip *ip) unlock() {
	if ip.lkID == utils.EmptyString {
		return
	}
	guardian.Guardian.UnguardIDs(ip.lkID)
	ip.lkID = utils.EmptyString
}

// isLocked returns the locks status of this ip
func (ip *ip) isLocked() bool {
	return ip.lkID != utils.EmptyString
}

// removeExpiredUnits removes units which are expired from the ip
func (ip *ip) removeExpiredUnits() {
	var firstActive int
	for _, usageID := range ip.IP.TTLIdx {
		if u, has := ip.IP.Usages[usageID]; has && u.IsActive(time.Now()) {
			break
		}
		firstActive++
	}
	if firstActive == 0 {
		return
	}
	for _, uID := range ip.IP.TTLIdx[:firstActive] {
		usage, has := ip.IP.Usages[uID]
		if !has {
			continue
		}
		delete(ip.IP.Usages, uID)
		if ip.tUsage != nil { //  total usage was not yet calculated so we do not need to update it
			*ip.tUsage -= usage.Units
			if *ip.tUsage < 0 { // something went wrong
				utils.Logger.Warning(
					fmt.Sprintf("resetting total usage for ipID: %s, usage smaller than 0: %f", ip.IP.ID, *ip.tUsage))
				ip.tUsage = nil
			}
		}
	}
	ip.IP.TTLIdx = ip.IP.TTLIdx[firstActive:]
	ip.tUsage = nil
}

// recordUsage records a new usage
func (ip *ip) recordUsage(usage *utils.IPUsage) error {
	if _, has := ip.IP.Usages[usage.ID]; has {
		return fmt.Errorf("duplicate ip usage with id: %s", usage.TenantID())
	}
	if ip.ttl != nil && *ip.ttl != -1 {
		if *ip.ttl == 0 {
			return nil // no recording for ttl of 0
		}
		usage = usage.Clone() // don't influence the initial ru
		usage.ExpiryTime = time.Now().Add(*ip.ttl)
	}
	ip.IP.Usages[usage.ID] = usage
	if ip.tUsage != nil {
		*ip.tUsage += usage.Units
	}
	if !usage.ExpiryTime.IsZero() {
		ip.IP.TTLIdx = append(ip.IP.TTLIdx, usage.ID)
	}
	return nil
}

// clearUsage clears the usage for an ID
func (ip *ip) clearUsage(usageID string) error {
	usage, has := ip.IP.Usages[usageID]
	if !has {
		return fmt.Errorf("cannot find usage record with id: %s", usageID)
	}
	if !usage.ExpiryTime.IsZero() {
		for i, uIDIdx := range ip.IP.TTLIdx {
			if uIDIdx == usageID {
				ip.IP.TTLIdx = slices.Delete(ip.IP.TTLIdx, i, i+1)
				break
			}
		}
	}
	if ip.tUsage != nil {
		*ip.tUsage -= usage.Units
	}
	delete(ip.IP.Usages, usageID)
	return nil
}

// IPs is a collection of IP objects.
type IPs []*ip

// unlock will unlock ips part of this slice
func (ips IPs) unlock() {
	for _, ip := range ips {
		ip.unlock()
		if ip.rPrf != nil {
			ip.rPrf.unlock()
		}
	}
}

// ids returns a map of ip IDs which is used for caching
func (ips IPs) ids() utils.StringSet {
	ids := make(utils.StringSet)
	for _, ip := range ips {
		ids.Add(ip.IP.ID)
	}
	return ids
}

// NewIPService  returns a new IPService
func NewIPService(dm *engine.DataManager, cfg *config.CGRConfig,
	fltrs *engine.FilterS, cm *engine.ConnManager) *IPService {
	return &IPService{dm: dm,
		storedIPs:   make(utils.StringSet),
		cfg:         cfg,
		cm:          cm,
		fltrs:       fltrs,
		loopStopped: make(chan struct{}),
		stopBackup:  make(chan struct{}),
	}

}

// IPService is the service handling resources
type IPService struct {
	dm           *engine.DataManager // So we can load the data in cache and index it
	fltrs        *engine.FilterS
	storedIPsMux sync.RWMutex    // protects storedIPs
	storedIPs    utils.StringSet // keep a record of resources which need saving, map[resID]bool
	cfg          *config.CGRConfig
	stopBackup   chan struct{} // control storing process
	loopStopped  chan struct{}
	cm           *engine.ConnManager
}

// Reload stops the backupLoop and restarts it
func (s *IPService) Reload(ctx *context.Context) {
	close(s.stopBackup)
	<-s.loopStopped // wait until the loop is done
	s.stopBackup = make(chan struct{})
	go s.runBackup(ctx)
}

// StartLoop starts the gorutine with the backup loop
func (s *IPService) StartLoop(ctx *context.Context) {
	go s.runBackup(ctx)
}

// Shutdown is called to shutdown the service
func (s *IPService) Shutdown(ctx *context.Context) {
	close(s.stopBackup)
	s.storeIPs(ctx)
}

// backup will regularly store resources changed to dataDB
func (s *IPService) runBackup(ctx *context.Context) {
	storeInterval := s.cfg.IPsCfg().StoreInterval
	if storeInterval <= 0 {
		s.loopStopped <- struct{}{}
		return
	}
	for {
		s.storeIPs(ctx)
		select {
		case <-s.stopBackup:
			s.loopStopped <- struct{}{}
			return
		case <-time.After(storeInterval):
		}
	}
}

// storeIPs represents one task of complete backup
func (s *IPService) storeIPs(ctx *context.Context) {
	var failedRIDs []string
	for { // don't stop until we store all dirty resources
		s.storedIPsMux.Lock()
		rID := s.storedIPs.GetOne()
		if rID != "" {
			s.storedIPs.Remove(rID)
		}
		s.storedIPsMux.Unlock()
		if rID == "" {
			break // no more keys, backup completed
		}
		rIf, ok := engine.Cache.Get(utils.CacheIPs, rID)
		if !ok || rIf == nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> failed retrieving from cache resource with ID: %s", utils.IPs, rID))
			continue
		}
		r := &ip{
			IP: rIf.(*utils.IP),

			// NOTE: dirty is hardcoded to true, otherwise resources would
			// never be stored.
			// Previously, dirty was part of the cached resource.
			dirty: utils.BoolPointer(true),
		}
		r.lock(utils.EmptyString)
		if err := s.storeIP(ctx, r); err != nil {
			failedRIDs = append(failedRIDs, rID) // record failure so we can schedule it for next backup
		}
		r.unlock()
		// randomize the CPU load and give up thread control
		runtime.Gosched()
	}
	if len(failedRIDs) != 0 { // there were errors on save, schedule the keys for next backup
		s.storedIPsMux.Lock()
		s.storedIPs.AddSlice(failedRIDs)
		s.storedIPsMux.Unlock()
	}
}

// StoreIP stores the resource in DB and corrects dirty flag
func (s *IPService) storeIP(ctx *context.Context, r *ip) (err error) {
	if r.dirty == nil || !*r.dirty {
		return
	}
	if err = s.dm.SetIP(ctx, r.IP); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<IPs> failed saving IP with ID: %s, error: %s",
				r.IP.ID, err.Error()))
		return
	}
	//since we no longer handle cache in DataManager do here a manual caching
	if tntID := r.IP.TenantID(); engine.Cache.HasItem(utils.CacheIPs, tntID) { // only cache if previously there
		if err = engine.Cache.Set(ctx, utils.CacheIPs, tntID, r.IP, nil,
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

// storeMatchedIPs will store the list of resources based on the StoreInterval
func (s *IPService) storeMatchedIPs(ctx *context.Context, mtcRLs IPs) (err error) {
	if s.cfg.IPsCfg().StoreInterval == 0 {
		return
	}
	if s.cfg.IPsCfg().StoreInterval > 0 {
		s.storedIPsMux.Lock()
		defer s.storedIPsMux.Unlock()
	}
	for _, r := range mtcRLs {
		if r.dirty != nil {
			*r.dirty = true // mark it to be saved
			if s.cfg.IPsCfg().StoreInterval > 0 {
				s.storedIPs.Add(r.IP.TenantID())
				continue
			}
			if err = s.storeIP(ctx, r); err != nil {
				return
			}
		}

	}
	return
}

// matchingIPsForEvent returns ordered list of matching resources which are active by the time of the call
func (s *IPService) matchingIPsForEvent(ctx *context.Context, tnt string, ev *utils.CGREvent,
	evUUID string, ttl *time.Duration) (ips IPs, err error) {
	var rIDs utils.StringSet
	evNm := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
	}
	if x, ok := engine.Cache.Get(utils.CacheEventIPs, evUUID); ok { // The IPIDs were cached as utils.StringSet{"resID":bool}
		if x == nil {
			return nil, utils.ErrNotFound
		}
		rIDs = x.(utils.StringSet)
		defer func() { // make sure we uncache if we find errors
			if err != nil {
				// TODO: Consider using RemoveWithoutReplicate instead, as
				// partitions with Replicate=true call ReplicateRemove in
				// onEvict by default.
				if errCh := engine.Cache.Remove(ctx, utils.CacheEventIPs, evUUID,
					true, utils.NonTransactional); errCh != nil {
					err = errCh
				}
			}
		}()

	} else { // select the resourceIDs out of dataDB
		rIDs, err = engine.MatchingItemIDsForEvent(ctx, evNm,
			s.cfg.IPsCfg().StringIndexedFields,
			s.cfg.IPsCfg().PrefixIndexedFields,
			s.cfg.IPsCfg().SuffixIndexedFields,
			s.cfg.IPsCfg().ExistsIndexedFields,
			s.cfg.IPsCfg().NotExistsIndexedFields,
			s.dm, utils.CacheIPFilterIndexes, tnt,
			s.cfg.IPsCfg().IndexedSelects,
			s.cfg.IPsCfg().NestedFields,
		)
		if err != nil {
			if err == utils.ErrNotFound {
				if errCh := engine.Cache.Set(ctx, utils.CacheEventIPs, evUUID, nil, nil, true, ""); errCh != nil { // cache negative match
					return nil, errCh
				}
			}
			return
		}
	}
	ips = make(IPs, 0, len(rIDs))
	weights := make(map[string]float64) // stores sorting weights by resource ID
	for resName := range rIDs {
		lkPrflID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			utils.IPProfileLockKey(tnt, resName))
		var rp *utils.IPProfile
		if rp, err = s.dm.GetIPProfile(ctx, tnt, resName,
			true, true, utils.NonTransactional); err != nil {
			guardian.Guardian.UnguardIDs(lkPrflID)
			if err == utils.ErrNotFound {
				continue
			}
			ips.unlock()
			return
		}
		rPrf := &ipProfile{
			IPProfile: rp,
		}
		rPrf.lock(lkPrflID)
		var pass bool
		if pass, err = s.fltrs.Pass(ctx, tnt, rPrf.IPProfile.FilterIDs,
			evNm); err != nil {
			rPrf.unlock()
			ips.unlock()
			return nil, err
		} else if !pass {
			rPrf.unlock()
			continue
		}
		lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
			config.CgrConfig().GeneralCfg().LockingTimeout,
			utils.IPLockKey(rPrf.IPProfile.Tenant, rPrf.IPProfile.ID))
		var res *utils.IP
		if res, err = s.dm.GetIP(ctx, rPrf.IPProfile.Tenant, rPrf.IPProfile.ID, true, true, ""); err != nil {
			guardian.Guardian.UnguardIDs(lkID)
			rPrf.unlock()
			ips.unlock()
			return nil, err
		}
		r := &ip{
			IP: res,
		}
		r.lock(lkID) // pass the lock into resource so we have it as reference
		if rPrf.IPProfile.Stored && r.dirty == nil {
			r.dirty = utils.BoolPointer(false)
		}
		if ttl != nil {
			if *ttl != 0 {
				r.ttl = ttl
			}
		} else if rPrf.IPProfile.TTL >= 0 {
			r.ttl = utils.DurationPointer(rPrf.IPProfile.TTL)
		}
		r.rPrf = rPrf
		weight, err := engine.WeightFromDynamics(ctx, rPrf.IPProfile.Weights, s.fltrs, tnt, evNm)
		if err != nil {
			return nil, err
		}
		weights[r.IP.ID] = weight
		ips = append(ips, r)
	}

	if len(ips) == 0 {
		return nil, utils.ErrNotFound
	}

	// Sort by weight (higher values first).
	slices.SortFunc(ips, func(a, b *ip) int {
		return cmp.Compare(weights[b.IP.ID], weights[a.IP.ID])
	})

	if err = engine.Cache.Set(ctx, utils.CacheEventIPs, evUUID, ips.ids(), nil, true, ""); err != nil {
		ips.unlock()
	}
	return
}
