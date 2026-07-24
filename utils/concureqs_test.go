// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"testing"
)

func TestConcureqsNewConReqs(t *testing.T) {
	expected := &ConcReqs{
		limit:    5,
		strategy: MetaBusy,
		aReqs:    make(chan struct{}, 5),
	}
	for i := 0; i < 5; i++ {
		expected.aReqs <- struct{}{}
	}
	rcv := NewConReqs(5, MetaBusy)

	if rcv.limit != expected.limit || rcv.strategy != expected.strategy || len(rcv.aReqs) != len(expected.aReqs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expected, rcv)
	}
}

func TestConcureqsAllocate(t *testing.T) {

	cases := []struct {
		name         string
		limit        int
		rcv          int
		emptyChannel bool
		err          error
	}{
		{
			name:         "if limit is equal to 0",
			limit:        0,
			rcv:          0,
			emptyChannel: false,
			err:          nil,
		},
		{
			name:         "limit more than 0",
			limit:        5,
			rcv:          4,
			emptyChannel: false,
			err:          nil,
		},
		{
			name:         "limit more than 0 but channel is empty",
			limit:        5,
			rcv:          0,
			emptyChannel: true,
			err:          errDeny,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cr := NewConReqs(c.limit, MetaBusy)

			if c.emptyChannel {
				for i := 0; i < c.limit; i++ {
					<-cr.aReqs
				}
			}

			err := cr.Allocate()

			if err != c.err {
				t.Fatal(err)
			}

			if len(cr.aReqs) != c.rcv {
				t.Errorf("expected: <%+v>, received: <%+v>", len(cr.aReqs), c.rcv)
			}
		})
	}
}

func TestConcureqsDeallocate(t *testing.T) {

	cases := []struct {
		name  string
		limit int
	}{
		{
			name:  "limit is 0",
			limit: 0,
		},
		{
			name:  "limit is not 0",
			limit: 5,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cr := NewConReqs(c.limit, MetaBusy)

			if c.limit != 0 {
				go cr.Deallocate()
				_, ok := <-cr.aReqs

				if !ok {
					t.Error("didn't recive from the channel")
				}
			} else {
				cr.Deallocate()

				if len(cr.aReqs) != c.limit {
					t.Errorf("expected: <%+v>, received: <%+v>", len(cr.aReqs), c.limit)
				}
			}

		})
	}
}
