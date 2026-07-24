// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v2

import (
	"github.com/cgrates/cgrates/utils"
)

// Creates a new destination within a tariff plan
func (self *APIerSv2) SetTPDestination(attrs utils.TPDestination, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Prefixes"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPDestinations([]*utils.TPDestination{&attrs}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPDestination struct {
	TPid string // Tariff plan id
	Tag  string // Destination id
}

// Queries a specific destination
func (self *APIerSv2) GetTPDestination(attrs AttrGetTPDestination, reply *utils.TPDestination) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Tag"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if tpDsts, err := self.StorDb.GetTPDestinations(attrs.TPid, attrs.Tag); err != nil {
		return utils.APIErrorHandler(err)
	} else if len(tpDsts) == 0 {
		return utils.ErrNotFound
	} else {
		*reply = *tpDsts[0]
	}
	return nil
}

func (self *APIerSv2) RemoveTPDestination(attrs AttrGetTPDestination, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Tag"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBLTPDestinations, attrs.TPid, map[string]string{"tag": attrs.Tag}); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = utils.OK
	}
	return nil
}
