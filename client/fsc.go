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
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ubccr/denssweb/model"
)

// Create Fourier Shell Correlation (FSC) curve
func plotFSC(log *logrus.Logger, job *model.Job, workDir string) error {
	outputPrefix := fmt.Sprintf("output_%d", job.ID)
    fscData := filepath.Join(workDir, outputPrefix, "spt_avg_01", "fsc_0.txt")
	fscPNG := filepath.Join(workDir, "fsc.png")

	log.WithFields(logrus.Fields{
		"id":   job.ID,
		"data": fscData,
	}).Info("Plotting fsc curve")

	args := []string{
		"--input",
		fscData,
		"--output",
		fscPNG,
	}

	cmd := exec.Command(viper.GetString("fsc_path"), args...)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(logrus.Fields{
			"error":  err.Error(),
			"id":     job.ID,
			"data":   fscData,
			"png":    fscPNG,
			"output": string(out),
		}).Error("Failed to create FSC curve PNG")
		return err
	}

	job.FSCChart, err = ioutil.ReadFile(fscPNG)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error":  err.Error(),
			"id":     job.ID,
			"data":   fscData,
			"png":    fscPNG,
			"output": string(out),
		}).Error("Failed to read fsc curve PNG file")
		return err
	}

	log.WithFields(logrus.Fields{
		"id":   job.ID,
		"data": fscData,
		"png":  fscPNG,
	}).Info("Successfully created FSC curve")

	return nil
}
