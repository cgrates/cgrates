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
	"testing"
)

func TestPathItemListNext(t *testing.T) {
	list := NewPathItemList()
	node1 := NewPathItems([]string{"path1"})
	node2 := NewPathItems([]string{"path2"})
	node3 := NewPathItems([]string{"path3"})
	list.PushFront(node3)
	list.PushFront(node2)
	list.PushFront(node1)
	if !reflect.DeepEqual("path2", list.Front().Next().Value.String()) {
		t.Errorf("Expecting: <path1>, received: <%+v>", list.Front().Value.String())
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
	node1 := NewPathItems([]string{"path1"})
	node2 := NewPathItems([]string{"path2"})
	node3 := NewPathItems([]string{"path3"})
	list.PushFront(node3)
	list.PushFront(node2)
	list.PushFront(node1)
	if !reflect.DeepEqual("path2", list.Back().Prev().Value.String()) {
		t.Errorf("Expecting: <path1>, received: <%+v>", list.Front().Value.String())
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
	node1 := NewPathItems([]string{"path1"})
	node2 := NewPathItems([]string{"path2"})
	node3 := NewPathItems([]string{"path3"})
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
	node1 := NewPathItems([]string{"path1"})
	node2 := NewPathItems([]string{"path2"})
	node3 := NewPathItems([]string{"path3"})
	list.PushFront(node3)
	list.PushFront(node2)
	list.PushFront(node1)
	if !reflect.DeepEqual(3, list.Len()) {
		t.Errorf("Expecting: <3>, received: <%+v>", list.Init())
	}
}

func TestPathItemListFront(t *testing.T) {
	list := NewPathItemList()
	node1 := NewPathItems([]string{"path1"})
	node2 := NewPathItems([]string{"path2"})
	list.PushBack(node1)
	list.PushBack(node2)
	if !reflect.DeepEqual("path1", list.Front().Value.String()) {
		t.Errorf("Expecting: <path1>, received: <%+v>", list.Front().Value.String())
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
	node1 := NewPathItems([]string{"path1"})
	node2 := NewPathItems([]string{"path2"})
	list.PushBack(node1)
	list.PushBack(node2)
	if !reflect.DeepEqual("path2", list.Back().Value.String()) {
		t.Errorf("Expecting: <path2>, received: <%+v>", list.Back().Value.String())
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
	node1 := NewPathItems([]string{"path1"})
	node2 := NewPathItems([]string{"path2"})
	list.PushBack(node1)
	list.PushBack(node2)
	list.insert(list.Front(), list.Back())
	if !reflect.DeepEqual("path1", list.Back().Value.String()) {
		t.Errorf("Expecting: <path1>, received: <%+v>", list.Back().Value.String())
	}
}
