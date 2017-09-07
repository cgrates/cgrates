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
package v1

import (
	"reflect"
	"strings"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/stats"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewStatSV1 initializes StatSV1
func NewStatSV1(sts *stats.StatService) *StatSV1 {
	return &StatSV1{sts: sts}
}

// Exports RPC from RLs
type StatSV1 struct {
	sts *stats.StatService
}

// Call implements rpcclient.RpcClientConnection interface for internal RPC
func (stsv1 *StatSV1) Call(serviceMethod string, args interface{}, reply interface{}) error {
	methodSplit := strings.Split(serviceMethod, ".")
	if len(methodSplit) != 2 {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	method := reflect.ValueOf(stsv1).MethodByName(methodSplit[1])
	if !method.IsValid() {
		return rpcclient.ErrUnsupporteServiceMethod
	}
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

// ProcessEvent returns processes a new Event
func (stsv1 *StatSV1) ProcessEvent(ev engine.StatsEvent, reply *string) error {
	return stsv1.sts.V1ProcessEvent(ev, reply)
}

// GetQueueIDs returns the list of queues IDs in the system
func (stsv1 *StatSV1) GetQueueIDs(ignored struct{}, reply *[]string) (err error) {
	return stsv1.sts.V1GetQueueIDs(ignored, reply)
}

// GetStatMetrics returns the metrics for a queueID
func (stsv1 *StatSV1) GetStringMetrics(queueID string, reply *map[string]string) (err error) {
	return stsv1.sts.V1GetStringMetrics(queueID, reply)
}

// GetStatMetrics returns the metrics for a queueID
func (stsv1 *StatSV1) GetFloatMetrics(queueID string, reply *map[string]float64) (err error) {
	return stsv1.sts.V1GetFloatMetrics(queueID, reply)
}

// LoadQueues loads from dataDB into statsService the queueIDs specified
// loads all when qIDs is nil
func (stsv1 *StatSV1) LoadQueues(args stats.ArgsLoadQueues, reply *string) (err error) {
	return stsv1.sts.V1LoadQueues(args, reply)
}

type AttrGetStatsCfg struct {
	ID string
}

//GetStatConfig returns a stat configuration
func (apierV1 *ApierV1) GetStatQueueProfile(attr AttrGetStatsCfg, reply *engine.StatQueueProfile) error {
	if missing := utils.MissingStructFields(&attr, []string{"ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if sCfg, err := apierV1.DataDB.GetStatQueueProfile(attr.ID); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *sCfg
	}
	return nil
}

//SetStatConfig add a new stat configuration
func (apierV1 *ApierV1) SetStatQueueProfile(attr *engine.StatQueueProfile, reply *string) error {
	if missing := utils.MissingStructFields(attr, []string{"ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataDB.SetStatQueueProfile(attr); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil

}

//Remove a specific stat configuration
func (apierV1 *ApierV1) RemStatQueueProfile(attrs AttrGetStatsCfg, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierV1.DataDB.RemStatQueueProfile(attrs.ID); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = utils.OK
	return nil

}
