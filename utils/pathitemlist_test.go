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
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestPathItemListNext(t *testing.T) {
	list := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	node3 := []string{"path3"}
	list.PushFront(node3)
	list.PushFront(node2)
	list.PushFront(node1)
	if !reflect.DeepEqual("path2", strings.Join(list.Front().Next().Value, NestingSep)) {
		t.Errorf("Expecting: <path2>, received: <%+v>", strings.Join(list.Front().Next().Value, NestingSep))
	}

}

func TestPathItemListNextNil(t *testing.T) {
	var e PathItemElement
	received := e.Next()
	if received != nil {
		t.Errorf("Expecting: <nil>, received: <%+v>", received)
	}
}

func TestPathItemListPrev(t *testing.T) {
	list := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	node3 := []string{"path3"}
	list.PushFront(node3)
	list.PushFront(node2)
	list.PushFront(node1)
	if !reflect.DeepEqual("path2", strings.Join(list.Back().Prev().Value, NestingSep)) {
		t.Errorf("Expecting: <path2>, received: <%+v>", strings.Join(list.Back().Prev().Value, NestingSep))
	}

}

func TestPathItemListPrevNil(t *testing.T) {
	var e PathItemElement
	received := e.Prev()
	if received != nil {
		t.Errorf("Expecting: <nil>, received: <%+v>", received)
	}
}

func TestPathItemListInit(t *testing.T) {
	list := NewPathItemList()
	list2 := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	node3 := []string{"path3"}
	list.PushFront(node3)
	list.PushFront(node2)
	list.PushFront(node1)
	if !reflect.DeepEqual(list2, list.Init()) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", list2, list.Init())
	}
}

func TestPathItemListNewPathItemList(t *testing.T) {
	list := NewPathItemList()
	list2 := new(PathItemList).Init()
	if !reflect.DeepEqual(list2, list) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", list2, list)
	}
}

func TestPathItemListLen(t *testing.T) {
	list := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	node3 := []string{"path3"}
	list.PushFront(node3)
	list.PushFront(node2)
	list.PushFront(node1)
	if !reflect.DeepEqual(3, list.Len()) {
		t.Errorf("Expecting: <3>, received: <%+v>", list.Len())
	}
}

func TestPathItemListFront(t *testing.T) {
	list := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	list.PushBack(node1)
	list.PushBack(node2)
	if !reflect.DeepEqual("path1", strings.Join(list.Front().Value, NestingSep)) {
		t.Errorf("Expecting: <path1>, received: <%+v>", strings.Join(list.Front().Value, NestingSep))
	}

}

func TestPathItemListFrontNil(t *testing.T) {
	var e PathItemList
	received := e.Front()
	if received != nil {
		t.Errorf("Expecting: <nil>, received: <%+v>", received)
	}
}

func TestPathItemListBack(t *testing.T) {
	list := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	list.PushBack(node1)
	list.PushBack(node2)
	if !reflect.DeepEqual("path2", strings.Join(list.Back().Value, NestingSep)) {
		t.Errorf("Expecting: <path2>, received: <%+v>", strings.Join(list.Back().Value, NestingSep))
	}

}

func TestPathItemListBackNil(t *testing.T) {
	var e PathItemList
	received := e.Back()
	if received != nil {
		t.Errorf("Expecting: <nil>, received: <%+v>", received)
	}
}

func TestPathItemListLazyInit(t *testing.T) {
	list := NewPathItemList()
	list2 := NewPathItemList()
	list.root.next = nil
	list.lazyInit()
	if !reflect.DeepEqual(list2, list) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", list2, list)
	}
}

func TestPathItemListInsert(t *testing.T) {
	list := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	list.PushBack(node1)
	list.PushBack(node2)
	list.insert(list.Front(), list.Back())
	if !reflect.DeepEqual("path1", strings.Join(list.Back().Value, NestingSep)) {
		t.Errorf("Expecting: <path1>, received: <%+v>", strings.Join(list.Back().Value, NestingSep))
	}
}

