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
	"errors"
	"fmt"
	"strings"
)

// NMType the type used for navigable Map
type NMType byte

// posible NMType
const (
	NMDataType NMType = iota
	NMMapType
	NMSliceType
)

// DataProvider is a data source from multiple formats
type DataProvider interface {
	String() string // printable version of data
	FieldAsInterface(fldPath []string) (interface{}, error)
	FieldAsString(fldPath []string) (string, error) // remove this
}

// RWDataProvider is a DataProvider with write methods on it
type RWDataProvider interface {
	DataProvider

	Set(fldPath []string, val interface{}) (err error)
	Remove(fldPath []string) (err error)
}

// NavigableMapper is the interface supported by replies convertible to CGRReply
type NavigableMapper interface {
	AsNavigableMap() map[string]*DataNode
}

// DPDynamicInterface returns the value of the field if the path is dynamic
func DPDynamicInterface(dnVal string, dP DataProvider) (interface{}, error) {
	if strings.HasPrefix(dnVal, DynamicDataPrefix) &&
		dnVal != DynamicDataPrefix {
		dnVal = strings.TrimPrefix(dnVal, DynamicDataPrefix)
		return dP.FieldAsInterface(SplitPath(dnVal, NestingSep[0], -1))
	}
	return StringToInterface(dnVal), nil
}

// DPDynamicString returns the string value of the field if the path is dynamic
func DPDynamicString(dnVal string, dP DataProvider) (string, error) {
	if strings.HasPrefix(dnVal, DynamicDataPrefix) &&
		dnVal != DynamicDataPrefix {
		dnVal = strings.TrimPrefix(dnVal, DynamicDataPrefix)
		return dP.FieldAsString(SplitPath(dnVal, NestingSep[0], -1))
	}
	return dnVal, nil
}

func IsPathValid(path string) (err error) {
	if !strings.HasPrefix(path, DynamicDataPrefix) {
		return nil
	}
	paths := SplitPath(path, NestingSep[0], -1)
	if len(paths) <= 1 {
		return errors.New("Path is missing ")
	}
	for _, path := range paths {
		if strings.TrimSpace(path) == EmptyString {
			return errors.New("Empty field path ")
		}
	}
	return nil
}

func IsPathValidForExporters(path string) (err error) {
	if !strings.HasPrefix(path, DynamicDataPrefix) {
		return nil
	}
	paths := SplitPath(path, NestingSep[0], -1)
	for _, newPath := range paths {
		if strings.TrimSpace(newPath) == EmptyString {
			return errors.New("Empty field path ")
		}
	}
	return nil
}

func CheckInLineFilter(fltrs []string) (err error) {
	for _, fltr := range fltrs {
		if strings.HasPrefix(fltr, Meta) {
			rules := strings.SplitN(fltr, InInFieldSep, 3)
			if len(rules) < 3 {
				return fmt.Errorf("inline parse error for string: <%s>", fltr)
			}
			valFunc := IsPathValid
			if rules[0] == MetaEmpty || rules[0] == MetaExists {
				valFunc = IsPathValidForExporters
			}
			if err = valFunc(rules[1]); err != nil {
				return fmt.Errorf("%s for <%s>", err, fltr) //encapsulated error
			}
			for _, val := range strings.Split(rules[2], PipeSep) {
				if err = valFunc(val); err != nil {
					return fmt.Errorf("%s for <%s>", err, fltr) //encapsulated error
				}
			}
		}
	}
	return nil
}
