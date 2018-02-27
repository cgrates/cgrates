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
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// Individual session run
type SessionRun struct {
	DerivedCharger *utils.DerivedCharger // Needed in reply
	CallDescriptor *CallDescriptor
	CallCosts      []*CallCost
}

type AttrGetLcr struct {
	*CallDescriptor
	*LCRFilter
	*utils.Paginator
}

type Responder struct {
	ExitChan         chan bool
	CdrStats         rpcclient.RpcClientConnection
	AttributeS       rpcclient.RpcClientConnection
	Timeout          time.Duration
	Timezone         string
	MaxComputedUsage map[string]time.Duration
	responseCache    *utils.ResponseCache
}

func (rs *Responder) SetTimeToLive(timeToLive time.Duration, out *int) error {
	rs.responseCache = utils.NewResponseCache(timeToLive)
	return nil
}

func (rs *Responder) getCache() *utils.ResponseCache {
	if rs.responseCache == nil {
		rs.responseCache = utils.NewResponseCache(0)
	}
	return rs.responseCache
}

// usageAllowed checks requested usage against configured MaxComputedUsage
func (rs *Responder) usageAllowed(tor string, reqUsage time.Duration) (allowed bool) {
	mcu, has := rs.MaxComputedUsage[tor]
	if !has {
		mcu = rs.MaxComputedUsage[utils.ANY]
	}
	if reqUsage <= mcu {
		allowed = true
	}
	return
}

/*
RPC method thet provides the external RPC interface for getting the rating information.
*/
func (rs *Responder) GetCost(arg *CallDescriptor, reply *CallCost) (err error) {
	if arg.Subject == "" {
		arg.Subject = arg.Account
	}
	// replace user profile fields
	if err := LoadUserProfile(arg, utils.EXTRA_FIELDS); err != nil {
		return err
	}
	// replace aliases
	if err := LoadAlias(
		&AttrMatchingAlias{
			Destination: arg.Destination,
			Direction:   arg.Direction,
			Tenant:      arg.Tenant,
			Category:    arg.Category,
			Account:     arg.Account,
			Subject:     arg.Subject,
			Context:     utils.MetaRating,
		}, arg, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		return err
	}
	if !rs.usageAllowed(arg.TOR, arg.GetDuration()) {
		return utils.ErrMaxUsageExceeded
	}
	r, e := guardian.Guardian.Guard(func() (interface{}, error) {
		return arg.GetCost()
	}, 0, arg.GetAccountKey())
	if r != nil {
		*reply = *r.(*CallCost)
	}
	if e != nil {
		return e
	}
	return
}

func (rs *Responder) Debit(arg *CallDescriptor, reply *CallCost) (err error) {
	if arg.Subject == "" {
		arg.Subject = arg.Account
	}
	// replace user profile fields
	if err := LoadUserProfile(arg, utils.EXTRA_FIELDS); err != nil {
		return err
	}
	// replace aliases
	if err := LoadAlias(
		&AttrMatchingAlias{
			Destination: arg.Destination,
			Direction:   arg.Direction,
			Tenant:      arg.Tenant,
			Category:    arg.Category,
			Account:     arg.Account,
			Subject:     arg.Subject,
			Context:     utils.MetaRating,
		}, arg, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		return err
	}
	if !rs.usageAllowed(arg.TOR, arg.GetDuration()) {
		return utils.ErrMaxUsageExceeded
	}
	r, e := arg.Debit()
	if e != nil {
		return e
	} else if r != nil {
		*reply = *r
	}
	return
}

