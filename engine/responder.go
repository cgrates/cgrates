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
	RequestType    string
	DerivedCharger *utils.DerivedCharger // Needed in reply
	CallDescriptor *CallDescriptor
	CallCosts      []*CallCost
}

type Responder struct {
	ExitChan         chan bool
	CdrStats         rpcclient.RpcClientConnection
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

func (rs *Responder) RefundIncrements(arg *CallDescriptor, reply *Account) (err error) {
	cacheKey := utils.REFUND_INCR_CACHE_PREFIX + arg.CgrID + arg.RunID
	if item, err := rs.getCache().Get(cacheKey); err == nil && item != nil {
		if item.Value != nil {
			*reply = *(item.Value.(*Account))
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
		err = utils.ErrMaxUsageExceeded
		rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{
			Err: err,
		})
		return
	}
	if acnt, err := arg.RefundIncrements(); err != nil {
		rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{
			Err: err,
		})
		return err
	} else if acnt != nil {
		*reply = *acnt
	}
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

func (rs *Responder) GetMaxSessionTime(arg *CallDescriptor, reply *time.Duration) (err error) {
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
	*reply, err = r, e
	return
}

// Returns MaxSessionTime for an event received in sessions, considering DerivedCharging for it
func (rs *Responder) GetDerivedMaxSessionTime(ev *CDR, reply *time.Duration) (err error) {
	cacheKey := utils.GET_DERIV_MAX_SESS_TIME + ev.CGRID + ev.RunID
	if item, err := rs.getCache().Get(cacheKey); err == nil && item != nil {
		if item.Value != nil {
			*reply = *(item.Value.(*time.Duration))
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
	maxCallDuration := time.Duration(-1.0)
	attrsDC := &utils.AttrDerivedChargers{Tenant: ev.Tenant,
		Category: ev.Category, Direction: utils.OUT,
		Account: ev.Account, Subject: ev.Subject}
	dcs := &utils.DerivedChargers{}
	if err := rs.GetDerivedChargers(attrsDC, dcs); err != nil {
		rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{Err: err})
		return err
	}
	dcs, _ = dcs.AppendDefaultRun()
	for _, dc := range dcs.Chargers {
		runFilters, _ := utils.ParseRSRFields(dc.RunFilters, utils.INFIELD_SEP)
		matchingAllFilters := true
		for _, dcRunFilter := range runFilters {
			if _, err := ev.FieldAsStringWithRSRField(dcRunFilter); err != nil {
				matchingAllFilters = false
				break
			}
		}
		if !matchingAllFilters { // Do not process the derived charger further if not all filters were matched
			continue
		}
		forkedCDR, err := ev.ForkCdr(dc.RunID, utils.NewRSRFieldMustCompile(dc.RequestTypeField),
			utils.NewRSRFieldMustCompile(dc.TenantField), utils.NewRSRFieldMustCompile(dc.CategoryField),
			utils.NewRSRFieldMustCompile(dc.AccountField), utils.NewRSRFieldMustCompile(dc.SubjectField),
			utils.NewRSRFieldMustCompile(dc.DestinationField), utils.NewRSRFieldMustCompile(dc.SetupTimeField),
			utils.NewRSRFieldMustCompile(dc.AnswerTimeField), utils.NewRSRFieldMustCompile(dc.UsageField),
			utils.NewRSRFieldMustCompile(dc.PreRatedField), utils.NewRSRFieldMustCompile(dc.CostField),
			nil, false, rs.Timezone)
		if err != nil {
			return err
		}
		if !utils.IsSliceMember([]string{utils.META_PREPAID, utils.PREPAID,
			utils.META_PSEUDOPREPAID, utils.PSEUDOPREPAID},
			forkedCDR.RequestType) { // Only consider prepaid and pseudoprepaid for MaxSessionTime
			continue
		}
		if forkedCDR.Usage == 0 {
			forkedCDR.Usage = config.CgrConfig().MaxCallDuration
		}
		setupTime := forkedCDR.SetupTime
		if setupTime.IsZero() {
			setupTime = forkedCDR.AnswerTime
		}
		cd := &CallDescriptor{
			CgrID:       forkedCDR.CGRID,
			RunID:       forkedCDR.RunID,
			TOR:         forkedCDR.ToR,
			Direction:   utils.OUT,
			Tenant:      forkedCDR.Tenant,
			Category:    forkedCDR.Category,
			Subject:     forkedCDR.Subject,
			Account:     forkedCDR.Account,
			Destination: forkedCDR.Destination,
			TimeStart:   setupTime,
			TimeEnd:     setupTime.Add(forkedCDR.Usage),
		}
		var remainingDuration time.Duration
		err = rs.GetMaxSessionTime(cd, &remainingDuration)
		if err != nil {
			*reply = time.Duration(0)
			rs.getCache().Cache(cacheKey, &utils.ResponseCacheItem{Err: err})
			return err
		}
		// Set maxCallDuration, smallest out of all forked sessions
		if maxCallDuration == time.Duration(-1) { // first time we set it /not initialized yet
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
	attrsDC := &utils.AttrDerivedChargers{Tenant: ev.Tenant,
		Category: ev.Category, Direction: utils.OUT,
		Account: ev.Account, Subject: ev.Subject,
		Destination: ev.Destination}
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
		forkedCDR, err := ev.ForkCdr(dc.RunID, utils.NewRSRFieldMustCompile(dc.RequestTypeField),
			utils.NewRSRFieldMustCompile(dc.TenantField), utils.NewRSRFieldMustCompile(dc.CategoryField),
			utils.NewRSRFieldMustCompile(dc.AccountField), utils.NewRSRFieldMustCompile(dc.SubjectField),
			utils.NewRSRFieldMustCompile(dc.DestinationField), utils.NewRSRFieldMustCompile(dc.SetupTimeField),
			utils.NewRSRFieldMustCompile(dc.AnswerTimeField), utils.NewRSRFieldMustCompile(dc.UsageField),
			utils.NewRSRFieldMustCompile(dc.PreRatedField), utils.NewRSRFieldMustCompile(dc.CostField),
			nil, false, rs.Timezone)
		if err != nil {
			return err
		}
		startTime := forkedCDR.AnswerTime
		if startTime.IsZero() { // AnswerTime not parsable, try SetupTime
			startTime = forkedCDR.SetupTime
		}
		cd := &CallDescriptor{
			CgrID:       forkedCDR.CGRID,
			RunID:       forkedCDR.RunID,
			TOR:         forkedCDR.ToR,
			Direction:   utils.OUT,
			Tenant:      forkedCDR.Tenant,
			Category:    forkedCDR.Category,
			Subject:     forkedCDR.Subject,
			Account:     forkedCDR.Account,
			Destination: forkedCDR.Destination,
			TimeStart:   startTime,
			TimeEnd:     startTime.Add(forkedCDR.Usage),
			ExtraFields: ev.ExtraFields}
		if flagsStr, hasFlags := ev.ExtraFields[utils.CGRFlags]; hasFlags { // Force duration from extra fields
			flags := utils.StringMapFromSlice(strings.Split(flagsStr, utils.INFIELD_SEP))
			if _, hasFD := flags[utils.FlagForceDuration]; hasFD {
				cd.ForceDuration = true
			}
		}
		sesRuns = append(sesRuns, &SessionRun{RequestType: forkedCDR.RequestType, DerivedCharger: dc, CallDescriptor: cd})
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

func (rs *Responder) Status(arg string, reply *map[string]interface{}) (err error) {
	if arg != "" { // Introduce  delay in answer, used in some automated tests
		if delay, err := utils.ParseDurationWithNanosecs(arg); err == nil {
			time.Sleep(delay)
		}
	}
	memstats := new(runtime.MemStats)
	runtime.ReadMemStats(memstats)
	response := make(map[string]interface{})
	response[utils.NodeID] = config.CgrConfig().GeneralCfg().NodeID
	response["MemoryUsage"] = utils.SizeFmt(float64(memstats.HeapAlloc), "")
	response[utils.ActiveGoroutines] = runtime.NumGoroutine()
	response["Footprint"] = utils.SizeFmt(float64(memstats.Sys), "")
	response[utils.Version] = utils.GetCGRVersion
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
