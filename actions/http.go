// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package actions

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// reqDetails holds HTTP request details: URL, method, headers and static body
type reqDetails struct {
	StaticBody  map[string]any
	ContentType string
	Method      string
	Headers     http.Header
	URL         string
}

// newActHTTPGeneric creates a new action that performs HTTP requests
func newActHTTP(ctx *context.Context, tnt string, cgrEv *utils.CGREvent,
	_ *engine.CacheS, fltrS *engine.FilterS, cfg *config.CGRConfig, aCfg *utils.APAction) (aL *actHTTP, err error) {
	weights := make(map[string]float64)
	diktats := make([]*utils.APDiktat, 0)
	data := cgrEv.AsDataProvider()
	for _, diktat := range aCfg.Diktats {
		if pass, err := fltrS.Pass(ctx, cfg.GeneralCfg().DefaultTenant, diktat.FilterIDs, data); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		weight, err := engine.WeightFromDynamics(ctx, diktat.Weights, fltrS, cfg.GeneralCfg().DefaultTenant, data)
		if err != nil {
			return nil, err
		}
		weights[diktat.ID] = weight
		diktats = append(diktats, diktat)
	}
	slices.SortFunc(diktats, func(a, b *utils.APDiktat) int {
		return cmp.Compare(weights[b.ID], weights[a.ID])
	})

	client := &http.Client{Transport: cfg.HTTPCfg().ClientOpts, Timeout: cfg.GeneralCfg().ReplyTimeout} // shared http.Client for connection reuse

	aL = &actHTTP{
		config: cfg,
		aCfg:   aCfg,
		fltrS:  fltrS,
		client: client,
	}

	dP := cgrEv.AsDataProvider() // convert event to data provider
	reqD := make([]*reqDetails, 0)
	for _, actD := range diktats {
		// get value of opts *httpURL and parse it
		urlOpts, err := engine.GetStringOpts(ctx, tnt, dP, actD.Opts, fltrS, cfg.ActionSCfg().Opts.HttpURL,
			utils.MetaHTTPurl)
		if err != nil {
			return nil, err
		}
		processedURL, err := utils.ParseParamForDataProvider(urlOpts, data, false)
		if err != nil {
			return nil, err
		}
		// get value of opts *httpHdrs and parse it
		h := make(http.Header)
		hdrOpts, err := engine.GetInterfaceOpts(ctx, tnt, dP, actD.Opts, fltrS, cfg.ActionSCfg().Opts.HttpHdrs, utils.MetaHTTPHdrs)
		if err != nil {
			return nil, err
		}
		if hdrOpts != nil {
			if headersIf, ok := hdrOpts.(map[string]any); ok {
				for hdrKey, hdrsValIf := range headersIf {
					// append per header values per key
					switch hdrVal := hdrsValIf.(type) {
					case []string:
						for _, hdrV := range hdrVal {
							hdrVProcessed, err := utils.ParseParamForDataProvider(hdrV, data, false)
							if err != nil {
								return nil, err
							}
							h.Add(hdrKey, hdrVProcessed)
						}
					case string:
						hdrVProcessed, err := utils.ParseParamForDataProvider(hdrVal, data, false)
						if err != nil {
							return nil, err
						}
						h.Add(hdrKey, hdrVProcessed)
					case []any:
						for _, hdrVIf := range hdrVal {
							if hdrV, ok := hdrVIf.(string); ok {
								hdrVProcessed, err := utils.ParseParamForDataProvider(hdrV, data, false)
								if err != nil {
									return nil, err
								}
								h.Add(hdrKey, hdrVProcessed)
							}
						}
					}
				}
			}
		}
		// get value of opts *httpMethod
		method, err := engine.GetStringOpts(ctx, tnt, dP, actD.Opts, fltrS, cfg.ActionSCfg().Opts.HttpMethod, utils.MetaHTTPMethod)
		if err != nil {
			return nil, err
		}
		// get value of opts *httpContentType
		contentType, err := engine.GetStringOpts(ctx, tnt, dP, actD.Opts, fltrS, cfg.ActionSCfg().Opts.HttpContentType, utils.MetaHTTPContentType)
		if err != nil {
			return nil, err
		}
		// get value of opts *httpBody, if populated, its value will be inserted in the request body instead of the event that triggered this action
		var staticBody map[string]any
		staticBodyAny, err := engine.GetInterfaceOpts(ctx, tnt, dP, actD.Opts, fltrS, cfg.ActionSCfg().Opts.HttpBody, utils.MetaHTTPBody)
		if err != nil {
			return nil, err
		}
		if sb, canCast := staticBodyAny.(map[string]any); canCast {
			staticBody = sb
		} else {
			return nil, fmt.Errorf("failed to cast body <%#v> of type <%T>", staticBodyAny, staticBodyAny)
		}
		// populate request details
		aL.reqDetails = append(reqD, &reqDetails{
			StaticBody:  staticBody,
			ContentType: contentType,
			Method:      strings.ToUpper(method),
			Headers:     h,
			URL:         processedURL,
		})
		if blocker, err := engine.BlockerFromDynamics(ctx, actD.Blockers, aL.fltrS, aL.config.GeneralCfg().DefaultTenant, data); err != nil {
			return nil, err
		} else if blocker {
			break
		}
	}
	return aL, nil
}