func (rs *Responder) MaxDebit(arg *CallDescriptor, reply *CallCost) (err error) {
	cacheKey := utils.MAX_DEBIT_CACHE_PREFIX + arg.CgrID + arg.RunID + arg.DurationIndex.String()
	if item, err := rs.getCache().Get(cacheKey); err == nil && item != nil {
		if item.Value != nil {
			*reply = *(item.Value.(*CallCost))
		}
		return item.Err
	}
	if arg.Subject == "" {
		arg.Subject = arg.Account
	}
	// replace user profile fields
	if err := LoadUserProfile(arg, utils.EXTRA_FIELDS); err != nil {
		return err
	}
	// replace aliases
	if err := LoadAlias(
		&AttrMatchingAlias{
			Destination: arg.Destination,
			Direction:   arg.Direction,
			Tenant:      arg.Tenant,
			Category:    arg.Category,
			Account:     arg.Account,
			Subject:     arg.Subject,
			Context:     utils.MetaRating,
		}, arg, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		return err
	}
	if !rs.usageAllowed(arg.TOR, arg.GetDuration()) {
		return utils.ErrMaxUsageExceeded
	}
	r, e := arg.MaxDebit()
	if e != nil {
		rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{
			Err: e,
		})
		return e
	} else if r != nil {
		*reply = *r
	}
	rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{
		Value: reply,
		Err:   err,
	})
	return
}

func (rs *Responder) RefundIncrements(arg *CallDescriptor, reply *float64) (err error) {
	cacheKey := utils.REFUND_INCR_CACHE_PREFIX + arg.CgrID + arg.RunID
	if item, err := rs.getCache().Get(cacheKey); err == nil && item != nil {
		if item.Value != nil {
			*reply = *(item.Value.(*float64))
		}
		return item.Err
	}
	if arg.Subject == "" {
		arg.Subject = arg.Account
	}
	// replace user profile fields
	if err := LoadUserProfile(arg, utils.EXTRA_FIELDS); err != nil {
		return err
	}
	// replace aliases
	if err := LoadAlias(
		&AttrMatchingAlias{
			Destination: arg.Destination,
			Direction:   arg.Direction,
			Tenant:      arg.Tenant,
			Category:    arg.Category,
			Account:     arg.Account,
			Subject:     arg.Subject,
			Context:     utils.MetaRating,
		}, arg, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{
			Err: err,
		})
		return err
	}
	if !rs.usageAllowed(arg.TOR, arg.GetDuration()) {
		return utils.ErrMaxUsageExceeded
	}
	err = arg.RefundIncrements()
	rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{
		Value: reply,
		Err:   err,
	})
	return
}

func (rs *Responder) RefundRounding(arg *CallDescriptor, reply *float64) (err error) {
	cacheKey := utils.REFUND_ROUND_CACHE_PREFIX + arg.CgrID + arg.RunID + arg.DurationIndex.String()
	if item, err := rs.getCache().Get(cacheKey); err == nil && item != nil {
		if item.Value != nil {
			*reply = *(item.Value.(*float64))
		}
		return item.Err
	}
	if arg.Subject == "" {
		arg.Subject = arg.Account
	}
	// replace user profile fields
	if err := LoadUserProfile(arg, utils.EXTRA_FIELDS); err != nil {
		return err
	}
	// replace aliases
	if err := LoadAlias(
		&AttrMatchingAlias{
			Destination: arg.Destination,
			Direction:   arg.Direction,
			Tenant:      arg.Tenant,
			Category:    arg.Category,
			Account:     arg.Account,
			Subject:     arg.Subject,
			Context:     utils.MetaRating,
		}, arg, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{
			Err: err,
		})
		return err
	}
	if !rs.usageAllowed(arg.TOR, arg.GetDuration()) {
		return utils.ErrMaxUsageExceeded
	}
	err = arg.RefundRounding()
	rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{
		Value: reply,
		Err:   err,
	})
	return
}

func (rs *Responder) GetMaxSessionTime(arg *CallDescriptor, reply *float64) (err error) {
	if arg.Subject == "" {
		arg.Subject = arg.Account
	}
	// replace user profile fields
	if err := LoadUserProfile(arg, utils.EXTRA_FIELDS); err != nil {
		return err
	}
	// replace aliases
	if err := LoadAlias(
		&AttrMatchingAlias{
			Destination: arg.Destination,
			Direction:   arg.Direction,
			Tenant:      arg.Tenant,
			Category:    arg.Category,
			Account:     arg.Account,
			Subject:     arg.Subject,
			Context:     utils.MetaRating,
		}, arg, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		return err
	}
	if !rs.usageAllowed(arg.TOR, arg.GetDuration()) {
		return utils.ErrMaxUsageExceeded
	}
	r, e := arg.GetMaxSessionDuration()
	*reply, err = float64(r), e
	return
}

