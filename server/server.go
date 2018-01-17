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
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/dchest/captcha"
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
	viper.SetDefault("show_job_list", true)
	viper.SetDefault("enable_captcha", false)
}

func middleware(ctx *app.AppContext) *negroni.Negroni {
	router := mux.NewRouter()

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.RenderNotFound(w)
	})

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(fmt.Sprintf("%s/static", ctx.Tmpldir)))))
	router.Path("/about").Handler(AboutHandler(ctx)).Methods("GET")

	if viper.GetBool("show_job_list") {
		router.Path("/jobs").Handler(JobListHandler(ctx)).Methods("GET")
	}
	if viper.GetBool("enable_captcha") {
		router.Path(fmt.Sprintf("/captcha/{cid:%s}.png", TokenRegex)).Handler(captcha.Server(captcha.StdWidth, captcha.StdHeight))
	}

	router.Path("/submit").Handler(SubmitHandler(ctx)).Methods("GET", "POST")
	router.Path(fmt.Sprintf("/job/{id:%s}", TokenRegex)).Handler(JobHandler(ctx)).Methods("GET")
	router.Path(fmt.Sprintf("/job/{id:%s}/status", TokenRegex)).Handler(StatusHandler(ctx)).Methods("GET")
	router.Path(fmt.Sprintf("/job/{id:%s}/density-map.ccp4", TokenRegex)).Handler(DensityMapHandler(ctx)).Methods("GET")
	router.Path(fmt.Sprintf("/job/{id:%s}/fsc.png", TokenRegex)).Handler(FSCChartHandler(ctx)).Methods("GET")
	router.Path(fmt.Sprintf("/job/{id:%s}/summary.png", TokenRegex)).Handler(SummaryChartHandler(ctx)).Methods("GET")
	router.Path(fmt.Sprintf("/job/{id:%s}/denss-{jid:[0-9]+}-output.zip", TokenRegex)).Handler(RawDataHandler(ctx)).Methods("GET")
	router.Path("/").Handler(IndexHandler(ctx)).Methods("GET")

	n := negroni.New(negroni.NewRecovery())
	n.UseHandler(router)

	return n
}

func RunServer(ctx *app.AppContext) {
	mw := middleware(ctx)

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Addr:         fmt.Sprintf("%s:%d", viper.GetString("bind"), viper.GetInt("port")),
		Handler:      mw,
	}

	certFile := viper.GetString("cert")
	keyFile := viper.GetString("key")

	if certFile != "" && keyFile != "" {
		cfg := &tls.Config{
			MinVersion: tls.VersionTLS12,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		}

		srv.TLSConfig = cfg

		log.Printf("Running on https://%s:%d", viper.GetString("bind"), viper.GetInt("port"))
		log.Fatal(srv.ListenAndServeTLS(certFile, keyFile))
	} else {
		log.Warn("**WARNING*** SSL/TLS not enabled. HTTP communication will not be encrypted and vulnerable to snooping.")
		log.Printf("Running on http://%s:%d", viper.GetString("bind"), viper.GetInt("port"))
		log.Fatal(srv.ListenAndServe())
	}
}
