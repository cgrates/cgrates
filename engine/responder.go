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
	CdrSrv        *CdrServer
	Stats         StatsInterface
	Timeout       time.Duration
	Timezone      string
	cnt           int64
	responseCache *cache2go.ResponseCache
}

func NewResponder(exitChan chan bool, cdrSrv *CdrServer, stats StatsInterface, timeout, timeToLive time.Duration) *Responder {
	return &Responder{
		ExitChan:      exitChan,
		Stats:         stats,
		Timeout:       timeToLive,
		responseCache: cache2go.NewResponseCache(timeToLive),
	}
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

func (rs *Responder) FakeDebit(arg *CallDescriptor, reply *CallCost) (err error) {
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
		r, e := rs.getCallCost(arg, "Responder.FakeDebit")
		*reply, err = *r, e
	} else {
		r, e := arg.FakeDebit()
		if e != nil {
			return e
		} else if r != nil {
			*reply = *r
		}
	}
	return
}

func (rs *Responder) MaxDebit(arg *CallDescriptor, reply *CallCost) (err error) {
	if item, err := rs.getCache().Get(utils.MAX_DEBIT_CACHE_PREFIX + arg.CgrId); err == nil && item != nil {
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
		return err
	}
	// replace user profile fields
	if err := LoadUserProfile(arg, utils.EXTRA_FIELDS); err != nil {
		return err
	}
	if rs.Bal != nil {
		r, e := rs.getCallCost(arg, "Responder.MaxDebit")
		*reply, err = *r, e
	} else {
		r, e := arg.MaxDebit()
		if e != nil {
			rs.getCache().Cache(utils.MAX_DEBIT_CACHE_PREFIX+arg.CgrId, &cache2go.CacheItem{
				Err: e,
			})
			return e
		} else if r != nil {
			*reply = *r
		}
	}
	rs.getCache().Cache(utils.MAX_DEBIT_CACHE_PREFIX+arg.CgrId, &cache2go.CacheItem{
		Value: reply,
		Err:   err,
	})
	return
}

