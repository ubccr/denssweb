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
	"github.com/urfave/negroni"
)

func init() {
	viper.SetDefault("port", 8080)
	viper.SetDefault("bind", "")
	viper.SetDefault("driver", "mysql")
	viper.SetDefault("dsn", "/denssweb?parseTime=true")
}

func middleware(ctx *AppContext) *negroni.Negroni {
	router := mux.NewRouter()

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.RenderNotFound(w)
	})

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(fmt.Sprintf("%s/static", ctx.Tmpldir)))))
	router.Path("/about").Handler(AboutHandler(ctx)).Methods("GET")
	router.Path("/").Handler(IndexHandler(ctx)).Methods("GET")

	n := negroni.New(negroni.NewRecovery())
	n.UseHandler(router)

	return n
}

func RunServer() {
	ctx, err := NewAppContext()
	if err != nil {
		log.Fatal(err.Error())
	}

	mw := middleware(ctx)

	log.Printf("Running on http://%s:%d", viper.GetString("bind"), viper.GetInt("port"))
	log.Printf("IPA server: %s", viper.GetString("ipahost"))

	certFile := viper.GetString("cert")
	keyFile := viper.GetString("key")

	if certFile != "" && keyFile != "" {
		http.ListenAndServeTLS(fmt.Sprintf("%s:%d", viper.GetString("bind"), viper.GetInt("port")), certFile, keyFile, mw)
	} else {
		log.Warn("**WARNING*** SSL/TLS not enabled. HTTP communication will not be encrypted and vulnerable to snooping.")
		http.ListenAndServe(fmt.Sprintf("%s:%d", viper.GetString("bind"), viper.GetInt("port")), mw)
	}
}
