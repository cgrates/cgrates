/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"errors"
	"fmt"

	"github.com/cgrates/cgrates/utils"
)

// Used as generic function logic for various fields

// Attributes
//  source - the base source
//  width - the field width
//  strip - if present it will specify the strip strategy, when missing strip will not be allowed
//  padding - if present it will specify the padding strategy to use, left, right, zeroleft, zeroright
func FmtFieldWidth(source string, width int, strip, padding string, mandatory bool) (string, error) {
	if mandatory && len(source) == 0 {
		return "", errors.New("Empty source value")
	}
	if width == 0 { // Disable width processing if not defined
		return source, nil
	}
	if len(source) == width { // the source is exactly the maximum length
		return source, nil
	}
	if len(source) > width { //the source is bigger than allowed
		if len(strip) == 0 {
			return "", fmt.Errorf("Source %s is bigger than the width %d, no strip defied", source, width)
		}
		if strip == "right" {
			return source[:width], nil
		} else if strip == "xright" {
			return source[:width-1] + "x", nil // Suffix with x to mark prefix
		} else if strip == "left" {
			diffIndx := len(source) - width
			return source[diffIndx:], nil
		} else if strip == "xleft" { // Prefix one x to mark stripping
			diffIndx := len(source) - width
			return "x" + source[diffIndx+1:], nil
		}
	} else { //the source is smaller as the maximum allowed
		if len(padding) == 0 {
			return "", fmt.Errorf("Source %s is smaller than the width %d, no padding defined", source, width)
		}
		var paddingFmt string
		switch padding {
		case "right":
			paddingFmt = fmt.Sprintf("%%-%ds", width)
		case "left":
			paddingFmt = fmt.Sprintf("%%%ds", width)
		case "zeroleft":
			paddingFmt = fmt.Sprintf("%%0%ds", width)
		}
		if len(paddingFmt) != 0 {
			return fmt.Sprintf(paddingFmt, source), nil
		}
	}
	return source, nil
}

// Mask a number of characters in the suffix of the destination
func MaskDestination(dest string, maskLen int) string {
	destLen := len(dest)
	if maskLen < 0 {
		return dest
	} else if maskLen > destLen {
		maskLen = destLen
	}
	dest = dest[:destLen-maskLen]
	for i := 0; i < maskLen; i++ {
		dest += utils.MASK_CHAR
	}
	return dest
}
