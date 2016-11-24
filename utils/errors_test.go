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

func TestCGRErrorActivate(t *testing.T) {
	ctx := "TEST_CONTEXT"
	apiErr := "TEST_API_ERR"
	shortErr := "short error"
	longErr := "long error which is good for debug"
	err := NewCGRError(ctx, apiErr, shortErr, longErr)
	if ctxRcv := err.Context(); ctxRcv != ctx {
		t.Errorf("Context: <%s>", ctxRcv)
	}
	if err.Error() != shortErr {
		t.Error(err)
	}
	err.ActivateAPIError()
	if err.Error() != apiErr {
		t.Error(err)
	}
	err.ActivateLongError()
	if err.Error() != longErr {
		t.Error(err)
	}
}

func TestAPIErrorHandler(t *testing.T) {
	if err := APIErrorHandler(ErrNotImplemented); err.Error() != NewErrServerError(ErrNotImplemented).Error() {
		t.Error(err)
	}
	if err := APIErrorHandler(ErrNotFound); err.Error() != ErrNotFound.Error() {
		t.Error(err)
	}
	cgrErr := NewCGRError("TEST_CONTEXT", "TEST_API_ERR", "short error", "long error which is good for debug")
	if err := APIErrorHandler(cgrErr); err.Error() != cgrErr.apiError {
		t.Error(err)
	}
}
