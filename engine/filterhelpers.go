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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

const matchedItemsWarningThreshold int = 100

// MatchingItemIDsForEvent returns the list of item IDs matching fieldName/fieldValue for an event
// fieldIDs limits the fields which are checked against indexes
// helper on top of dataDB.GetIndexes, adding utils.MetaAny to list of fields queried
func MatchingItemIDsForEvent(ev utils.MapStorage, stringFldIDs, prefixFldIDs, suffixFldIDs, existsFldIDs *[]string,
	dm *DataManager, cacheID, itemIDPrefix string, indexedSelects, nestedFields bool) (itemIDs utils.StringSet, err error) {
	itemIDs = make(utils.StringSet)
	var allFieldIDs []string
	if indexedSelects && (stringFldIDs == nil || prefixFldIDs == nil || suffixFldIDs == nil || existsFldIDs == nil) {
		allFieldIDs = ev.GetKeys(nestedFields, 2, utils.EmptyString)
	}
	// Guard will protect the function with automatic locking
	lockID := utils.CacheInstanceToPrefix[cacheID] + itemIDPrefix
	guardian.Guardian.Guard(func() (_ error) {
		if !indexedSelects {
			var keysWithID []string
			if keysWithID, err = dm.DataDB().GetKeysForPrefix(utils.CacheIndexesToPrefix[cacheID]); err != nil {
				return
			}
			var sliceIDs []string
			for _, id := range keysWithID {
				sliceIDs = append(sliceIDs, utils.SplitConcatenatedKey(id)[1])
			}
			itemIDs = utils.NewStringSet(sliceIDs)
			return
		}
		stringFieldVals := map[string]string{utils.MetaAny: utils.MetaAny}                                                   // cache here field string values, start with default one
		filterIndexTypes := []string{utils.MetaString, utils.MetaPrefix, utils.MetaSuffix, utils.MetaExists, utils.MetaNone} // the MetaNone is used for all items that do not have filters
		for i, fieldIDs := range []*[]string{stringFldIDs, prefixFldIDs, suffixFldIDs, existsFldIDs, {utils.MetaAny}} {      // same routine for both string and prefix filter types
			if fieldIDs == nil {
				fieldIDs = &allFieldIDs
			}
			for _, fldName := range *fieldIDs {
				var dbItemIDs utils.StringSet // list of items matched in DB
				if filterIndexTypes[i] == utils.MetaExists {
					var dbIndexes map[string]utils.StringSet // list of items matched in DB
					key := utils.ConcatenatedKey(filterIndexTypes[i], fldName)
					if dbIndexes, err = dm.GetIndexes(cacheID, itemIDPrefix, true, true, key); err != nil {
						if err == utils.ErrNotFound {
							err = nil
							continue
						}
						return
					}
					dbItemIDs = dbIndexes[key]
					for itemID := range dbItemIDs {
						if _, hasIt := itemIDs[itemID]; !hasIt { // Add it to list if not already there
							itemIDs[itemID] = dbItemIDs[itemID]
						}
					}
					continue // no need to look at values for *exists indexes
				}
				var fieldValIf any
				fieldValIf, err = ev.FieldAsInterface(utils.SplitPath(fldName, utils.NestingSep[0], -1))
				if err != nil && filterIndexTypes[i] != utils.MetaNone {
					continue
				}
				if _, cached := stringFieldVals[fldName]; !cached {
					stringFieldVals[fldName] = utils.IfaceAsString(fieldValIf)
				}
				fldVal := stringFieldVals[fldName]
				fldVals := []string{fldVal}
				// default is only one fieldValue checked
				if filterIndexTypes[i] == utils.MetaPrefix {
					fldVals = utils.SplitPrefix(fldVal, 1) // all prefixes till last digit
				} else if filterIndexTypes[i] == utils.MetaSuffix {
					fldVals = utils.SplitSuffix(fldVal) // all suffix till first digit
				}
				for _, val := range fldVals {
					var dbIndexes map[string]utils.StringSet // list of items matched in DB
					key := utils.ConcatenatedKey(filterIndexTypes[i], fldName, val)
					if dbIndexes, err = dm.GetIndexes(cacheID, itemIDPrefix, true, true, key); err != nil {
						if err == utils.ErrNotFound {
							err = nil
							continue
						}
						return
					}
					dbItemIDs = dbIndexes[key]
					break // we got at least one answer back, longest prefix wins
				}
				for itemID := range dbItemIDs {
					if _, hasIt := itemIDs[itemID]; !hasIt { // Add it to list if not already there
						itemIDs[itemID] = dbItemIDs[itemID]
					}
				}
			}
		}
		return
	}, config.CgrConfig().GeneralCfg().LockingTimeout, lockID)
	if err != nil {
		return nil, err
	}
	if len(itemIDs) == 0 {
		return nil, utils.ErrNotFound
	}
	if len(itemIDs) > matchedItemsWarningThreshold {
		utils.Logger.Warning(fmt.Sprintf(
			"Matched %d %s items. Performance may be affected", len(itemIDs), cacheID))
	}
	return
}

