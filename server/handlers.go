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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/ubccr/denssweb/app"
	"github.com/ubccr/denssweb/model"
)

const (
	MaxFileSize = 1 << (10 * 2) // 1MB
)

func IndexHandler(ctx *app.AppContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.RenderTemplate(w, "index.html", nil)
	})
}

func AboutHandler(ctx *app.AppContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.RenderTemplate(w, "about.html", nil)
	})
}

func JobListHandler(ctx *app.AppContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status, _ := strconv.Atoi(r.FormValue("status"))
		offset, _ := strconv.Atoi(r.FormValue("offset"))
		if offset <= 0 {
			offset = 0
		}

		prev := offset - 20
		if prev <= 0 {
			prev = 0
		}
		next := offset + 20

		jobs, err := model.FetchAllJobs(ctx.DB, status, 20, offset)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("Failed to fetch jobs from db")
			ctx.RenderError(w, http.StatusInternalServerError)
			return
		}

		vars := map[string]interface{}{
			"offset": offset,
			"prev":   prev,
			"next":   next,
			"status": status,
			"jobs":   jobs}
		ctx.RenderTemplate(w, "job-list.html", vars)
	})
}

func JobHandler(ctx *app.AppContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		job, err := model.FetchJob(ctx.DB, id)
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

func SubmitHandler(ctx *app.AppContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		message := ""

		if r.Method == "POST" {
			r.Body = http.MaxBytesReader(w, r.Body, MaxFileSize)
			err := r.ParseMultipartForm(MaxFileSize)
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

			job, err := submitJob(ctx, inputData, r)

			if err == nil {
				http.Redirect(w, r, job.URL(), 302)
				return
			}

			message = err.Error()
		}

		vars := map[string]interface{}{
			"message": message}

		ctx.RenderTemplate(w, "submit.html", vars)
	})
}

func submitJob(ctx *app.AppContext, data []byte, r *http.Request) (*model.Job, error) {
	if len(data) == 0 {
		return nil, errors.New("Please provide an input data file")
	}

	dmax := float64(0)

	if version, err := parseGNOMHeader(data); err == nil {
		// Convert GNOM to DAT
		dat, dm, err := convertGNOM(data, version)
		if err != nil {
			return nil, err
		}
		data = dat
		dmax = dm
	} else {
		// Check 3-column DAT file
		err := validateDAT(data)
		if err != nil {
			return nil, err
		}
	}

	if dmax == 0 {
		var err error
		dmax, err = strconv.ParseFloat(r.FormValue("dmax"), 64)
		if err != nil {
			return nil, errors.New("Please provide a float for the maximum particle dimension")
		}
	}

	name := r.FormValue("name")
	if len(name) > 255 {
		return nil, errors.New("Job name must be less than 255 characters")
	}

	job := &model.Job{InputData: data, Dmax: dmax, Name: name}

	// Set optional parameters
	var err error
	job.NumSamples, err = parseInt(r.FormValue("num_samples"), "Samples")
	if err != nil {
		return nil, err
	}
	job.Oversampling, err = parseFloat(r.FormValue("oversampling"), "Oversampling")
	if err != nil {
		return nil, err
	}
	job.VoxelSize, err = parseFloat(r.FormValue("voxel_size"), "Voxel Size")
	if err != nil {
		return nil, err
	}
	job.Electrons, err = parseInt(r.FormValue("electrons"), "Electrons")
	if err != nil {
		return nil, err
	}
	job.MaxSteps, err = parseInt(r.FormValue("max_steps"), "Max Steps")
	if err != nil {
		return nil, err
	}
	job.MaxRuns, err = parseInt(r.FormValue("max_runs"), "Max Runs")
	if err != nil {
		return nil, err
	}

	err = model.QueueJob(ctx.DB, job)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Failed to queue job")
		return nil, errors.New("Failed to submit job. Please contact system administrator")
	}

	log.WithFields(log.Fields{
		"ID":           job.ID,
		"URL":          job.URL(),
		"Dmax":         job.Dmax,
		"NumSamples":   job.NumSamples,
		"Oversampling": job.Oversampling,
		"VoxelSize":    job.VoxelSize,
		"Electrons":    job.Electrons,
		"MaxSteps":     job.MaxSteps,
		"MaxRuns":      job.MaxRuns,
	}).Info("Job queued successfully")

	return job, nil
}

func parseFloat(n, label string) (float64, error) {
	if n == "" {
		return float64(0.0), nil
	}

	f, err := strconv.ParseFloat(n, 64)
	if err != nil {
		return float64(0.0), fmt.Errorf("Please provide a float for %s", label)
	}

	return f, nil
}

func parseInt(n, label string) (int64, error) {
	if n == "" {
		return int64(0), nil
	}

	i, err := strconv.Atoi(n)
	if err != nil {
		return int64(0), fmt.Errorf("Please provide an integer for %s", label)
	}

	return int64(i), nil
}

func DensityMapHandler(ctx *app.AppContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		job, err := model.FetchDensityMap(ctx.DB, id)
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

		if job.DensityMap != nil {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(job.DensityMap)
			return
		}

		ctx.RenderNotFound(w)
	})
}

func FSCChartHandler(ctx *app.AppContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		job, err := model.FetchFSCChart(ctx.DB, id)
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

		if job.FSCChart != nil {
			w.Header().Set("Content-Type", "image/png")
			w.Write(job.FSCChart)
			return
		}

		ctx.RenderNotFound(w)
	})
}

func RawDataHandler(ctx *app.AppContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		job, err := model.FetchRawData(ctx.DB, id)
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

		if job.RawData != nil {
			w.Header().Set("Content-Type", "application/zip")
			w.Write(job.RawData)
			return
		}

		ctx.RenderNotFound(w)
	})
}

func StatusHandler(ctx *app.AppContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		job, err := model.FetchJob(ctx.DB, id)
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

		if job.StatusID == model.StatusPending {
			job.Time = job.WaitTime()
		} else {
			job.Time = job.RunTime()
		}

		out, err := json.Marshal(job)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
				"id":    job.ID,
			}).Error("Error encoding job as json")
			ctx.RenderError(w, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
	})
}

func SummaryChartHandler(ctx *app.AppContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		job, err := model.FetchSummaryChart(ctx.DB, id)
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

		if job.SummaryChart != nil {
			w.Header().Set("Content-Type", "image/png")
			w.Write(job.SummaryChart)
			return
		}

		ctx.RenderNotFound(w)
	})
}
