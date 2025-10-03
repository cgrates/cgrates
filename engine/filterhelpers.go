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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

const matchedItemsWarningThreshold int = 100

// MatchingItemIDsForEvent returns the list of item IDs matching fieldName/fieldValue for an event
// fieldIDs limits the fields which are checked against indexes
// helper on top of dataDB.GetIndexes, adding utils.MetaAny to list of fields queried
func MatchingItemIDsForEvent(ctx *context.Context, ev utils.MapStorage, stringFldIDs, prefixFldIDs, suffixFldIDs, existsFldIDs, notExistsFldIDs *[]string,
	dm *DataManager, cacheID, itemIDPrefix string, indexedSelects, nestedFields bool) (itemIDs utils.StringSet, err error) {
	itemIDs = make(utils.StringSet)
	var allFieldIDs []string
	if indexedSelects && (stringFldIDs == nil || prefixFldIDs == nil || suffixFldIDs == nil || existsFldIDs == nil || notExistsFldIDs == nil) {
		allFieldIDs = ev.GetKeys(nestedFields, 2, utils.EmptyString)
	}
	// Guard will protect the function with automatic locking
	lockID := utils.CacheInstanceToPrefix[cacheID] + itemIDPrefix
	guardian.Guardian.Guard(ctx, func(ctx *context.Context) (_ error) {
		if !indexedSelects {
			var keysWithID []string
			if keysWithID, err = dm.DataDB().GetKeysForPrefix(ctx, utils.CacheIndexesToPrefix[cacheID]); err != nil {
				return
			}
			var sliceIDs []string
			for _, id := range keysWithID {
				sliceIDs = append(sliceIDs, utils.SplitConcatenatedKey(id)[1])
			}
			itemIDs = utils.NewStringSet(sliceIDs)
			return
		}
		stringFieldVals := map[string]string{utils.MetaAny: utils.MetaAny}                                                                        // cache here field string values, start with default one
		filterIndexTypes := []string{utils.MetaString, utils.MetaPrefix, utils.MetaSuffix, utils.MetaExists, utils.MetaNotExists, utils.MetaNone} // the MetaNone is used for all items that do not have filters
		for i, fieldIDs := range []*[]string{stringFldIDs, prefixFldIDs, suffixFldIDs, existsFldIDs, notExistsFldIDs, {utils.MetaAny}} {          // same routine for  filter indexes types
			if fieldIDs == nil {
				fieldIDs = &allFieldIDs
			}
			for _, fldName := range *fieldIDs {
				var fieldValIf any
				fieldValIf, err = ev.FieldAsInterface(utils.SplitPath(fldName, utils.NestingSep[0], -1))
				if err == nil && filterIndexTypes[i] == utils.MetaNotExists {
					continue // field should not exist in our event in order to check index
				}
				if err != nil && filterIndexTypes[i] != utils.MetaNone {
					continue
				}
				if _, cached := stringFieldVals[fldName]; !cached {
					stringFieldVals[fldName] = utils.IfaceAsString(fieldValIf)
				}
				fldVal := stringFieldVals[fldName]
				fldVals := []string{fldVal}
				// default is only one fieldValue checked
				var dbItemIDs utils.StringSet // list of items matched in DB
				switch filterIndexTypes[i] {
				case utils.MetaPrefix:
					fldVals = utils.SplitPrefix(fldVal, 1) // all prefixes till last digit
				case utils.MetaSuffix:
					fldVals = utils.SplitSuffix(fldVal)
				case utils.MetaExists:
					fldVals = []string{utils.MetaAny} // for *exists, we will use *any value
				case utils.MetaNotExists:
					fldVals = []string{utils.MetaNone} // for *notexists, we will use *none
				}
				for _, val := range fldVals {
					var dbIndexes map[string]utils.StringSet // list of items matched in DB
					key := utils.ConcatenatedKey(filterIndexTypes[i], fldName, val)
					if dbIndexes, err = dm.GetIndexes(ctx, cacheID, itemIDPrefix, key, utils.NonTransactional, true, true); err != nil {
						if err == utils.ErrNotFound {
							err = nil
							continue
						}
						return
					}
					dbItemIDs = dbIndexes[key]
					break // we got at least one answer back, longest prefix wins
				}
				itemIDs = utils.JoinStringSet(itemIDs, dbItemIDs)
			}
		}
		return
	},
		config.CgrConfig().GeneralCfg().LockingTimeout, lockID)
	if err != nil {
		return nil, err
	}
	if len(itemIDs) == 0 {
		return nil, utils.ErrNotFound
	}
	if len(itemIDs) > matchedItemsWarningThreshold {
		utils.Logger.Warning(fmt.Sprintf(
			"Matched %d %s items. Performance may be affected.\nevent = %s",
			len(itemIDs), cacheID, utils.ToJSON(ev)))
	}
	return
}

