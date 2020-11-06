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
	"testing"
)

func TestPathItemListNextNil(t *testing.T) {
	var e PathItemElement
	received := e.Next()
	if received != nil {
		t.Errorf("Expecting: nil, received: %+v", received)
	}
}

func TestPathItemListPrevNil(t *testing.T) {
	var e PathItemElement
	received := e.Prev()
	if received != nil {
		t.Errorf("Expecting: nil, received: %+v", received)
	}
}

func TestPathItemListFrontNil(t *testing.T) {
	var e PathItemList
	received := e.Front()
	if received != nil {
		t.Errorf("Expecting: nil, received: %+v", received)
	}
}

func TestPathItemListBackNil(t *testing.T) {
	var e PathItemList
	received := e.Back()
	if received != nil {
		t.Errorf("Expecting: nil, received: %+v", received)
	}
}

/*
func TestPathItemListNext(t *testing.T) {
	list := NewPathItemList()
	node1 := NewPathItems([]string{"path1"})
	node2 := NewPathItems([]string{"path2"})
	node3 := NewPathItems([]string{"path3"})
	list.PushFront(node1)
	list.PushFront(node2)
	list.PushFront(node3)
	fmt.Println(list.Back().Value.String())
	fmt.Println(list.Back().Prev().Value.String())

}
*/
