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
	"fmt"
	"net"
	"net/rpc"
	"strings"
	"syscall"
)

var (
	ErrNoMoreData               = errors.New("NO_MORE_DATA")
	ErrNotImplemented           = errors.New("NOT_IMPLEMENTED")
	ErrNotFound                 = errors.New("NOT_FOUND")
	ErrTimedOut                 = errors.New("TIMED_OUT")
	ErrServerError              = errors.New("SERVER_ERROR")
	ErrMaxRecursionDepth        = errors.New("MAX_RECURSION_DEPTH")
	ErrMandatoryIeMissing       = errors.New("MANDATORY_IE_MISSING")
	ErrExists                   = errors.New("EXISTS")
	ErrBrokenReference          = errors.New("BROKEN_REFERENCE")
	ErrParserError              = errors.New("PARSER_ERROR")
	ErrInvalidPath              = errors.New("INVALID_PATH")
	ErrInvalidKey               = errors.New("INVALID_KEY")
	ErrUnauthorizedDestination  = errors.New("UNAUTHORIZED_DESTINATION")
	ErrRatingPlanNotFound       = errors.New("RATING_PLAN_NOT_FOUND")
	ErrAccountNotFound          = errors.New("ACCOUNT_NOT_FOUND")
	ErrAccountDisabled          = errors.New("ACCOUNT_DISABLED")
	ErrUserNotFound             = errors.New("USER_NOT_FOUND")
	ErrInsufficientCredit       = errors.New("INSUFFICIENT_CREDIT")
	ErrNotConvertible           = errors.New("NOT_CONVERTIBLE")
	ErrResourceUnavailable      = errors.New("RESOURCE_UNAVAILABLE")
	ErrResourceUnauthorized     = errors.New("RESOURCE_UNAUTHORIZED")
	ErrNoActiveSession          = errors.New("NO_ACTIVE_SESSION")
	ErrPartiallyExecuted        = errors.New("PARTIALLY_EXECUTED")
	ErrMaxUsageExceeded         = errors.New("MAX_USAGE_EXCEEDED")
	ErrUnallocatedResource      = errors.New("UNALLOCATED_RESOURCE")
	ErrNotFoundNoCaps           = errors.New("not found")
	ErrFilterNotPassingNoCaps   = errors.New("filter not passing")
	ErrNotConvertibleNoCaps     = errors.New("not convertible")
	ErrMandatoryIeMissingNoCaps = errors.New("mandatory information missing")
	ErrUnauthorizedApi          = errors.New("UNAUTHORIZED_API")
	ErrUnknownApiKey            = errors.New("UNKNOWN_API_KEY")
	ErrIncompatible             = errors.New("INCOMPATIBLE")
	ErrReqUnsynchronized        = errors.New("REQ_UNSYNCHRONIZED")
	ErrUnsupporteServiceMethod  = errors.New("UNSUPPORTED_SERVICE_METHOD")
	ErrWrongArgsType            = errors.New("WRONG_ARGS_TYPE")
	ErrWrongReplyType           = errors.New("WRONG_REPLY_TYPE")
	ErrDisconnected             = errors.New("DISCONNECTED")
	ErrReplyTimeout             = errors.New("REPLY_TIMEOUT")
	ErrFailedReconnect          = errors.New("FAILED_RECONNECT")
	ErrInternallyDisconnected   = errors.New("INTERNALLY_DISCONNECTED")
	ErrUnsupportedCodec         = errors.New("UNSUPPORTED_CODEC")
	ErrSessionNotFound          = errors.New("SESSION_NOT_FOUND")
	ErrJsonIncompleteComment    = errors.New("JSON_INCOMPLETE_COMMENT")
	ErrCDRCNoProfileID          = errors.New("CDRC_PROFILE_WITHOUT_ID")
	ErrCDRCNoInDir              = errors.New("CDRC_PROFILE_WITHOUT_IN_DIR")
	ErrNotEnoughParameters      = errors.New("NotEnoughParameters")
	ErrNotConnected             = errors.New("NOT_CONNECTED")
	RalsErrorPrfx               = "RALS_ERROR"
	DispatcherErrorPrefix       = "DISPATCHER_ERROR"
	ErrUnsupportedFormat        = errors.New("UNSUPPORTED_FORMAT")
	ErrNoDatabaseConn           = errors.New("NO_DATA_BASE_CONNECTION")
)

