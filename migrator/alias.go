// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (

	//"log"
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type v1Alias struct {
	Direction string
	Tenant    string
	Category  string
	Account   string
	Subject   string
	Context   string
	Values    v1AliasValues
}

var (
	ALIASES_PREFIX = "als_"
	Alias          = "Alias"
)

type v1AliasValues []*v1AliasValue

type v1AliasValue struct {
	DestinationId string //=Destination
	Pairs         v1AliasPairs
	Weight        float64
}

type v1AliasPairs map[string]map[string]string

func (al *v1Alias) SetId(id string) error {
	vals := strings.Split(id, utils.CONCATENATED_KEY_SEP)
	if len(vals) != 6 {
		return utils.ErrInvalidKey
	}
	al.Direction = vals[0]
	al.Tenant = vals[1]
	al.Category = vals[2]
	al.Account = vals[3]
	al.Subject = vals[4]
	al.Context = vals[5]
	return nil
}

func (al *v1Alias) GetId() string {
	return utils.ConcatenatedKey(al.Direction, al.Tenant, al.Category, al.Account, al.Subject, al.Context)
}

func alias2AtttributeProfile(alias *v1Alias, defaultTenant string) *engine.AttributeProfile {
	out := &engine.AttributeProfile{
		Tenant:             alias.Tenant,
		ID:                 alias.GetId(),
		Contexts:           []string{utils.META_ANY},
		FilterIDs:          make([]string, 0),
		ActivationInterval: nil,
		Attributes:         make([]*engine.Attribute, 0),
		Blocker:            false,
		Weight:             20, // should have prio against attributes out of users
	}
	if len(out.Tenant) == 0 || out.Tenant == utils.META_ANY {
		out.Tenant = defaultTenant
	}
	if len(alias.Category) != 0 && alias.Category != utils.META_ANY {
		out.FilterIDs = append(out.FilterIDs,
			fmt.Sprintf("%s:~%s:%s", utils.MetaString, utils.MetaReq+utils.NestingSep+utils.Category, alias.Category))
	}
	if len(alias.Account) != 0 && alias.Account != utils.META_ANY {
		out.FilterIDs = append(out.FilterIDs,
			fmt.Sprintf("%s:~%s:%s", utils.MetaString, utils.MetaReq+utils.NestingSep+utils.Account, alias.Account))
	}
	if len(alias.Subject) != 0 && alias.Subject != utils.META_ANY {
		out.FilterIDs = append(out.FilterIDs,
			fmt.Sprintf("%s:~%s:%s", utils.MetaString, utils.MetaReq+utils.NestingSep+utils.Subject, alias.Subject))
	}
	var destination string
	for _, av := range alias.Values {
		if len(destination) == 0 || destination == utils.META_ANY {
			destination = av.DestinationId
		}
		for fieldName, vals := range av.Pairs {
			for initial, substitute := range vals {
				var fld string
				if fieldName == utils.Tenant {
					fieldName = utils.MetaTenant
					fld = utils.MetaTenant
				} else {
					fld = utils.MetaReq + utils.NestingSep + fieldName
				}
				attr := &engine.Attribute{
					Path:  fld,
					Type:  utils.MetaVariable, //default type for Attribute
					Value: config.NewRSRParsersMustCompile(substitute, true, utils.INFIELD_SEP),
				}
				out.Attributes = append(out.Attributes, attr)
				// Add attribute filters if needed
				if initial == "" || initial == utils.META_ANY {
					continue
				}
				if fieldName == utils.MetaTenant { // no filter for tenant
					continue
				}
				if fieldName == utils.Category && alias.Category == initial {
					continue
				}
				if fieldName == utils.Account && alias.Account == initial {
					continue
				}
				if fieldName == utils.Subject && alias.Subject == initial {
					continue
				}
				attr.FilterIDs = append(attr.FilterIDs,
					fmt.Sprintf("%s:~%s:%s", utils.MetaString, utils.MetaReq+utils.NestingSep+fieldName, initial))
			}
		}
	}
	if len(destination) != 0 && destination != utils.META_ANY {
		out.FilterIDs = append(out.FilterIDs,
			fmt.Sprintf("%s:~%s:%s", utils.MetaDestinations, utils.MetaReq+utils.NestingSep+utils.Destination, destination))
	}
	return out
}

func (m *Migrator) migrateAlias2Attributes() (err error) {
	cfg := config.CgrConfig()

	defaultTenant := cfg.GeneralCfg().DefaultTenant
	for {
		alias, err := m.dmIN.getV1Alias()
		if err == utils.ErrNoMoreData {
			break
		}
		if err != nil {
			return err
		}
		if alias == nil || m.dryRun {
			continue
		}
		attr := alias2AtttributeProfile(alias, defaultTenant)
		if len(attr.Attributes) == 0 {
			continue
		}
		if err := m.dmIN.remV1Alias(alias.GetId()); err != nil {
			return err
		}
		if err := m.dmOut.DataManager().SetAttributeProfile(attr, true); err != nil {
			return err
		}
		m.stats[Alias] += 1
	}
	if m.dryRun {
		return
	}
	// All done, update version with current one
	vrs := engine.Versions{Alias: 0} //engine.CurrentDataDBVersions()[utils.Alias]}
	if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating Alias version into dataDB", err.Error()))
	}
	return
}

func (m *Migrator) migrateAlias() (err error) {
	if err = m.migrateAlias2Attributes(); err != nil {
		return err
	}
	return m.ensureIndexesDataDB(engine.ColAttr)
}
