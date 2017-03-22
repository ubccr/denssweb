// Copyright 2017 DENSSWeb Authors. All rights reserved.
//
// This file is part of DENSSWeb.
//
// DENSSWeb is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// DENSSWeb is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with DENSSWeb.  If not, see <http://www.gnu.org/licenses/>.

package server

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/ubccr/denssweb/app"
	"github.com/urfave/negroni"
)

const (
	TokenRegex = `[ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789\-\_]+`
)

func init() {
	viper.SetDefault("port", 8080)
	viper.SetDefault("bind", "127.0.0.1")
	viper.SetDefault("base_url", "http://localhost:8080")
}

func middleware(ctx *app.AppContext) *negroni.Negroni {
	router := mux.NewRouter()

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.RenderNotFound(w)
	})

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(fmt.Sprintf("%s/static", ctx.Tmpldir)))))
	router.Path("/about").Handler(AboutHandler(ctx)).Methods("GET")
	router.Path("/submit").Handler(SubmitHandler(ctx)).Methods("GET", "POST")
	router.Path(fmt.Sprintf("/job/{id:%s}", TokenRegex)).Handler(JobHandler(ctx)).Methods("GET")
	router.Path(fmt.Sprintf("/job/{id:%s}/status", TokenRegex)).Handler(StatusHandler(ctx)).Methods("GET")
	router.Path(fmt.Sprintf("/job/{id:%s}/density-map.ccp4", TokenRegex)).Handler(DensityMapHandler(ctx)).Methods("GET")
	router.Path(fmt.Sprintf("/job/{id:%s}/fsc.png", TokenRegex)).Handler(FSCChartHandler(ctx)).Methods("GET")
	router.Path(fmt.Sprintf("/job/{id:%s}/denss-{jid:[0-9]+}-output.zip", TokenRegex)).Handler(RawDataHandler(ctx)).Methods("GET")
	router.Path("/").Handler(IndexHandler(ctx)).Methods("GET")

	n := negroni.New(negroni.NewRecovery())
	n.UseHandler(router)

	return n
}

func RunServer(ctx *app.AppContext) {
	mw := middleware(ctx)

	log.Printf("Running on http://%s:%d", viper.GetString("bind"), viper.GetInt("port"))

	certFile := viper.GetString("cert")
	keyFile := viper.GetString("key")

	if certFile != "" && keyFile != "" {
		log.Warn("SSL/TLS enabled. HTTP communication will be encrypted")
		http.ListenAndServeTLS(fmt.Sprintf("%s:%d", viper.GetString("bind"), viper.GetInt("port")), certFile, keyFile, mw)
	} else {
		http.ListenAndServe(fmt.Sprintf("%s:%d", viper.GetString("bind"), viper.GetInt("port")), mw)
	}
}
