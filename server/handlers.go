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
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/ubccr/denssweb/model"
)

const (
	MaxFileSize = 1 << (10 * 2) // 1MB
)

func IndexHandler(ctx *AppContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.RenderTemplate(w, "index.html", nil)
	})
}

func AboutHandler(ctx *AppContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.RenderTemplate(w, "about.html", nil)
	})
}

func JobHandler(ctx *AppContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(mux.Vars(r)["id"])

		job, err := model.FetchJob(ctx.DB, int64(id))
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
				"id":    id,
			}).Error("Failed to fetch job from database")

			if err == sql.ErrNoRows {
				ctx.RenderNotFound(w)
			} else {
				ctx.RenderError(w, http.StatusInternalServerError)
			}

			return
		}
		vars := map[string]interface{}{
			"job": job}
		ctx.RenderTemplate(w, "job.html", vars)
	})
}

func SubmitHandler(ctx *AppContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		message := ""

		if r.Method == "POST" {
			r.Body = http.MaxBytesReader(w, r.Body, MaxFileSize)
			err := r.ParseMultipartForm(4096)
			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("Failed to parse multipart form")
				ctx.RenderError(w, http.StatusInternalServerError)
				return
			}
			files := r.MultipartForm.File["inputFile"]
			inputData := []byte("")
			if len(files) > 0 {
				// Only use first file
				file, err := files[0].Open()
				defer file.Close()
				if err != nil {
					log.WithFields(log.Fields{
						"err": err,
					}).Error("Failed to open input data file")
					ctx.RenderError(w, http.StatusInternalServerError)
					return
				}
				inputData, err = ioutil.ReadAll(file)
				if err != nil {
					log.WithFields(log.Fields{
						"err": err,
					}).Error("Failed to read input data file")
					ctx.RenderError(w, http.StatusInternalServerError)
					return
				}
			}

			job, err := submitJob(ctx, inputData, r.FormValue("dmax"))

			if err == nil {
				http.Redirect(w, r, fmt.Sprintf("/job/%d", job.ID), 302)
				return
			}

			message = err.Error()
		}

		vars := map[string]interface{}{
			"message": message}

		ctx.RenderTemplate(w, "submit.html", vars)
	})
}

func submitJob(ctx *AppContext, data []byte, dmax string) (*model.Job, error) {
	if len(data) == 0 {
		return nil, errors.New("Please provide an input data file")
	}

	d, err := strconv.Atoi(dmax)
	if err != nil {
		return nil, errors.New("Please an integer for the maximum particle dimension")
	}

	job := &model.Job{InputData: data, Dmax: d}

	err = model.QueueJob(ctx.DB, job)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Failed to queue job")
		return nil, errors.New("Failed to submit job. Please contact system administrator")
	}

	return job, nil
}
