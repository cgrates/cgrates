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
	"errors"
	"reflect"
	"testing"
)

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

func TestNewErrRates(t *testing.T) {
	err := errors.New("ErrorRates")
	if rcv := NewErrRateS(err); rcv.Error() != "RATES_ERROR:ErrorRates" {
		t.Errorf("Expecting: RATES_ERROR:ErrorRates, received: %+v", rcv)
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

func TestNewCGRError(t *testing.T) {
	context := "*sessions"
	apiErr := "API Error"
	shortErr := "Short Error"
	longErr := "Long Error"
	exp := &CGRError{
		context:      context,
		apiError:     apiErr,
		shortError:   shortErr,
		longError:    longErr,
		errorMessage: "Short Error",
	}
	rcv := NewCGRError(context, apiErr, shortErr, longErr)
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received %v\n", exp, rcv)
	}
}

func TestGetContext(t *testing.T) {
	err := &CGRError{
		context: "*sessions",
	}
	exp := "*sessions"
	if rcv := err.Context(); !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received %v\n", exp, rcv)
	}
}

func TestGetError(t *testing.T) {
	err := &CGRError{
		errorMessage: "ERROR MESSAGE IN errorMessage field",
	}
	exp := "ERROR MESSAGE IN errorMessage field"
	if rcv := err.Error(); !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received %v\n", exp, rcv)
	}
}

func TestActivateAPIError(t *testing.T) {
	err := &CGRError{
		apiError:     "API Error",
		errorMessage: "ERROR MESSAGE IN errorMessage field",
	}
	exp := "API Error"
	err.ActivateAPIError()
	if !reflect.DeepEqual(err.errorMessage, exp) {
		t.Errorf("Expected %v \n but received %v\n", exp, err.errorMessage)
	}
}

func TestActivateShortError(t *testing.T) {
	err := &CGRError{
		shortError:   "Short Error",
		errorMessage: "ERROR MESSAGE IN errorMessage field",
	}
	exp := "Short Error"
	err.ActivateShortError()
	if !reflect.DeepEqual(err.errorMessage, exp) {
		t.Errorf("Expected %v \n but received %v\n", exp, err.errorMessage)
	}
}

func TestActivateLongError(t *testing.T) {
	err := &CGRError{
		longError:    "Long Error",
		errorMessage: "ERROR MESSAGE IN errorMessage field",
	}
	exp := "Long Error"
	err.ActivateLongError()
	if !reflect.DeepEqual(err.errorMessage, exp) {
		t.Errorf("Expected %v \n but received %v\n", exp, err.errorMessage)
	}
}

func TestNewErrServerError(t *testing.T) {
	err := ErrNotFound
	exp := "SERVER_ERROR: NOT_FOUND"
	if rcv := NewErrServerError(err); rcv.Error() != exp {
		t.Errorf("Expected %v \n but received %v\n", exp, rcv)
	}
}

func TestNewErrNotConnected(t *testing.T) {
	serv := "localhost:8080"
	exp := "NOT_CONNECTED: localhost:8080"
	if rcv := NewErrNotConnected(serv); rcv.Error() != exp {
		t.Errorf("Expected %v \n but received %v\n", exp, rcv)
	}
}
