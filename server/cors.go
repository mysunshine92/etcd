/*
Copyright 2013 CoreOS Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"fmt"
	"net/http"
	"net/url"
)

type corsInfo map[string]bool

func NewCORSInfo(origins []string) (*corsInfo, error) {
	// Construct a lookup of all origins.
	m := make(map[string]bool)
	for _, v := range origins {
		if v != "*" {
			if _, err := url.Parse(v); err != nil {
				return nil, fmt.Errorf("Invalid CORS origin: %s", err)
			}
		}
		m[v] = true
	}

	info := corsInfo(m)
	return &info, nil
}

// OriginAllowed determines whether the server will allow a given CORS origin.
func (c corsInfo) OriginAllowed(origin string) bool {
	return c["*"] || c[origin]
}

type CORSHTTPMiddleware struct {
	Handler http.Handler
	Info    *corsInfo
}

// addHeader adds the correct cors headers given an origin
func (h *CORSHTTPMiddleware) addHeader(w http.ResponseWriter, origin string) {
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Add("Access-Control-Allow-Origin", origin)
}

// ServeHTTP adds the correct CORS headers based on the origin and returns immediatly
// with a 200 OK if the method is OPTIONS.
func (h *CORSHTTPMiddleware) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Write CORS header.
	if h.Info.OriginAllowed("*") {
		h.addHeader(w, "*")
	} else if origin := req.Header.Get("Origin"); h.Info.OriginAllowed(origin) {
		h.addHeader(w, origin)
	}

	if req.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	h.Handler.ServeHTTP(w, req)
}
