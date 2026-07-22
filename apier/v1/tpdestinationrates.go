// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

// This file deals with tp_destination_rates management over APIs

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// SetTPDestinationRate creates a new DestinationRate profile within a tariff plan
func (apierSv1 *APIerSv1) SetTPDestinationRate(ctx *context.Context, attrs *utils.TPDestinationRate, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID, utils.DestinationRates}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierSv1.StorDb.SetTPDestinationRates([]*utils.TPDestinationRate{attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPDestinationRate struct {
	TPid string // Tariff plan id
	ID   string // Rate id
	utils.Paginator
}

// GetTPDestinationRate queries specific DestinationRate profile on tariff plan
func (apierSv1 *APIerSv1) GetTPDestinationRate(ctx *context.Context, attrs *AttrGetTPDestinationRate, reply *utils.TPDestinationRate) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	drs, err := apierSv1.StorDb.GetTPDestinationRates(attrs.TPid, attrs.ID, &attrs.Paginator)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *drs[0]
	return nil
}

type AttrTPDestinationRateIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPDestinationRateIds queries DestinationRate identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPDestinationRateIds(ctx *context.Context, attrs *AttrGetTPRateIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPDestinationRates,
		utils.TPDistinctIds{utils.TagCfg}, nil, &attrs.PaginatorWithSearch)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = ids
	return nil
}

// RemoveTPDestinationRate removes specific DestinationRate on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPDestinationRate(ctx *context.Context, attrs *AttrGetTPDestinationRate, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierSv1.StorDb.RemTpData(utils.TBLTPDestinationRates, attrs.TPid, map[string]string{utils.TagCfg: attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