func (rs *Responder) RefundIncrements(arg *CallDescriptor, reply *float64) (err error) {
	if item, err := rs.getCache().Get(utils.REFUND_INCR_CACHE_PREFIX + arg.CgrId); err == nil && item != nil {
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
		return err
	}
	// replace user profile fields
	if err := LoadUserProfile(arg, utils.EXTRA_FIELDS); err != nil {
		return err
	}
	if rs.Bal != nil {
		*reply, err = rs.callMethod(arg, "Responder.RefundIncrements")
	} else {
		*reply, err = arg.RefundIncrements()
	}
	rs.getCache().Cache(utils.REFUND_INCR_CACHE_PREFIX+arg.CgrId, &cache2go.CacheItem{
		Value: reply,
		Err:   err,
	})
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
func (rs *Responder) GetDerivedMaxSessionTime(ev *CDR, reply *float64) error {
	if rs.Bal != nil {
		return errors.New("unsupported method on the balancer")
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
		return err
	}
	// replace user profile fields
	if err := LoadUserProfile(ev, utils.EXTRA_FIELDS); err != nil {
		return err
	}
	maxCallDuration := -1.0
	attrsDC := &utils.AttrDerivedChargers{Tenant: ev.GetTenant(utils.META_DEFAULT), Category: ev.GetCategory(utils.META_DEFAULT), Direction: ev.GetDirection(utils.META_DEFAULT),
		Account: ev.GetAccount(utils.META_DEFAULT), Subject: ev.GetSubject(utils.META_DEFAULT)}
	dcs := &utils.DerivedChargers{}
	if err := rs.GetDerivedChargers(attrsDC, dcs); err != nil {
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
			return err
		}
		usage, err := ev.GetDuration(utils.META_DEFAULT)
		if err != nil {
			return err
		}
		if usage == 0 {
			usage = config.CgrConfig().MaxCallDuration
		}
		cd := &CallDescriptor{
			CgrId:       ev.GetCgrId(rs.Timezone),
			TOR:         ev.ToR,
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
	return nil
}

// Used by SM to get all the prepaid CallDescriptors attached to a session
func (rs *Responder) GetSessionRuns(ev *CDR, sRuns *[]*SessionRun) error {
	if rs.Bal != nil {
		return errors.New("Unsupported method on the balancer")
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
		return err
	}
	// replace user profile fields
	if err := LoadUserProfile(ev, utils.EXTRA_FIELDS); err != nil {
		return err
	}
	attrsDC := &utils.AttrDerivedChargers{Tenant: ev.GetTenant(utils.META_DEFAULT), Category: ev.GetCategory(utils.META_DEFAULT), Direction: ev.GetDirection(utils.META_DEFAULT),
		Account: ev.GetAccount(utils.META_DEFAULT), Subject: ev.GetSubject(utils.META_DEFAULT)}
	dcs := &utils.DerivedChargers{}
	if err := rs.GetDerivedChargers(attrsDC, dcs); err != nil {
		rs.getCache().Cache(utils.GET_SESS_RUNS_CACHE_PREFIX+ev.CGRID, &cache2go.CacheItem{
			Err: err,
		})
		return err
	}
	dcs, _ = dcs.AppendDefaultRun()
	sesRuns := make([]*SessionRun, 0)
	for _, dc := range dcs.Chargers {
		if !utils.IsSliceMember([]string{utils.META_PREPAID, utils.PREPAID}, ev.GetReqType(dc.RequestTypeField)) {
			continue // We only consider prepaid sessions
		}
		startTime, err := ev.GetAnswerTime(dc.AnswerTimeField, rs.Timezone)
		if err != nil {
			rs.getCache().Cache(utils.GET_SESS_RUNS_CACHE_PREFIX+ev.CGRID, &cache2go.CacheItem{
				Err: err,
			})
			return errors.New("Error parsing answer event start time")
		}
		endTime, err := ev.GetEndTime("", rs.Timezone)
		if err != nil {
			rs.getCache().Cache(utils.GET_SESS_RUNS_CACHE_PREFIX+ev.CGRID, &cache2go.CacheItem{
				Err: err,
			})
			return errors.New("Error parsing answer event end time")
		}
		cd := &CallDescriptor{
			CgrId:       ev.GetCgrId(rs.Timezone),
			TOR:         ev.ToR,
			Direction:   ev.GetDirection(dc.DirectionField),
			Tenant:      ev.GetTenant(dc.TenantField),
			Category:    ev.GetCategory(dc.CategoryField),
			Subject:     ev.GetSubject(dc.SubjectField),
			Account:     ev.GetAccount(dc.AccountField),
			Destination: ev.GetDestination(dc.DestinationField),
			TimeStart:   startTime,
			endTime:     endTime,
			ExtraFields: ev.GetExtraFields()}
		sesRuns = append(sesRuns, &SessionRun{DerivedCharger: dc, CallDescriptor: cd})
	}
	*sRuns = sesRuns
	rs.getCache().Cache(utils.GET_SESS_RUNS_CACHE_PREFIX+ev.CGRID, &cache2go.CacheItem{
		Value: sRuns,
	})
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

func (rs *Responder) ProcessCdr(cdr *CDR, reply *string) error {
	if rs.CdrSrv == nil {
		return errors.New("CDR_SERVER_NOT_RUNNING")
	}
	if err := rs.CdrSrv.ProcessCdr(cdr); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rs *Responder) LogCallCost(ccl *CallCostLog, reply *string) error {
	if item, err := rs.getCache().Get(utils.LOG_CALL_COST_CACHE_PREFIX + ccl.CgrId); err == nil && item != nil {
		*reply = item.Value.(string)
		return item.Err
	}
	if rs.CdrSrv == nil {
		err := errors.New("CDR_SERVER_NOT_RUNNING")
		rs.getCache().Cache(utils.LOG_CALL_COST_CACHE_PREFIX+ccl.CgrId, &cache2go.CacheItem{
			Err: err,
		})
		return err
	}
	if err := rs.CdrSrv.LogCallCost(ccl); err != nil {
		rs.getCache().Cache(utils.LOG_CALL_COST_CACHE_PREFIX+ccl.CgrId, &cache2go.CacheItem{
			Err: err,
		})
		return err
	}
	*reply = utils.OK
	rs.getCache().Cache(utils.LOG_CALL_COST_CACHE_PREFIX+ccl.CgrId, &cache2go.CacheItem{
		Value: utils.OK,
	})
	return nil
}

func (rs *Responder) GetLCR(attrs *AttrGetLcr, reply *LCRCost) error {
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
		return err
	}
	// replace user profile fields
	if err := LoadUserProfile(attrs.CallDescriptor, utils.EXTRA_FIELDS); err != nil {
		return err
	}
	lcrCost, err := attrs.CallDescriptor.GetLCR(rs.Stats, attrs.Paginator)
	if err != nil {
		return err
	}
	if lcrCost.Entry.Strategy == LCR_STRATEGY_LOAD {
		for _, suppl := range lcrCost.SupplierCosts {
			suppl.Cost = -1 // In case of load distribution we don't calculate costs
		}
	}
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

func (rs *Responder) GetTimeout(i int, d *time.Duration) error {
	*d = rs.Timeout
	return nil
}

// Reflection worker type for not standalone balancer
type ResponderWorker struct{}

func (rw *ResponderWorker) Call(serviceMethod string, args interface{}, reply interface{}) error {
	methodName := strings.TrimLeft(serviceMethod, "Responder.")
	switch args.(type) {
	case CallDescriptor:
		cd := args.(CallDescriptor)
		switch reply.(type) {
		case *CallCost:
			rep := reply.(*CallCost)
			method := reflect.ValueOf(&cd).MethodByName(methodName)
			ret := method.Call([]reflect.Value{})
			*rep = *(ret[0].Interface().(*CallCost))
		case *float64:
			rep := reply.(*float64)
			method := reflect.ValueOf(&cd).MethodByName(methodName)
			ret := method.Call([]reflect.Value{})
			*rep = *(ret[0].Interface().(*float64))
		}
	case string:
		switch methodName {
		case "Status":
			*(reply.(*string)) = "Local!"
		case "Shutdown":
			*(reply.(*string)) = "Done!"
		}

	}
	return nil
}

func (rw *ResponderWorker) Close() error {
	return nil
}

type Connector interface {
	GetCost(*CallDescriptor, *CallCost) error
	Debit(*CallDescriptor, *CallCost) error
	MaxDebit(*CallDescriptor, *CallCost) error
	RefundIncrements(*CallDescriptor, *float64) error
	GetMaxSessionTime(*CallDescriptor, *float64) error
	GetDerivedChargers(*utils.AttrDerivedChargers, *utils.DerivedChargers) error
	GetDerivedMaxSessionTime(*CDR, *float64) error
	GetSessionRuns(*CDR, *[]*SessionRun) error
	ProcessCdr(*CDR, *string) error
	LogCallCost(*CallCostLog, *string) error
	GetLCR(*AttrGetLcr, *LCRCost) error
	GetTimeout(int, *time.Duration) error
}

type RPCClientConnector struct {
	Client  *rpcclient.RpcClient
	Timeout time.Duration
}

func (rcc *RPCClientConnector) GetCost(cd *CallDescriptor, cc *CallCost) error {
	return rcc.Client.Call("Responder.GetCost", cd, cc)
}

func (rcc *RPCClientConnector) Debit(cd *CallDescriptor, cc *CallCost) error {
	return rcc.Client.Call("Responder.Debit", cd, cc)
}

func (rcc *RPCClientConnector) MaxDebit(cd *CallDescriptor, cc *CallCost) error {
	return rcc.Client.Call("Responder.MaxDebit", cd, cc)
}

func (rcc *RPCClientConnector) RefundIncrements(cd *CallDescriptor, resp *float64) error {
	return rcc.Client.Call("Responder.RefundIncrements", cd, resp)
}

func (rcc *RPCClientConnector) GetMaxSessionTime(cd *CallDescriptor, resp *float64) error {
	return rcc.Client.Call("Responder.GetMaxSessionTime", cd, resp)
}

func (rcc *RPCClientConnector) GetDerivedMaxSessionTime(ev *CDR, reply *float64) error {
	return rcc.Client.Call("Responder.GetDerivedMaxSessionTime", ev, reply)
}

func (rcc *RPCClientConnector) GetSessionRuns(ev *CDR, sRuns *[]*SessionRun) error {
	return rcc.Client.Call("Responder.GetSessionRuns", ev, sRuns)
}

func (rcc *RPCClientConnector) GetDerivedChargers(attrs *utils.AttrDerivedChargers, dcs *utils.DerivedChargers) error {
	return rcc.Client.Call("ApierV1.GetDerivedChargers", attrs, dcs)
}

func (rcc *RPCClientConnector) ProcessCdr(cdr *CDR, reply *string) error {
	return rcc.Client.Call("CdrsV1.ProcessCdr", cdr, reply)
}

func (rcc *RPCClientConnector) LogCallCost(ccl *CallCostLog, reply *string) error {
	return rcc.Client.Call("CdrsV1.LogCallCost", ccl, reply)
}

func (rcc *RPCClientConnector) GetLCR(attrs *AttrGetLcr, reply *LCRCost) error {
	return rcc.Client.Call("Responder.GetLCR", attrs, reply)
}

func (rcc *RPCClientConnector) GetTimeout(i int, d *time.Duration) error {
	*d = rcc.Timeout
	return nil
}

type ConnectorPool []Connector

func (cp ConnectorPool) GetCost(cd *CallDescriptor, cc *CallCost) error {
	for _, con := range cp {
		c := make(chan error, 1)
		callCost := &CallCost{}

		var timeout time.Duration
		con.GetTimeout(0, &timeout)

		go func() { c <- con.GetCost(cd, callCost) }()
		select {
		case err := <-c:
			*cc = *callCost
			return err
		case <-time.After(timeout):
			// call timed out, continue
		}
	}
	return utils.ErrTimedOut
}

func (cp ConnectorPool) Debit(cd *CallDescriptor, cc *CallCost) error {
	for _, con := range cp {
		c := make(chan error, 1)
		callCost := &CallCost{}

		var timeout time.Duration
		con.GetTimeout(0, &timeout)

		go func() { c <- con.Debit(cd, callCost) }()
		select {
		case err := <-c:
			*cc = *callCost
			return err
		case <-time.After(timeout):
			// call timed out, continue
		}
	}
	return utils.ErrTimedOut
}

func (cp ConnectorPool) MaxDebit(cd *CallDescriptor, cc *CallCost) error {
	for _, con := range cp {
		c := make(chan error, 1)
		callCost := &CallCost{}

		var timeout time.Duration
		con.GetTimeout(0, &timeout)

		go func() { c <- con.MaxDebit(cd, callCost) }()
		select {
		case err := <-c:
			*cc = *callCost
			return err
		case <-time.After(timeout):
			// call timed out, continue
		}
	}
	return utils.ErrTimedOut
}

func (cp ConnectorPool) RefundIncrements(cd *CallDescriptor, resp *float64) error {
	for _, con := range cp {
		c := make(chan error, 1)
		var r float64

		var timeout time.Duration
		con.GetTimeout(0, &timeout)

		go func() { c <- con.RefundIncrements(cd, &r) }()
		select {
		case err := <-c:
			*resp = r
			return err
		case <-time.After(timeout):
			// call timed out, continue
		}
	}
	return utils.ErrTimedOut
}

func (cp ConnectorPool) GetMaxSessionTime(cd *CallDescriptor, resp *float64) error {
	for _, con := range cp {
		c := make(chan error, 1)
		var r float64

		var timeout time.Duration
		con.GetTimeout(0, &timeout)

		go func() { c <- con.GetMaxSessionTime(cd, &r) }()
		select {
		case err := <-c:
			*resp = r
			return err
		case <-time.After(timeout):
			// call timed out, continue
		}
	}
	return utils.ErrTimedOut
}

func (cp ConnectorPool) GetDerivedMaxSessionTime(ev *CDR, reply *float64) error {
	for _, con := range cp {
		c := make(chan error, 1)
		var r float64

		var timeout time.Duration
		con.GetTimeout(0, &timeout)

		go func() { c <- con.GetDerivedMaxSessionTime(ev, &r) }()
		select {
		case err := <-c:
			*reply = r
			return err
		case <-time.After(timeout):
			// call timed out, continue
		}
	}
	return utils.ErrTimedOut
}

func (cp ConnectorPool) GetSessionRuns(ev *CDR, sRuns *[]*SessionRun) error {
	for _, con := range cp {
		c := make(chan error, 1)
		sr := make([]*SessionRun, 0)

		var timeout time.Duration
		con.GetTimeout(0, &timeout)

		go func() { c <- con.GetSessionRuns(ev, &sr) }()
		select {
		case err := <-c:
			*sRuns = sr
			return err
		case <-time.After(timeout):
			// call timed out, continue
		}
	}
	return utils.ErrTimedOut
}

func (cp ConnectorPool) GetDerivedChargers(attrs *utils.AttrDerivedChargers, dcs *utils.DerivedChargers) error {
	for _, con := range cp {
		c := make(chan error, 1)
		derivedChargers := utils.DerivedChargers{}

		var timeout time.Duration
		con.GetTimeout(0, &timeout)

		go func() { c <- con.GetDerivedChargers(attrs, &derivedChargers) }()
		select {
		case err := <-c:
			*dcs = derivedChargers
			return err
		case <-time.After(timeout):
			// call timed out, continue
		}
	}
	return utils.ErrTimedOut
}

func (cp ConnectorPool) ProcessCdr(cdr *CDR, reply *string) error {
	for _, con := range cp {
		c := make(chan error, 1)
		var r string

		var timeout time.Duration
		con.GetTimeout(0, &timeout)

		go func() { c <- con.ProcessCdr(cdr, &r) }()
		select {
		case err := <-c:
			*reply = r
			return err
		case <-time.After(timeout):
			// call timed out, continue
		}
	}
	return utils.ErrTimedOut
}

func (cp ConnectorPool) LogCallCost(ccl *CallCostLog, reply *string) error {
	for _, con := range cp {
		c := make(chan error, 1)
		var r string

		var timeout time.Duration
		con.GetTimeout(0, &timeout)

		go func() { c <- con.LogCallCost(ccl, &r) }()
		select {
		case err := <-c:
			*reply = r
			return err
		case <-time.After(timeout):
			// call timed out, continue
		}
	}
	return utils.ErrTimedOut
}

func (cp ConnectorPool) GetLCR(attr *AttrGetLcr, reply *LCRCost) error {
	for _, con := range cp {
		c := make(chan error, 1)
		lcrCost := &LCRCost{}

		var timeout time.Duration
		con.GetTimeout(0, &timeout)

		go func() { c <- con.GetLCR(attr, lcrCost) }()
		select {
		case err := <-c:
			*reply = *lcrCost
			return err
		case <-time.After(timeout):
			// call timed out, continue
		}
	}
	return utils.ErrTimedOut
}

func (cp ConnectorPool) GetTimeout(i int, d *time.Duration) error {
	*d = 0
	return nil
}
