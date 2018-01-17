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
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ubccr/denssweb/model"
)

// Exec single denss.py process
func execDenss(log *logrus.Logger, job *model.Job, workDir, inputFile string, thread int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(viper.GetInt64("max_seconds"))*time.Second)
	defer cancel()

	outputPrefix := filepath.Join(workDir, fmt.Sprintf("output_%d", thread))
	args := []string{
		"-f",
		inputFile,
		"--oversampling",
		fmt.Sprintf("%.4f", job.Oversampling),
		"-o",
		outputPrefix,
		"--plot_off",
	}

	if job.Dmax > 0 {
		args = append(args, "-d")
		args = append(args, fmt.Sprintf("%.4f", job.Dmax))
	}
	if job.VoxelSize > 0 {
		args = append(args, "--voxel")
		args = append(args, fmt.Sprintf("%.4f", job.VoxelSize))
	}

	log.WithFields(logrus.Fields{
		"id":     job.ID,
		"thread": thread,
	}).Info("Running denss")

	cmd := exec.CommandContext(ctx, viper.GetString("denss_path"), args...)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(logrus.Fields{
			"error":  err.Error(),
			"id":     job.ID,
			"output": string(out),
		}).Error("Failed to run denss job")
		return err
	}

	log.WithFields(logrus.Fields{
		"id":     job.ID,
		"thread": thread,
	}).Info("denss completed successfully")

	return nil
}

// Run denss.py in parallel
func runDenss(log *logrus.Logger, job *model.Job, workDir string) error {

	inputFile := filepath.Join(workDir, fmt.Sprintf("input.%s", job.FileType))
	err := ioutil.WriteFile(inputFile, job.InputData, 0640)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    job.ID,
		}).Error("Failed to write input data file")
		return err
	}

	var wg sync.WaitGroup
	errChannel := make(chan error, 1)

	maxRuns := int(job.MaxRuns)

	wg.Add(maxRuns)
	finished := make(chan bool, 1)

	log.WithFields(logrus.Fields{
		"id":      job.ID,
		"maxRuns": job.MaxRuns,
	}).Info("Spawning denss.py runs")

	for i := 0; i < maxRuns; i++ {
		go func(thread int) {
			err = execDenss(log, job, workDir, inputFile, thread)
			if err != nil {
				errChannel <- err
			}

			wg.Done()
		}(i)
	}

	go func() {
		wg.Wait()
		close(finished)
	}()

	select {
	case <-finished:
	case err := <-errChannel:
		if err != nil {
			return err
		}
	}

	log.WithFields(logrus.Fields{
		"id":      job.ID,
		"maxRuns": job.MaxRuns,
	}).Info("denss.py runs completed successfully")

	return nil
}
