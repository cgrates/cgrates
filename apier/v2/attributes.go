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

package v2

/*
import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

//SetAttributeProfile add/update a new Attribute Profile
func (apierV1 *ApierV2) SetAttributeProfile(alsPrf *engine.AttributeProfile, reply *string) error {
	if missing := utils.MissingStructFields(alsPrf, []string{"Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if len(alsPrf.Attributes) != 0 {
		for _, attr := range alsPrf.Attributes {
			for _, sub := range attr.Substitute {
				if sub.Rules == "" {
					return utils.NewErrMandatoryIeMissing("Rules")
				}
				if err := sub.Compile(); err != nil {
					return utils.NewErrServerError(err)
				}
			}
		}
	}

	if err := apierV1.DataManager.SetAttributeProfile(alsPrf, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
*/
