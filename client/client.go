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

package client

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ubccr/denssweb/app"
	"github.com/ubccr/denssweb/model"
)

func init() {
	// Try and set sensible defaults here
	wd, err := os.Getwd()
	if err != nil {
		wd = os.TempDir()
	}

	viper.SetDefault("work_dir", filepath.Join(wd, "denssweb-work"))
	viper.SetDefault("denss_path", "/usr/local/bin/denss.py")
	viper.SetDefault("eman2dir", filepath.Join(os.Getenv("HOME"), "EMAN2"))
	viper.SetDefault("fsc_path", filepath.Join(wd, "scripts", "denssweb-fsc-chart.py"))
	viper.SetDefault("summary_path", filepath.Join(wd, "scripts", "denssweb-summary-chart.py"))
	// Defaults to 10 minutes
	viper.SetDefault("max_seconds", 3600)
}

func processJob(ctx *app.AppContext, job *model.Job, threads int) error {
	// TODO make the percent complete more accurate

	os.Setenv("LD_LIBRARY_PATH", filepath.Join(viper.GetString("eman2dir"), "lib"))
	os.Setenv("PYTHONPATH", filepath.Join(viper.GetString("eman2dir"), "lib"))

	logrus.WithFields(logrus.Fields{
		"id": job.ID,
	}).Info("Creating job directory")

	model.LogJobMessage(ctx.DB, job, "Setup", "Creating job directory", 0)
	workDir := filepath.Join(viper.GetString("work_dir"), fmt.Sprintf("denss%d-%s", job.ID, job.Name))
	os.RemoveAll(workDir)
	err := os.MkdirAll(workDir, 0700)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    job.ID,
		}).Error("Failed to create working directory")
		model.LogJobMessage(ctx.DB, job, "Setup Failed", "Failed to create job directory", 0)
		return err
	}

	logrus.WithFields(logrus.Fields{
		"id": job.ID,
	}).Info("Creating log file")

	model.LogJobMessage(ctx.DB, job, "Setup", "Creating log file", 5)

	log := logrus.New()

	logPath := filepath.Join(workDir, fmt.Sprintf("denss-%d.log", job.ID))
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    job.ID,
		}).Error("Failed to create job log file")
		model.LogJobMessage(ctx.DB, job, "Setup Failed", "Failed to create job log file", 0)
		return err
	}
	defer logFile.Close()

	log.Out = logFile

	logrus.WithFields(logrus.Fields{
		"id": job.ID,
	}).Info("Running DENSS All")

	model.LogJobMessage(ctx.DB, job, "Run DENSS All", "Performing parallel DENSS runs", 25)
	err = runDenssAll(log, job, workDir, threads)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    job.ID,
		}).Error("Failed to run denss.all.py")
		model.LogJobMessage(ctx.DB, job, "Run DENSS All Failed", "Failed to run DENSS All", 0)
		return err
	}

	logrus.WithFields(logrus.Fields{
		"id": job.ID,
	}).Info("Saving MRC file")

	job.DensityMap, err = ioutil.ReadFile(filepath.Join(workDir, fmt.Sprintf("output_%d", job.ID), fmt.Sprintf("output_%d_avg.mrc", job.ID)))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    job.ID,
		}).Error("Failed to read MRC file")
		model.LogJobMessage(ctx.DB, job, "Reading MRC file failed", "Failed to read MRC file", 0)
		return err
	}

	logrus.WithFields(logrus.Fields{
		"id": job.ID,
	}).Info("Creating FSC Curve")
	model.LogJobMessage(ctx.DB, job, "FSC Curve", "Plotting FSC Cruve", 85)
	err = plotFSC(log, job, workDir)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    job.ID,
		}).Error("Failed to plot FSC curve")
		model.LogJobMessage(ctx.DB, job, "FSC Curve Failed", "Failed to plot FSC curve", 0)
		return err
	}

	logrus.WithFields(logrus.Fields{
		"id": job.ID,
	}).Info("Creating Summary Chart")
	model.LogJobMessage(ctx.DB, job, "Summary Chart", "Plotting Summary Stats", 90)
	err = plotSummary(log, job, workDir)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    job.ID,
		}).Error("Failed to plot Summary chart")
		model.LogJobMessage(ctx.DB, job, "Summary Chart Failed", "Failed to plot summary stats", 0)
		return err
	}

	logrus.WithFields(logrus.Fields{
		"id": job.ID,
	}).Info("Creating zip archive")
	model.LogJobMessage(ctx.DB, job, "Creating ZIP", "Building zip archive of raw data", 95)
	err = createZIP(job, workDir)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    job.ID,
		}).Error("Failed to create zip archive")
		model.LogJobMessage(ctx.DB, job, "Create ZIP Failed", "Failed to create zip archive", 0)
		return err
	}

	return nil
}