// Returns MaxSessionTime for an event received in sessions, considering DerivedCharging for it
func (rs *Responder) GetDerivedMaxSessionTime(ev *CDR, reply *float64) (err error) {
	cacheKey := utils.GET_DERIV_MAX_SESS_TIME + ev.CGRID + ev.RunID
	if item, err := rs.getCache().Get(cacheKey); err == nil && item != nil {
		if item.Value != nil {
			*reply = *(item.Value.(*float64))
		}
		return item.Err
	}
	if ev.Subject == "" {
		ev.Subject = ev.Account
	}
	// replace user profile fields
	if err := LoadUserProfile(ev, utils.EXTRA_FIELDS); err != nil {
		rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{Err: err})
		return err
	}
	// replace aliases
	if err := LoadAlias(
		&AttrMatchingAlias{
			Destination: ev.Destination,
			Direction:   utils.OUT,
			Tenant:      ev.Tenant,
			Category:    ev.Category,
			Account:     ev.Account,
			Subject:     ev.Subject,
			Context:     utils.MetaRating,
		}, ev, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{Err: err})
		return err
	}
	if !rs.usageAllowed(ev.ToR, ev.Usage) {
		return utils.ErrMaxUsageExceeded
	}
	maxCallDuration := -1.0
	attrsDC := &utils.AttrDerivedChargers{Tenant: ev.GetTenant(utils.META_DEFAULT),
		Category: ev.GetCategory(utils.META_DEFAULT), Direction: utils.OUT,
		Account: ev.GetAccount(utils.META_DEFAULT), Subject: ev.GetSubject(utils.META_DEFAULT)}
	dcs := &utils.DerivedChargers{}
	if err := rs.GetDerivedChargers(attrsDC, dcs); err != nil {
		rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{Err: err})
		return err
	}
	dcs, _ = dcs.AppendDefaultRun()
	for _, dc := range dcs.Chargers {
		if utils.IsSliceMember([]string{utils.META_RATED, utils.RATED}, ev.GetReqType(dc.RequestTypeField)) { // Only consider prepaid and pseudoprepaid for MaxSessionTime
			continue
		}
		runFilters, _ := utils.ParseRSRFields(dc.RunFilters, utils.INFIELD_SEP)
		matchingAllFilters := true
		for _, dcRunFilter := range runFilters {
			if !dcRunFilter.FilterPasses(ev.FieldAsString(dcRunFilter)) {
				matchingAllFilters = false
				break
			}
		}
		if !matchingAllFilters { // Do not process the derived charger further if not all filters were matched
			continue
		}
		startTime, err := ev.GetSetupTime(utils.META_DEFAULT, rs.Timezone)
		if err != nil {
			rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{Err: err})
			return err
		}
		usage, err := ev.GetDuration(utils.META_DEFAULT)
		if err != nil {
			rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{Err: err})
			return err
		}
		if usage == 0 {
			usage = config.CgrConfig().MaxCallDuration
		}
		cd := &CallDescriptor{
			CgrID:       ev.GetCgrId(rs.Timezone),
			RunID:       dc.RunID,
			TOR:         ev.ToR,
			Direction:   utils.OUT,
			Tenant:      ev.GetTenant(dc.TenantField),
			Category:    ev.GetCategory(dc.CategoryField),
			Subject:     ev.GetSubject(dc.SubjectField),
			Account:     ev.GetAccount(dc.AccountField),
			Destination: ev.GetDestination(dc.DestinationField),
			TimeStart:   startTime,
			TimeEnd:     startTime.Add(usage),
		}
		var remainingDuration float64
		err = rs.GetMaxSessionTime(cd, &remainingDuration)
		if err != nil {
			*reply = 0
			rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{Err: err})
			return err
		}
		if utils.IsSliceMember([]string{utils.META_POSTPAID, utils.POSTPAID}, ev.GetReqType(dc.RequestTypeField)) {
			// Only consider prepaid and pseudoprepaid for MaxSessionTime, do it here for unauthorized destination error check
			continue
		}
		// Set maxCallDuration, smallest out of all forked sessions
		if maxCallDuration == -1.0 { // first time we set it /not initialized yet
			maxCallDuration = remainingDuration
		} else if maxCallDuration > remainingDuration {
			maxCallDuration = remainingDuration
		}
	}
	*reply = maxCallDuration
	rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{Value: reply})
	return nil
}

