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
	"regexp"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	valid "github.com/asaskevich/govalidator"
	"github.com/dchest/captcha"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/spf13/viper"
	"github.com/ubccr/denssweb/app"
	"github.com/ubccr/denssweb/model"
)

const (
	MaxFileSize = 1 << (10 * 2) // 1MB
)

var (
	JobNameRegexp = regexp.MustCompile(`^[A-Za-z0-9\-\_]+$`)
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

func TutorialHandler(ctx *app.AppContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.RenderTemplate(w, "tutorial.html", nil)
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
			"emailEnabled": viper.GetBool("enable_notifications"),
			"message":      message,
		}

		if viper.GetBool("enable_captcha") {
			vars["captchaID"] = captcha.New()
		}

		ctx.RenderTemplate(w, "submit.html", vars)
	})
}

func submitJob(ctx *app.AppContext, data []byte, r *http.Request) (*model.Job, error) {
	if len(data) == 0 {
		return nil, errors.New("Please provide an input data file")
	}

	dmax, err := parseFloat(r.FormValue("dmax"), "Dmax")
	if err != nil {
		return nil, err
	}

	fileType := "out"
	if version, err := parseGNOMHeader(data); err == nil {
		log.WithFields(log.Fields{
			"version": version,
		}).Info("Input data appears to be GNOM")
	} else {
		// Check 3-column DAT file
		err := validateDAT(data)
		if err != nil {
			return nil, err
		}
		log.Info("Input data appears to be 3-column DAT file")

		// DAT files require Dmax
		if dmax == 0.0 {
			return nil, errors.New("Please provide a float for the maximum particle dimension")
		}

		fileType = "dat"
	}

	captchaID := r.FormValue("captcha_id")
	captchaSol := r.FormValue("captcha_sol")
	if viper.GetBool("enable_captcha") {
		if len(captchaID) == 0 {
			return nil, errors.New("Invalid captcha provided")
		}
		if len(captchaSol) == 0 {
			return nil, errors.New("Please type in the numbers you see in the picture")
		}

		if !captcha.VerifyString(captchaID, captchaSol) {
			return nil, errors.New("The numbers you typed in do not match the image")
		}
	}

	job := &model.Job{InputData: data, FileType: fileType}

	err = ctx.Decoder.Decode(job, r.PostForm)
	if err != nil {
		switch serr := err.(type) {
		case schema.ConversionError:
			return nil, errors.New(fmt.Sprintf("Invalid data for %s", serr.Key))
		case schema.MultiError:
			msg := make([]string, 0)
			for k, _ := range serr {
				msg = append(msg, fmt.Sprintf("Invalid data for %s", k))
			}
			return nil, errors.New(strings.Join(msg, ";"))
		default:
			log.WithFields(log.Fields{
				"err": err,
			}).Error("Failed to decode form input")
			return nil, errors.New("The input data you provided is invalid")
		}
	}

	if len(job.Name) == 0 {
		return nil, errors.New("Job name is required")
	}

	if len(job.Name) > 255 {
		return nil, errors.New("Job name must be less than 255 characters")
	}

	if !JobNameRegexp.MatchString(job.Name) {
		return nil, errors.New("Job name must be alphanumeric")
	}

	if len(job.Email) > 0 && !valid.IsEmail(job.Email) {
		return nil, errors.New("Please provide a valid email address")
	}

	// Validate parameters to sane default ranges
	// *Note* range validator tag is not currently working.
	// See: https://github.com/asaskevich/govalidator/issues/223
	if job.Dmax > 0 && !valid.InRangeFloat64(job.Dmax, 10, 1000) {
		return nil, errors.New("Dmax should be between 10 and 1000")
	}
	if job.MaxSteps > 0 && !valid.InRangeInt(job.MaxSteps, 100, 10000) {
		return nil, errors.New("Max Steps should be between 10 and 10000")
	}
	if job.NumSamples > 0 && !valid.InRangeInt(job.NumSamples, 2, 500) {
		return nil, errors.New("Num Samples should be between 2 and 500")
	}
	if job.Oversampling > 0 && !valid.InRangeFloat64(job.Oversampling, 2, 50) {
		return nil, errors.New("Oversampling should be between 2 and 50")
	}
	if job.MaxRuns > 0 && !valid.InRangeInt(job.MaxRuns, 2, 1000) {
		return nil, errors.New("Max Runs should be between 2 and 1000")
	}
	if job.VoxelSize > 0 && !valid.InRangeFloat64(job.VoxelSize, 1, 100) {
		return nil, errors.New("Voxel Size should be between 1 and 100")
	}
	if job.Electrons > 0 && !valid.InRangeInt(job.Electrons, 1, 100000000) {
		return nil, errors.New("Electrons should be between 1 and 1e8")
	}

	if viper.GetBool("restrict_params") {
		// Force setting default parameters
		job.MaxSteps = 3000
		job.MaxRuns = 20
		job.NumSamples = 32
		job.VoxelSize = 0
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
		"FileType":     job.FileType,
		"Dmax":         job.Dmax,
		"NumSamples":   job.NumSamples,
		"Oversampling": job.Oversampling,
		"VoxelSize":    job.VoxelSize,
		"Electrons":    job.Electrons,
		"MaxSteps":     job.MaxSteps,
		"MaxRuns":      job.MaxRuns,
	}).Info("Job queued successfully")

	if len(job.Email) > 0 {
		err = ctx.SendEmail(job.Email, "SUBMITTED", job.URL(), job.ID)
		if err != nil {
			log.WithFields(log.Fields{
				"job_id": job.ID,
				"email":  job.Email,
				"url":    job.URL(),
				"status": "SUBMITTED",
				"error":  err,
			}).Error("Failed to send email")
		}
	}

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