// execute implements actioner interface
func (aL *actHTTP) execute(ctx *context.Context, data utils.MapStorage, _ string) (err error) {
	var partExec bool
	for _, reqD := range aL.reqDetails {
		if reqD.StaticBody != nil {
			body, err := utils.GetDynamicPathValueFromMapStringAny(reqD.StaticBody, data, false)
			if err != nil {
				return err
			}
			data = body
		}
		var body io.Reader
		// build body depending on ContentType
		if reqD.ContentType == utils.ContentForm {
			urlVals := url.Values{}
			for k, v := range data {
				urlVals.Set(k, utils.IfaceAsString(v))
			}
			body = strings.NewReader(urlVals.Encode())
			reqD.Headers.Set("Content-Type", "application/x-www-form-urlencoded")
		} else {
			jsonBody, err := json.Marshal(data)
			if err != nil {
				return err
			}
			body = bytes.NewBuffer(jsonBody)
			reqD.Headers.Set("Content-Type", "application/json")
		}

		req, rerr := http.NewRequestWithContext(context.Background(), reqD.Method, reqD.URL, body)
		if rerr != nil {
			utils.Logger.Warning(fmt.Sprintf("<ActionS> failed to create <%v> request <%v> to URL <%v> with error <%v>", reqD.Method, data, reqD.URL, rerr))
			partExec = true
			continue
		}
		req.Header = reqD.Headers

		// dispatch request
		if async, has := aL.cfg().Opts[utils.MetaAsync]; has && utils.IfaceAsString(async) == utils.TrueStr {
			go func(r *http.Request) {
				resp, err := aL.client.Do(r)
				if err != nil {
					utils.Logger.Warning("<ActionS> *async http request failed: " + err.Error())
					return
				}
				// cleanup response to reuse the connection
				_, err = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				if err != nil {
					utils.Logger.Warning("<ActionS> *http request error: " + err.Error())
				}
				if resp.StatusCode > 299 {
					utils.Logger.Warning(fmt.Sprintf("<ActionS> unexpected status code received: <%d>", resp.StatusCode))
				}
			}(req)
		} else {
			resp, err := aL.client.Do(req)
			if err != nil {
				utils.Logger.Warning("<ActionS> *http client request error: " + err.Error())
				partExec = true
				continue
			}
			// cleanup response to reuse the connection
			_, err = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if err != nil {
				utils.Logger.Warning("<ActionS> *http request error: " + err.Error())
				partExec = true
				continue
			}
			if resp.StatusCode > 299 {
				partExec = true
				utils.Logger.Warning(fmt.Sprintf("<ActionS> unexpected status code received: <%d>", resp.StatusCode))
				continue
			}
		}
	}
	if partExec {
		err = utils.ErrPartiallyExecuted
	}
	return
}