// Weight returns the weight specified by the first matching DynamicWeight
func WeightFromDynamics(ctx *context.Context, dWs []*utils.DynamicWeight,
	fltrS *FilterS, tnt string, ev utils.DataProvider) (wg float64, err error) {
	for _, dW := range dWs {
		var pass bool
		if pass, err = fltrS.Pass(ctx, tnt, dW.FilterIDs, ev); err != nil {
			return
		} else if pass {
			return dW.Weight, nil
		}
	}
	return 0.0, nil
}

// BlockerFromDynamics returns the value of the blocker specified by the first matching DynamicBlocker
func BlockerFromDynamics(ctx *context.Context, dBs []*utils.DynamicBlocker,
	fltrS *FilterS, tnt string, ev utils.DataProvider) (blckr bool, err error) {
	for _, dB := range dBs {
		var pass bool
		if pass, err = fltrS.Pass(ctx, tnt, dB.FilterIDs, ev); err != nil {
			return
		} else if pass {
			return dB.Blocker, nil
		}
	}
	return false, nil
}

func GetSentryPeer(ctx *context.Context, val string, sentryPeerCfg *config.SentryPeerCfg, dataType string) (found bool, err error) {
	itemId := utils.ConcatenatedKey(dataType, val)
	var (
		isCached bool
		apiUrl   string
		token    string
	)
	if x, ok := Cache.Get(utils.MetaSentryPeer, itemId); ok && x != nil { // Attempt to find in cache first
		return x.(bool), nil
	}
	var cachedToken any
	if cachedToken, isCached = Cache.Get(utils.MetaSentryPeer,
		utils.MetaToken); isCached && cachedToken != nil {
		token = cachedToken.(string)
	}
	switch dataType {
	case utils.MetaIP:
		apiUrl, err = url.JoinPath(sentryPeerCfg.IpsUrl, val)
	case utils.MetaNumber:
		apiUrl, err = url.JoinPath(sentryPeerCfg.NumbersUrl, val)
	}
	if err != nil {
		return
	}
	if !isCached {
		if token, err = sentrypeerGetToken(sentryPeerCfg.TokenUrl, sentryPeerCfg.ClientID, sentryPeerCfg.ClientSecret,
			sentryPeerCfg.Audience, sentryPeerCfg.GrantType); err != nil {
			utils.Logger.Err(fmt.Sprintf("sentrypeer token auth got err <%v> ", err.Error()))
			return
		}
		if err = Cache.Set(ctx, utils.MetaSentryPeer, utils.MetaToken,
			token, nil, true, utils.NonTransactional); err != nil {
			return
		}
	}

	for i := 0; i < 2; i++ {
		if found, err = sentrypeerHasData(itemId, token, apiUrl); err == nil {
			if err = Cache.Set(ctx, utils.MetaSentryPeer, itemId, found,
				nil, true, ""); err != nil {
				return
			}
			break
		} else if err != utils.ErrNotAuthorized {
			utils.Logger.Err(err.Error())
			break
		}
		utils.Logger.Warning("Sentrypeer token expired !Getting new one.")
		Cache.Remove(ctx, utils.MetaSentryPeer, utils.MetaToken, true, utils.EmptyString)
		if token, err = sentrypeerGetToken(sentryPeerCfg.TokenUrl, sentryPeerCfg.ClientID, sentryPeerCfg.ClientSecret,
			sentryPeerCfg.Audience, sentryPeerCfg.GrantType); err != nil {
			return
		}
		if err = Cache.Set(ctx, utils.MetaSentryPeer, utils.MetaToken, token,
			nil, true, utils.NonTransactional); err != nil {
			return
		}
	}
	return
}

