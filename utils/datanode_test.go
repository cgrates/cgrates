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
	"reflect"
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
	dnRgxp        = regexp.MustCompile(`([^\.\[\]]+)`)
	dn2Rgxp       = regexp.MustCompile(`([^\[\]]+)`)
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

// Unit tests
func TestCompilePathSlice(t *testing.T) {
	path := CompilePathSlice(pathSliceBnch)
	pathExpect := []string{"Field1", "*raw", "0", "Field2", "0", "Field3", "*new", "Field5"}
	if !reflect.DeepEqual(path, pathExpect) {
		t.Errorf("Expected %q but received %q", pathExpect, path)
	}
}

func TestCompilePath(t *testing.T) {
	path := CompilePath(pathBnch)
	pathExpect := []string{"Field1", "*raw", "0", "Field2", "0", "Field3", "*new", "Field5"}
	if !reflect.DeepEqual(path, pathExpect) {
		t.Errorf("Expected %q but received %q", pathExpect, path)
	}
}

func TestField(t *testing.T) {
	//dn - DataNode
	//dl - DataLeaf
	var errExpect error
	dn := new(DataNode)
	dn.Type = NMSliceType
	_, err := dn.Field(pathSliceBnch)
	errExpectStr := `strconv.Atoi: parsing "Field1[*raw][0]": invalid syntax`
	if err == nil || err.Error() != errExpectStr {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}

	///
	path := []string{"0"}
	dn.Type = NMSliceType
	_, err = dn.Field(path)
	errExpect = ErrNotFound
	if err == nil || err != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}

	///
	path = []string{"-2"}
	dn.Type = NMSliceType
	_, err = dn.Field(path)
	errExpect = ErrNotFound
	if err == nil || err != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}

	///
	dn.Type = 3
	_, err = dn.Field(path)
	errExpect = ErrWrongPath
	if err == nil || err != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}

	///
	dn.Type = NMSliceType
	testPath := []string{"Length", "3", "test"}
	dl, err := dn.Field(testPath)
	if err != nil {
		t.Error(err)
	}
	dlExpect := &DataLeaf{Data: 0}
	if !reflect.DeepEqual(dl, dlExpect) {
		t.Errorf("Expected %q but received %q", dlExpect, dl)
	}

	///
	testPath = make([]string, 0)
	_, err = dn.Field(testPath)
	if err == nil || err != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestFieldAsInterface(t *testing.T) {
	var errExpect error
	dn := new(DataNode)
	dn.Type = NMDataType
	testPath := make([]string, 1)
	errExpect = ErrNotFound
	_, err := dn.fieldAsInterface(testPath)
	if err == nil || err != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}

	///
	dn.Type = NMMapType
	testPath = make([]string, 0)
	rcvExpect := dn.Map
	rcv, err := dn.fieldAsInterface(testPath)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, rcvExpect) {
		t.Errorf("Expected %v but received %v", rcvExpect, rcv)
	}

	///
	testPath = []string{"path"}
	dn.Map = map[string]*DataNode{
		"notPath": &DataNode{
			Type: 1,
		},
	}
	_, err = dn.fieldAsInterface(testPath)
	if err == nil || err != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}

	///
	dn.Type = NMSliceType
	testPath = []string{"Length"}
	n, err := dn.fieldAsInterface(testPath)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(n, len(dn.Slice)) {
		t.Errorf("Expected %+v but received %+v", len(dn.Slice), n)
	}

	///
	testPath = []string{"nonIntAscii"}
	_, err = dn.fieldAsInterface(testPath)
	errExpectStr := `strconv.Atoi: parsing "nonIntAscii": invalid syntax`
	if err == nil || err.Error() != errExpectStr {
		t.Errorf("Expected %+v\n but received %+v\n", errExpectStr, err)
	}

	///
	testPath = []string{"-2"}
	dn.Type = NMSliceType
	_, err = dn.fieldAsInterface(testPath)
	if err == nil || err != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}

	///
	dn.Type = 3
	_, err = dn.fieldAsInterface(testPath)
	errExpect = ErrWrongPath
	if err == nil || err != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestSet(t *testing.T) {
	dn := new(DataNode)
	testPath := make([]string, 0)
	val1 := map[string]*DataNode{
		"test": &DataNode{
			Type: 1,
		},
	}
	_, err := dn.Set(testPath, val1)
	if err != nil {
		t.Error(err)
	}
	if dn.Type != NMMapType {
		t.Errorf("Expected %+v but received %+v", NMMapType, dn.Type)
	} else if !reflect.DeepEqual(dn.Map, val1) {
		t.Errorf("Expected %+v but received %+v", val1, dn.Map)
	}

	///
	val2 := &DataLeaf{}
	_, err = dn.Set(testPath, val2)
	if err != nil {
		t.Error(err)
	}
	if dn.Type != NMDataType {
		t.Errorf("[1] Expected %+v but received %+v", NMDataType, dn.Type)
	} else if !reflect.DeepEqual(dn.Value, val2) {
		t.Errorf("[2] Expected %+v but received %+v", val2, dn.Value)
	}

	///
	val3 := 3
	tmpVal3 := &DataLeaf{
		Data: val3,
	}
	_, err = dn.Set(testPath, val3)
	if err != nil {
		t.Error(err)
	}
	if dn.Type != NMDataType {
		t.Errorf("[1] Expected %+v but received %+v", NMDataType, dn.Type)
	} else if !reflect.DeepEqual(dn.Value, tmpVal3) {
		t.Errorf("[2] Expected %+v but received %+v", tmpVal3, dn.Value)
	}

	///
	dn.Type = NMSliceType
	testPath = []string{"nonIntAscii"}
	_, err = dn.Set(testPath, val2)
	errExpectStr := `strconv.Atoi: parsing "nonIntAscii": invalid syntax`
	if err == nil || err.Error() != errExpectStr {
		t.Errorf("Expected %+v\n but received %+v\n", errExpectStr, err)
	}

	///
	dn.Type = 3
	_, err = dn.Set(testPath, val2)
	errExpect := ErrWrongPath
	if err == nil || err != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestRemovePath(t *testing.T) {
	dn := new(DataNode)
	testPath := []string{"-2"}
	dn.Type = NMDataType
	errExpect := ErrWrongPath
	if err := dn.Remove(testPath); err == nil || err != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}

	///
	dn.Slice = []*DataNode{
		{
			Value: &DataLeaf{
				Data: "testValue",
			},
		},
		{},
	}
	dn.Type = NMSliceType
	rcvExpect := "0"
	if err := dn.Remove(testPath); err != nil {
		t.Error(err)
	} else if testPath[0] != rcvExpect {
		t.Errorf("Expected %s but received %s", rcvExpect, testPath[0])
	}

	testPath = []string{"1", "path", "test"} // reminder: look at this again
	if err := dn.Remove(testPath); err != nil {
		t.Error(err)
	}

	///
	dn.Type = 3
	if err := dn.Remove(testPath); err == nil || err != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestAppend1(t *testing.T) {
	var errExpect error
	dn := new(DataNode)
	testPath := make([]string, 0)
	dn.Type = NMMapType
	val1 := &DataLeaf{
		AttributeID: "ID",
	}
	errExpect = ErrWrongPath
	if _, err := dn.Append(testPath, val1); err == nil || err != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}

	///
	dn.Type = NMDataType
	if _, err := dn.Append(testPath, val1); err != nil {
		t.Error(err)
	}

	///
	dn.Value = &DataLeaf{
		Data: "testValue",
	}
	dn.Type = NMDataType
	if _, err := dn.Append(testPath, val1); err == nil || err != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}