func TestPathItemListInsertValue(t *testing.T) {
	list := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	list.PushBack(node1)
	list.PushBack(node2)
	list.insertValue([]string{"path3"}, list.Back())
	if !reflect.DeepEqual("path3", strings.Join(list.Back().Value, NestingSep)) {
		t.Errorf("Expecting: <path3>, received: <%+v>", strings.Join(list.Back().Value, NestingSep))
	}
}

func TestPathItemListRemoveLowerCase(t *testing.T) {
	list := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	node3 := []string{"path3"}
	list.PushBack(node1)
	list.PushBack(node2)
	list.PushBack(node3)
	list.remove(list.Back())
	if !reflect.DeepEqual("path2", strings.Join(list.Back().Value, NestingSep)) {
		t.Errorf("Expecting: <path2>, received: <%+v>", strings.Join(list.Back().Value, NestingSep))
	}
}

func TestPathItemListRemoveUpperCase1(t *testing.T) {
	list := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	node3 := []string{"path3"}
	list.PushBack(node1)
	list.PushBack(node2)
	list.PushBack(node3)
	list.Remove(list.Front().Next())
	if !reflect.DeepEqual("path3", strings.Join(list.Front().Next().Value, NestingSep)) {
		t.Errorf("Expecting: <path3>, received: <%+v>", strings.Join(list.Front().Next().Value, NestingSep))
	}
}

func TestPathItemListRemoveUpperCase2(t *testing.T) {
	list1 := NewPathItemList()
	list2 := NewPathItemList()
	node1 := []string{"pathA"}
	node2 := []string{"pathB"}
	list1.PushBack(node1)
	list2.PushBack(node2)
	list1.Remove(list2.Front())
	if !reflect.DeepEqual("pathA", strings.Join(list1.Front().Value, NestingSep)) {
		t.Errorf("Expecting: <pathA>, received: <%+v>", strings.Join(list1.Front().Value, NestingSep))
	}

}

func TestPathItemListMoveEqual(t *testing.T) {
	list := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	node3 := []string{"path3"}
	list.PushBack(node1)
	list.PushBack(node2)
	list.PushBack(node3)
	list.move(list.Back(), list.Back())
	if !reflect.DeepEqual("path3", strings.Join(list.Back().Value, NestingSep)) {
		t.Errorf("Expecting: <path3>, received: <%+v>", strings.Join(list.Back().Value, NestingSep))
	}
}

func TestPathItemListMove(t *testing.T) {
	list := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	node3 := []string{"path3"}
	list.PushBack(node1)
	list.PushBack(node2)
	list.PushBack(node3)
	list.move(list.Front(), list.Back())
	if !reflect.DeepEqual("path3", strings.Join(list.Front().next.Value, NestingSep)) {
		t.Errorf("Expecting: <path3>, received: <%+v>", strings.Join(list.Front().next.Value, NestingSep))
	}
	if !reflect.DeepEqual("path1", strings.Join(list.Back().Value, NestingSep)) {
		t.Errorf("Expecting: <path1>, received: <%+v>", strings.Join(list.Back().Value, NestingSep))
	}
	if !reflect.DeepEqual("path2", strings.Join(list.Front().Value, NestingSep)) {
		t.Errorf("Expecting: <path2>, received: <%+v>", strings.Join(list.Front().Value, NestingSep))
	}
}

func TestPathItemListPushFront(t *testing.T) {
	list := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	node3 := []string{"path3"}
	list.PushBack(node1)
	list.PushBack(node2)
	list.PushBack(node3)
	list.PushFront([]string{"path4"})
	if !reflect.DeepEqual("path4", strings.Join(list.Front().Value, NestingSep)) {
		t.Errorf("Expecting: <path4>, received: <%+v>", strings.Join(list.Front().Value, NestingSep))
	}
}

