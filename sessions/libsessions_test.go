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

package sessions

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/dgrijalva/jwt-go"
)

func TestLibSessionSGetSetCGRID(t *testing.T) {
	sEv := engine.NewMapEvent(map[string]interface{}{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.ToR:              "*voice",
		utils.OriginID:         "12345",
		utils.Account:          "account1",
		utils.Subject:          "subject1",
		utils.Destination:      "+4986517174963",
		utils.Category:         "call",
		utils.Tenant:           "cgrates.org",
		utils.RequestType:      "*prepaid",
		utils.SetupTime:        "2015-11-09 14:21:24",
		utils.AnswerTime:       "2015-11-09 14:22:02",
		utils.Usage:            "1m23s",
		utils.LastUsed:         "21s",
		utils.PDD:              "300ms",
		utils.SUPPLIER:         "supplier1",
		utils.DISCONNECT_CAUSE: "NORMAL_DISCONNECT",
		utils.OriginHost:       "127.0.0.1",
	})
	//Empty CGRID in event
	cgrID := GetSetCGRID(sEv)
	if len(cgrID) == 0 {
		t.Errorf("Unexpected cgrID: %+v", cgrID)
	}
	//populate CGRID in event
	sEv[utils.CGRID] = "someRandomVal"
	cgrID = GetSetCGRID(sEv)
	if cgrID != "someRandomVal" {
		t.Errorf("Expecting: someRandomVal, received: %+v", cgrID)
	}
}

func TestLibSessionSgetSessionTTL(t *testing.T) {
	sEv := engine.NewMapEvent(map[string]interface{}{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.ToR:              "*voice",
		utils.OriginID:         "12345",
		utils.Account:          "account1",
		utils.Subject:          "subject1",
		utils.Destination:      "+4986517174963",
		utils.Category:         "call",
		utils.Tenant:           "cgrates.org",
		utils.RequestType:      "*prepaid",
		utils.SetupTime:        "2015-11-09 14:21:24",
		utils.AnswerTime:       "2015-11-09 14:22:02",
		utils.Usage:            "1m23s",
		utils.LastUsed:         "21s",
		utils.PDD:              "300ms",
		utils.SUPPLIER:         "supplier1",
		utils.DISCONNECT_CAUSE: "NORMAL_DISCONNECT",
		utils.OriginHost:       "127.0.0.1",
	})

	sEv[utils.SessionTTL] = "notanumber"
	if _, err := getSessionTTL(&sEv, time.Duration(0), nil); err == nil {
		t.Errorf("Expecting: NOT_FOUND, received: %+v", err)
	}
	sEv[utils.SessionTTL] = 0
	if ttl, err := getSessionTTL(&sEv, time.Duration(0), nil); err != nil {
		t.Error(err)
	} else if ttl != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, ttl)
	}
	sEv[utils.SessionTTL] = "2s"
	sEv[utils.SessionTTLMaxDelay] = "notanumber"
	if _, err := getSessionTTL(&sEv, time.Duration(0), nil); err == nil {
		t.Error("Expecting: invalid duration, received: ", err)
	}
	sEv[utils.SessionTTLMaxDelay] = 0

	//ttl is taken from event
	if ttl, err := getSessionTTL(&sEv, time.Duration(0), nil); err != nil {
		t.Error(err)
	} else if ttl != time.Duration(2*time.Second) {
		t.Errorf("Expecting: %+v, received: %+v",
			time.Duration(2*time.Second), ttl)
	}
	//remove ttl from event
	delete(sEv, utils.SessionTTL)
	if ttl, err := getSessionTTL(&sEv, time.Duration(4*time.Second), nil); err != nil {
		t.Error(err)
	} else if ttl != time.Duration(4*time.Second) {
		t.Errorf("Expecting: %+v, received: %+v",
			time.Duration(4*time.Second), ttl)
	}

	//add sessionTTLMaxDelay in event
	sEv[utils.SessionTTLMaxDelay] = "1s"
	if ttl, err := getSessionTTL(&sEv, time.Duration(4*time.Second), nil); err != nil {
		t.Error(err)
	} else if ttl <= time.Duration(4*time.Second) {
		t.Errorf("Unexpected ttl : %+v", ttl)
	}

	//remove sessionTTLMaxDelay from event
	delete(sEv, utils.SessionTTLMaxDelay)
	if ttl, err := getSessionTTL(&sEv, time.Duration(7*time.Second),
		utils.DurationPointer(time.Duration(2*time.Second))); err != nil {
		t.Error(err)
	} else if ttl <= time.Duration(7*time.Second) {
		t.Errorf("Unexpected ttl : %+v", ttl)
	}
}

