/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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
package timespans

import (
	"testing"
)

func TestGetDestination(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	mb := &MinuteBucket{DestinationId: "nationale"}
	d := mb.getDestination(getter)
	if d.Id != "nationale" || len(d.Prefixes) != 4 {
		t.Error("Got wrong destination: ", d)
	}
}

func TestMultipleGetDestination(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	mb := &MinuteBucket{DestinationId: "nationale"}
	d := mb.getDestination(getter)
	d = mb.getDestination(getter)
	d = mb.getDestination(getter)
	if d.Id != "nationale" || len(d.Prefixes) != 4 {
		t.Error("Got wrong destination: ", d)
	}
	mb = &MinuteBucket{DestinationId: "retea"}
	d = mb.getDestination(getter)
	d = mb.getDestination(getter)
	d = mb.getDestination(getter)
	if d.Id != "retea" || len(d.Prefixes) != 2 {
		t.Error("Got wrong destination: ", d)
	}
	mb = &MinuteBucket{DestinationId: "mobil"}
	d = mb.getDestination(getter)
	d = mb.getDestination(getter)
	d = mb.getDestination(getter)
	if d.Id != "mobil" || len(d.Prefixes) != 2 {
		t.Error("Got wrong destination: ", d)
	}
}
