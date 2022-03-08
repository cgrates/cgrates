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

package cores

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

// use provides a cleaner interface for chaining middleware for single routes.
// Middleware functions are simple HTTP handlers (w http.ResponseWriter, r *http.Request)
//
//  r.HandleFunc("/login", use(loginHandler, rateLimit, csrf))
//  r.HandleFunc("/form", use(formHandler, csrf))
//  r.HandleFunc("/about", aboutHandler)
//
// From https://gist.github.com/elithrar/9146306
// See https://gist.github.com/elithrar/7600878#comment-955958 for how to extend it to suit simple http.Handler's
func use(h http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, m := range middleware {
		h = m(h)
	}

	return h
}

type basicAuthMiddleware func(h http.HandlerFunc) http.HandlerFunc

// basicAuth returns a middleware function to intercept the request and validate
func basicAuth(userList map[string]string) basicAuthMiddleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

			authHeader := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
			if len(authHeader) != 2 {
				utils.Logger.Warning("<BasicAuth> Missing authorization header value")
				http.Error(w, "Not authorized", http.StatusUnauthorized)
				return
			}

			authHeaderDecoded, err := base64.StdEncoding.DecodeString(authHeader[1])
			if err != nil {
				utils.Logger.Warning("<BasicAuth> Unable to decode authorization header")
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			userPass := strings.SplitN(string(authHeaderDecoded), ":", 2)
			if len(userPass) != 2 {
				utils.Logger.Warning("<BasicAuth> Unauthorized API access. Missing or extra credential components")
				http.Error(w, "Not authorized", http.StatusUnauthorized)
				return
			}

			valid := verifyCredential(userPass[0], userPass[1], userList)
			if !valid {
				utils.Logger.Warning(fmt.Sprintf("<BasicAuth> Unauthorized API access by user '%s'", userPass[0]))
				http.Error(w, "Not authorized", http.StatusUnauthorized)
				return
			}

			h.ServeHTTP(w, r)
		}
	}
}

// verifyCredential validates the incoming username and password against the authorized user list
func verifyCredential(username string, password string, userList map[string]string) bool {
	hash, ok := userList[username]
	if !ok {
		return false
	}

	storedPass, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return false
	}

	return string(storedPass[:]) == password
}
