/*
Real-time Charging System for Telecom & ISP environments
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
	"net/rpc"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/balancer2go"
	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/config"
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
	*utils.Paginator
}

type Responder struct {
	Bal           *balancer2go.Balancer
	ExitChan      chan bool
	Stats         rpcclient.RpcClientConnection
	Timezone      string
	cnt           int64
	responseCache *cache2go.ResponseCache
}

func (rs *Responder) SetTimeToLive(timeToLive time.Duration, out *int) error {
	rs.responseCache = cache2go.NewResponseCache(timeToLive)
	return nil
}

func (rs *Responder) getCache() *cache2go.ResponseCache {
	if rs.responseCache == nil {
		rs.responseCache = cache2go.NewResponseCache(0)
	}
	return rs.responseCache
}

/*
RPC method thet provides the external RPC interface for getting the rating information.
*/
func (rs *Responder) GetCost(arg *CallDescriptor, reply *CallCost) (err error) {
	rs.cnt += 1
	if arg.Subject == "" {
		arg.Subject = arg.Account
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
			Context:     utils.ALIAS_CONTEXT_RATING,
		}, arg, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		return err
	}
	// replace user profile fields
	if err := LoadUserProfile(arg, utils.EXTRA_FIELDS); err != nil {
		return err
	}
	if rs.Bal != nil {
		r, e := rs.getCallCost(arg, "Responder.GetCost")
		*reply, err = *r, e
	} else {
		r, e := Guardian.Guard(func() (interface{}, error) {
			return arg.GetCost()
		}, 0, arg.GetAccountKey())
		if r != nil {
			*reply = *r.(*CallCost)
		}
		if e != nil {
			return e
		}
	}
	return
}

func (rs *Responder) Debit(arg *CallDescriptor, reply *CallCost) (err error) {
	if arg.Subject == "" {
		arg.Subject = arg.Account
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
			Context:     utils.ALIAS_CONTEXT_RATING,
		}, arg, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		return err
	}
	// replace user profile fields
	if err := LoadUserProfile(arg, utils.EXTRA_FIELDS); err != nil {
		return err
	}
	if rs.Bal != nil {
		r, e := rs.getCallCost(arg, "Responder.Debit")
		*reply, err = *r, e
	} else {
		r, e := arg.Debit()
		if e != nil {
			return e
		} else if r != nil {
			*reply = *r
		}
	}
	return
}

func (rs *Responder) MaxDebit(arg *CallDescriptor, reply *CallCost) (err error) {
	cacheKey := "MaxDebit" + arg.CgrId + strconv.FormatFloat(arg.LoopIndex, 'f', -1, 64)
	if item, err := rs.getCache().Get(cacheKey); err == nil && item != nil {
		*reply = *(item.Value.(*CallCost))
		return item.Err
	}
	if arg.Subject == "" {
		arg.Subject = arg.Account
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
			Context:     utils.ALIAS_CONTEXT_RATING,
		}, arg, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
		return err
	}
	// replace user profile fields
	if err := LoadUserProfile(arg, utils.EXTRA_FIELDS); err != nil {
		rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
		return err
	}
	if rs.Bal != nil {
		r, e := rs.getCallCost(arg, "Responder.MaxDebit")
		*reply, err = *r, e
	} else {
		r, e := arg.MaxDebit()
		if e != nil {
			rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: e})
			return e
		} else if r != nil {
			*reply = *r
		}
	}
	rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Value: reply, Err: err})
	return
}

func (rs *Responder) RefundIncrements(arg *CallDescriptor, reply *float64) (err error) {
	cacheKey := "RefundIncrements" + arg.CgrId
	if item, err := rs.getCache().Get(cacheKey); err == nil && item != nil {
		*reply = *(item.Value.(*float64))
		return item.Err
	}
	if arg.Subject == "" {
		arg.Subject = arg.Account
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
			Context:     utils.ALIAS_CONTEXT_RATING,
		}, arg, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
		return err
	}
	// replace user profile fields
	if err := LoadUserProfile(arg, utils.EXTRA_FIELDS); err != nil {
		rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
		return err
	}
	if rs.Bal != nil {
		*reply, err = rs.callMethod(arg, "Responder.RefundIncrements")
	} else {
		r, e := Guardian.Guard(func() (interface{}, error) {
			return arg.RefundIncrements()
		}, 0, arg.GetAccountKey())
		*reply, err = r.(float64), e
	}
	rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Value: reply, Err: err})
	return
}