func TestPathItemListPushBack(t *testing.T) {
	list := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	node3 := []string{"path3"}
	list.PushBack(node1)
	list.PushBack(node2)
	list.PushBack(node3)
	list.PushBack([]string{"path4"})
	if !reflect.DeepEqual("path4", strings.Join(list.Back().Value, NestingSep)) {
		t.Errorf("Expecting: <path4>, received: <%+v>", strings.Join(list.Back().Value, NestingSep))
	}
}

func TestPathItemListInsertBefore(t *testing.T) {
	list := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	node3 := []string{"path3"}
	list.PushBack(node1)
	list.PushBack(node2)
	list.PushBack(node3)
	list.InsertBefore([]string{"path4"}, list.Back())
	if !reflect.DeepEqual("path4", strings.Join(list.Back().Prev().Value, NestingSep)) {
		t.Errorf("Expecting: <path4>, received: <%+v>", strings.Join(list.Back().Prev().Value, NestingSep))
	}

}

func TestPathItemListInsertBeforeNil(t *testing.T) {
	list1 := NewPathItemList()
	list2 := NewPathItemList()
	var received *PathItemElement
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	list1.PushBack(node1)
	list2.PushBack(node2)
	received = list1.InsertBefore([]string{"path4"}, list2.Back())
	if received != nil {
		t.Errorf("Expecting: <%+v>, received: <%+v>", nil, received)
	}
}

func TestPathItemListInsertAfter(t *testing.T) {
	list := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	node3 := []string{"path3"}
	list.PushBack(node1)
	list.PushBack(node2)
	list.PushBack(node3)
	list.InsertAfter([]string{"path4"}, list.Front())
	if !reflect.DeepEqual("path4", strings.Join(list.Front().Next().Value, NestingSep)) {
		t.Errorf("Expecting: <path4>, received: <%+v>", strings.Join(list.Front().Next().Value, NestingSep))
	}
}

func TestPathItemListInsertAfterNil(t *testing.T) {
	list1 := NewPathItemList()
	list2 := NewPathItemList()
	var received *PathItemElement
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	list1.PushBack(node1)
	list2.PushBack(node2)
	received = list1.InsertAfter([]string{"path4"}, list2.Back())
	if received != nil {
		t.Errorf("Expecting: <%+v>, received: <%+v>", nil, received)
	}
}

func TestPathItemListMoveToFrontCase1(t *testing.T) {
	list1 := NewPathItemList()
	list2 := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	list1.PushBack(node1)
	list2.PushBack(node2)
	list1.MoveToFront(list2.Back())
	if !reflect.DeepEqual("path1", strings.Join(list1.Front().Value, NestingSep)) {
		t.Errorf("Expecting: <path3>, received: <%+v>", strings.Join(list1.Front().Value, NestingSep))
	}

}

func TestPathItemListMoveToFrontCase2(t *testing.T) {
	list := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	node3 := []string{"path3"}
	list.PushBack(node1)
	list.PushBack(node2)
	list.PushBack(node3)
	list.MoveToFront(list.Back())
	if !reflect.DeepEqual("path3", strings.Join(list.Front().Value, NestingSep)) {
		t.Errorf("Expecting: <path3>, received: <%+v>", strings.Join(list.Front().Value, NestingSep))
	}
}

func TestPathItemListMoveToBackCase1(t *testing.T) {
	list1 := NewPathItemList()
	list2 := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	list1.PushBack(node1)
	list2.PushBack(node2)
	list1.MoveToBack(list2.Back())
	if !reflect.DeepEqual("path1", strings.Join(list1.Back().Value, NestingSep)) {
		t.Errorf("Expecting: <path1>, received: <%+v>", strings.Join(list1.Back().Value, NestingSep))
	}

}