func RunClient(ctx *app.AppContext, maxThreads int) {
	logrus.Info("--------------------------------------------")
	logrus.Info("Client config")
	logrus.Info("--------------------------------------------")
	logrus.Infof("Path to denss.py: %s", viper.GetString("denss_path"))
	logrus.Infof("Path to EMAN2: %s", viper.GetString("eman2dir"))
	logrus.Infof("Path to denssweb-fsc-chart.py: %s", viper.GetString("fsc_path"))
	logrus.Infof("Path to denss-summary-chart.py: %s", viper.GetString("summary_path"))
	logrus.Infof("Max number of seconds: %d", viper.GetInt("max_seconds"))
	logrus.Infof("Job Work directory: %s", viper.GetString("work_dir"))
	logrus.Infof("Max threads: %d", maxThreads)
	logrus.Info("--------------------------------------------")
	runtime.GOMAXPROCS(maxThreads)

	for {
		time.Sleep(3 * time.Second)

		job, err := model.FetchNextPending(ctx.DB)
		if err != nil {
			if err != sql.ErrNoRows {
				logrus.WithFields(logrus.Fields{
					"error": err.Error(),
				}).Error("Failed to fetch pending job")
			}
			continue
		}

		logrus.WithFields(logrus.Fields{
			"id":  job.ID,
			"url": job.URL(),
		}).Info("Processing new job")

		workDir := filepath.Join(viper.GetString("work_dir"), fmt.Sprintf("denss%d-%s", job.ID, job.Name))

		err = processJob(ctx, job, maxThreads)
		if err != nil {
			// create zip of logs if job failed
			logrus.WithFields(logrus.Fields{
				"id": job.ID,
			}).Info("Creating zip archive for failed job")
			zerr := createZIP(job, workDir)
			if zerr != nil {
				logrus.WithFields(logrus.Fields{
					"error": zerr.Error(),
					"id":    job.ID,
				}).Error("Failed to create zip archive for failed job")
			}

			cerr := model.CompleteJob(ctx.DB, job, model.StatusError)
			if cerr != nil {
				logrus.WithFields(logrus.Fields{
					"error": cerr.Error(),
					"url":   job.URL(),
					"id":    job.ID,
				}).Error("Failed save failed job to database")
			}

			logrus.WithFields(logrus.Fields{
				"error": err.Error(),
				"url":   job.URL(),
				"id":    job.ID,
			}).Error("Failed to process job")

			err = os.RemoveAll(workDir)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error":   err.Error(),
					"url":     job.URL(),
					"id":      job.ID,
					"workDir": workDir,
				}).Error("Failed to clean up work dir")
			}

			if len(job.Email) > 0 {
				err = ctx.SendEmail(job.Email, "FAILED", job.URL(), job.ID)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"job_id": job.ID,
						"email":  job.Email,
						"url":    job.URL(),
						"status": "FAILED",
						"error":  err,
					}).Error("Failed to send email")
				}
			}
			continue
		}

		if len(job.Email) > 0 {
			err = ctx.SendEmail(job.Email, "COMPLETED", job.URL(), job.ID)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"job_id": job.ID,
					"email":  job.Email,
					"url":    job.URL(),
					"status": "COMPLETED",
					"error":  err,
				}).Error("Failed to send email")
			}
		}

		model.LogJobMessage(ctx.DB, job, "Complete", "Job completed successfully", 100)
		err = model.CompleteJob(ctx.DB, job, model.StatusComplete)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err.Error(),
				"url":   job.URL(),
				"id":    job.ID,
			}).Error("Failed to save completed job")
			continue
		}

		logrus.WithFields(logrus.Fields{
			"id":  job.ID,
			"url": job.URL(),
		}).Info("Job processed succesfully")

		err = os.RemoveAll(workDir)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error":   err.Error(),
				"url":     job.URL(),
				"id":      job.ID,
				"workDir": workDir,
			}).Error("Failed to clean up work dir")
		}
	}
}