func TestAppend2(t *testing.T) {
	dn := new(DataNode)
	dn.Type = NMDataType
	testPath := []string{"0", "testPath"}
	val1 := &DataLeaf{
		Data: "data",
	}
	if rcv, err := dn.Append(testPath, val1); err != ErrWrongPath {
		t.Errorf("Expected %v but received %v", ErrWrongPath, err)
	} else if rcv != -1 {
		t.Errorf("Expected %v but received %v", -1, rcv)
	}

	///
	dn.Type = NMMapType
	dn.Slice = nil
	dn.Map = map[string]*DataNode{}

	if _, err := dn.Append(testPath, val1); err != nil {
		t.Error(err)
	}

	///
	dn.Type = NMSliceType
	if rcv, err := dn.Append(testPath, val1); err != nil {
		t.Error(err)
	} else if rcv == -1 {
		t.Error(err)
	}

	///
	testPath = []string{"notIntAscii", "path"}
	errExpectStr := `strconv.Atoi: parsing "notIntAscii": invalid syntax`
	if rcv, err := dn.Append(testPath, val1); err == nil || err.Error() != errExpectStr {
		t.Errorf("Expected %v but received %v", errExpectStr, err)
	} else if rcv != -1 {
		t.Errorf("Expected %v but received %v", -1, rcv)
	}

	///
	testPath = []string{"-2", "path"}
	if rcv, err := dn.Append(testPath, val1); err != ErrNotFound {
		t.Errorf("Expected %v but received %v", ErrNotFound, err)
	} else if rcv != -1 {
		t.Errorf("Expected %v but received %v", -1, rcv)
	}

	///
	testPath = []string{"0", "testPath"}
	if rcv, err := dn.Append(testPath, val1); err != nil {
		t.Error(err)
	} else if rcv == -1 {
		t.Errorf("Expected %v but received %v", -1, rcv)
	}

	///
	dn.Type = 3
	if rcv, err := dn.Append(testPath, val1); err != ErrWrongPath {
		t.Errorf("Expected %v but received %v", ErrWrongPath, err)
	} else if rcv != -1 {
		t.Errorf("Expected %v but received %v", -1, rcv)
	}
}

