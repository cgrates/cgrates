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
	URL        string
	Method     string
	Headers    http.Header
	StaticBody map[string]any
}

// newActHTTPGeneric creates a new action that performs HTTP requests
func newActHTTP(ctx *context.Context, tnt string, cgrEv *utils.CGREvent,
	cache *engine.CacheS, fltrS *engine.FilterS, cfg *config.CGRConfig, aCfg *utils.APAction) (aL *actHTTP, err error) {
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
		// get value of opts *url and parse it
		urlOpts, err := engine.GetStringOpts(ctx, tnt, dP, actD.Opts, fltrS, cfg.ActionSCfg().Opts.Url,
			utils.MetaURL)
		if err != nil {
			return nil, err
		}
		processedURL, err := utils.ParseParamForDataProvider(urlOpts, data, false)
		if err != nil {
			return nil, err
		}
		// get value of opts *hdrs and parse it
		h := make(http.Header)
		hdrOpts, err := engine.GetInterfaceOpts(ctx, tnt, dP, actD.Opts, fltrS, cfg.ActionSCfg().Opts.Hdrs, utils.MetaHdrs)
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
		// get value of opts *method
		method, err := engine.GetStringOpts(ctx, tnt, dP, actD.Opts, fltrS, cfg.ActionSCfg().Opts.Method, utils.MetaMethod)
		if err != nil {
			return nil, err
		}
		// get value of opts *body, if populated, its value will be inserted in the request body instead of the event that triggered this action
		var staticBody map[string]any
		staticBodyAny, err := engine.GetInterfaceOpts(ctx, tnt, dP, actD.Opts, fltrS, cfg.ActionSCfg().Opts.Body, utils.MetaBody)
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
			URL:        processedURL,
			Method:     strings.ToUpper(method),
			Headers:    h,
			StaticBody: staticBody,
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
			data = reqD.StaticBody
		}
		var body []byte
		if body, err = json.Marshal(data); err != nil {
			return
		}
		req, rerr := http.NewRequestWithContext(context.Background(), reqD.Method, reqD.URL, bytes.NewBuffer(body))
		if rerr != nil {
			utils.Logger.Warning(fmt.Sprintf("<ActionS> failed to send <%v> request <%v> to URL <%v> with error <%v>", reqD.Method, data, reqD.URL, rerr))
			partExec = true
			continue
		}
		req.Header = reqD.Headers
		req.Header.Set("Content-Type", "application/json")

		// dispatch request
		if async, has := aL.cfg().Opts[utils.MetaAsync]; has && utils.IfaceAsString(async) == utils.TrueStr {
			go func(r *http.Request) {
				resp, err := aL.client.Do(r)
				if err != nil {
					utils.Logger.Warning("async http request failed: " + err.Error())
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
		processedURL, err := utils.ParseParamForDataProvider(utils.IfaceAsString(actD.Opts[utils.MetaURL]), data, false)
		if err != nil {
			return nil, err
		}
		eeCfg := config.NewEventExporterCfg(aL.id(), utils.EmptyString, processedURL,
			cfg.EEsCfg().ExporterCfg(utils.MetaDefault).FailedPostsDir,
			attempts, nil)
		aL.pstrs[i], _ = ees.NewHTTPjsonMapEE(eeCfg, cfg, cache, nil, nil)
		// append headers from Opts, if a key exists it will append the values to it
		if headers, has := actD.Opts[utils.MetaHdrs]; has {
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

type actHTTPPost struct {
	config *config.CGRConfig
	aCfg   *utils.APAction
	fltrS  *engine.FilterS

	pstrs []*ees.HTTPjsonMapEE
}

func (aL *actHTTPPost) id() string {
	return aL.aCfg.ID
}

func (aL *actHTTPPost) cfg() *utils.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actHTTPPost) execute(ctx *context.Context, data utils.MapStorage, _ string) (err error) {
	var body []byte
	if body, err = json.Marshal(data); err != nil {
		return
	}
	var partExec bool
	for _, pstr := range aL.pstrs {
		if async, has := aL.cfg().Opts[utils.MetaAsync]; has && utils.IfaceAsString(async) == utils.TrueStr {
			go ees.ExportWithAttempts(context.Background(), pstr, &ees.HTTPPosterRequest{Body: body, Header: pstr.Headers()}, utils.EmptyString,
				nil, aL.config.GeneralCfg().DefaultTenant, aL.fltrS)
		} else if err = ees.ExportWithAttempts(ctx, pstr, &ees.HTTPPosterRequest{Body: body, Header: pstr.Headers()}, utils.EmptyString,
			nil, aL.config.GeneralCfg().DefaultTenant, aL.fltrS); err != nil {
			if pstr.Cfg().FailedPostsDir != utils.MetaNone {
				err = nil
			} else {
				partExec = true
			}
		}
	}
	if partExec {
		err = utils.ErrPartiallyExecuted
	}
	return
}

type actExport struct {
	tnt     string
	config  *config.CGRConfig
	connMgr *engine.ConnManager
	fltrS   *engine.FilterS
	aCfg    *utils.APAction
}

func (aL *actExport) id() string {
	return aL.aCfg.ID
}

func (aL *actExport) cfg() *utils.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actExport) execute(ctx *context.Context, data utils.MapStorage, _ string) (err error) {
	var exporterIDs []string
	if expIDs, has := aL.cfg().Opts[utils.MetaExporterIDs]; has {
		exporterIDs = strings.Split(utils.IfaceAsString(expIDs), utils.InfieldSep)
	}
	var eesConns []string
	if eesConns, err = engine.GetConnIDs(ctx, aL.config.ActionSCfg().Conns, utils.MetaEEs, aL.tnt, data, nil, aL.fltrS); err != nil {
		return
	}
	var rply map[string]map[string]any
	return aL.connMgr.Call(ctx, eesConns,
		utils.EeSv1ProcessEvent, &utils.CGREventWithEeIDs{
			EeIDs: exporterIDs,
			CGREvent: &utils.CGREvent{
				Tenant:  aL.tnt,
				ID:      utils.GenUUID(),
				Event:   data[utils.MetaReq].(map[string]any),
				APIOpts: data[utils.MetaOpts].(map[string]any),
			},
		}, &rply)
}
