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

func TestNewCGRError(t *testing.T) {
	eOut := &CGRError{}
	if rcv := NewCGRError(EmptyString, EmptyString, EmptyString, EmptyString); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected: %+v, received: %+v ", ToJSON(eOut), ToJSON(rcv))
	}
	eOut = &CGRError{
		context:      "context",
		apiError:     "apiError",
		shortError:   "shortError",
		longError:    "longError",
		errorMessage: "shortError",
	}
	if rcv := NewCGRError("context", "apiError", "shortError", "longError"); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected: %+v, received: %+v ", ToJSON(eOut), ToJSON(rcv))
	}
}

func TestCGRErrorActivate(t *testing.T) {
	ctx := "TEST_CONTEXT"
	apiErr := "TEST_API_ERR"
	shortErr := "short error"
	longErr := "long error which is good for debug"
	cgrError := NewCGRError(ctx, apiErr, shortErr, longErr)

	if rcv := cgrError.Context(); !reflect.DeepEqual(rcv, ctx) {
		t.Errorf("Expected: %+q, received: %+q ", ctx, rcv)
	}
	if rcv := cgrError.Error(); !reflect.DeepEqual(rcv, shortErr) {
		t.Errorf("Expected: %+q, received: %+q ", shortErr, rcv)
	}
	cgrError.ActivateAPIError()
	if !reflect.DeepEqual(apiErr, cgrError.errorMessage) {
		t.Errorf("Expected: %+q, received: %+q ", apiErr, cgrError.errorMessage)
	}
	cgrError.ActivateShortError()
	if !reflect.DeepEqual(shortErr, cgrError.errorMessage) {
		t.Errorf("Expected: %+q, received: %+q ", shortErr, cgrError.errorMessage)
	}
	cgrError.ActivateLongError()
	if !reflect.DeepEqual(longErr, cgrError.errorMessage) {
		t.Errorf("Expected: %+q, received: %+q ", longErr, cgrError.errorMessage)
	}
}

func TestNewErrMandatoryIeMissing(t *testing.T) {
	if rcv := NewErrMandatoryIeMissing(EmptyString); rcv.Error() != "MANDATORY_IE_MISSING: []" {
		t.Errorf("Expecting: MANDATORY_IE_MISSING: [], received: %+v", rcv)
	}
	if rcv := NewErrMandatoryIeMissing("string1", "string2"); rcv.Error() != "MANDATORY_IE_MISSING: [string1 string2]" {
		t.Errorf("Expecting: MANDATORY_IE_MISSING: [string1 string2], received: %+b", rcv)
	}
	if rcv := NewErrMandatoryIeMissing("test"); rcv.Error() != "MANDATORY_IE_MISSING: [test]" {
		t.Errorf("Expecting: MANDATORY_IE_MISSING: [test], received: %+v", rcv)
	}
}

func TestNewErrServerError(t *testing.T) {
	cgrError := NewCGRError("context", "apiError", "shortError", "longError")
	if rcv := NewErrServerError(cgrError); rcv.Error() != "SERVER_ERROR: shortError" {
		t.Errorf("Expecting: SERVER_ERROR: shortError, received: %+v", rcv)
	}
}

func TestNewErrServiceNotOperational(t *testing.T) {
	if rcv := NewErrServiceNotOperational("Error"); rcv.Error() != "SERVICE_NOT_OPERATIONAL: Error" {
		t.Errorf("Expecting: SERVICE_NOT_OPERATIONAL: Error, received: %+v", rcv)
	}
}

func TestNewErrNotConnected(t *testing.T) {
	if rcv := NewErrNotConnected("Error"); rcv.Error() != "NOT_CONNECTED: Error" {
		t.Errorf("Expecting: NOT_CONNECTED: Error, received: %+v", rcv)
	}
}

func TestNewErrRALs(t *testing.T) {
	cgrError := NewCGRError("context", "apiError", "shortError", "longError")
	if rcv := NewErrRALs(cgrError); rcv.Error() != "RALS_ERROR:shortError" {
		t.Errorf("Expecting: RALS_ERROR:shortError, received: %+v", rcv)
	}
}

func TestNewErrResourceS(t *testing.T) {
	cgrError := NewCGRError("context", "apiError", "shortError", "longError")
	if rcv := NewErrResourceS(cgrError); rcv.Error() != "RESOURCES_ERROR:shortError" {
		t.Errorf("Expecting: RESOURCES_ERROR:shortError, received: %+v", rcv)
	}
}

func TestNewErrSupplierS(t *testing.T) {
	cgrError := NewCGRError("context", "apiError", "shortError", "longError")
	if rcv := NewErrRouteS(cgrError); rcv.Error() != "ROUTES_ERROR:shortError" {
		t.Errorf("Expecting: ROUTES_ERROR:shortError, received: %+v", rcv)
	}
}
func TestNewErrAttributeS(t *testing.T) {
	cgrError := NewCGRError("context", "apiError", "shortError", "longError")
	if rcv := NewErrAttributeS(cgrError); rcv.Error() != "ATTRIBUTES_ERROR:shortError" {
		t.Errorf("Expecting: ATTRIBUTES_ERROR:shortError, received: %+v", rcv)
	}
}

