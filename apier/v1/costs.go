// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	*utils.ArgDispatcher
}

func (apier *APIerSv1) GetCost(attrs AttrGetCost, ec *engine.EventCost) error {
	usage, err := utils.ParseDurationWithNanosecs(attrs.Usage)
	if err != nil {
		return err
	}
	aTime, err := utils.ParseTimeDetectLayout(attrs.AnswerTime,
		apier.Config.GeneralCfg().DefaultTimezone)
	if err != nil {
		return err
	}

	cd := &engine.CallDescriptor{
		Category:      attrs.Category,
		Tenant:        attrs.Tenant,
		Subject:       attrs.Subject,
		Destination:   attrs.Destination,
		TimeStart:     aTime,
		TimeEnd:       aTime.Add(usage),
		DurationIndex: usage,
	}
	var cc engine.CallCost
	if err := apier.Responder.GetCost(&engine.CallDescriptorWithArgDispatcher{CallDescriptor: cd,
		ArgDispatcher: attrs.ArgDispatcher}, &cc); err != nil {
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
	*utils.ArgDispatcher
}

func (apier *APIerSv1) GetDataCost(attrs AttrGetDataCost, reply *engine.DataCost) error {
	aTime, err := utils.ParseTimeDetectLayout(attrs.AnswerTime,
		apier.Config.GeneralCfg().DefaultTimezone)
	if err != nil {
		return err
	}
	cd := &engine.CallDescriptor{
		Category:      attrs.Category,
		Tenant:        attrs.Tenant,
		Subject:       attrs.Subject,
		TimeStart:     aTime,
		TimeEnd:       aTime.Add(attrs.Usage),
		DurationIndex: attrs.Usage,
		ToR:           utils.DATA,
	}
	var cc engine.CallCost
	if err := apier.Responder.GetCost(&engine.CallDescriptorWithArgDispatcher{CallDescriptor: cd,
		ArgDispatcher: attrs.ArgDispatcher}, &cc); err != nil {
		return utils.NewErrServerError(err)
	}
	if dc, err := cc.ToDataCost(); err != nil {
		return utils.NewErrServerError(err)
	} else if dc != nil {
		*reply = *dc
	}
	return nil
}