func (rs *Responder) GetMaxSessionTime(arg *CallDescriptor, reply *float64) (err error) {
	if arg.Subject == "" {
		arg.Subject = arg.Account
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
			Context:     utils.ALIAS_CONTEXT_RATING,
		}, arg, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		return err
	}
	// replace user profile fields
	if err := LoadUserProfile(arg, utils.EXTRA_FIELDS); err != nil {
		return err
	}
	if rs.Bal != nil {
		*reply, err = rs.callMethod(arg, "Responder.GetMaxSessionTime")
	} else {
		r, e := arg.GetMaxSessionDuration()
		*reply, err = float64(r), e
	}
	return
}

// Returns MaxSessionTime for an event received in SessionManager, considering DerivedCharging for it
func (rs *Responder) GetDerivedMaxSessionTime(ev *StoredCdr, reply *float64) error {
	cacheKey := "GetDerivedMaxSessionTime" + ev.CgrId
	if item, err := rs.getCache().Get(cacheKey); err == nil && item != nil {
		*reply = item.Value.(float64)
		return item.Err
	}
	if rs.Bal != nil {
		err := errors.New("unsupported method on the balancer")
		rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
		return err
	}
	if ev.Subject == "" {
		ev.Subject = ev.Account
	}
	// replace aliases
	if err := LoadAlias(
		&AttrMatchingAlias{
			Destination: ev.Destination,
			Direction:   ev.Direction,
			Tenant:      ev.Tenant,
			Category:    ev.Category,
			Account:     ev.Account,
			Subject:     ev.Subject,
			Context:     utils.ALIAS_CONTEXT_RATING,
		}, ev, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
		return err
	}
	// replace user profile fields
	if err := LoadUserProfile(ev, utils.EXTRA_FIELDS); err != nil {
		rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
		return err
	}
	maxCallDuration := -1.0
	attrsDC := &utils.AttrDerivedChargers{Tenant: ev.GetTenant(utils.META_DEFAULT), Category: ev.GetCategory(utils.META_DEFAULT), Direction: ev.GetDirection(utils.META_DEFAULT),
		Account: ev.GetAccount(utils.META_DEFAULT), Subject: ev.GetSubject(utils.META_DEFAULT)}
	dcs := &utils.DerivedChargers{}
	if err := rs.GetDerivedChargers(attrsDC, dcs); err != nil {
		rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
		return err
	}
	dcs, _ = dcs.AppendDefaultRun()
	for _, dc := range dcs.Chargers {
		if utils.IsSliceMember([]string{utils.META_RATED, utils.RATED}, ev.GetReqType(dc.ReqTypeField)) { // Only consider prepaid and pseudoprepaid for MaxSessionTime
			continue
		}
		runFilters, _ := utils.ParseRSRFields(dc.RunFilters, utils.INFIELD_SEP)
		matchingAllFilters := true
		for _, dcRunFilter := range runFilters {
			if fltrPass, _ := ev.PassesFieldFilter(dcRunFilter); !fltrPass {
				matchingAllFilters = false
				break
			}
		}
		if !matchingAllFilters { // Do not process the derived charger further if not all filters were matched
			continue
		}
		startTime, err := ev.GetSetupTime(utils.META_DEFAULT, rs.Timezone)
		if err != nil {
			rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
			return err
		}
		usage, err := ev.GetDuration(utils.META_DEFAULT)
		if err != nil {
			rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
			return err
		}
		if usage == 0 {
			usage = config.CgrConfig().MaxCallDuration
		}
		cd := &CallDescriptor{
			Direction:   ev.GetDirection(dc.DirectionField),
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
			rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
			return err
		}
		if utils.IsSliceMember([]string{utils.META_POSTPAID, utils.POSTPAID}, ev.GetReqType(dc.ReqTypeField)) {
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
	rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Value: maxCallDuration})
	*reply = maxCallDuration
	return nil
}

