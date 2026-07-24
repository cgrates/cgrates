// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

// This file deals with tp_rate_profiles management over APIs

import (
	"github.com/cgrates/cgrates/utils"
)

// SetTPRatingProfile creates a new RatingProfile within a tariff plan
func (api *APIerSv1) SetTPRatingProfile(attrs utils.TPRatingProfile, reply *string) error {
	if missing := utils.MissingStructFields(&attrs,
		[]string{"TPid", "LoadId", "Tenant", "Category", "Subject", "RatingPlanActivations"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := api.StorDb.SetTPRatingProfiles([]*utils.TPRatingProfile{&attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// GetTPRatingProfilesByLoadID queries specific RatingProfile on tariff plan
func (api *APIerSv1) GetTPRatingProfilesByLoadID(attrs utils.TPRatingProfile, reply *[]*utils.TPRatingProfile) error {
	mndtryFlds := []string{"TPid", "LoadId"}
	if len(attrs.Subject) != 0 { // If Subject provided as filter, make all related fields mandatory
		mndtryFlds = append(mndtryFlds, "Tenant", "Category", "Subject")
	}
	if missing := utils.MissingStructFields(&attrs, mndtryFlds); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	rps, err := api.StorDb.GetTPRatingProfiles(&attrs)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = rps
	return nil
}

// GetTPRatingProfileLoadIds queries RatingProfile identities on specific tariff plan.
func (api *APIerSv1) GetTPRatingProfileLoadIds(attrs utils.AttrTPRatingProfileIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := api.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPRateProfiles,
		utils.TPDistinctIds{"loadid"}, map[string]string{
			"tenant":   attrs.Tenant,
			"category": attrs.Category,
			"subject":  attrs.Subject,
		}, new(utils.PaginatorWithSearch))
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = ids
	return nil
}

// AttrGetTPRatingProfile arguments used by GetTPRatingProfile and RemoveTPRatingProfile
type AttrGetTPRatingProfile struct {
	TPid            string // Tariff plan id
	RatingProfileID string // RatingProfile id
}

// GetTPRatingProfile queries specific RatingProfile on tariff plan
func (api *APIerSv1) GetTPRatingProfile(attrs AttrGetTPRatingProfile, reply *utils.TPRatingProfile) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RatingProfileID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tmpRpf := &utils.TPRatingProfile{TPid: attrs.TPid}
	if err := tmpRpf.SetRatingProfileID(attrs.RatingProfileID); err != nil {
		return err
	}
	rpfs, err := api.StorDb.GetTPRatingProfiles(tmpRpf)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *rpfs[0]
	return nil
}

// AttrGetTPRatingProfileIds arguments used by GetTPRatingProfileIds
type AttrGetTPRatingProfileIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPRatingProfileIds queries RatingProfiles identities on specific tariff plan.
func (api *APIerSv1) GetTPRatingProfileIds(attrs AttrGetTPRatingProfileIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := api.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPRateProfiles,
		utils.TPDistinctIds{"loadid", "tenant", "category", "subject"},
		nil, &attrs.PaginatorWithSearch)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = ids
	return nil
}

// RemoveTPRatingProfile removes specific RatingProfiles on Tariff plan
func (api *APIerSv1) RemoveTPRatingProfile(attrs AttrGetTPRatingProfile, reply *string) (err error) {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RatingProfileID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tmpRpf := new(utils.TPRatingProfile)
	if err = tmpRpf.SetRatingProfileID(attrs.RatingProfileID); err != nil {
		return
	}
	err = api.StorDb.RemTpData(utils.TBLTPRateProfiles,
		attrs.TPid, map[string]string{
			"loadid":   tmpRpf.LoadId,
			"tenant":   tmpRpf.Tenant,
			"category": tmpRpf.Category,
			"subject":  tmpRpf.Subject,
		})
	if err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return
}
