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
	if !rs.usageAllowed(arg.TOR, arg.GetDuration()) {
		return utils.ErrMaxUsageExceeded
	}
	r, e := guardian.Guardian.Guard(func() (interface{}, error) {
		return arg.GetCost()
	}, config.CgrConfig().GeneralCfg().LockingTimeout, arg.GetAccountKey())
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
	if !rs.usageAllowed(arg.TOR, arg.GetDuration()) {
		return utils.ErrMaxUsageExceeded
	}
	r, e := arg.GetMaxSessionDuration()
	*reply, err = r, e
	return
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
	response[utils.Version] = utils.GetCGRVersion()
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
