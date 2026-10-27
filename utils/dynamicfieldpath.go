// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	for path := range strings.SplitSeq(fldPath[startIdx+1:endIdx], sep) { // proccess the found path
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
	return GetValueFromDynamicPath(param, dP, onlyEncapsulatead)
}

// GetValueFromDynamicPath will parse a param string with prefix "~*req.", "~*otps.", "*tenant" or "*now"; or strings containing dynamic paths between "<" ">" signs. If none match, it returns the original string
func GetValueFromDynamicPath(param string, dP DataProvider, onlyEncapsulatead bool) (string, error) {
	if !onlyEncapsulatead { // dont check of non-encapsulated dynamic paths if onlyEncapsulatead is true
		switch { // check for dynamic paths with special prefixes, if not found any, fall down to encapsulated dynamic paths
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

// GetDynamicPathValueFromMapStringAny will parse a map[string]any for dynamic paths in keys and values, and return a new map[string]any with the parsed values. It will recursively parse nested maps. If onlyEncapsulatead is true, it will only parse dynamic paths that are encapsulated between "<" and ">" signs.
func GetDynamicPathValueFromMapStringAny(param map[string]any, dP DataProvider, onlyEncapsulatead bool) (map[string]any, error) {
	for key, val := range param {
		// parse the key for dynamic paths
		keyProcessed, err := GetValueFromDynamicPath(key, dP, onlyEncapsulatead)
		if err != nil {
			return nil, err
		}
		// parse the value for dynamic paths, if the value is a map[string]any, recursively call this function to process the nested map
		if valMap, ok := val.(map[string]any); ok {
			valProcessed, err := GetDynamicPathValueFromMapStringAny(valMap, dP, onlyEncapsulatead)
			if err != nil {
				return nil, err
			}
			param[keyProcessed] = valProcessed
			if keyProcessed != key {
				delete(param, key) // remove the old key if it was changed
			}
		} else if valStr, ok := val.(string); ok { // parse the value for dynamic paths if it is a string
			valProcessed, err := GetValueFromDynamicPath(valStr, dP, onlyEncapsulatead)
			if err != nil {
				return nil, err
			}
			param[keyProcessed] = valProcessed
			if keyProcessed != key {
				delete(param, key) // remove the old key if it was changed
			}
		} else { // if the value is not a string or a map[string]any, just keep it as is
			param[keyProcessed] = val
			if keyProcessed != key {
				delete(param, key) // remove the old key if it was changed
			}
		}
	}
	return param, nil
}
