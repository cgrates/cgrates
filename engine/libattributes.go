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

package engine

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// Attribute used by AttributeProfile to describe a single attribute
type Attribute struct {
	FilterIDs []string
	Path      string
	Type      string
	Value     config.RSRParsers
}

// AttributeProfile the profile definition for the attributes
type AttributeProfile struct {
	Tenant             string
	ID                 string
	Contexts           []string // bind this AttributeProfile to multiple contexts
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // Activation interval
	Attributes         []*Attribute
	Blocker            bool // blocker flag to stop processing on multiple runs
	Weight             float64
}

// AttributeProfileWithOpts is used in replicatorV1 for dispatcher
type AttributeProfileWithOpts struct {
	*AttributeProfile
	Opts map[string]interface{}
}

func (ap *AttributeProfile) compileSubstitutes() (err error) {
	for _, attr := range ap.Attributes {
		if err = attr.Value.Compile(); err != nil {
			return
		}
	}
	return
}

// Compile is a wrapper for convenience setting up the AttributeProfile
func (ap *AttributeProfile) Compile() error {
	return ap.compileSubstitutes()
}

// TenantID returns the tenant wit the ID
func (ap *AttributeProfile) TenantID() string {
	return utils.ConcatenatedKey(ap.Tenant, ap.ID)
}

// AttributeProfiles is a sortable list of Attribute profiles
type AttributeProfiles []*AttributeProfile

// Sort is part of sort interface, sort based on Weight
func (aps AttributeProfiles) Sort() {
	sort.Slice(aps, func(i, j int) bool { return aps[i].Weight > aps[j].Weight })
}

// ExternalAttribute the attribute for external profile
type ExternalAttribute struct {
	FilterIDs []string
	Path      string
	Type      string
	Value     string
}

// APIAttributeProfile used by APIs
type APIAttributeProfile struct {
	Tenant             string
	ID                 string
	Contexts           []string // bind this AttributeProfile to multiple contexts
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // Activation interval
	Attributes         []*ExternalAttribute
	Blocker            bool // blocker flag to stop processing on multiple runs
	Weight             float64
}

// AsAttributeProfile converts the external attribute format to the actual AttributeProfile
func (ext *APIAttributeProfile) AsAttributeProfile() (attr *AttributeProfile, err error) {
	attr = new(AttributeProfile)
	if len(ext.Attributes) == 0 {
		return nil, utils.NewErrMandatoryIeMissing("Attributes")
	}
	attr.Attributes = make([]*Attribute, len(ext.Attributes))
	for i, extAttr := range ext.Attributes {
		if extAttr.Path == utils.EmptyString {
			return nil, utils.NewErrMandatoryIeMissing("Path")
		}
		if len(extAttr.Value) == 0 {
			return nil, utils.NewErrMandatoryIeMissing("Value")
		}
		attr.Attributes[i] = new(Attribute)
		if attr.Attributes[i].Value, err = config.NewRSRParsers(extAttr.Value, utils.InfieldSep); err != nil {
			return nil, err
		}
		attr.Attributes[i].Type = extAttr.Type
		attr.Attributes[i].FilterIDs = extAttr.FilterIDs
		attr.Attributes[i].Path = extAttr.Path
	}
	attr.Tenant = ext.Tenant
	attr.ID = ext.ID
	attr.Contexts = ext.Contexts
	attr.FilterIDs = ext.FilterIDs
	attr.ActivationInterval = ext.ActivationInterval
	attr.Blocker = ext.Blocker
	attr.Weight = ext.Weight
	return
}

// NewAttributeFromInline parses an inline rule into a compiled AttributeProfile
func NewAttributeFromInline(tenant, inlnRule string) (attr *AttributeProfile, err error) {
	attr = &AttributeProfile{
		Tenant:   tenant,
		ID:       inlnRule,
		Contexts: []string{utils.MetaAny},
	}
	for _, rule := range strings.Split(inlnRule, utils.InfieldSep) {
		ruleSplt := strings.SplitN(rule, utils.InInFieldSep, 3)
		if len(ruleSplt) < 3 {
			return nil, fmt.Errorf("inline parse error for string: <%s>", rule)
		}
		var vals config.RSRParsers
		if vals, err = config.NewRSRParsers(ruleSplt[2], utils.ANDSep); err != nil {
			return nil, err
		}
		if len(ruleSplt[1]) == 0 {
			err = fmt.Errorf("empty path in inline AttributeProfile <%s>", inlnRule)
			return
		}
		attr.Attributes = append(attr.Attributes, &Attribute{
			Path:  ruleSplt[1],
			Type:  ruleSplt[0],
			Value: vals,
		})
	}
	return
}
