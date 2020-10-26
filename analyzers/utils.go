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

package analyzers

import (
	"time"
)

type extraInfo struct {
	enc  string
	from string
	to   string
}

type InfoRPC struct {
	RequestDuration  time.Duration
	RequestStartTime time.Time
	// EndTime          time.Time

	RequestEncoding    string
	RequestSource      string
	RequestDestination string

	RequestID     uint64
	RequestMethod string
	RequestParams interface{}
	Reply         interface{}
	ReplyError    interface{}
}

type rpcAPI struct {
	ID     uint64      `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`

	StartTime time.Time
}