type actHTTP struct {
	config *config.CGRConfig
	aCfg   *utils.APAction
	fltrS  *engine.FilterS

	client     *http.Client
	reqDetails []*reqDetails
}

func (aL *actHTTP) id() string {
	return aL.aCfg.ID
}

func (aL *actHTTP) cfg() *utils.APAction {
	return aL.aCfg
}

func newActHTTPPost(ctx *context.Context, tnt string, cgrEv *utils.CGREvent,
	cache *engine.CacheS, fltrS *engine.FilterS, cfg *config.CGRConfig, aCfg *utils.APAction) (aL *actHTTPPost, err error) {
	weights := make(map[string]float64)   // stores sorting weights by Diktat ID
	diktats := make([]*utils.APDiktat, 0) // list of diktats which have *balancePath in opts, will be weight sorted later
	data := cgrEv.AsDataProvider()
	for _, diktat := range aCfg.Diktats {
		if pass, err := fltrS.Pass(ctx, cfg.GeneralCfg().DefaultTenant, diktat.FilterIDs, data); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		weight, err := engine.WeightFromDynamics(ctx, diktat.Weights, fltrS, cfg.GeneralCfg().DefaultTenant, data)
		if err != nil {
			return nil, err
		}
		weights[diktat.ID] = weight
		diktats = append(diktats, diktat)
	}
	// Sort by weight (higher values first).
	slices.SortFunc(diktats, func(a, b *utils.APDiktat) int {
		return cmp.Compare(weights[b.ID], weights[a.ID])
	})
	aL = &actHTTPPost{
		config: cfg,
		aCfg:   aCfg,
		pstrs:  make([]*ees.HTTPjsonMapEE, len(diktats)),
	}
	for i, actD := range diktats {
		attempts, err := engine.GetIntOpts(ctx, tnt, cgrEv.AsDataProvider(), nil, fltrS, cfg.ActionSCfg().Opts.PosterAttempts,
			utils.MetaPosterAttempts)
		if err != nil {
			return nil, err
		}
		processedURL, err := utils.ParseParamForDataProvider(utils.IfaceAsString(actD.Opts[utils.MetaHTTPurl]), data, false)
		if err != nil {
			return nil, err
		}
		eeCfg := config.NewEventExporterCfg(aL.id(), utils.EmptyString, processedURL,
			cfg.EEsCfg().ExporterCfg(utils.MetaDefault).FailedPostsDir,
			attempts, nil)
		aL.pstrs[i], _ = ees.NewHTTPjsonMapEE(eeCfg, cfg, cache, nil, nil)
		// append headers from Opts, if a key exists it will append the values to it
		if headers, has := actD.Opts[utils.MetaHTTPHdrs]; has {
			if headersIf, ok := headers.(map[string]any); ok {
				for hdrKey, hdrsValIf := range headersIf {
					switch hdrVal := hdrsValIf.(type) {
					case []string:
						for _, hdrV := range hdrVal {
							hdrVProcessed, err := utils.ParseParamForDataProvider(hdrV, data, false)
							if err != nil {
								return nil, err
							}
							aL.pstrs[i].AppendHeader(hdrKey, hdrVProcessed)
						}
					case string:
						hdrVProcessed, err := utils.ParseParamForDataProvider(hdrVal, data, false)
						if err != nil {
							return nil, err
						}
						aL.pstrs[i].AppendHeader(hdrKey, hdrVProcessed)
					case []any:
						for _, hdrVIf := range hdrVal {
							if hdrV, ok := hdrVIf.(string); ok {
								hdrVProcessed, err := utils.ParseParamForDataProvider(hdrV, data, false)
								if err != nil {
									return nil, err
								}
								aL.pstrs[i].AppendHeader(hdrKey, hdrVProcessed)
							}
						}
					}
				}
			}
		}
		if blocker, err := engine.BlockerFromDynamics(ctx, actD.Blockers, aL.fltrS, aL.config.GeneralCfg().DefaultTenant, data); err != nil {
			return nil, err
		} else if blocker {
			break
		}
	}
	return aL, nil
}
