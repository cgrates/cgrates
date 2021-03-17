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
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type AttrGetCost struct {
	Tenant      string
	Category    string
	Subject     string
	AnswerTime  string
	Destination string
	Usage       string
	Opts        map[string]interface{}
}

func (apierSv1 *APIerSv1) GetCost(attrs *AttrGetCost, ec *engine.EventCost) error {
	if apierSv1.Responder == nil {
		return utils.NewErrNotConnected(utils.RALService)
	}
	usage, err := utils.ParseDurationWithNanosecs(attrs.Usage)
	if err != nil {
		return err
	}
	aTime, err := utils.ParseTimeDetectLayout(attrs.AnswerTime,
		apierSv1.Config.GeneralCfg().DefaultTimezone)
	if err != nil {
		return err
	}

	cd := &engine.CallDescriptor{
		Category:      attrs.Category,
		Tenant:        utils.FirstNonEmpty(attrs.Tenant, apierSv1.Config.GeneralCfg().DefaultTenant),
		Subject:       attrs.Subject,
		Destination:   attrs.Destination,
		TimeStart:     aTime,
		TimeEnd:       aTime.Add(usage),
		DurationIndex: usage,
	}
	var cc engine.CallCost
	if err := apierSv1.Responder.GetCost(
		&engine.CallDescriptorWithAPIOpts{
			CallDescriptor: cd,
			APIOpts:        attrs.Opts,
		}, &cc); err != nil {
		return utils.NewErrServerError(err)
	}
	*ec = *engine.NewEventCostFromCallCost(&cc, "", "")
	ec.Compute()
	return nil
}

type AttrGetDataCost struct {
	Tenant     string
	Category   string
	Subject    string
	AnswerTime string
	Usage      time.Duration // the call duration so far (till TimeEnd)
	Opts       map[string]interface{}
}

func (apierSv1 *APIerSv1) GetDataCost(attrs *AttrGetDataCost, reply *engine.DataCost) error {
	if apierSv1.Responder == nil {
		return utils.NewErrNotConnected(utils.RALService)
	}
	aTime, err := utils.ParseTimeDetectLayout(attrs.AnswerTime,
		apierSv1.Config.GeneralCfg().DefaultTimezone)
	if err != nil {
		return err
	}
	cd := &engine.CallDescriptor{
		Category:      attrs.Category,
		Tenant:        utils.FirstNonEmpty(attrs.Tenant, apierSv1.Config.GeneralCfg().DefaultTenant),
		Subject:       attrs.Subject,
		TimeStart:     aTime,
		TimeEnd:       aTime.Add(attrs.Usage),
		DurationIndex: attrs.Usage,
		ToR:           utils.MetaData,
	}
	var cc engine.CallCost
	if err := apierSv1.Responder.GetCost(&engine.CallDescriptorWithAPIOpts{
		CallDescriptor: cd,
		APIOpts:        attrs.Opts,
	}, &cc); err != nil {
		return utils.NewErrServerError(err)
	}
	if dc, err := cc.ToDataCost(); err != nil {
		return utils.NewErrServerError(err)
	} else if dc != nil {
		*reply = *dc
	}
	return nil
}
