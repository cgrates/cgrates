/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package config

import (
	"code.google.com/p/goconf/conf"
	"errors"
	"strings"
)


// Adds support for slice values in config
func ConfigSlice( c *conf.ConfigFile, section, valName string ) ([]string, error) {
	sliceStr, errGet := c.GetString(section, valName)
	if errGet != nil {
		return nil, errGet
	}
	cfgValStrs := strings.Split(sliceStr, ",") // If need arrises, we can make the separator configurable
	if len(cfgValStrs)==1 && cfgValStrs[0]=="" { // Prevents returning iterable with empty value
		return []string{}, nil
	}
	for _,elm := range cfgValStrs {
		if elm == "" { //One empty element is presented when splitting empty string
			return nil, errors.New("Empty values in config slice")
			
		}
	}
	return cfgValStrs, nil
}
