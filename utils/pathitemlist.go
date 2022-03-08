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

// PathItemElement is an element of a linked list.
// Inspired by Go's double linked list
type PathItemElement struct {
	// Next and previous pointers in the doubly-linked list of elements.
	// To simplify the implementation, internally a list l is implemented
	// as a ring, such that &l.root is both the next element of the last
	// list element (l.Back()) and the previous element of the first list
	// element (l.Front()).
	next, prev *PathItemElement

	// The list to which this element belongs.
	list *PathItemList

	// The value stored with this element.
	Value []string
}

// Next returns the next list element or nil.
func (e *PathItemElement) Next() *PathItemElement {
	if p := e.next; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// Prev returns the previous list element or nil.
func (e *PathItemElement) Prev() *PathItemElement {
	if p := e.prev; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// PathItemList represents a doubly linked list.
// The zero value for PathItemList is an empty list ready to use.
type PathItemList struct {
	root PathItemElement // sentinel list element, only &root, root.prev, and root.next are used
	len  int             // current list length excluding (this) sentinel element
}

// Init initializes or clears list l.
func (l *PathItemList) Init() *PathItemList {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

// NewPathItemList returns an initialized list.
func NewPathItemList() *PathItemList { return new(PathItemList).Init() }

// Len returns the number of elements of list l.
// The complexity is O(1).
func (l *PathItemList) Len() int { return l.len }

// Front returns the first element of list l or nil if the list is empty.
func (l *PathItemList) Front() *PathItemElement {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

// Back returns the last element of list l or nil if the list is empty.
func (l *PathItemList) Back() *PathItemElement {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// lazyInit lazily initializes a zero PathItemList value.
func (l *PathItemList) lazyInit() {
	if l.root.next == nil {
		l.Init()
	}
}

// insert inserts e after at, increments l.len, and returns e.
func (l *PathItemList) insert(e, at *PathItemElement) *PathItemElement {
	n := at.next
	at.next = e
	e.prev = at
	e.next = n
	n.prev = e
	e.list = l
	l.len++
	return e
}

// insertValue is a convenience wrapper for insert(&PathItemElement{Value: v}, at).
func (l *PathItemList) insertValue(v []string, at *PathItemElement) *PathItemElement {
	return l.insert(&PathItemElement{Value: v}, at)
}

// remove removes e from its list, decrements l.len, and returns e.
func (l *PathItemList) remove(e *PathItemElement) *PathItemElement {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	e.list = nil
	l.len--
	return e
}

// move moves e to next to at and returns e.
func (l *PathItemList) move(e, at *PathItemElement) *PathItemElement {
	if e == at {
		return e
	}
	e.prev.next = e.next
	e.next.prev = e.prev

	n := at.next
	at.next = e
	e.prev = at
	e.next = n
	n.prev = e

	return e
}

// Remove removes e from l if e is an element of list l.
// It returns the element value e.Value.
// The element must not be nil.
func (l *PathItemList) Remove(e *PathItemElement) []string {
	if e.list == l {
		// if e.list == l, l must have been initialized when e was inserted
		// in l or l == nil (e is a zero PathItemElement) and l.remove will crash
		l.remove(e)
	}
	return e.Value
}

// PushFront inserts a new element e with value v at the front of list l and returns e.
func (l *PathItemList) PushFront(v []string) *PathItemElement {
	l.lazyInit()
	return l.insertValue(v, &l.root)
}

// PushBack inserts a new element e with value v at the back of list l and returns e.
func (l *PathItemList) PushBack(v []string) *PathItemElement {
	l.lazyInit()
	return l.insertValue(v, l.root.prev)
}

// InsertBefore inserts a new element e with value v immediately before mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *PathItemList) InsertBefore(v []string, mark *PathItemElement) *PathItemElement {
	if mark.list != l {
		return nil
	}
	// see comment in PathItemList.Remove about initialization of l
	return l.insertValue(v, mark.prev)
}

// InsertAfter inserts a new element e with value v immediately after mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *PathItemList) InsertAfter(v []string, mark *PathItemElement) *PathItemElement {
	if mark.list != l {
		return nil
	}
	// see comment in PathItemList.Remove about initialization of l
	return l.insertValue(v, mark)
}

// MoveToFront moves element e to the front of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *PathItemList) MoveToFront(e *PathItemElement) {
	if e.list != l || l.root.next == e {
		return
	}
	// see comment in PathItemList.Remove about initialization of l
	l.move(e, &l.root)
}

// MoveToBack moves element e to the back of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *PathItemList) MoveToBack(e *PathItemElement) {
	if e.list != l || l.root.prev == e {
		return
	}
	// see comment in PathItemList.Remove about initialization of l
	l.move(e, l.root.prev)
}

// MoveBefore moves element e to its new position before mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *PathItemList) MoveBefore(e, mark *PathItemElement) {
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.move(e, mark.prev)
}

// MoveAfter moves element e to its new position after mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *PathItemList) MoveAfter(e, mark *PathItemElement) {
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.move(e, mark)
}

// PushBackList inserts a copy of an other list at the back of list l.
// The lists l and other may be the same. They must not be nil.
func (l *PathItemList) PushBackList(other *PathItemList) {
	l.lazyInit()
	for i, e := other.Len(), other.Front(); i > 0; i, e = i-1, e.Next() {
		l.insertValue(e.Value, l.root.prev)
	}
}

// PushFrontList inserts a copy of an other list at the front of list l.
// The lists l and other may be the same. They must not be nil.
func (l *PathItemList) PushFrontList(other *PathItemList) {
	l.lazyInit()
	for i, e := other.Len(), other.Back(); i > 0; i, e = i-1, e.Prev() {
		l.insertValue(e.Value, &l.root)
	}
}
