/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package actions

import (
	"context"
	"fmt"

	"github.com/cgrates/cgrates/utils"
)

// Action is implemented by each action type
type actioner interface {
	execute(ctx context.Context, data interface{}) (err error)
}

// newAction is the constructor to create actioner
func newActioner(typ string) (act actioner, err error) {
	switch typ {
	case utils.LOG:
		return new(actLog), nil
	default:
		return nil, fmt.Errorf("unsupported action type: <%s>", typ)

	}
	return
}

// actLogger will log data to CGRateS logger
type actLog struct{}

// execute implements actioner interface
func (aL *actLog) execute(ctx context.Context, data interface{}) (err error) {
	return
}
