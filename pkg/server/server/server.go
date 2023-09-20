/*
Copyright 2023 EscherCloud.

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
	"github.com/eschercloudai/baski/pkg/server/generated"
	"github.com/eschercloudai/baski/pkg/server/handler"
	"github.com/gorilla/mux"
	"net/http"
)

type Options struct {
	ListenAddress string
	ListenPort    int32
	Endpoint      string
	AccessKey     string
	SecretKey     string
	Bucket        string
}

type Server struct {
	Options Options
}

func (s *Server) NewServer(dev bool) (*http.Server, error) {

	middleware := []generated.MiddlewareFunc{}
	r := mux.NewRouter()

	// Here we decide how to handle CORS.
	// If it's dev, we allow the lot if it's not, we only allow the defaults for any OPTION methods
	if !dev {
		r.Use(mux.CORSMethodMiddleware(r))
	} else {
		r.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w = OptionsPreflightAllow(w)
			w.WriteHeader(http.StatusOK)
		})
		middleware = append(middleware, CORSAllowOriginAllMiddleware)
	}

	handlers := handler.New(s.Options.Endpoint, s.Options.AccessKey, s.Options.SecretKey, s.Options.Bucket)

	options := generated.GorillaServerOptions{
		BaseRouter:  r,
		Middlewares: middleware,
	}

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.Options.ListenAddress, s.Options.ListenPort),
		Handler: generated.HandlerWithOptions(handlers, options),
	}

	return server, nil
}

// CORSAllowOriginAllMiddleware sets the header for Access-Control-Allow-Origin = "*"
func CORSAllowOriginAllMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = OptionsPreflightAllow(w)
		next(w, r)
	}
}

// OptionsPreflightAllow sets the header for Access-Control-Allow-Origin = "*"
func OptionsPreflightAllow(w http.ResponseWriter) http.ResponseWriter {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	return w
}
