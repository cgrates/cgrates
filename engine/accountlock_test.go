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

package engine

import (
	"log"
	"testing"
	"time"
)

func ATestAccountLock(t *testing.T) {
	go AccLock.Guard(func() (float64, error) {
		log.Print("first 1")
		time.Sleep(1 * time.Second)
		log.Print("end first 1")
		return 0, nil
	}, "1")
	go AccLock.Guard(func() (float64, error) {
		log.Print("first 2")
		time.Sleep(1 * time.Second)
		log.Print("end first 2")
		return 0, nil
	}, "2")
	go AccLock.Guard(func() (float64, error) {
		log.Print("second 1")
		time.Sleep(1 * time.Second)
		log.Print("end second 1")
		return 0, nil
	}, "1")
	time.Sleep(3 * time.Second)
}
