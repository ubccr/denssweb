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
	"io/ioutil"
	"os/exec"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ubccr/denssweb/model"
)

// Create summary chart
func plotSummary(log *logrus.Logger, job *model.Job, workDir string) error {
	summaryPNG := filepath.Join(workDir, "summary.png")

	log.WithFields(logrus.Fields{
		"id": job.ID,
	}).Info("Plotting summary chart")

	args := []string{
		"--input",
		workDir,
		"--output",
		summaryPNG,
	}

	cmd := exec.Command(viper.GetString("summary_path"), args...)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(logrus.Fields{
			"error":   err.Error(),
			"id":      job.ID,
			"workDir": workDir,
			"png":     summaryPNG,
			"output":  string(out),
		}).Error("Failed to create Summary chart PNG")
		return err
	}

	job.SummaryChart, err = ioutil.ReadFile(summaryPNG)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error":   err.Error(),
			"id":      job.ID,
			"workDir": workDir,
			"png":     summaryPNG,
			"output":  string(out),
		}).Error("Failed to read Summary chart PNG file")
		return err
	}

	log.WithFields(logrus.Fields{
		"id":  job.ID,
		"png": summaryPNG,
	}).Info("Successfully created Summary chart")

	return nil
}