// Used by SM to get all the prepaid CallDescriptors attached to a session
func (rs *Responder) GetSessionRuns(ev *StoredCdr, sRuns *[]*SessionRun) error {
	cacheKey := "GetSessionRuns" + ev.CgrId
	if item, err := rs.getCache().Get(cacheKey); err == nil && item != nil {
		*sRuns = item.Value.([]*SessionRun)
		return item.Err
	}
	if rs.Bal != nil {
		err := errors.New("Unsupported method on the balancer")
		rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
		return err
	}
	if ev.Subject == "" {
		ev.Subject = ev.Account
	}
	// replace aliases
	if err := LoadAlias(
		&AttrMatchingAlias{
			Destination: ev.Destination,
			Direction:   ev.Direction,
			Tenant:      ev.Tenant,
			Category:    ev.Category,
			Account:     ev.Account,
			Subject:     ev.Subject,
			Context:     utils.ALIAS_CONTEXT_RATING,
		}, ev, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
		return err
	}
	// replace user profile fields
	if err := LoadUserProfile(ev, utils.EXTRA_FIELDS); err != nil {
		rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
		return err
	}
	attrsDC := &utils.AttrDerivedChargers{Tenant: ev.GetTenant(utils.META_DEFAULT), Category: ev.GetCategory(utils.META_DEFAULT), Direction: ev.GetDirection(utils.META_DEFAULT),
		Account: ev.GetAccount(utils.META_DEFAULT), Subject: ev.GetSubject(utils.META_DEFAULT)}
	dcs := &utils.DerivedChargers{}
	if err := rs.GetDerivedChargers(attrsDC, dcs); err != nil {
		rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
		return err
	}
	dcs, _ = dcs.AppendDefaultRun()
	sesRuns := make([]*SessionRun, 0)
	for _, dc := range dcs.Chargers {
		if !utils.IsSliceMember([]string{utils.META_PREPAID, utils.PREPAID}, ev.GetReqType(dc.ReqTypeField)) {
			continue // We only consider prepaid sessions
		}
		startTime, err := ev.GetAnswerTime(dc.AnswerTimeField, rs.Timezone)
		if err != nil {
			err := errors.New("Error parsing answer event start time")
			rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
			return err
		}
		cd := &CallDescriptor{
			CgrId:       ev.GetCgrId(rs.Timezone),
			Direction:   ev.GetDirection(dc.DirectionField),
			Tenant:      ev.GetTenant(dc.TenantField),
			Category:    ev.GetCategory(dc.CategoryField),
			Subject:     ev.GetSubject(dc.SubjectField),
			Account:     ev.GetAccount(dc.AccountField),
			Destination: ev.GetDestination(dc.DestinationField),
			TimeStart:   startTime,
			ExtraFields: ev.GetExtraFields()}
		sesRuns = append(sesRuns, &SessionRun{DerivedCharger: dc, CallDescriptor: cd})
	}
	*sRuns = sesRuns
	rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Value: sRuns})
	return nil
}

func (rs *Responder) GetDerivedChargers(attrs *utils.AttrDerivedChargers, dcs *utils.DerivedChargers) error {
	if rs.Bal != nil {
		return errors.New("BALANCER_UNSUPPORTED_METHOD")
	}
	if dcsH, err := HandleGetDerivedChargers(ratingStorage, attrs); err != nil {
		return err
	} else if dcsH != nil {
		*dcs = *dcsH
	}
	return nil
}

func (rs *Responder) GetLCR(attrs *AttrGetLcr, reply *LCRCost) error {
	cacheKey := "GetLCR" + attrs.CgrId
	if item, err := rs.getCache().Get(cacheKey); err == nil && item != nil {
		*reply = *(item.Value.(*LCRCost))
		return item.Err
	}
	if attrs.CallDescriptor.Subject == "" {
		attrs.CallDescriptor.Subject = attrs.CallDescriptor.Account
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
			Context:     utils.ALIAS_CONTEXT_RATING,
		}, cd, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
		return err
	}
	// replace user profile fields
	if err := LoadUserProfile(attrs.CallDescriptor, utils.EXTRA_FIELDS); err != nil {
		return err
	}
	lcrCost, err := attrs.CallDescriptor.GetLCR(rs.Stats, attrs.Paginator)
	if err != nil {
		rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
		return err
	}
	if lcrCost.Entry.Strategy == LCR_STRATEGY_LOAD {
		for _, suppl := range lcrCost.SupplierCosts {
			suppl.Cost = -1 // In case of load distribution we don't calculate costs
		}
	}
	rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Value: lcrCost})
	*reply = *lcrCost
	return nil
}

