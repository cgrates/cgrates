/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package utils

import (
	"strings"
)

// GetFullFieldPath returns the full path for the
func GetFullFieldPath(fldPath string, dP DataProvider) (fpath *FullPath, err error) {
	var newPath string
	if newPath, err = ProcessFieldPath(fldPath, InfieldSep, dP); err != nil || newPath == EmptyString {
		return
	}
	return NewFullPath(newPath), nil
}

// replaces the dynamic path between <> seperated by sep
func ProcessFieldPath(fldPath, sep string, dP DataProvider) (newPath string, err error) {
	startIdx := strings.IndexByte(fldPath, RSRDynStartChar)
	if startIdx == -1 {
		return // no proccessing requred
	}
	endIdx := strings.IndexByte(fldPath, RSRDynEndChar)
	if endIdx == -1 {
		err = ErrWrongPath
		newPath = EmptyString
		return
	}
	newPath = fldPath[:startIdx]
	for _, path := range strings.Split(fldPath[startIdx+1:endIdx], sep) { // proccess the found path
		var val string
		if strings.HasPrefix(path, MetaNow) || strings.HasPrefix(path, MetaTenant) {
			if val, err = dP.FieldAsString(SplitPath(path, NestingSep[0], -1)); err != nil {
				newPath = EmptyString
				return
			}
		} else {
			if val, err = DPDynamicString(path, dP); err != nil {
				newPath = EmptyString
				return
			}
		}
		newPath += val
	}
	newPath += fldPath[endIdx+1:]
	return
}

// ParseParamForDataProvider will parse a param string with prefix "~*req.", "~*otps.", "*tenant" or "*now"; or strings containing dynamic paths between "<" ">" signs. If none match, it returns the original string
func ParseParamForDataProvider(param string, dP DataProvider, onlyEncapsulatead bool) (string, error) {
	// Check if the string contains any "&" characters
	if strings.Contains(param, ANDSep) {
		parts := strings.Split(param, ANDSep)
		var results []string
		// Process each part individually
		for _, part := range parts {
			result, err := ParseParamForDataProvider(part, dP, onlyEncapsulatead)
			if err != nil {
				return EmptyString, err
			}
			results = append(results, result)
		}
		// Join all processed parts with &
		return strings.Join(results, ANDSep), nil
	}
	if !onlyEncapsulatead {
		switch {
		case strings.HasPrefix(param, MetaDynReq) || strings.HasPrefix(param, DynamicDataPrefix+MetaOpts):
			return DPDynamicString(param, dP)
		case strings.HasPrefix(param, MetaNow) || strings.HasPrefix(param, MetaTenant):
			return dP.FieldAsString(SplitPath(param, NestingSep[0], -1))
		}
	}

	// look for dynamic paths in the param string, and parse it
	if newParam, err := ProcessFieldPath(param, PlusChar, dP); err != nil {
		return EmptyString, err
	} else if newParam != EmptyString {
		return newParam, nil
	}

	return param, nil
}
