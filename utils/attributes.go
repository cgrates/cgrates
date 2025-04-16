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

package utils

import (
	"fmt"
	"strings"
)

// AttributeProfile defines the configuration of attributes for processing.
type AttributeProfile struct {
	Tenant     string
	ID         string
	FilterIDs  []string
	Weights    DynamicWeights
	Blockers   DynamicBlockers // Blockers flag to stop processing on multiple runs
	Attributes []*Attribute
}

// AttributeProfileWithAPIOpts wraps AttributeProfile with APIOpts.
type AttributeProfileWithAPIOpts struct {
	*AttributeProfile
	APIOpts map[string]any
}

// NewAttributeFromInline parses an inline rule into a compiled AttributeProfile.
func NewAttributeFromInline(tenant, inlnRule string) (attr *AttributeProfile, err error) {
	attr = &AttributeProfile{
		Tenant: tenant,
		ID:     inlnRule,
	}
	for _, rule := range strings.Split(inlnRule, InfieldSep) {
		ruleSplt := SplitPath(rule, InInFieldSep[0], 3)
		if len(ruleSplt) < 3 {
			return nil, fmt.Errorf("inline parse error for string: <%s>", rule)
		}
		var vals RSRParsers
		if vals, err = NewRSRParsers(ruleSplt[2], ANDSep); err != nil {
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

// compileSubstitutes processes all attribute value substitutes for the profile.
func (ap *AttributeProfile) compileSubstitutes() (err error) {
	for _, attr := range ap.Attributes {
		if err = attr.Value.Compile(); err != nil {
			return
		}
	}
	return
}

// Compile is a wrapper for convenience setting up the AttributeProfile.
func (ap *AttributeProfile) Compile() error {
	return ap.compileSubstitutes()
}

// TenantID returns the concatenated tenant and ID.
func (ap *AttributeProfile) TenantID() string {
	return ConcatenatedKey(ap.Tenant, ap.ID)
}

// TenantIDInline returns the ID for inline processing, keeping meta prefixes
// intact.
func (ap *AttributeProfile) TenantIDInline() string {
	if strings.HasPrefix(ap.ID, Meta) {
		return ap.ID
	}
	return ap.TenantID()
}

// Set implements the profile interface, setting values in AttributeProfile
// based on path.
func (ap *AttributeProfile) Set(path []string, val any, newBranch bool) (err error) {
	switch len(path) {
	case 1:
		switch path[0] {
		case Tenant:
			ap.Tenant = IfaceAsString(val)
		case ID:
			ap.ID = IfaceAsString(val)
		case FilterIDs:
			var valA []string
			valA, err = IfaceAsStringSlice(val)
			ap.FilterIDs = append(ap.FilterIDs, valA...)
		case Blockers:
			if val != EmptyString {
				ap.Blockers, err = NewDynamicBlockersFromString(IfaceAsString(val), InfieldSep, ANDSep)
			}
		case Weights:
			if val != EmptyString {
				ap.Weights, err = NewDynamicWeightsFromString(IfaceAsString(val), InfieldSep, ANDSep)
			}
		default:
			return ErrWrongPath
		}
	case 2:
		if path[0] != Attributes {
			return ErrWrongPath
		}
		if len(ap.Attributes) == 0 || newBranch {
			ap.Attributes = append(ap.Attributes, new(Attribute))
		}
		switch path[1] {
		case FilterIDs:
			var valA []string
			valA, err = IfaceAsStringSlice(val)
			ap.Attributes[len(ap.Attributes)-1].FilterIDs = append(ap.Attributes[len(ap.Attributes)-1].FilterIDs, valA...)
		case Blockers:
			if val != EmptyString {
				ap.Attributes[len(ap.Attributes)-1].Blockers, err = NewDynamicBlockersFromString(IfaceAsString(val), InfieldSep, ANDSep)
			}
		case Path:
			ap.Attributes[len(ap.Attributes)-1].Path = IfaceAsString(val)
		case Type:
			ap.Attributes[len(ap.Attributes)-1].Type = IfaceAsString(val)
		case Value:
			ap.Attributes[len(ap.Attributes)-1].Value, err = NewRSRParsers(IfaceAsString(val), RSRSep)
		default:
			return ErrWrongPath
		}
	default:
		return ErrWrongPath
	}
	return
}

// Merge implements the profile interface, merging values from another AttributeProfile.
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
		if attr.Type != EmptyString {
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

// String implements the DataProvider interface, returning the AttributeProfile in JSON format.
func (ap *AttributeProfile) String() string { return ToJSON(ap) }

// FieldAsString implements the DataProvider interface, retrieving field value as string.
func (ap *AttributeProfile) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = ap.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}

// FieldAsInterface implements the DataProvider interface, retrieving field value as interface.
func (ap *AttributeProfile) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) == 1 {
		switch fldPath[0] {
		default:
			fld, idx := GetPathIndex(fldPath[0])
			if idx != nil {
				switch fld {
				case Attributes:
					if *idx < len(ap.Attributes) {
						return ap.Attributes[*idx], nil
					}
				case FilterIDs:
					if *idx < len(ap.FilterIDs) {
						return ap.FilterIDs[*idx], nil
					}
				}
			}
			return nil, ErrNotFound
		case Tenant:
			return ap.Tenant, nil
		case ID:
			return ap.ID, nil
		case FilterIDs:
			return ap.FilterIDs, nil
		case Blockers:
			return ap.Blockers, nil
		case Weights:
			return ap.Weights, nil
		case Attributes:
			return ap.Attributes, nil
		}
	}
	if len(fldPath) == 0 {
		return nil, ErrNotFound
	}
	fld, idx := GetPathIndex(fldPath[0])
	if fld != Attributes || idx == nil {
		return nil, ErrNotFound
	}
	if *idx >= len(ap.Attributes) {
		return nil, ErrNotFound
	}
	return ap.Attributes[*idx].FieldAsInterface(fldPath[1:])
}

// Attribute defines a single attribute.
type Attribute struct {
	FilterIDs []string
	Blockers  DynamicBlockers // Blockers flag to stop processing on multiple attributes from a profile
	Path      string
	Type      string
	Value     RSRParsers
}

// String returns the Attribute in JSON format.
func (a *Attribute) String() string { return ToJSON(a) }

// FieldAsString retrieves field value as string from Attribute.
func (a *Attribute) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = a.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}