func (rs *Responder) FlushCache(arg *CallDescriptor, reply *float64) (err error) {
	if rs.Bal != nil {
		*reply, err = rs.callMethod(arg, "Responder.FlushCache")
	} else {
		r, e := Guardian.Guard(func() (interface{}, error) {
			return 0, arg.FlushCache()
		}, 0, arg.GetAccountKey())
		*reply, err = r.(float64), e
	}
	return
}

func (rs *Responder) Status(arg string, reply *map[string]interface{}) (err error) {
	memstats := new(runtime.MemStats)
	runtime.ReadMemStats(memstats)
	response := make(map[string]interface{})
	if rs.Bal != nil {
		response["Raters"] = rs.Bal.GetClientAddresses()
	}
	response["memstat"] = memstats.HeapAlloc / 1024
	response["footprint"] = memstats.Sys / 1024
	*reply = response
	return
}

func (rs *Responder) Shutdown(arg string, reply *string) (err error) {
	if rs.Bal != nil {
		rs.Bal.Shutdown("Responder.Shutdown")
	}
	ratingStorage.Close()
	accountingStorage.Close()
	storageLogger.Close()
	cdrStorage.Close()
	defer func() { rs.ExitChan <- true }()
	*reply = "Done!"
	return
}

/*
The function that gets the information from the raters using balancer.
*/
func (rs *Responder) getCallCost(key *CallDescriptor, method string) (reply *CallCost, err error) {
	err = errors.New("") //not nil value
	for err != nil {
		client := rs.Bal.Balance()
		if client == nil {
			utils.Logger.Info("<Balancer> Waiting for raters to register...")
			time.Sleep(1 * time.Second) // wait one second and retry
		} else {
			_, err = Guardian.Guard(func() (interface{}, error) {
				err = client.Call(method, *key, reply)
				return reply, err
			}, 0, key.GetAccountKey())
			if err != nil {
				utils.Logger.Err(fmt.Sprintf("<Balancer> Got en error from rater: %v", err))
			}
		}
	}
	return
}

/*
The function that gets the information from the raters using balancer.
*/
func (rs *Responder) callMethod(key *CallDescriptor, method string) (reply float64, err error) {
	err = errors.New("") //not nil value
	for err != nil {
		client := rs.Bal.Balance()
		if client == nil {
			utils.Logger.Info("Waiting for raters to register...")
			time.Sleep(1 * time.Second) // wait one second and retry
		} else {
			_, err = Guardian.Guard(func() (interface{}, error) {
				err = client.Call(method, *key, &reply)
				return reply, err
			}, 0, key.GetAccountKey())
			if err != nil {
				utils.Logger.Info(fmt.Sprintf("Got en error from rater: %v", err))
			}
		}
	}
	return
}

/*
RPC method that receives a rater address, connects to it and ads the pair to the rater list for balancing
*/
func (rs *Responder) RegisterRater(clientAddress string, replay *int) error {
	utils.Logger.Info(fmt.Sprintf("Started rater %v registration...", clientAddress))
	time.Sleep(2 * time.Second) // wait a second for Rater to start serving
	client, err := rpc.Dial("tcp", clientAddress)
	if err != nil {
		utils.Logger.Err("Could not connect to client!")
		return err
	}
	rs.Bal.AddClient(clientAddress, client)
	utils.Logger.Info(fmt.Sprintf("Rater %v registered succesfully.", clientAddress))
	return nil
}

/*
RPC method that recives a rater addres gets the connections and closes it and removes the pair from rater list.
*/
func (rs *Responder) UnRegisterRater(clientAddress string, replay *int) error {
	client, ok := rs.Bal.GetClient(clientAddress)
	if ok {
		client.Close()
		rs.Bal.RemoveClient(clientAddress)
		utils.Logger.Info(fmt.Sprintf("Rater %v unregistered succesfully.", clientAddress))
	} else {
		utils.Logger.Info(fmt.Sprintf("Server %v was not on my watch!", clientAddress))
	}
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
