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

package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	janus "github.com/cgrates/janusgo"
)

var (
	janSessionPath      = regexp.MustCompile(`^\/janus/\d+?$`)
	janPluginHandlePath = regexp.MustCompile(`^\/janus/.*/.*`)
)

// NewJanusAgent will construct a JanusAgent
func NewJanusAgent(cgrCfg *config.CGRConfig,
	connMgr *engine.ConnManager,
	filterS *engine.FilterS) *JanusAgent {
	return &JanusAgent{
		cgrCfg:  cgrCfg,
		connMgr: connMgr,
		filterS: filterS,
	}
}

// JanusAgent is a gateway between HTTP and Janus Server over Websocket
type JanusAgent struct {
	cgrCfg  *config.CGRConfig
	connMgr *engine.ConnManager
	filterS *engine.FilterS
	jnsConn *janus.Gateway
}

// Connect will create the connection to the Janus Server
func (ja *JanusAgent) Connect() (err error) {
	ja.jnsConn, err = janus.Connect(
		fmt.Sprintf("ws://%s", ja.cgrCfg.JanusAgentCfg().JanusConns[0].Address))
	return
}

// Shutdown will close the connection to the Janus Server
func (ja *JanusAgent) Shutdown() error {
	return ja.jnsConn.Close()
}

// ServeHTTP implements http.Handler interface
func (ja *JanusAgent) CORSOptions(w http.ResponseWriter, req *http.Request) {
	janusAccessControlHeaders(w, req)
}

// CreateSession will create a new session within janusgo
func (ja *JanusAgent) CreateSession(w http.ResponseWriter, req *http.Request) {
	janusAccessControlHeaders(w, req)
	var msg janus.BaseMsg
	if err := json.NewDecoder(req.Body).Decode(&msg); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := ja.jnsConn.CreateSession(ctx, msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(&resp)
}

// SessioNKeepalive sends keepalive once OPTIONS are coming for the session from HTTP
func (ja *JanusAgent) SessioNKeepalive(w http.ResponseWriter, r *http.Request) {
	janusAccessControlHeaders(w, r)
	sessionID, err := strconv.ParseUint(r.PathValue("sessionID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}
	ja.jnsConn.RLock()
	session, has := ja.jnsConn.Sessions[sessionID]
	ja.jnsConn.RUnlock()
	if !has {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	msg := janus.BaseMsg{
		Session: session.ID,
		Type:    "keepalive",
	}
	var resp any
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err = session.KeepAlive(ctx, msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}
	json.NewEncoder(w).Encode(resp)
}

// PollSession will create a long-poll request to be notified about events and incoming messages from session
func (ja *JanusAgent) PollSession(w http.ResponseWriter, req *http.Request) {
	janusAccessControlHeaders(w, req)
	sessionID, err := strconv.ParseUint(req.PathValue("sessionID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}
	ja.jnsConn.RLock()
	session, has := ja.jnsConn.Sessions[sessionID]
	ja.jnsConn.RUnlock()
	if !has {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	maxEvs, err := strconv.Atoi(req.URL.Query().Get("maxev"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid maxev, err: %s", err.Error()),
			http.StatusBadRequest)
		return
	}
	msg := janus.BaseMsg{
		Session: session.ID,
		Type:    "keepalive",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	events, err := session.LongPoll(ctx, maxEvs, msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(events)
}

// AttachPlugin will attach a plugin to a session
func (ja *JanusAgent) AttachPlugin(w http.ResponseWriter, r *http.Request) {
	janusAccessControlHeaders(w, r)
	sessionID, err := strconv.ParseUint(r.PathValue("sessionID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}
	ja.jnsConn.RLock()
	session, has := ja.jnsConn.Sessions[sessionID]
	ja.jnsConn.RUnlock()
	if !has {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	var msg janus.BaseMsg
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	msg.Session = session.ID

	var resp any
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if msg.Type == "destroy" {
		resp, err = session.DestroySession(ctx, msg)
	} else {
		resp, err = session.AttachSession(ctx, msg)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}
	json.NewEncoder(w).Encode(resp)
}

// HandlePlugin will handle requests towards a plugin
func (ja *JanusAgent) HandlePlugin(w http.ResponseWriter, r *http.Request) {
	janusAccessControlHeaders(w, r)
	sessionID, err := strconv.ParseUint(r.PathValue("sessionID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}
	ja.jnsConn.RLock()
	session, has := ja.jnsConn.Sessions[sessionID]
	ja.jnsConn.RUnlock()
	if !has {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	handleID, err := strconv.ParseUint(r.PathValue("handleID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid handle ID", http.StatusBadRequest)
		return
	}
	handle, has := session.Handles[handleID]
	if !has {
		if !has {
			http.Error(w, "Handle not found", http.StatusNotFound)
			return
		}
	}
	rBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Cannot read body", http.StatusBadRequest)
		return
	}
	var msg janus.BaseMsg
	if err := json.Unmarshal(rBody, &msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var resp any
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// handle message, depending on it's type
	switch msg.Type {
	case "message":
		var hMsg janus.HandlerMessageJsep
		if err := json.Unmarshal(rBody, &hMsg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		hMsg.Session = session.ID
		hMsg.BaseMsg.Handle = handle.ID
		hMsg.Handle = handle.ID
		resp, err = handle.Message(ctx, hMsg)
	case "trickle":
		var hMsg janus.TrickleOne
		if err := json.Unmarshal(rBody, &hMsg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		hMsg.Session = session.ID
		hMsg.Handle = handle.ID
		hMsg.HandleR = handle.ID
		resp, err = handle.Trickle(ctx, hMsg)
	default:
		if err != nil {
			http.Error(w, "Invalid message type", http.StatusBadRequest)
			return
		}

	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}
	json.NewEncoder(w).Encode(resp)
}