func TestNewErrDispatcherS(t *testing.T) {
	cgrError := NewCGRError("context", "apiError", "shortError", "longError")
	if rcv := NewErrDispatcherS(cgrError); rcv.Error() != "DISPATCHER_ERROR:shortError" {
		t.Errorf("Expecting: DISPATCHER_ERROR:shortError, received: %+v", rcv)
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

func TestNewErrStringCast(t *testing.T) {
	if rcv := NewErrStringCast("test"); rcv.Error() != "cannot cast value: test to string" {
		t.Errorf("Expecting: cannot cast value: test to string, received: %+v", rcv)
	}
}

func TestNewErrFldStringCast(t *testing.T) {
	if rcv := NewErrFldStringCast("test1", "test2"); rcv.Error() != "cannot cast field: test1 with value: test2 to string" {
		t.Errorf("Expecting: cannot cast field: test1 with value: test2 to string, received: %+v", rcv)
	}
}

func TestErrHasPrefix(t *testing.T) {
	if ErrHasPrefix(nil, EmptyString) {
		t.Error("Expecting false, received: true")
	}
	if !ErrHasPrefix(&CGRError{errorMessage: "test_errorMessage"}, "test") {
		t.Error("Expecting true, received: false")
	}
}

func TestErrPrefix(t *testing.T) {
	cgrError := NewCGRError("context", "apiError", "shortError", "longError")
	if rcv := ErrPrefix(cgrError, "notaprefix"); rcv.Error() != "shortError:notaprefix" {
		t.Errorf("Expecting: shortError:notaprefix, received: %+v", rcv)
	}
}

func TestErrPrefixNotFound(t *testing.T) {
	if rcv := ErrPrefixNotFound("test_string"); rcv.Error() != "NOT_FOUND:test_string" {
		t.Errorf("Expecting: NOT_FOUND:test_string, received: %+v", rcv)
	}
}

func TestErrPrefixNotErrNotImplemented(t *testing.T) {
	if rcv := ErrPrefixNotErrNotImplemented("test_string"); rcv.Error() != "NOT_IMPLEMENTED:test_string" {
		t.Errorf("Expecting: NOT_IMPLEMENTED:test_string, received: %+v", rcv)
	}
}

func TestErrEnvNotFound(t *testing.T) {
	if rcv := ErrEnvNotFound("test_string"); rcv.Error() != "NOT_FOUND:ENV_VAR:test_string" {
		t.Errorf("Expecting: NOT_FOUND:ENV_VAR:test_string, received: %+v", rcv)
	}
}

func TestErrPathNotReachable(t *testing.T) {
	if rcv := ErrPathNotReachable("test/path"); rcv.Error() != `path:"test/path" is not reachable` {
		t.Errorf("Expecting: path:'test/path' is not reachable, received: %+v", rcv)
	}
}

func TestErrNotConvertibleTF(t *testing.T) {
	if rcv := ErrNotConvertibleTF("test_type1", "test_type2"); rcv.Error() != `not convertible : from: test_type1 to:test_type2` {
		t.Errorf("Expecting: not convertible : from: test_type1 to:test_type2, received: %+v", rcv)
	}
}

func TestNewErrChargerS(t *testing.T) {
	expected := `CHARGERS_ERROR:NOT_FOUND`
	if rcv := NewErrChargerS(ErrNotFound); rcv.Error() != expected {
		t.Errorf("Expecting: %q, received: %q", expected, rcv.Error())
	}
}

func TestNewErrStatS(t *testing.T) {
	expected := "STATS_ERROR:NOT_FOUND"
	if rcv := NewErrStatS(ErrNotFound); rcv.Error() != expected {
		t.Errorf("Expected %+q, receiveed %+q", expected, rcv.Error())
	}
}

func TestNewErrCDRS(t *testing.T) {
	expected := "CDRS_ERROR:NOT_FOUND"
	if rcv := NewErrCDRS(ErrNotFound); rcv.Error() != expected {
		t.Errorf("Expected %+q, received %+q", expected, rcv.Error())
	}
}

func TestNewErrThresholdS(t *testing.T) {
	expected := "THRESHOLDS_ERROR:NOT_FOUND"
	if rcv := NewErrThresholdS(ErrNotFound); rcv.Error() != expected {
		t.Errorf("Expected %+q, received %+q", expected, rcv.Error())
	}
}

func TestNewSTIRError(t *testing.T) {
	expected := `*stir_authenticate: wrong header`
	if rcv := NewSTIRError("wrong header"); rcv.Error() != expected {
		t.Errorf("Expecting: %q, received: %q", expected, rcv.Error())
	}
}
