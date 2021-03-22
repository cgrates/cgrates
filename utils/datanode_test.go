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
	"regexp"
	"strings"
	"testing"
)

// Bench result
/*
goos: linux
goarch: amd64
pkg: github.com/cgrates/cgrates/utils
cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
BenchmarkCompilePathRegex
BenchmarkCompilePathRegex      	 3744520	      3252 ns/op	     416 B/op	       9 allocs/op
BenchmarkCompilePath
BenchmarkCompilePath           	17784284	       669.0 ns/op	     288 B/op	       4 allocs/op
BenchmarkCompilePathSliceRegex
BenchmarkCompilePathSliceRegex 	 2804264	      4358 ns/op	    1088 B/op	      14 allocs/op
BenchmarkCompilePathSlice
BenchmarkCompilePathSlice      	24116252	       501.2 ns/op	     224 B/op	       3 allocs/op
PASS
ok  	github.com/cgrates/cgrates/utils	57.174s

*/
const pathBnch = "Field1[*raw][0].Field2[0].Field3[*new].Field5"

var (
	pathSliceBnch = strings.Split("Field1[*raw][0].Field2[0].Field3[*new].Field5", NestingSep)

	dnRgxp  = regexp.MustCompile(`([^\.\[\]]+)`)
	dn2Rgxp = regexp.MustCompile(`([^\[\]]+)`)
)

func CompilePathR(spath string) (path []string) {
	return dnRgxp.FindAllString(spath, -1)
}

func CompilePathSliceR(spath []string) (path []string) {
	path = make([]string, 0, len(spath))
	for _, p := range spath {
		path = append(path, dn2Rgxp.FindAllString(p, -1)...)
	}
	return
}

func BenchmarkCompilePathRegex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CompilePathR(pathBnch)
	}
}

func BenchmarkCompilePath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CompilePath(pathBnch)
	}
}

func BenchmarkCompilePathSliceRegex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CompilePathSliceR(pathSliceBnch)
	}
}

func BenchmarkCompilePathSlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CompilePathSlice(pathSliceBnch)
	}
}
