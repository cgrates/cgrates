// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type v1UserProfile struct {
	Tenant   string
	UserName string
	Masked   bool //disable if true
	Profile  map[string]string
	Weight   float64
}

func (ud *v1UserProfile) GetId() string {
	return utils.ConcatenatedKey(ud.Tenant, ud.UserName)
}

func (ud *v1UserProfile) SetId(id string) error {
	vals := strings.Split(id, utils.CONCATENATED_KEY_SEP)
	if len(vals) != 2 {
		return utils.ErrInvalidKey
	}
	ud.Tenant = vals[0]
	ud.UserName = vals[1]
	return nil
}

func userProfile2attributeProfile(user *v1UserProfile) (attr *engine.AttributeProfile) {
	usrFltr := config.CgrConfig().MigratorCgrCfg().UsersFilters
	attr = &engine.AttributeProfile{
		Tenant:             config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:                 user.UserName,
		Contexts:           []string{utils.META_ANY},
		FilterIDs:          make([]string, 0),
		ActivationInterval: nil,
		Attributes:         make([]*engine.Attribute, 0),
		Blocker:            false,
		Weight:             user.Weight,
	}
	if user.Tenant != attr.Tenant {
		attr.Attributes = append(attr.Attributes, &engine.Attribute{
			Path:  utils.MetaTenant,
			Value: config.NewRSRParsersMustCompile(user.Tenant, true, utils.INFIELD_SEP),
			Type:  utils.META_CONSTANT,
		})
	}
	for fieldName, substitute := range user.Profile {
		if fieldName == "ReqType" { // old style
			fieldName = utils.RequestType
		}
		if utils.IsSliceMember(usrFltr, fieldName) {
			attr.FilterIDs = append(attr.FilterIDs, fmt.Sprintf("*string:~%s:%s", fieldName, substitute))
			continue
		}
		attr.Attributes = append(attr.Attributes, &engine.Attribute{
			Path:  utils.MetaReq + utils.NestingSep + fieldName,
			Value: config.NewRSRParsersMustCompile(substitute, true, utils.INFIELD_SEP),
			Type:  utils.MetaVariable,
		})
	}
	return
}

func (m *Migrator) migrateV1User2AttributeProfile() (err error) {
	for {
		user, err := m.dmIN.getV1User()
		if err == utils.ErrNoMoreData {
			break
		}
		if err != nil {
			return err
		}
		if user == nil || user.Masked || m.dryRun {
			continue
		}
		attr := userProfile2attributeProfile(user)
		if len(attr.Attributes) == 0 {
			continue
		}
		if err := m.dmIN.remV1User(user.GetId()); err != nil {
			return err
		}
		if err := m.dmOut.DataManager().SetAttributeProfile(attr, true); err != nil {
			return err
		}
		m.stats[utils.User]++
	}
	if m.dryRun {
		return
	}
	// All done, update version with current one
	vrs := engine.Versions{utils.User: engine.CurrentDataDBVersions()[utils.User]}
	if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating Alias version into dataDB", err.Error()))
	}
	return
}

func (m *Migrator) migrateUser() (err error) {
	if err = m.migrateV1User2AttributeProfile(); err != nil {
		return err
	}
	return m.ensureIndexesDataDB(engine.ColAttr)
}