// Used by SM to get all the prepaid CallDescriptors attached to a session
func (rs *Responder) GetSessionRuns(ev *CDR, sRuns *[]*SessionRun) (err error) {
	cacheKey := utils.GET_SESS_RUNS_CACHE_PREFIX + ev.CGRID
	if item, err := rs.getCache().Get(cacheKey); err == nil && item != nil {
		if item.Value != nil {
			*sRuns = *(item.Value.(*[]*SessionRun))
		}
		return item.Err
	}
	if ev.Subject == "" {
		ev.Subject = ev.Account
	}
	//utils.Logger.Info(fmt.Sprintf("DC before: %+v", ev))
	// replace user profile fields
	if err := LoadUserProfile(ev, utils.EXTRA_FIELDS); err != nil {
		return err
	}
	// replace aliases
	if err := LoadAlias(
		&AttrMatchingAlias{
			Destination: ev.Destination,
			Direction:   utils.OUT,
			Tenant:      ev.Tenant,
			Category:    ev.Category,
			Account:     ev.Account,
			Subject:     ev.Subject,
			Context:     utils.MetaRating,
		}, ev, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		return err
	}

	//utils.Logger.Info(fmt.Sprintf("DC after: %+v", ev))
	attrsDC := &utils.AttrDerivedChargers{Tenant: ev.GetTenant(utils.META_DEFAULT),
		Category: ev.GetCategory(utils.META_DEFAULT), Direction: utils.OUT,
		Account: ev.GetAccount(utils.META_DEFAULT), Subject: ev.GetSubject(utils.META_DEFAULT),
		Destination: ev.GetDestination(utils.META_DEFAULT)}
	//utils.Logger.Info(fmt.Sprintf("Derived chargers for: %+v", attrsDC))
	dcs := &utils.DerivedChargers{}
	if err := rs.GetDerivedChargers(attrsDC, dcs); err != nil {
		rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{
			Err: err,
		})
		return err
	}
	dcs, _ = dcs.AppendDefaultRun()
	//utils.Logger.Info(fmt.Sprintf("DCS: %v", len(dcs.Chargers)))
	sesRuns := make([]*SessionRun, 0)
	for _, dc := range dcs.Chargers {
		if !utils.IsSliceMember([]string{utils.META_PREPAID, utils.PREPAID}, ev.GetReqType(dc.RequestTypeField)) {
			continue // We only consider prepaid sessions
		}
		startTime, err := ev.GetAnswerTime(dc.AnswerTimeField, rs.Timezone)
		if err != nil || startTime.IsZero() { // AnswerTime not parsable, try SetupTime
			startTime, err = ev.GetSetupTime(dc.SetupTimeField, rs.Timezone)
			if err != nil {
				rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{Err: err})
				return errors.New("Error parsing answer event start time")
			}
		}
		endTime, err := ev.GetEndTime("", rs.Timezone)
		if err != nil {
			rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{Err: err})
			return errors.New("Error parsing answer event end time")
		}
		extraFields := ev.GetExtraFields()
		cd := &CallDescriptor{
			CgrID:       ev.GetCgrId(rs.Timezone),
			RunID:       dc.RunID,
			TOR:         ev.ToR,
			Direction:   utils.OUT,
			Tenant:      ev.GetTenant(dc.TenantField),
			Category:    ev.GetCategory(dc.CategoryField),
			Subject:     ev.GetSubject(dc.SubjectField),
			Account:     ev.GetAccount(dc.AccountField),
			Destination: ev.GetDestination(dc.DestinationField),
			TimeStart:   startTime,
			TimeEnd:     endTime,
			ExtraFields: extraFields}
		if flagsStr, hasFlags := extraFields[utils.CGRFlags]; hasFlags { // Force duration from extra fields
			flags := utils.StringMapFromSlice(strings.Split(flagsStr, utils.INFIELD_SEP))
			if _, hasFD := flags[utils.FlagForceDuration]; hasFD {
				cd.ForceDuration = true
			}
		}
		sesRuns = append(sesRuns, &SessionRun{DerivedCharger: dc, CallDescriptor: cd})
	}
	//utils.Logger.Info(fmt.Sprintf("RUNS: %v", len(sesRuns)))
	*sRuns = sesRuns
	rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{Value: sRuns})
	return nil
}