// Returns a new token from sentrypeer api
func sentrypeerGetToken(tokenUrl, clientID, clientSecret, audience, grantType string) (token string, err error) {
	var resp *http.Response
	payload := map[string]string{
		utils.ClientIDCfg:     clientID,
		utils.ClientSecretCfg: clientSecret,
		utils.AudienceCfg:     audience,
		utils.GrantTypeCfg:    grantType,
	}
	jsonPayload, _ := json.Marshal(payload)
	resp, err = getHTTP(http.MethodPost, tokenUrl, bytes.NewBuffer(jsonPayload), map[string]string{"Content-Type": "application/json"})
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var m struct {
		AccessToken string `json:"access_token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&m)
	if err != nil {
		return
	}
	token = m.AccessToken
	return
}

// sentrypeerHasData return a boolean based on query response on finding ip/number
func sentrypeerHasData(itemId, token, url string) (found bool, err error) {
	var resp *http.Response
	resp, err = getHTTP(http.MethodGet, url, nil, map[string]string{"Authorization": fmt.Sprintf("Bearer %s", token)})
	if err != nil {
		return
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return false, utils.ErrNotAuthorized
	case resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices:
		return false, nil
	case resp.StatusCode == http.StatusNotFound:
		return true, nil
	case resp.StatusCode >= http.StatusBadRequest && resp.StatusCode < http.StatusInternalServerError:
		err = fmt.Errorf("sentrypeer api got client err <%v>", resp.Status)
	case resp.StatusCode >= http.StatusInternalServerError:
		err = fmt.Errorf("sentrypeer api got server err <%s>", resp.Status)
	default:
		err = fmt.Errorf("sentrypeer api got unexpected err <%s>", resp.Status)
	}
	return false, err
}
func filterHTTP(httpType string, dDP utils.DataProvider, fieldname, value string) (bool, error) {
	var (
		parsedURL *url.URL
		resp      string
		err       error
	)
	urlS, err := ExtractURLFromHTTPType(httpType)
	if err != nil {
		return false, err
	}

	parsedURL, err = url.Parse(urlS)
	if err != nil {
		return false, err
	}
	if fieldname != utils.MetaAny {
		queryParams := parsedURL.Query()
		queryParams.Set(fieldname, value)
		parsedURL.RawQuery = queryParams.Encode()
		resp, err = MakeExternalAPIRequest(parsedURL.String(), nil)
	} else {
		resp, err = MakeExternalAPIRequest(parsedURL.String(), bytes.NewReader([]byte(dDP.String())))
	}
	if err != nil {
		return false, err
	}
	return utils.IfaceAsBool(resp)
}

// MakeExternalAPIRequest makes an HTTP GET request to the specified URL with
// the provided request body and returns the response body as a string.
func MakeExternalAPIRequest(url string, reader io.Reader) (string, error) {
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	resp, err := getHTTP(http.MethodGet, url, reader, headers)
	if err != nil {
		return "", fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusMultipleChoices {
		return "", fmt.Errorf("HTTP request failed with status code %d: %s",
			resp.StatusCode, string(body))
	}

	return string(body), nil
}

// constructs an request via parameters provided,url,header and payload ,uses defaultclient for sending the request
func getHTTP(method, url string, payload io.Reader, headers map[string]string) (resp *http.Response, err error) {
	var req *http.Request
	if req, err = http.NewRequest(method, url, payload); err != nil {
		return
	}
	for k, hVal := range headers {
		req.Header.Add(k, hVal)
	}
	return http.DefaultClient.Do(req)
}

// ExtractURLFromHTTPType parses a type string in the format "prefix#[url]" and
// returns the URL portion.
func ExtractURLFromHTTPType(typeStr string) (string, error) {
	parts := strings.Split(typeStr, utils.HashtagSep)
	if len(parts) != 2 {
		return "", errors.New("invalid format: URL portion not found")
	}

	url := strings.Trim(parts[1], utils.IdxStart+utils.IdxEnd)
	return url, nil
}
