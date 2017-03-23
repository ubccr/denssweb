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
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ubccr/denssweb/app"
	"github.com/ubccr/denssweb/model"
)

func init() {
	// Try and set sensible defaults here
	wd, err := os.Getwd()
	if err != nil {
		wd = "/tmp"
	}

	viper.SetDefault("work_dir", filepath.Join(wd, "denssweb-work"))
	viper.SetDefault("denss_path", "/usr/local/bin/denss.py")
	viper.SetDefault("map2map_path", filepath.Join(os.Getenv("HOME"), "Situs_2.8", "bin", "map2map"))
	viper.SetDefault("eman2dir", filepath.Join(os.Getenv("HOME"), "EMAN2"))
	viper.SetDefault("fsc_path", filepath.Join(wd, "scripts", "fsc-chart.py"))
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
	workDir := filepath.Join(viper.GetString("work_dir"), fmt.Sprintf("denss-%d", job.ID))
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
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0600)
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
	}).Info("Running DENSS")

	model.LogJobMessage(ctx.DB, job, "Run DENSS", fmt.Sprintf("Performing %d parallel DENSS runs", job.MaxRuns), 25)
	err = runDenss(log, job, workDir)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    job.ID,
		}).Error("Failed to run denss")
		model.LogJobMessage(ctx.DB, job, "Run DENSS Failed", "Failed to run DENSS", 0)
		return err
	}

	logrus.WithFields(logrus.Fields{
		"id": job.ID,
	}).Info("Building HDF stack")

	model.LogJobMessage(ctx.DB, job, "Process Output", "Combine DENSS output files into hdf", 50)
	err = buildStack(log, job, workDir)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    job.ID,
		}).Error("Failed to build stack hdf")
		model.LogJobMessage(ctx.DB, job, "Process Output Failed", "Failed to build hdf stack", 0)
		return err
	}

	logrus.WithFields(logrus.Fields{
		"id": job.ID,
	}).Info("Running Averaging")

	model.LogJobMessage(ctx.DB, job, "Run Averaging", "Run parallel averaging using EMAN2", 75)
	err = runAveraging(log, job, workDir, threads)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    job.ID,
		}).Error("Failed to run averaging")
		model.LogJobMessage(ctx.DB, job, "Run Averaging Failed", "Failed to run averaging using EMAN2", 0)
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
	}).Info("Creating zip archive")
	model.LogJobMessage(ctx.DB, job, "Creating ZIP", "Building zip archive of raw data", 95)
	err = createZIP(log, job, workDir)
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
	logrus.Infof("Path to map2map: %s", viper.GetString("map2map_path"))
	logrus.Infof("Path to EMAN2: %s", viper.GetString("eman2dir"))
	logrus.Infof("Path to fsc-chart.py: %s", viper.GetString("fsc_path"))
	logrus.Infof("Max number of seconds: %d", viper.GetInt("max_seconds"))
	logrus.Infof("Job Work directory: %s", viper.GetString("work_dir"))
	logrus.Infof("Max threads: %d", maxThreads)
	logrus.Info("--------------------------------------------")
	runtime.GOMAXPROCS(maxThreads)

	for {
		time.Sleep(3 * time.Second)

		job, err := model.FetchNextPending(ctx.DB)
		if err != nil {
			if err == sql.ErrNoRows {
				logrus.Info("No pending jobs found")
			} else {
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

		err = processJob(ctx, job, maxThreads)
		if err != nil {
			cerr := model.CompleteJob(ctx.DB, job, model.StatusError)
			if cerr != nil {
				logrus.WithFields(logrus.Fields{
					"error": cerr.Error(),
					"url":   job.URL,
					"id":    job.ID,
				}).Error("Failed save failed job to database")
			}

			logrus.WithFields(logrus.Fields{
				"error": err.Error(),
				"url":   job.URL,
				"id":    job.ID,
			}).Error("Failed to process job")
			continue
		}

		model.LogJobMessage(ctx.DB, job, "Complete", "Job completed successfully", 100)
		err = model.CompleteJob(ctx.DB, job, model.StatusComplete)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err.Error(),
				"url":   job.URL,
				"id":    job.ID,
			}).Error("Failed to save completed job")
			continue
		}

		logrus.WithFields(logrus.Fields{
			"id":  job.ID,
			"url": job.URL,
		}).Info("Job processed succesfully")
	}
}