// Weight returns weight of the first matching DynamicWeight
func WeightFromDynamics(dWs []*utils.DynamicWeight,
	fltrS *FilterS, tnt string, ev utils.DataProvider) (wg float64, err error) {
	for _, dW := range dWs {
		var pass bool
		if pass, err = fltrS.Pass(tnt, dW.FilterIDs, ev); err != nil {
			return
		} else if pass {
			return dW.Weight, nil
		}
	}
	return 0.0, nil
}

// fail or pass the filter based on sentrypeer server response
func GetSentryPeer(val string, sentryPeerCfg *config.SentryPeerCfg, dataType string) (found bool, err error) {
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
	case utils.MetaIp:
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
		if err = Cache.Set(utils.MetaSentryPeer, utils.MetaToken,
			token, nil, true, utils.NonTransactional); err != nil {
			return
		}
	}

	for i := 0; i < 2; i++ {
		if found, err = sentrypeerHasData(itemId, token, apiUrl); err == nil {
			if err = Cache.Set(utils.MetaSentryPeer, itemId, found,
				nil, true, utils.NonTransactional); err != nil {
				return
			}
			break
		} else if err != utils.ErrNotAuthorized {
			utils.Logger.Err(err.Error())
			break
		}
		utils.Logger.Warning("Sentrypeer token expired !Getting new one.")
		Cache.Remove(utils.MetaSentryPeer, utils.MetaToken, true, utils.EmptyString)
		if token, err = sentrypeerGetToken(sentryPeerCfg.TokenUrl, sentryPeerCfg.ClientID, sentryPeerCfg.ClientSecret,
			sentryPeerCfg.Audience, sentryPeerCfg.GrantType); err != nil {
			return
		}
		if err = Cache.Set(utils.MetaSentryPeer, utils.MetaToken, token,
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
		utils.ClientIdCfg:     clientID,
		utils.ClientSecretCfg: clientSecret,
		utils.AudienceCfg:     audience,
		utils.GrantTypeCfg:    grantType,
	}
	jsonPayload, _ := json.Marshal(payload)
	resp, err = getHTTP(http.MethodPost, tokenUrl, bytes.NewBuffer(jsonPayload), map[string]string{utils.ContentType: utils.JsonBody})
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
	resp, err = getHTTP(http.MethodGet, url, nil, map[string]string{utils.AuthorizationHdr: fmt.Sprintf("%s %s", utils.BearerAuth, token)})
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

// send an http get request to a server
// expects a boolean reply
// when element is set to *any the  CGREvent is sent as JSON body
// when the element is  specified  as a path e.g ~*req.Account  is sent as  query string pair  ,the path being the key with the value extracted from dataprovider
func filterHTTP(httpType string, dDP utils.DataProvider, fieldname, value string) (bool, error) {
	var (
		parsedURL *url.URL
		resp      string
		err       error
	)
	urlS, err := extractUrlFromType(httpType)
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
		resp, err = externalAPI(parsedURL.String(), nil)
	} else {
		resp, err = externalAPI(parsedURL.String(), bytes.NewReader([]byte(dDP.String())))
	}
	if err != nil {
		return false, err
	}
	return utils.IfaceAsBool(resp)
}

func externalAPI(url string, rdr io.Reader) (string, error) {
	hdr := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := getHTTP(http.MethodGet, url, rdr, hdr)
	if err != nil {
		return "", fmt.Errorf("error processing the request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusMultipleChoices {
		body, err := io.ReadAll(resp.Body)
		return "", fmt.Errorf("http request returned non-OK status code: %d ,body: %v ,err: %w", resp.StatusCode, string(body), err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
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

func extractUrlFromType(httpType string) (string, error) {
	parts := strings.Split(httpType, utils.HashtagSep)
	if len(parts) != 2 {
		return "", errors.New("url is not specified")
	}
	//extracting  the url from the type
	url := strings.Trim(parts[1], utils.IdxStart+utils.IdxEnd)
	return url, nil
}