func TestCompose(t *testing.T) {
	dn := new(DataNode)
	testPath := make([]string, 0)
	val := &DataLeaf{
		Data: "test",
	}
	dn.Type = NMMapType
	if err := dn.Compose(testPath, val); err != ErrWrongPath {
		t.Errorf("Expected %v but received %v", ErrWrongPath, err)
	}

	///
	dn.Type = NMDataType
	dn.Value = val
	if err := dn.Compose(testPath, val); err != nil {
		t.Error(err)
	}

	///
	dn.Type = 3
	if err := dn.Compose(testPath, val); err != nil {
		t.Error(err)
	}

	///
	dn.Type = NMDataType
	testPath = []string{"0", "testPath"}
	if err := dn.Compose(testPath, val); err != ErrWrongPath {
		t.Errorf("Expected %v but received %v", ErrWrongPath, err)
	}

	///
	dn.Type = NMSliceType
	testPath = []string{"notIntAscii", "path"}
	errExpectStr := `strconv.Atoi: parsing "notIntAscii": invalid syntax`
	if err := dn.Compose(testPath, val); err == nil || err.Error() != errExpectStr {
		t.Errorf("Expected %v but received %v", errExpectStr, err)
	}

	///
	dn.Slice = nil
	testPath = []string{"0", "testPath"}
	if err := dn.Compose(testPath, val); err != nil {
		t.Error(err)
	}
	if err := dn.Compose(testPath, val); err != nil {
		t.Error(err)
	}

	///
	dn.Type = 3
	if err := dn.Compose(testPath, val); err != ErrWrongPath {
		t.Errorf("Expected %v but received %v", ErrWrongPath, err)
	}
}

func TestCompose2(t *testing.T) {
	dn := new(DataNode)
	dn.Type = NMDataType
	val := &DataLeaf{
		Data: "test",
	}
	testPath := []string{"0", "testPath"}
	if err := dn.Compose(testPath, val); err != ErrWrongPath {
		t.Errorf("Expected %v but received %v", ErrWrongPath, err)
	}

	///
	dn.Type = NMSliceType
	testPath = []string{"notIntAscii", "path"}
	errExpectStr := `strconv.Atoi: parsing "notIntAscii": invalid syntax`
	if err := dn.Compose(testPath, val); err == nil || err.Error() != errExpectStr {
		t.Errorf("Expected %v but received %v", errExpectStr, err)
	}

	///
	dn.Slice = nil
	testPath = []string{"0", "testPath"}
	if err := dn.Compose(testPath, val); err != nil {
		t.Error(err)
	}
	if err := dn.Compose(testPath, val); err != nil {
		t.Error(err)
	}

	///
	dn.Type = 3
	if err := dn.Compose(testPath, val); err != ErrWrongPath {
		t.Errorf("Expected %v but received %v", ErrWrongPath, err)
	}
}
