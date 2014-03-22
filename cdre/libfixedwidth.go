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

package cdre

import (
	"fmt"
	"strconv"
)

// Used as generic function logic for various fields

// Attributes
//  source - the base source
//  maxLen - the maximum field lenght
//  stripAllowed - whether we allow stripping of chars in case of source bigger than the maximum allowed
//  lStrip - if true, strip from beginning of the string
//  lPadding - if true, add chars at the beginning of the string
//  paddingChar - the character wich will be used to fill the existing
func filterField(source string, maxLen int, stripAllowed, lStrip, lPadding, padWithZero bool) (string, error) {
	if len(source) == maxLen { // the source is exactly the maximum length
		return source, nil
	}
	if len(source) > maxLen { //the source is bigger than allowed
		if !stripAllowed {
			return "", fmt.Errorf("source %s is bigger than the maximum allowed length %d", source, maxLen)
		}
		if !lStrip {
			return source[:maxLen], nil
		} else {
			diffIndx := len(source) - maxLen
			return source[diffIndx:], nil
		}
	} else { //the source is smaller as the maximum allowed
		paddingString := "%"
		if padWithZero {
			paddingString += "0" // it will not work for rPadding but this is not needed
		}
		if !lPadding {
			paddingString += "-"
		}
		paddingString += strconv.Itoa(maxLen) + "s"
		return fmt.Sprintf(paddingString, source), nil
	}
}

/*
type XmlCdreConfig struct {
	XMLName xml.Name        `xml:"configuration"`
	Name    string          `xml:"name,attr"`
	Type    string          `xml:"type,attr"`
	Header  XMLFWCdrHeader  `xml:"header"`
	Content XMLFWCdrContent `xml:"content"`
	Footer  XMLFWCdrFooter  `xml:"footer"`
}
*/
