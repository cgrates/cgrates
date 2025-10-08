/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package utils

import (
	"fmt"
)

type PaginatorWithSearch struct {
	*Paginator
	Search string // Global matching pattern in items returned, partially used in some APIs
}

// Paginate stuff around items returned
type Paginator struct {
	Limit    *int // Limit the number of items returned
	Offset   *int // Offset of the first item returned (eg: use Limit*Page in case of PerPage items)
	MaxItems *int
}

// Clone creates a clone of the object
func (pgnt Paginator) Clone() Paginator {
	var limit *int
	if pgnt.Limit != nil {
		limit = new(int)
		*limit = *pgnt.Limit
	}

	var offset *int
	if pgnt.Offset != nil {
		offset = new(int)
		*offset = *pgnt.Offset
	}

	var maxItems *int
	if pgnt.MaxItems != nil {
		maxItems = new(int)
		*maxItems = *pgnt.MaxItems
	}
	return Paginator{
		Limit:    limit,
		Offset:   offset,
		MaxItems: maxItems,
	}
}

// GetPaginateOpts retrieves paginate options from the APIOpts map
func GetPaginateOpts(opts map[string]any) (limit, offset, maxItems int, err error) {
	if limitIface, has := opts[PageLimitOpt]; has {
		if limit, err = IfaceAsInt(limitIface); err != nil {
			return
		}
	}
	if offsetIface, has := opts[PageOffsetOpt]; has {
		if offset, err = IfaceAsInt(offsetIface); err != nil {
			return
		}
	}
	if maxItemsIface, has := opts[PageMaxItemsOpt]; has {
		if maxItems, err = IfaceAsInt(maxItemsIface); err != nil {
			return
		}
	}
	return
}

// Paginate returns a modified input sting based on the paginate options provided
func Paginate[E any](in []E, limit, offset, maxItems int) (out []E, err error) {
	if len(in) == 0 {
		return
	}
	if maxItems != 0 && maxItems < limit+offset {
		return nil, fmt.Errorf("SERVER_ERROR: maximum number of items exceeded")
	}
	if offset > len(in) {
		return
	}
	if limit == 0 && offset == 0 {
		out = in
	} else {
		if offset != 0 && limit != 0 {
			limit = limit + offset
		}
		if limit == 0 || limit > len(in) {
			limit = len(in)
		}
		out = in[offset:limit]
	}
	if maxItems != 0 && maxItems < len(out)+offset {
		return nil, fmt.Errorf("SERVER_ERROR: maximum number of items exceeded")
	}
	return
}
