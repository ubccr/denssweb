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

package app

import (
	"bytes"
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ubccr/denssweb/model"
)

func init() {
	viper.SetDefault("driver", "sqlite3")
	viper.SetDefault("dsn", "/tmp/denssweb.db?_busy_timeout=5000&cache=shared")
}

type AppContext struct {
	DB        *sqlx.DB
	Tmpldir   string
	dsn       string
	templates map[string]*template.Template
}

func NewAppContext() (*AppContext, error) {
	db, err := model.NewDB(viper.GetString("driver"), viper.GetString("dsn"))
	if err != nil {
		return nil, err
	}

	tmpldir := viper.GetString("templates")
	if len(tmpldir) == 0 {
		// default to directory of current executable
		ex, err := os.Executable()
		if err != nil {
			log.Fatal(err)
		}

		path, err := filepath.EvalSymlinks(ex)
		if err != nil {
			log.Fatal(err)
		}

		dir, err := filepath.Abs(filepath.Dir(path))
		if err != nil {
			log.Fatal(err)
		}

		tmpldir = filepath.Join(dir, "dist", "templates")
		if _, err := os.Stat(tmpldir); err != nil {
			tmpldir = filepath.Join(dir, "templates")
			if _, err := os.Stat(tmpldir); err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Warn("failed to find template directory")
			}
		}
	}

	log.Printf("Using template dir: %s", tmpldir)

	tmpls, err := filepath.Glob(tmpldir + "/*.html")
	if err != nil {
		log.Fatal(err)
	}

	templates := make(map[string]*template.Template)
	for _, t := range tmpls {
		base := filepath.Base(t)
		if base != "layout.html" {
			templates[base] = template.Must(template.New("layout").ParseFiles(t,
				tmpldir+"/layout.html"))
		}
	}

	app := &AppContext{}
	app.Tmpldir = tmpldir
	app.DB = db
	app.templates = templates

	return app, nil
}

// Render 404 template
func (app *AppContext) RenderNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)

	app.RenderTemplate(w, "404.html", nil)
}

// Render template t using template parameters in data.
func (app *AppContext) RenderTemplate(w http.ResponseWriter, name string, data interface{}) {
	t := app.templates[name]

	var buf bytes.Buffer
	err := t.ExecuteTemplate(&buf, "layout", data)

	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to render template")
		http.Error(w, "Fatal error rendering template", http.StatusInternalServerError)
		return
	}

	buf.WriteTo(w)
}

// Render error template and write HTTP status
func (app *AppContext) RenderError(w http.ResponseWriter, status int) {
	w.WriteHeader(status)

	app.RenderTemplate(w, "error.html", nil)
}
