// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"time"

	"github.com/cgrates/birpc/context"
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
	APIOpts     map[string]any
}

func (apierSv1 *APIerSv1) GetCost(ctx *context.Context, attrs *AttrGetCost, ec *engine.EventCost) error {
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
	if err := apierSv1.Responder.GetCost(context.Background(),
		&engine.CallDescriptorWithAPIOpts{
			CallDescriptor: cd,
			APIOpts:        attrs.APIOpts,
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
	Opts       map[string]any
}

func (apierSv1 *APIerSv1) GetDataCost(ctx *context.Context, attrs *AttrGetDataCost, reply *engine.DataCost) error {
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
	if err := apierSv1.Responder.GetCost(
		context.Background(),
		&engine.CallDescriptorWithAPIOpts{
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

// GetAccountCost returns a simulated cost of an account without debiting from it (dryrun)
func (apierSv1 *APIerSv1) GetAccountCost(ctx *context.Context, args *utils.CGREvent, ec *engine.EventCost) (err error) {
	cd, err := engine.NewCallDescriptorFromCGREvent(args, apierSv1.Config.GeneralCfg().DefaultTimezone)
	if err != nil {
		return
	}
	cd.DryRun = true
	var cc engine.CallCost
	if err := apierSv1.Responder.Debit(context.Background(),
		&engine.CallDescriptorWithAPIOpts{
			CallDescriptor: cd,
			APIOpts:        args.APIOpts,
		}, &cc); err != nil {
		return utils.NewErrServerError(err)
	}
	*ec = *engine.NewEventCostFromCallCost(&cc, cd.CgrID, cd.RunID)
	ec.Compute()
	return nil
}
