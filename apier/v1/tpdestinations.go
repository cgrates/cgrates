// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// SetTPDestination creates a new destination within a tariff plan
func (apierSv1 *APIerSv1) SetTPDestination(ctx *context.Context, attrs *utils.TPDestination, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID, utils.Prefixes}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierSv1.StorDb.SetTPDestinations([]*utils.TPDestination{attrs}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPDestination struct {
	TPid string // Tariff plan id
	ID   string // Destination id
}

// GetTPDestination queries a specific destination
func (apierSv1 *APIerSv1) GetTPDestination(ctx *context.Context, attrs *AttrGetTPDestination, reply *utils.TPDestination) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tpDsts, err := apierSv1.StorDb.GetTPDestinations(attrs.TPid, attrs.ID)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	if len(tpDsts) == 0 {
		return utils.ErrNotFound
	}
	tpDst := tpDsts[0]
	*reply = utils.TPDestination{TPid: tpDst.TPid,
		ID: tpDst.ID, Prefixes: tpDst.Prefixes}
	return nil
}

type AttrGetTPDestinationIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPDestinationIDs queries destination identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPDestinationIDs(ctx *context.Context, attrs *AttrGetTPDestinationIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPDestinations,
		utils.TPDistinctIds{utils.TagCfg}, nil, &attrs.PaginatorWithSearch)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	if ids == nil {
		return utils.ErrNotFound
	}
	*reply = ids
	return nil
}

// RemoveTPDestination removes specific Destination on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPDestination(ctx *context.Context, attrs *AttrGetTPDestination, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierSv1.StorDb.RemTpData(utils.TBLTPDestinations, attrs.TPid, map[string]string{utils.TagCfg: attrs.ID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
