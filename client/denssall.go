// Copyright 2020 DENSSWeb Authors. All rights reserved.
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
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ubccr/denssweb/model"
)

// Run denss.all.py
func runDenssAll(log *logrus.Logger, job *model.Job, workDir string, threads int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(viper.GetInt64("max_seconds"))*time.Second)
	defer cancel()

	inputFile := fmt.Sprintf("input.%s", job.FileType)
	err := ioutil.WriteFile(filepath.Join(workDir, inputFile), job.InputData, 0640)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    job.ID,
		}).Error("Failed to write input data file")
		return err
	}

	outputPrefix := fmt.Sprintf("output_%d", job.ID)
	args := []string{
		"-f",
		inputFile,
		"-o",
		outputPrefix,
		"-j",
		fmt.Sprintf("%d", threads),
		"--plot_off",
		"--quiet",
		"--mode",
		strings.ToUpper(job.Mode),
	}

	if viper.GetBool("enable_gpu") {
		args = append(args, "--gpu")
	}

	if job.Units != "" {
		args = append(args, "--units")
		args = append(args, job.Units)
	}

	if !job.Enantiomer {
		args = append(args, "-en_off")
	}

	if job.Dmax > 0 {
		args = append(args, "-d")
		args = append(args, fmt.Sprintf("%.4f", job.Dmax))
	}
	if job.Electrons > 0 {
		args = append(args, "--ne")
		args = append(args, fmt.Sprintf("%d", job.Electrons))
	}
	if job.Symmetry > 0 {
		args = append(args, "-ncs")
		args = append(args, fmt.Sprintf("%d", job.Symmetry))
		if job.SymmetryAxis > 0 {
			args = append(args, "-ncs_axis")
			args = append(args, fmt.Sprintf("%d", job.SymmetryAxis))
		}
		if job.SymmetrySteps != "" {
			args = append(args, "-ncs_steps")
			args = append(args, job.SymmetrySteps)
		}
	}

	log.WithFields(logrus.Fields{
		"id":      job.ID,
		"threads": threads,
	}).Info("Running denss.all.py")

	cmd := exec.CommandContext(ctx, viper.GetString("denssall_path"), args...)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(logrus.Fields{
			"error":  err.Error(),
			"id":     job.ID,
			"output": string(out),
		}).Error("Failed to run denss.all.py job")
		return err
	}

	log.WithFields(logrus.Fields{
		"id":      job.ID,
		"threads": threads,
	}).Info("denss.all.py completed successfully")

	return nil
}
