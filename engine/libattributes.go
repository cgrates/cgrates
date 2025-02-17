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
	"bytes"
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

type apWithWeight struct {
	*AttributeProfile
	weight float64
}

// Attribute used by AttributeProfile to describe a single attribute
type Attribute struct {
	FilterIDs []string
	Blockers  utils.DynamicBlockers // Blockers flag to stop processing on multiple attributes from a profile
	Path      string
	Type      string
	Value     utils.RSRParsers
}

// AttributeProfile the profile definition for the attributes
type AttributeProfile struct {
	Tenant     string
	ID         string
	FilterIDs  []string
	Weights    utils.DynamicWeights
	Blockers   utils.DynamicBlockers // Blockers flag to stop processing on multiple runs
	Attributes []*Attribute
}

// AttributeProfileWithAPIOpts is used in replicatorV1 for dispatcher
type AttributeProfileWithAPIOpts struct {
	*AttributeProfile
	APIOpts map[string]any
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

// ExternalAttribute the attribute for external profile
type ExternalAttribute struct {
	FilterIDs []string
	Blockers  utils.DynamicBlockers
	Path      string
	Type      string
	Value     string
}

// APIAttributeProfile used by APIs
type APIAttributeProfile struct {
	Tenant    string
	ID        string
	FilterIDs []string
	Blockers  utils.DynamicBlockers
	//Blocker    bool // blocker flag to stop processing on multiple runs
	Weights    utils.DynamicWeights
	Attributes []*ExternalAttribute
}

type APIAttributeProfileWithAPIOpts struct {
	*APIAttributeProfile
	APIOpts map[string]any
}

func NewAPIAttributeProfile(attr *AttributeProfile) (ext *APIAttributeProfile) {
	ext = &APIAttributeProfile{
		Tenant:     attr.Tenant,
		ID:         attr.ID,
		FilterIDs:  attr.FilterIDs,
		Attributes: make([]*ExternalAttribute, len(attr.Attributes)),
		Weights:    attr.Weights,
		Blockers:   attr.Blockers,
	}
	for i, at := range attr.Attributes {
		ext.Attributes[i] = &ExternalAttribute{
			FilterIDs: at.FilterIDs,
			Blockers:  at.Blockers,
			Path:      at.Path,
			Type:      at.Type,
			Value:     at.Value.GetRule(),
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
		if attr.Attributes[i].Value, err = utils.NewRSRParsers(extAttr.Value, utils.InfieldSep); err != nil {
			return nil, err
		}
		attr.Attributes[i].Blockers = extAttr.Blockers
		attr.Attributes[i].Type = extAttr.Type
		attr.Attributes[i].FilterIDs = extAttr.FilterIDs
		attr.Attributes[i].Path = extAttr.Path
	}
	attr.Tenant = ext.Tenant
	attr.ID = ext.ID
	attr.FilterIDs = ext.FilterIDs
	attr.Blockers = ext.Blockers
	attr.Weights = ext.Weights
	return
}

// NewAttributeFromInline parses an inline rule into a compiled AttributeProfile
func NewAttributeFromInline(tenant, inlnRule string) (attr *AttributeProfile, err error) {
	attr = &AttributeProfile{
		Tenant: tenant,
		ID:     inlnRule,
	}
	for _, rule := range strings.Split(inlnRule, utils.InfieldSep) {
		ruleSplt := utils.SplitPath(rule, utils.InInFieldSep[0], 3)
		if len(ruleSplt) < 3 {
			return nil, fmt.Errorf("inline parse error for string: <%s>", rule)
		}
		var vals utils.RSRParsers
		if vals, err = utils.NewRSRParsers(ruleSplt[2], utils.ANDSep); err != nil {
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
func externalAttributeAPI(httpType string, dDP utils.DataProvider) (string, error) {
	urlS, err := extractUrlFromType(httpType)
	if err != nil {
		return "", err
	}
	return externalAPI(urlS, bytes.NewReader([]byte(dDP.String())))
}

func (ap *AttributeProfile) Set(path []string, val any, newBranch bool) (err error) {
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
		case utils.Blockers:
			if val != utils.EmptyString {
				ap.Blockers, err = utils.NewDynamicBlockersFromString(utils.IfaceAsString(val), utils.InfieldSep, utils.ANDSep)
			}
		case utils.Weights:
			if val != utils.EmptyString {
				ap.Weights, err = utils.NewDynamicWeightsFromString(utils.IfaceAsString(val), utils.InfieldSep, utils.ANDSep)
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
		case utils.Blockers:
			if val != utils.EmptyString {
				ap.Attributes[len(ap.Attributes)-1].Blockers, err = utils.NewDynamicBlockersFromString(utils.IfaceAsString(val), utils.InfieldSep, utils.ANDSep)
			}
		case utils.Path:
			ap.Attributes[len(ap.Attributes)-1].Path = utils.IfaceAsString(val)
		case utils.Type:
			ap.Attributes[len(ap.Attributes)-1].Type = utils.IfaceAsString(val)
		case utils.Value:
			ap.Attributes[len(ap.Attributes)-1].Value, err = utils.NewRSRParsers(utils.IfaceAsString(val), utils.RSRSep)
		default:
			return utils.ErrWrongPath
		}
	default:
		return utils.ErrWrongPath
	}
	return
}

func (ap *AttributeProfile) Merge(v2 any) {
	vi := v2.(*AttributeProfile)
	if len(vi.Tenant) != 0 {
		ap.Tenant = vi.Tenant
	}
	if len(vi.ID) != 0 {
		ap.ID = vi.ID
	}
	ap.FilterIDs = append(ap.FilterIDs, vi.FilterIDs...)
	for _, attr := range vi.Attributes {
		if attr.Type != utils.EmptyString {
			ap.Attributes = append(ap.Attributes, attr)
		}
	}
	if vi.Blockers != nil {
		ap.Blockers = append(ap.Blockers, vi.Blockers...)
	}
	if vi.Weights != nil {
		ap.Weights = append(ap.Weights, vi.Weights...)
	}
}

func (ap *AttributeProfile) String() string { return utils.ToJSON(ap) }
func (ap *AttributeProfile) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = ap.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(val), nil
}
func (ap *AttributeProfile) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) == 1 {
		switch fldPath[0] {
		default:
			fld, idx := utils.GetPathIndex(fldPath[0])
			if idx != nil {
				switch fld {
				case utils.Attributes:
					if *idx < len(ap.Attributes) {
						return ap.Attributes[*idx], nil
					}
				case utils.FilterIDs:
					if *idx < len(ap.FilterIDs) {
						return ap.FilterIDs[*idx], nil
					}
				}
			}
			return nil, utils.ErrNotFound
		case utils.Tenant:
			return ap.Tenant, nil
		case utils.ID:
			return ap.ID, nil
		case utils.FilterIDs:
			return ap.FilterIDs, nil
		case utils.Blockers:
			return ap.Blockers, nil
		case utils.Weights:
			return ap.Weights, nil
		case utils.Attributes:
			return ap.Attributes, nil
		}
	}
	if len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}
	fld, idx := utils.GetPathIndex(fldPath[0])
	if fld != utils.Attributes || idx == nil {
		return nil, utils.ErrNotFound
	}
	if *idx >= len(ap.Attributes) {
		return nil, utils.ErrNotFound
	}
	return ap.Attributes[*idx].FieldAsInterface(fldPath[1:])
}

func (at *Attribute) String() string { return utils.ToJSON(at) }
func (at *Attribute) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = at.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(val), nil
}

func (at *Attribute) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	switch fldPath[0] {
	default:
		fld, idx := utils.GetPathIndex(fldPath[0])
		if idx != nil &&
			fld == utils.FilterIDs &&
			*idx < len(at.FilterIDs) {
			return at.FilterIDs[*idx], nil
		}
		return nil, utils.ErrNotFound
	case utils.FilterIDs:
		return at.FilterIDs, nil
	case utils.Blockers:
		return at.Blockers, nil
	case utils.Path:
		return at.Path, nil
	case utils.Type:
		return at.Type, nil
	case utils.Value:
		return at.Value.GetRule(), nil
	}
}