func TestPathItemListMoveToBackCase2(t *testing.T) {
	list := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	node3 := []string{"path3"}
	list.PushBack(node1)
	list.PushBack(node2)
	list.PushBack(node3)
	list.MoveToBack(list.Front())
	if !reflect.DeepEqual("path1", strings.Join(list.Back().Value, NestingSep)) {
		t.Errorf("Expecting: <path1>, received: <%+v>", strings.Join(list.Front().Value, NestingSep))
	}
}

func TestPathItemListMoveBeforeCase1(t *testing.T) {
	list1 := NewPathItemList()
	list2 := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	list1.PushBack(node1)
	list2.PushBack(node2)
	list1.MoveBefore(list2.Front(), list1.Back())
	if !reflect.DeepEqual("path1", strings.Join(list1.Front().Value, NestingSep)) {
		t.Errorf("Expecting: <path1>, received: <%+v>", strings.Join(list1.Front().Value, NestingSep))
	}

}

func TestPathItemListMoveBeforeCase2(t *testing.T) {
	list := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	node3 := []string{"path3"}
	list.PushBack(node1)
	list.PushBack(node2)
	list.PushBack(node3)
	list.MoveBefore(list.Front(), list.Back())
	if !reflect.DeepEqual("path1", strings.Join(list.Back().Prev().Value, NestingSep)) {
		t.Errorf("Expecting: <path1>, received: <%+v>", strings.Join(list.Back().Prev().Value, NestingSep))
	}

}

func TestPathItemListMoveAfterCase1(t *testing.T) {
	list1 := NewPathItemList()
	list2 := NewPathItemList()
	node1 := []string{"pathA"}
	node2 := []string{"pathB"}
	list1.PushBack(node1)
	list2.PushBack(node2)
	list1.MoveAfter(list2.Front(), list1.Back())
	if !reflect.DeepEqual("pathA", strings.Join(list1.Back().Value, NestingSep)) {
		t.Errorf("Expecting: <pathA>, received: <%+v>", strings.Join(list1.Back().Value, NestingSep))
	}

}

func TestPathItemListMoveAfterCase2(t *testing.T) {
	list := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	node3 := []string{"path3"}
	list.PushBack(node1)
	list.PushBack(node2)
	list.PushBack(node3)
	list.MoveAfter(list.Front(), list.Back())
	if !reflect.DeepEqual("path1", strings.Join(list.Back().Value, NestingSep)) {
		t.Errorf("Expecting: <path1>, received: <%+v>", strings.Join(list.Back().Value, NestingSep))
	}

}

func TestPathItemListPushBackList(t *testing.T) {
	list1 := NewPathItemList()
	list2 := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	nodeA := []string{"pathA"}
	nodeB := []string{"pathB"}
	list1.PushBack(node1)
	list1.PushBack(node2)
	list2.PushBack(nodeA)
	list2.PushBack(nodeB)
	list1.PushBackList(list2)
	if !reflect.DeepEqual("pathB", strings.Join(list1.Back().Value, NestingSep)) {
		t.Errorf("Expecting: <pathB>, received: <%+v>", strings.Join(list1.Back().Value, NestingSep))
	}
	if !reflect.DeepEqual("pathA", strings.Join(list1.Back().Prev().Value, NestingSep)) {
		t.Errorf("Expecting: <pathA>, received: <%+v>", strings.Join(list1.Back().Prev().Value, NestingSep))
	}
}

func TestPathItemListPushFrontList(t *testing.T) {
	list1 := NewPathItemList()
	list2 := NewPathItemList()
	node1 := []string{"path1"}
	node2 := []string{"path2"}
	nodeA := []string{"pathA"}
	nodeB := []string{"pathB"}
	list1.PushBack(node1)
	list1.PushBack(node2)
	list2.PushBack(nodeA)
	list2.PushBack(nodeB)
	list1.PushFrontList(list2)
	if !reflect.DeepEqual("pathA", strings.Join(list1.Front().Value, NestingSep)) {
		t.Errorf("Expecting: <pathA>, received: <%+v>", strings.Join(list1.Front().Value, NestingSep))
	}
	if !reflect.DeepEqual("pathB", strings.Join(list1.Front().Next().Value, NestingSep)) {
		t.Errorf("Expecting: <pathB>, received: <%+v>", strings.Join(list1.Front().Next().Value, NestingSep))
	}
}

