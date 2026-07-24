// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"reflect"
	"testing"
)

func TestPathItemListPrev(t *testing.T) {

	pr := PathItemElement{}

	pl := PathItemList{
		root: pr,
		len:  3,
	}

	p2 := PathItemElement{
		prev: &pr,
	}

	p3 := PathItemElement{
		prev: &p2,
		list: &pl,
	}

	p2.next = &p3

	rcv := p3.Prev()

	if !reflect.DeepEqual(rcv, &p2) {
		t.Errorf("recived %v, expected %v", rcv, &p2)
	}

	t.Run("return nil", func(t *testing.T) {
		rcv := p2.Prev()

		if rcv != nil {
			t.Errorf("recived %v, expected nil", rcv)
		}
	})

}

func TestPathItemListBack(t *testing.T) {
	pr := PathItemElement{}

	pr.prev = &pr

	pl := PathItemList{
		root: pr,
		len:  3,
	}

	p2 := PathItemElement{
		prev: &pr,
	}

	p3 := PathItemElement{
		prev: &p2,
		list: &pl,
	}

	p2.next = &p3

	rcv := pl.Back()

	if !reflect.DeepEqual(rcv, &pr) {
		t.Errorf("recived %v, expected %v", rcv, p3)
	}

	pl.len = 0

	rcv = pl.Back()

	if rcv != nil {
		t.Error("expected nil")
	}
}

func TestPathItemListFront(t *testing.T) {

	pr := PathItemElement{}

	pr.prev = &pr

	pl := PathItemList{
		root: pr,
		len:  3,
	}

	p2 := PathItemElement{
		prev: &pr,
		list: &pl,
	}

	p3 := PathItemElement{
		prev: &p2,
		list: &pl,
	}

	p2.next = &p3

	pr.next = &p2

	pr.list = &pl

	rcv := pl.Front()

	if rcv != nil {
		t.Errorf("recived %v, expected %v", rcv, p2)
	}

	pl.len = 0

	rcv = pl.Front()

	if rcv != nil {
		t.Errorf("recived %v, expected nil", rcv)
	}
}

func TestPathItemListInsertBefore(t *testing.T) {

	pr := PathItemElement{}

	pl := PathItemList{
		root: pr,
		len:  3,
	}

	p2 := PathItemElement{
		prev: &pr,
	}

	p3 := PathItemElement{
		prev: &p2,
		list: &pl,
	}

	p2.next = &p3

	pi := PathItem{}

	ps := PathItems{pi}

	pTest := PathItemElement{
		list: nil,
	}

	type args struct {
		v    PathItems
		mark *PathItemElement
	}

	tests := []struct {
		name string
		args args
		exp  *PathItemElement
	}{
		{
			name: "different path item lists",
			args: args{ps, &pTest},
			exp:  nil,
		},
		{
			name: "returns new path item element",
			args: args{ps, &p3},
			exp:  nil,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rcv := pl.InsertBefore(tt.args.v, tt.args.mark)

			if i < 1 {
				if rcv != nil {
					t.Errorf("recived %v, expected nil", rcv)
				}
			} else {
				if rcv == nil {
					t.Errorf("recived %v", rcv)
				}
			}
		})
	}
}

func TestPathItemListInsertAfter(t *testing.T) {

	pr := PathItemElement{}

	pl := PathItemList{
		root: pr,
		len:  3,
	}

	pr.list = &pl

	p2 := PathItemElement{
		prev: &pr,
		list: &pl,
	}

	p3 := PathItemElement{
		prev: &p2,
		list: &pl,
	}

	p2.next = &p3

	pi := PathItem{}

	ps := PathItems{pi}

	pTest := PathItemElement{
		list: nil,
	}

	type args struct {
		v    PathItems
		mark *PathItemElement
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "different path item lists",
			args: args{ps, &pTest},
		},
		{
			name: "returns new path item element",
			args: args{ps, &p2},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rcv := pl.InsertAfter(tt.args.v, tt.args.mark)

			if i < 1 {
				if rcv != nil {
					t.Errorf("recived %v, expected nil", rcv)
				}
			} else {
				if rcv == nil {
					t.Errorf("recived %v", rcv)
				}
			}
		})
	}
}

func TestPathItemListPushBack(t *testing.T) {

	pr := PathItemElement{}

	pl := PathItemList{
		root: pr,
		len:  3,
	}

	pr.list = &pl

	p2 := PathItemElement{
		prev: &pr,
		list: &pl,
	}

	p3 := PathItemElement{
		prev: &p2,
		list: &pl,
	}

	p2.next = &p3

	pi := PathItem{}

	ps := PathItems{pi}

	rcv := pl.PushBack(ps)

	if rcv == nil {
		t.Errorf("recived %v", rcv)
	}
}

func TestPathItemListPushFront(t *testing.T) {

	pr := PathItemElement{}

	pl := PathItemList{
		root: pr,
		len:  3,
	}

	pr.list = &pl

	p2 := PathItemElement{
		prev: &pr,
		list: &pl,
	}

	p3 := PathItemElement{
		prev: &p2,
		list: &pl,
	}

	p2.next = &p3

	pi := PathItem{}

	ps := PathItems{pi}

	rcv := pl.PushFront(ps)

	if rcv == nil {
		t.Errorf("recived %v", rcv)
	}
}

func TestPathItemListLen(t *testing.T) {

	pl := PathItemList{
		len: 3,
	}

	rcv := pl.Len()

	if rcv != 3 {
		t.Errorf("recived %d, expected 3", rcv)
	}
}

func TestPathItemListMoveToFront(t *testing.T) {

	pr := PathItemElement{
		Value: PathItems{PathItem{Field: "test"}},
	}

	pl := PathItemList{
		root: pr,
		len:  3,
	}

	pr.list = &pl

	p2 := PathItemElement{
		prev: &pr,
		list: &pl,
	}

	p3 := PathItemElement{
		prev: &p2,
		list: &pl,
	}

	p2.next = &p3

	p := PathItemElement{
		Value: PathItems{PathItem{Field: "test"}},
	}

	pl.MoveToFront(&p)
}

func TestPathItemListMoveBefore(t *testing.T) {
	pr := PathItemElement{}

	pl := PathItemList{
		root: pr,
		len:  3,
	}

	pr.list = &pl

	p2 := PathItemElement{
		prev: &pr,
		list: &pl,
	}

	p3 := PathItemElement{
		prev: &p2,
		list: &pl,
	}

	p2.next = &p3

	p := PathItemElement{
		Value: PathItems{PathItem{Field: "test"}},
	}

	pl.MoveBefore(&p, &p2)

	if len(p2.prev.Value) != 0 {
		t.Error("moved before")
	}
}

func TestPathItemListMoveAfter(t *testing.T) {

	pr := PathItemElement{}

	pl := PathItemList{
		root: pr,
		len:  3,
	}

	pr.list = &pl

	p2 := PathItemElement{
		prev: &pr,
		list: &pl,
	}

	p3 := PathItemElement{
		prev: &p2,
		list: &pl,
	}

	p2.next = &p3

	p := PathItemElement{
		Value: PathItems{PathItem{Field: "test"}},
	}

	pl.MoveAfter(&p, &p2)

	if len(p2.prev.Value) != 0 {
		t.Error("moved after")
	}
}

func TestPathItemListmove(t *testing.T) {
	l := PathItemList{}

	rcv := l.move(nil, nil)

	if rcv != nil {
		t.Error(rcv)
	}
}
