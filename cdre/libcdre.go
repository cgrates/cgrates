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
	"github.com/cgrates/cgrates/utils"
)

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