func (rs *Responder) GetDerivedChargers(attrs *utils.AttrDerivedChargers, dcs *utils.DerivedChargers) error {
	if dcsH, err := HandleGetDerivedChargers(dm, attrs); err != nil {
		return err
	} else if dcsH != nil {
		*dcs = *dcsH
	}
	return nil
}

func (rs *Responder) GetLCR(attrs *AttrGetLcr, reply *LCRCost) (err error) {
	cacheKey := utils.LCRCachePrefix + attrs.CgrID + attrs.RunID
	if item, err := rs.getCache().Get(cacheKey); err == nil && item != nil {
		if item.Value != nil {
			*reply = *(item.Value.(*LCRCost))
		}
		return item.Err
	}
	if attrs.CallDescriptor.Subject == "" {
		attrs.CallDescriptor.Subject = attrs.CallDescriptor.Account
	}
	// replace user profile fields
	if err := LoadUserProfile(attrs.CallDescriptor, utils.EXTRA_FIELDS); err != nil {
		return err
	}
	// replace aliases
	cd := attrs.CallDescriptor
	if err := LoadAlias(
		&AttrMatchingAlias{
			Destination: cd.Destination,
			Direction:   cd.Direction,
			Tenant:      cd.Tenant,
			Category:    cd.Category,
			Account:     cd.Account,
			Subject:     cd.Subject,
			Context:     utils.MetaRating,
		}, cd, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{Err: err})
		return err
	}
	if !rs.usageAllowed(cd.TOR, cd.GetDuration()) {
		return utils.ErrMaxUsageExceeded
	}
	lcrCost, err := attrs.CallDescriptor.GetLCR(rs.CdrStats, attrs.LCRFilter, attrs.Paginator)
	if err != nil {
		rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{Err: err})
		return err
	}
	if lcrCost.Entry != nil && lcrCost.Entry.Strategy == LCR_STRATEGY_LOAD {
		for _, suppl := range lcrCost.SupplierCosts {
			suppl.Cost = -1 // In case of load distribution we don't calculate costs
		}
	}
	rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{Value: lcrCost})
	*reply = *lcrCost
	return nil
}

func (rs *Responder) Status(arg string, reply *map[string]interface{}) (err error) {
	if arg != "" { // Introduce  delay in answer, used in some automated tests
		if delay, err := utils.ParseDurationWithNanosecs(arg); err == nil {
			time.Sleep(delay)
		}
	}
	memstats := new(runtime.MemStats)
	runtime.ReadMemStats(memstats)
	response := make(map[string]interface{})
	response[utils.NodeID] = config.CgrConfig().NodeID
	response["MemoryUsage"] = utils.SizeFmt(float64(memstats.HeapAlloc), "")
	response[utils.ActiveGoroutines] = runtime.NumGoroutine()
	response["Footprint"] = utils.SizeFmt(float64(memstats.Sys), "")
	*reply = response
	return
}

func (rs *Responder) Shutdown(arg string, reply *string) (err error) {
	dm.DataDB().Close()
	cdrStorage.Close()
	defer func() { rs.ExitChan <- true }()
	*reply = "Done!"
	return
}

func (rs *Responder) GetTimeout(i int, d *time.Duration) error {
	*d = rs.Timeout
	return nil
}

func (rs *Responder) Call(serviceMethod string, args interface{}, reply interface{}) error {
	parts := strings.Split(serviceMethod, ".")
	if len(parts) != 2 {
		return utils.ErrNotImplemented
	}
	// get method
	method := reflect.ValueOf(rs).MethodByName(parts[1])
	if !method.IsValid() {
		return utils.ErrNotImplemented
	}
	// construct the params
	params := []reflect.Value{reflect.ValueOf(args), reflect.ValueOf(reply)}
	ret := method.Call(params)
	if len(ret) != 1 {
		return utils.ErrServerError
	}
	if ret[0].Interface() == nil {
		return nil
	}
	err, ok := ret[0].Interface().(error)
	if !ok {
		return utils.ErrServerError
	}
	return err
}
