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
	Tenant     string
	ID         string
	FilterIDs  []string
	Attributes []*Attribute
	Blocker    bool // blocker flag to stop processing on multiple runs
	Weight     float64
}

// AttributeProfileWithAPIOpts is used in replicatorV1 for dispatcher
type AttributeProfileWithAPIOpts struct {
	*AttributeProfile
	APIOpts map[string]interface{}
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

// TenantIDInline returns the id for inline
func (ap *AttributeProfile) TenantIDInline() string {
	if strings.HasPrefix(ap.ID, utils.Meta) {
		return ap.ID
	}
	return ap.TenantID()
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
	Tenant     string
	ID         string
	FilterIDs  []string
	Attributes []*ExternalAttribute
	Blocker    bool // blocker flag to stop processing on multiple runs
	Weight     float64
}

type APIAttributeProfileWithAPIOpts struct {
	*APIAttributeProfile
	APIOpts map[string]interface{}
}

func NewAPIAttributeProfile(attr *AttributeProfile) (ext *APIAttributeProfile) {
	ext = &APIAttributeProfile{
		Tenant:     attr.Tenant,
		ID:         attr.ID,
		FilterIDs:  attr.FilterIDs,
		Attributes: make([]*ExternalAttribute, len(attr.Attributes)),
		Blocker:    attr.Blocker,
		Weight:     attr.Weight,
	}
	for i, at := range attr.Attributes {
		ext.Attributes[i] = &ExternalAttribute{
			FilterIDs: at.FilterIDs,
			Path:      at.Path,
			Type:      at.Type,
			Value:     at.Value.GetRule(utils.InfieldSep),
		}
	}
	return
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
	attr.FilterIDs = ext.FilterIDs
	attr.Blocker = ext.Blocker
	attr.Weight = ext.Weight
	return
}

// NewAttributeFromInline parses an inline rule into a compiled AttributeProfile
func NewAttributeFromInline(tenant, inlnRule string) (attr *AttributeProfile, err error) {
	attr = &AttributeProfile{
		Tenant: tenant,
		ID:     inlnRule,
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

func (ap *AttributeProfile) Set(path []string, val interface{}, newBranch bool, rsrSep string) (err error) {
	switch len(path) {
	case 1:
		switch path[0] {
		case utils.Tenant:
			ap.Tenant = utils.IfaceAsString(val)
		case utils.ID:
			ap.ID = utils.IfaceAsString(val)
		case utils.FilterIDs:
			var valA []string
			valA, err = utils.IfaceAsStringSlice(val)
			ap.FilterIDs = append(ap.FilterIDs, valA...)
		case utils.Blocker:
			ap.Blocker, err = utils.IfaceAsBool(val)
		case utils.Weight:
			if val != utils.EmptyString {
				ap.Weight, err = utils.IfaceAsFloat64(val)
			}
		default:
			return utils.ErrWrongPath
		}
	case 2:
		if path[0] != utils.Attributes {
			return utils.ErrWrongPath
		}
		if len(ap.Attributes) == 0 || newBranch {
			ap.Attributes = append(ap.Attributes, new(Attribute))
		}
		switch path[1] {
		case utils.FilterIDs:
			var valA []string
			valA, err = utils.IfaceAsStringSlice(val)
			ap.Attributes[len(ap.Attributes)-1].FilterIDs = append(ap.Attributes[len(ap.Attributes)-1].FilterIDs, valA...)
		case utils.Path:
			ap.Attributes[len(ap.Attributes)-1].Path = utils.IfaceAsString(val)
		case utils.Type:
			ap.Attributes[len(ap.Attributes)-1].Type = utils.IfaceAsString(val)
		case utils.Value:
			ap.Attributes[len(ap.Attributes)-1].Value, err = config.NewRSRParsers(utils.IfaceAsString(val), rsrSep)
		default:
			return utils.ErrWrongPath
		}
	default:
		return utils.ErrWrongPath
	}
	return
}