// const benchPath = "~*req.Field1[0][-1][*raw][10][path1][path2][path3]"

// const benchPath = "~*req.Field1[*raw]"

const benchPath = "Field1[1000000000000000]"

// Benchmark results:
// goos: linux
// goarch: amd64
// pkg: github.com/cgrates/cgrates/utils
// cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
// BenchmarkGetPathIndexSlice
// BenchmarkGetPathIndexSlice             	16462084	       360.8 ns/op	     160 B/op	       2 allocs/op
// BenchmarkGetPathIndexSliceSplit
// BenchmarkGetPathIndexSliceSplit        	16291755	       359.4 ns/op	     112 B/op	       1 allocs/op
// BenchmarkGetPathIndexSliceStringsIndex
// BenchmarkGetPathIndexSliceStringsIndex 	10418586	       597.2 ns/op	     240 B/op	       4 allocs/op
// BenchmarkGetPathIndexString
// BenchmarkGetPathIndexString            	96080587	        61.19 ns/op	      16 B/op	       1 allocs/op
// PASS
// ok  	github.com/cgrates/cgrates/utils	25.372s

func GetPathIndexSlice1(spath string) (opath string, idx []string) {
	idxStart := strings.Index(spath, IdxStart)
	if idxStart == -1 || !strings.HasSuffix(spath, IdxEnd) {
		return spath, nil
	}
	idxVal := spath[idxStart+1 : len(spath)-1]
	opath = spath[:idxStart]
	if len(idxVal) <= 3 {
		return opath, []string{idxVal}
	}
	idxValB := []byte(idxVal)
	n := bytes.Count(idxValB, []byte{']'}) // we number only ] as an optimization
	if n <= 0 {
		return opath, []string{idxVal}
	}
	idx = make([]string, n+1) // alloc the memory for the slice
	for i := 0; i < n; i++ {  // expect a valid path
		ix := bytes.IndexByte(idxValB, ']') // safe to asume that ix is not -1 as we counted before
		if ix == len(idxValB)-1 ||
			idxValB[ix+1] != '[' { // this is clearly an error so stop
			return spath, nil
		}
		idx[i] = idxVal[:ix]
		idxValB = idxValB[ix+2:]
		idxVal = idxVal[ix+2:]
	}
	idx[n] = idxVal
	return
}

func GetPathIndexSliceStringsIndex(spath string) (opath string, idx []string) {
	idxStart := strings.Index(spath, IdxStart)
	if idxStart == -1 || !strings.HasSuffix(spath, IdxEnd) {
		return spath, nil
	}
	idxVal := spath[idxStart+1 : len(spath)-1]
	opath = spath[:idxStart]
	if len(idxVal) <= 3 {
		return opath, []string{idxVal}
	}

	for ix := strings.Index(idxVal, IdxCombination); ix != -1; ix = strings.Index(idxVal, IdxCombination) {
		idx = append(idx, idxVal[:ix])
		idxVal = idxVal[ix+1:]

	}
	idx = append(idx, string(idxVal))

	return
}

func BenchmarkGetPathIndexSlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetPathIndexSlice1(benchPath)
	}
}

/*
	func BenchmarkGetPathIndexSliceSplit(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GetPathIndexSlice(benchPath)
		}
	}
*/
func BenchmarkGetPathIndexSliceStringsIndex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetPathIndexSliceStringsIndex(benchPath)
	}
}
func BenchmarkGetPathIndexString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetPathIndexString(benchPath)
	}
}