func TestGetFlagIDs(t *testing.T) {
	//empty check
	rcv := getFlagIDs("")
	var eOut []string
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	//normal check
	rcv = getFlagIDs("*attributes:ATTR1;ATTR2")
	eOut = []string{"ATTR1", "ATTR2"}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestNewProcessedIdentity(t *testing.T) {
	if _, err := NewProcessedIdentity(""); err == nil ||
		err.Error() != "missing parts of the message header" {
		t.Errorf("Expected %q received: %v", "missing parts of the message header", err)
	}
	if _, err := NewProcessedIdentity(";"); err == nil ||
		err.Error() != "wrong header format" {
		t.Errorf("Expected %q received: %v", "wrong header format", err)
	}

	if _, err := NewProcessedIdentity("eyJhbGciOiJFUzI1NiIsInBwdCI6InNoYWtlbiIsInR,5cCI6InBhc3Nwb3J0IiwieDV1IjoiaHR0cHM6Ly93d3cuZXhhbXBsZS5vcmcvY2VydC5jZXIifQ.eyJhdHRlc3QiOiJBIiwiZGVzdCI6eyJ0biI6WyIxMDAyIl19LCJpYXQiOjE1ODcwMTk4MjIsIm9yaWc,iOnsidG4iOiIxMDAxIn0sIm9yaWdpZCI6IjEyMzQ1NiJ9.4ybtWmgqdkNyJLS9Iv3PuJV8ZxR7yZ_NEBhCpKCEu2WBiTchqwoqoWpI17Q_ALm38tbnpay32t95ZY_LhSgwJg;info=<https://www.example.org/cert.cer>;ppt=shaken"); err == nil {
		t.Errorf("Expected error")
	}

	if _, err := NewProcessedIdentity("eyJhbGciOiJFUzI1NiIsInBwdCI6InNoYWtlbiIsInR5cCI6InBhc3Nwb3J0IiwieDV1IjoiaHR0cHM6Ly93d3cuZXhhbXBsZS5vcmcvY2VydC5jZXIifQ.eyJhdHRlc3QiOiJBIiwiZGVzdCI6eyJ0biI6WyIxMDAyIl19LCJpYXQiOjE1ODcwMTk4MjIsIm9yaWc,iOnsidG4iOiIxMDAxIn0sIm9yaWdpZCI6IjEyMzQ1NiJ9.4ybtWmgqdkNyJLS9Iv3PuJV8ZxR7yZ_NEBhCpKCEu2WBiTchqwoqoWpI17Q_ALm38tbnpay32t95ZY_LhSgwJg;info=<https://www.example.org/cert.cer>;ppt=shaken"); err == nil {
		t.Errorf("Expected error")
	}

	expected := &ProcessedStirIdentity{
		Tokens:     []string{"info=<https://www.example.org/cert.cer>", "ppt=shaken"},
		SigningStr: "eyJhbGciOiJFUzI1NiIsInBwdCI6InNoYWtlbiIsInR5cCI6InBhc3Nwb3J0IiwieDV1IjoiaHR0cHM6Ly93d3cuZXhhbXBsZS5vcmcvY2VydC5jZXIifQ.eyJhdHRlc3QiOiJBIiwiZGVzdCI6eyJ0biI6WyIxMDAyIl19LCJpYXQiOjE1ODcwMTk4MjIsIm9yaWciOnsidG4iOiIxMDAxIn0sIm9yaWdpZCI6IjEyMzQ1NiJ9",
		Signature:  "4ybtWmgqdkNyJLS9Iv3PuJV8ZxR7yZ_NEBhCpKCEu2WBiTchqwoqoWpI17Q_ALm38tbnpay32t95ZY_LhSgwJg",
		Header: &utils.PASSporTHeader{
			Typ: utils.STIRTyp,
			Alg: utils.STIRAlg,
			Ppt: utils.STIRPpt,
			X5u: "https://www.example.org/cert.cer",
		},
		Payload: &utils.PASSporTPayload{
			ATTest: "A",
			Dest: utils.PASSporTDestinationsIdentity{
				Tn: []string{"1002"},
			},
			IAT: 1587019822,
			Orig: utils.PASSporTOriginsIdentity{
				Tn: "1001",
			},
			OrigID: "123456",
		},
	}
	if rply, err := NewProcessedIdentity("eyJhbGciOiJFUzI1NiIsInBwdCI6InNoYWtlbiIsInR5cCI6InBhc3Nwb3J0IiwieDV1IjoiaHR0cHM6Ly93d3cuZXhhbXBsZS5vcmcvY2VydC5jZXIifQ.eyJhdHRlc3QiOiJBIiwiZGVzdCI6eyJ0biI6WyIxMDAyIl19LCJpYXQiOjE1ODcwMTk4MjIsIm9yaWciOnsidG4iOiIxMDAxIn0sIm9yaWdpZCI6IjEyMzQ1NiJ9.4ybtWmgqdkNyJLS9Iv3PuJV8ZxR7yZ_NEBhCpKCEu2WBiTchqwoqoWpI17Q_ALm38tbnpay32t95ZY_LhSgwJg;info=<https://www.example.org/cert.cer>;ppt=shaken"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %s, received:%s", utils.ToJSON(expected), utils.ToJSON(rply))
	}
}

func TestProcessedIdentityVerifyHeader(t *testing.T) {
	args := &ProcessedStirIdentity{
		Tokens: []string{"info=<https://www.example.org/cert.cer>", "ppt=shaken", "extra", "alg=ES256"},
		Header: &utils.PASSporTHeader{
			Typ: utils.STIRTyp,
			Alg: utils.STIRAlg,
			Ppt: utils.STIRPpt,
			X5u: "https://www.example.org/cert.cer",
		},
	}
	if !args.VerifyHeader() {
		t.Errorf("Expected the header to be valid")
	}
	args.Header.Typ = "1"
	if args.VerifyHeader() {
		t.Errorf("Expected the header to not be valid")
	}

	args.Tokens = []string{"info=<https://www.example.org/cert.cer>", "ppt=shaken", "alg=wrongArg"}
	if args.VerifyHeader() {
		t.Errorf("Expected the header to not be valid")
	}

	args.Tokens = []string{"info=<https://www.example.org/cert.cer>", "ppt=wrongExtension"}
	if args.VerifyHeader() {
		t.Errorf("Expected the header to not be valid")
	}
	args.Tokens = []string{"info=<", "ppt=shaken"}
	if args.VerifyHeader() {
		t.Errorf("Expected the header to not be valid")
	}
}

func TestProcessedIdentityVerifyPayload(t *testing.T) {
	args := &ProcessedStirIdentity{
		Payload: &utils.PASSporTPayload{
			ATTest: "C",
			Dest: utils.PASSporTDestinationsIdentity{
				Tn: []string{"1002"},
			},
			IAT: 1587019822,
			Orig: utils.PASSporTOriginsIdentity{
				Tn: "1001",
			},
			OrigID: "123456",
		},
	}
	if err := args.VerifyPayload("1001", "", "1002", "", -1, utils.NewStringSet([]string{utils.META_ANY})); err != nil {
		t.Error(err)
	}
	if err := args.VerifyPayload("1001", "", "1003", "", -1, utils.NewStringSet([]string{utils.META_ANY})); err == nil ||
		err.Error() != "wrong destinationTn" {
		t.Errorf("Expected error: %s,receved %v", "wrong destinationTn", err)
	}
	if err := args.VerifyPayload("1001", "", "1003", "1002", -1, utils.NewStringSet([]string{utils.META_ANY})); err == nil ||
		err.Error() != "wrong destinationURI" {
		t.Errorf("Expected error: %s,receved %v", "wrong destinationURI", err)
	}
	if err := args.VerifyPayload("1002", "", "1003", "1002", -1, utils.NewStringSet([]string{utils.META_ANY})); err == nil ||
		err.Error() != "wrong originatorTn" {
		t.Errorf("Expected error: %s,receved %v", "wrong originatorTn", err)
	}
	if err := args.VerifyPayload("1002", "1001", "1003", "1002", -1, utils.NewStringSet([]string{utils.META_ANY})); err == nil ||
		err.Error() != "wrong originatorURI" {
		t.Errorf("Expected error: %s,receved %v", "wrong originatorURI", err)
	}
	if err := args.VerifyPayload("1001", "", "1002", "", time.Second, utils.NewStringSet([]string{utils.META_ANY})); err == nil ||
		err.Error() != "expired payload" {
		t.Errorf("Expected error: %s,receved %v", "expired payload", err)
	}
	if err := args.VerifyPayload("1001", "", "1002", "", time.Second, utils.NewStringSet([]string{"A"})); err == nil ||
		err.Error() != "wrong attest level" {
		t.Errorf("Expected error: %s,receved %v", "wrong attest level", err)
	}
}

func TestAuthStirShaken(t *testing.T) {
	if err := AuthStirShaken("", "1001", "", "1002", "", utils.NewStringSet([]string{utils.META_ANY}), -1); err == nil {
		t.Error("Expected invalid identity")
	}
	if err := AuthStirShaken(
		"eyJhbGciOiJFUzI1NiIsInBwdCI6InNoYWtlbiIsInR5cCI6InBhc3Nwb3J0IiwieDV1IjoiaHR0cHM6Ly93d3cuZXhhbXBsZS5vcmcvY2VydC5jZXIifQ.eyJhdHRlc3QiOiJBIiwiZGVzdCI6eyJ0biI6WyIxMDAyIl19LCJpYXQiOjE1ODcwMTk4MjIsIm9yaWciOnsidG4iOiIxMDAxIn0sIm9yaWdpZCI6IjEyMzQ1NiJ9.4ybtWmgqdkNyJLS9Iv3PuJV8ZxR7yZ_NEBhCpKCEu2WBiTchqwoqoWpI17Q_ALm38tbnpay32t95ZY_LhSgwJg;info=<https://www.example.org/cert.cer2>;ppt=shaken",
		"1001", "", "1002", "", utils.NewStringSet([]string{utils.META_ANY}), -1); err == nil {
		t.Error("Expected invalid identity")
	}
	engine.Cache.Set(utils.CacheSTIR, "https://www.example.org/cert.cer", nil,
		nil, true, utils.NonTransactional)
	if err := AuthStirShaken(
		"eyJhbGciOiJFUzI1NiIsInBwdCI6InNoYWtlbiIsInR5cCI6InBhc3Nwb3J0IiwieDV1IjoiaHR0cHM6Ly93d3cuZXhhbXBsZS5vcmcvY2VydC5jZXIifQ.eyJhdHRlc3QiOiJBIiwiZGVzdCI6eyJ0biI6WyIxMDAyIl19LCJpYXQiOjE1ODcwMTk4MjIsIm9yaWciOnsidG4iOiIxMDAxIn0sIm9yaWdpZCI6IjEyMzQ1NiJ9.4ybtWmgqdkNyJLS9Iv3PuJV8ZxR7yZ_NEBhCpKCEu2WBiTchqwoqoWpI17Q_ALm38tbnpay32t95ZY_LhSgwJg;info=<https://www.example.org/cert.cer>;ppt=shaken", "1001", "", "1002", "", utils.NewStringSet([]string{utils.META_ANY}), -1); err == nil {
		t.Error("Expected invalid identity")
	}

	pubkeyBuf := []byte(`-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAESt8sEh55Yc579vLHjFRWVQO27p4Y
aa+jqv4dwkr/FLEcN1zC76Y/IniI65fId55hVJvN3ORuzUqYEtzD3irmsw==
-----END PUBLIC KEY-----
`)
	pubKey, err := jwt.ParseECPublicKeyFromPEM(pubkeyBuf)
	if err != nil {
		t.Fatal(err)
	}
	engine.Cache.Set(utils.CacheSTIR, "https://www.example.org/cert.cer", pubKey,
		nil, true, utils.NonTransactional)

	if err := AuthStirShaken(
		"eyJhbGciOiJFUzI1NiIsInBwdCI6InNoYWtlbiIsInR5cCI6InBhc3Nwb3J0IiwieDV1IjoiaHR0cHM6Ly93d3cuZXhhbXBsZS5vcmcvY2VydC5jZXIifQ.eyJhdHRlc3QiOiJBIiwiZGVzdCI6eyJ0biI6WyIxMDAyIl19LCJpYXQiOjE1ODcwMTk4MjIsIm9yaWciOnsidG4iOiIxMDAxIn0sIm9yaWdpZCI6IjEyMzQ1NiJ9.4ybtWmgqdkNyJLS9Iv3PuJV8ZxR7yZ_NEBhCpKCEu2WBiTchqwoqoWpI17Q_ALm38tbnpay32t95ZY_LhSgwJg;info=<https://www.example.org/cert.cer>;ppt=shaken", "1001", "", "1003", "", utils.NewStringSet([]string{utils.META_ANY}), -1); err == nil {
		t.Error("Expected invalid identity")
	}

	if err := AuthStirShaken(
		"eyJhbGciOiJFUzI1NiIsInBwdCI6InNoYWtlbiIsInR5cCI6InBhc3Nwb3J0IiwieDV1IjoiaHR0cHM6Ly93d3cuZXhhbXBsZS5vcmcvY2VydC5jZXIifQ.eyJhdHRlc3QiOiJBIiwiZGVzdCI6eyJ0biI6WyIxMDAyIl19LCJpYXQiOjE1ODcwMTk4MjIsIm9yaWciOnsidG4iOiIxMDAxIn0sIm9yaWdpZCI6IjEyMzQ1NiJ9.4ybtWmgqdkNyJLS9Iv3PuJV8ZxR7yZ_NEBhCpKCEu2WBiTchqwoqoWpI17Q_ALm38tbnpay32t95ZY_LhSgwJg;info=<https://www.example.org/cert.cer>;ppt=shaken", "1001", "", "1002", "", utils.NewStringSet([]string{utils.META_ANY}), -1); err != nil {
		t.Fatal(err)
	}
}

func TestNewIdentity(t *testing.T) {
	payload := &utils.PASSporTPayload{
		ATTest: "A",
		Dest:   utils.PASSporTDestinationsIdentity{Tn: []string{"1002"}},
		IAT:    1587019822,
		Orig:   utils.PASSporTOriginsIdentity{Tn: "1001"},
		OrigID: "123456",
	}
	header := utils.NewPASSporTHeader("https://www.example.org/cert.cer")
	prvkeyBuf := []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIBIx2HW6dYy5S4wlJUY1J8VxO1un8xr4SHQlT7/UFkktoAoGCCqGSM49
AwEHoUQDQgAESt8sEh55Yc579vLHjFRWVQO27p4Yaa+jqv4dwkr/FLEcN1zC76Y/
IniI65fId55hVJvN3ORuzUqYEtzD3irmsw==
-----END EC PRIVATE KEY-----
`)
	engine.Cache.Set(utils.CacheSTIR, "https://www.example.org/private.pem", nil,
		nil, true, utils.NonTransactional)

	if _, err := NewSTIRIdentity(header, payload, "https://www.example.org/private.pem", time.Second); err == nil {
		t.Error("Expected error when creating new identity")
	}

	prvKey, err := jwt.ParseECPrivateKeyFromPEM(prvkeyBuf)
	if err != nil {
		t.Fatal(err)
	}

	pubkeyBuf := []byte(`-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAESt8sEh55Yc579vLHjFRWVQO27p4Y
aa+jqv4dwkr/FLEcN1zC76Y/IniI65fId55hVJvN3ORuzUqYEtzD3irmsw==
-----END PUBLIC KEY-----
`)
	pubKey, err := jwt.ParseECPublicKeyFromPEM(pubkeyBuf)
	if err != nil {
		t.Fatal(err)
	}
	engine.Cache.Set(utils.CacheSTIR, "https://www.example.org/cert.cer", pubKey,
		nil, true, utils.NonTransactional)
	engine.Cache.Set(utils.CacheSTIR, "https://www.example.org/private.pem", prvKey,
		nil, true, utils.NonTransactional)

	if rcv, err := NewSTIRIdentity(header, payload, "https://www.example.org/private.pem", time.Second); err != nil {
		t.Error(err)
	} else if err := AuthStirShaken(rcv, "1001", "", "1002", "", utils.NewStringSet([]string{utils.META_ANY}), -1); err != nil {
		t.Fatal(err)
	}
}
