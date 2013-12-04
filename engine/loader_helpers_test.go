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

package engine

import (
	"github.com/cgrates/cgrates/utils"
	"strings"
	"testing"
	"io"
	"reflect"
	"bufio"
)

var ratesTest = `#Tag,DestinationRatesTag,TimingTag,Weight
RT_1CENT,0,1,1s,1s,0s,*up,2
DUMMY,INVALID;DATA
`

func TestTPCSVFileParser(t *testing.T) {
	bfRdr := bufio.NewReader( strings.NewReader(ratesTest) )
	fParser := &TPCSVFileParser{FileValidators[utils.RATES_CSV], bfRdr}
	lineNr := 0
	for {
		lineNr++
		record, err := fParser.ParseNextLine()
		if err == io.EOF { // Reached end of the string
			break
		}
		switch lineNr {
		case 1:
			if err == nil || err.Error() != "Line starts with comment character." {
				t.Error("Failed to detect comment character")
			}
		case 2:
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual( record, []string{"RT_1CENT","0","1","1s","1s","0s","*up","2"}) {
				t.Error("Unexpected record extracted", record)
			}
		case 3:
			if err==nil {
				t.Error("Expecting invalid line at row 3")
			}
		}
	}
}
	