// NewCGRError initialises a new CGRError
func NewCGRError(context, apiErr, shortErr, longErr string) *CGRError {
	return &CGRError{context: context, apiError: apiErr,
		shortError: shortErr, longError: longErr, errorMessage: shortErr}
}

// CGRError is a context based error
// returns always errorMessage but this can be switched based on methods called on it
type CGRError struct {
	context      string
	apiError     string
	shortError   string
	longError    string
	errorMessage string
}

func (err *CGRError) Context() string {
	return err.context
}

func (err *CGRError) Error() string {
	return err.errorMessage
}

func (err *CGRError) ActivateAPIError() {
	err.errorMessage = err.apiError
}

func (err *CGRError) ActivateShortError() {
	err.errorMessage = err.shortError
}

func (err *CGRError) ActivateLongError() {
	err.errorMessage = err.longError
}

func NewErrMandatoryIeMissing(fields ...string) error {
	return fmt.Errorf("MANDATORY_IE_MISSING: %v", fields)
}

func NewErrServerError(err error) error {
	return fmt.Errorf("SERVER_ERROR: %s", err)
}

func NewErrServiceNotOperational(serv string) error {
	return fmt.Errorf("SERVICE_NOT_OPERATIONAL: %s", serv)
}

func NewErrNotConnected(serv string) error {
	return fmt.Errorf("NOT_CONNECTED: %s", serv)
}

func NewErrRALs(err error) error {
	return fmt.Errorf("%s:%s", RalsErrorPrfx, err)
}

func NewErrResourceS(err error) error {
	return fmt.Errorf("RESOURCES_ERROR:%s", err)
}

func NewErrSupplierS(err error) error {
	return fmt.Errorf("SUPPLIERS_ERROR:%s", err)
}

func NewErrAttributeS(err error) error {
	return fmt.Errorf("ATTRIBUTES_ERROR:%s", err)
}

func NewErrDispatcherS(err error) error {
	return fmt.Errorf("%s:%s", DispatcherErrorPrefix, err.Error())
}

// Centralized returns for APIs
func APIErrorHandler(errIn error) (err error) {
	cgrErr, ok := errIn.(*CGRError)
	if !ok {
		err = errIn
		if err != ErrNotFound {
			err = NewErrServerError(err)
		}
		return
	}
	cgrErr.ActivateAPIError()
	return cgrErr
}

func NewErrStringCast(valIface interface{}) error {
	return fmt.Errorf("cannot cast value: %v to string", valIface)
}

func NewErrFldStringCast(fldName string, valIface interface{}) error {
	return fmt.Errorf("cannot cast field: %s with value: %v to string", fldName, valIface)
}

func ErrHasPrefix(err error, prfx string) (has bool) {
	if err == nil {
		return
	}
	return strings.HasPrefix(err.Error(), prfx)
}

func ErrPrefix(err error, reason string) error {
	return fmt.Errorf("%s:%s", err.Error(), reason)
}

func ErrPrefixNotFound(reason string) error {
	return ErrPrefix(ErrNotFound, reason)
}

func ErrPrefixNotErrNotImplemented(reason string) error {
	return ErrPrefix(ErrNotImplemented, reason)
}

func ErrEnvNotFound(key string) error {
	return ErrPrefix(ErrNotFound, "ENV_VAR:"+key)
}

// IsNetworkError will decide if an error is network generated or RPC one
// used by Dispatcher to figure out whether it should try another connection
func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}
	if operr, ok := err.(*net.OpError); ok &&
		strings.HasSuffix(operr.Err.Error(),
			syscall.ECONNRESET.Error()) { // connection reset
		return true
	}
	return err.Error() == rpc.ErrShutdown.Error() ||
		err.Error() == ErrReqUnsynchronized.Error() ||
		err.Error() == ErrDisconnected.Error() ||
		err.Error() == ErrReplyTimeout.Error() ||
		err.Error() == ErrSessionNotFound.Error() ||
		strings.HasPrefix(err.Error(), "rpc: can't find service")
}

func ErrPathNotReachable(path string) error {
	return fmt.Errorf("path:%+q is not reachable", path)
}

func ErrNotConvertibleTF(from, to string) error {
	return fmt.Errorf("%s : from: %s to:%s", ErrNotConvertibleNoCaps.Error(), from, to)
}
