// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// SetTPTiming creates a new timing within a tariff plan
func (apierSv1 *APIerSv1) SetTPTiming(ctx *context.Context, attrs *utils.ApierTPTiming, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID, utils.YearsFieldName, utils.MonthsFieldName, utils.MonthDaysFieldName, utils.WeekDaysFieldName, utils.Time}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierSv1.StorDb.SetTPTimings([]*utils.ApierTPTiming{attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPTiming struct {
	TPid string // Tariff plan id
	ID   string // Timing id
}

// GetTPTiming queries specific Timing on Tariff plan
func (apierSv1 *APIerSv1) GetTPTiming(ctx *context.Context, attrs *AttrGetTPTiming, reply *utils.ApierTPTiming) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tms, err := apierSv1.StorDb.GetTPTimings(attrs.TPid, attrs.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *tms[0]
	return nil
}

type AttrGetTPTimingIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPTimingIds queries timing identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPTimingIds(ctx *context.Context, attrs *AttrGetTPTimingIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPTimings,
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

// RemoveTPTiming removes specific Timing on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPTiming(ctx *context.Context, attrs AttrGetTPTiming, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierSv1.StorDb.RemTpData(utils.TBLTPTimings, attrs.TPid, map[string]string{utils.TagCfg: attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
