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

package engine

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestFilterHelpersWeightFromDynamics(t *testing.T) {
	var expected float64 = 64
	ctx := context.Background()
	dWs := []*utils.DynamicWeight{
		{
			Weight: 64,
		},
	}
	fltrs := &FilterS{}
	tnt := utils.CGRateSorg
	ev := utils.MapStorage{}
	result, err := WeightFromDynamics(ctx, dWs, fltrs, tnt, ev)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestFilterHelpersWeightFromDynamicsErr(t *testing.T) {

	ctx := context.Background()
	dWs := []*utils.DynamicWeight{
		{
			FilterIDs: []string{"*stirng:~*req.Account:1001"},
			Weight:    64,
		},
	}
	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, nil)

	cM := NewConnManager(cfg)
	fltrs := NewFilterS(cfg, cM, dm)
	tnt := utils.CGRateSorg
	ev := utils.MapStorage{}

	expErr := "NOT_IMPLEMENTED:*stirng"
	_, err := WeightFromDynamics(ctx, dWs, fltrs, tnt, ev)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}

func TestBlockerFromDynamicsErr(t *testing.T) {

	ctx := context.Background()
	dBs := []*utils.DynamicBlocker{
		{
			FilterIDs: []string{"*stirng:~*req.Account:1001"},
			Blocker:   true,
		},
	}
	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg, nil)
	cM := NewConnManager(cfg)
	fltrs := NewFilterS(cfg, cM, dm)
	tnt := utils.CGRateSorg
	ev := utils.MapStorage{}

	expErr := "NOT_IMPLEMENTED:*stirng"
	if _, err := BlockerFromDynamics(ctx, dBs, fltrs, tnt, ev); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}

func TestMatchingItemIDsForEventGetKeysForPrefixErr(t *testing.T) {

	matchEV = utils.MapStorage{utils.MetaReq: map[string]any{
		utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
		"Field":          "profile",
	}}
	data := &DataDBMock{
		GetKeysForPrefixF: func(ctx *context.Context, s string) ([]string, error) { return []string{}, utils.ErrNotImplemented },
	}
	cfg := config.NewDefaultCGRConfig()
	dmMatch := NewDataManager(data, cfg, nil)

	tntCtx := utils.ConcatenatedKey(utils.CGRateSorg, utils.MetaRating)

	if _, err := MatchingItemIDsForEvent(context.Background(), matchEV, nil, nil, nil, nil, nil,
		dmMatch, utils.CacheAttributeFilterIndexes, tntCtx, false, false); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%+v>, received error <%+v>", utils.ErrNotImplemented, err)
	}

}

func TestMatchingItemIDsForEventFilterIndexTypeNotNone(t *testing.T) {
	matchEV = utils.MapStorage{utils.MetaReq: map[string]any{
		utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
		"Fiel..d":        "profile",
	}}
	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dmMatch := NewDataManager(data, cfg, nil)

	tntCtx := utils.ConcatenatedKey(utils.CGRateSorg, utils.MetaRating)

	if _, err := MatchingItemIDsForEvent(context.Background(), matchEV, nil, nil, nil, nil, nil,
		dmMatch, utils.CacheAttributeFilterIndexes, tntCtx, true, false); err != utils.ErrNotFound {
		t.Errorf("Expected error <%+v>, received error <%+v>", utils.ErrNotFound, err)
	}

}

func TestSentrypeerGetTokenErrorResponse(t *testing.T) {
	tokenUrl := "Url"
	clientID := "ID"
	clientSecret := "clientSecret"
	audience := "audience"
	grantType := "grantType"

	token, err := sentrypeerGetToken(tokenUrl, clientID, clientSecret, audience, grantType)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if token != "" {
		t.Errorf("Expected empty token, got %v", token)
	}
}

func TestFilterHelpersExtractURLFromHTTPType(t *testing.T) {
	tests := []struct {
		name     string
		httpType string
		wantURL  string
		wantErr  string
	}{
		{
			name:     "Valid input",
			httpType: "http#cgrates.com",
			wantURL:  "cgrates.com",
		},
		{
			name:     "Incorrect format",
			httpType: "http",
			wantURL:  "",
			wantErr:  "invalid format: URL portion not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, err := ExtractURLFromHTTPType(tt.httpType)

			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Errorf("Expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
			}

			if gotURL != tt.wantURL {
				t.Errorf("Expected URL %q, got %q", tt.wantURL, gotURL)
			}
		})
	}
}

func TestMatchingItemIDsForEventWarningThresholds(t *testing.T) {
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Fatal(dErr)
	}
	dm := NewDataManager(data, cfg, nil)

	tnt := cfg.GeneralCfg().DefaultTenant
	contextKey := utils.MetaRating
	tntCtx := utils.ConcatenatedKey(tnt, contextKey)

	for i := 0; i < matchedItemsWarningThreshold+5; i++ {
		fid := fmt.Sprintf("filter_%d", i)
		rule := &FilterRule{
			Type:    utils.MetaString,
			Element: "~*req.Field",
			Values:  []string{"trigger"},
		}
		if err := rule.CompileValues(); err != nil {
			t.Fatal(err)
		}
		f := &Filter{
			Tenant: tnt,
			ID:     fid,
			Rules:  []*FilterRule{rule},
		}
		if err := dm.SetFilter(context.Background(), f, true); err != nil {
			t.Fatal(err)
		}
		if err := addItemToFilterIndex(context.Background(), dm, utils.CacheAttributeFilterIndexes,
			tnt, contextKey, fid, []string{fid}); err != nil {
			t.Fatal(err)
		}
	}

	ev := utils.MapStorage{
		utils.MetaReq: map[string]any{
			"Field": "trigger",
		},
	}

	ids, err := MatchingItemIDsForEvent(context.Background(), ev, nil, nil, nil, nil, nil,
		dm, utils.CacheAttributeFilterIndexes, tntCtx, true, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(ids) <= matchedItemsWarningThreshold {
		t.Errorf("Expected more than %d matched items, got %d",
			matchedItemsWarningThreshold, len(ids))
	}

	expectedLog := "Matched 105 *attribute_filter_indexes items. Performance may be affected."
	logContent := buf.String()
	if !strings.Contains(logContent, expectedLog) {
		t.Errorf("Expected warning not found in logs:\n%s", logContent)
	}
}