// FieldAsInterface retrieves field value as interface from Attribute.
func (a *Attribute) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) != 1 {
		return nil, ErrNotFound
	}
	switch fldPath[0] {
	default:
		fld, idx := GetPathIndex(fldPath[0])
		if idx != nil &&
			fld == FilterIDs &&
			*idx < len(a.FilterIDs) {
			return a.FilterIDs[*idx], nil
		}
		return nil, ErrNotFound
	case FilterIDs:
		return a.FilterIDs, nil
	case Blockers:
		return a.Blockers, nil
	case Path:
		return a.Path, nil
	case Type:
		return a.Type, nil
	case Value:
		return a.Value.GetRule(), nil
	}
}

// APIAttributeProfile represents the external representation used by APIs.
type APIAttributeProfile struct {
	Tenant    string
	ID        string
	FilterIDs []string
	Blockers  DynamicBlockers
	//Blocker    bool // blocker flag to stop processing on multiple runs
	Weights    DynamicWeights
	Attributes []*ExternalAttribute
}

// ExternalAttribute represents the API-facing attribute structure.
type ExternalAttribute struct {
	FilterIDs []string
	Blockers  DynamicBlockers
	Path      string
	Type      string
	Value     string
}

// APIAttributeProfileWithAPIOpts wraps APIAttributeProfile with APIOpts.
type APIAttributeProfileWithAPIOpts struct {
	*APIAttributeProfile
	APIOpts map[string]any
}

// NewAPIAttributeProfile creates an external representation from an AttributeProfile.
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

// AsAttributeProfile converts the external attribute format to the actual AttributeProfile.
func (ext *APIAttributeProfile) AsAttributeProfile() (attr *AttributeProfile, err error) {
	attr = new(AttributeProfile)
	if len(ext.Attributes) == 0 {
		return nil, NewErrMandatoryIeMissing("Attributes")
	}
	attr.Attributes = make([]*Attribute, len(ext.Attributes))
	for i, extAttr := range ext.Attributes {
		if extAttr.Path == EmptyString {
			return nil, NewErrMandatoryIeMissing("Path")
		}
		if len(extAttr.Value) == 0 {
			return nil, NewErrMandatoryIeMissing("Value")
		}
		attr.Attributes[i] = new(Attribute)
		if attr.Attributes[i].Value, err = NewRSRParsers(extAttr.Value, InfieldSep); err != nil {
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
