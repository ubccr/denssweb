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
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ubccr/denssweb/app"
	"github.com/ubccr/denssweb/model"
)

const (
	// Maxium number of seconds to run a command
	MaxSeconds = 3600
)

func init() {
	viper.SetDefault("work_dir", "/tmp")
	viper.SetDefault("denss_path", "/usr/local/bin/denss.py")
}

// Exec single denss.py process
func execDenss(job *model.Job, workDir, inputFile string, thread int) error {
	ctx, cancel := context.WithTimeout(context.Background(), MaxSeconds*time.Second)
	defer cancel()

	outputPrefix := filepath.Join(workDir, fmt.Sprintf("output_%d", thread))
	args := []string{
		"-f",
		inputFile,
		"-d",
		fmt.Sprintf("%.4f", job.Dmax),
		"--oversampling",
		fmt.Sprintf("%.4f", job.Oversampling),
		"--voxel",
		fmt.Sprintf("%.4f", job.VoxelSize),
		"-o",
		outputPrefix,
		"--plot-off",
	}
	log.Infof("Starting denss thread %d for job %d", thread, job.ID)
	cmd := exec.CommandContext(ctx, viper.GetString("denss_path"), args...)
	err := cmd.Start()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"id":    job.ID,
		}).Error("Failed to start denss job")
		return err
	}

	log.Infof("Waiting for denss thread %d for job %d to finish", thread, job.ID)
	err = cmd.Wait()
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err.Error(),
			"id":     job.ID,
			"thread": thread,
		}).Error("denss job failed")
		return err
	}

	log.Infof("Denss thread %d for job %d completed", thread, job.ID)

	return nil
}

// Run denss.py in parallel
func runDenss(job *model.Job, workDir string) error {
	inputFile := filepath.Join(workDir, "input.dat")
	err := ioutil.WriteFile(inputFile, job.InputData, 0700)
	if err != nil {
		log.WithFields(log.Fields{
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

	for i := 0; i < maxRuns; i++ {
		go func(thread int) {
			err = execDenss(job, workDir, inputFile, thread)
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

	// All denss.py process finished successfully. Need to convert xplor output
	// files to mrc using map2map
	for i := 0; i < maxRuns; i++ {
		err := convertToMRC(workDir, i)
		if err != nil {
			return err
		}
	}

	return nil
}

func convertToMRC(workDir string, num int) error {
	xplorFile := filepath.Join(workDir, fmt.Sprintf("output_%d.xplor", num))
	mrcFile := filepath.Join(workDir, fmt.Sprintf("output_%d.mrc", num))

	args := []string{
		xplorFile,
		mrcFile,
	}
	log.Infof("Converting %s to mrc", xplorFile)
	cmd := exec.Command(viper.GetString("map2map_path"), args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.WithFields(log.Fields{
			"error":     err.Error(),
			"xplorFile": xplorFile,
		}).Error("Failed stdin pipe")
		return err
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, "2\n")
	}()

	err = cmd.Run()
	if err != nil {
		log.WithFields(log.Fields{
			"error":     err.Error(),
			"xplorFile": xplorFile,
		}).Error("map2map command failed")
		return err
	}

	_, err = os.Stat(mrcFile)
	if os.IsNotExist(err) {
		log.WithFields(log.Fields{
			"error":     err.Error(),
			"xplorFile": xplorFile,
			"mrcFile":   mrcFile,
		}).Error("mrc file does not exists. map2map failed")
		return err
	} else if err != nil {
		log.WithFields(log.Fields{
			"error":     err.Error(),
			"xplorFile": xplorFile,
			"mrcFile":   mrcFile,
		}).Error("Failed to read mrc. map2map failed")
		return err
	}

	log.Infof("%s successfully converted to mrc", xplorFile)

	return nil
}

func processJob(job *model.Job) error {
	workDir := filepath.Join(viper.GetString("work_dir"), fmt.Sprintf("denss-%d", job.ID))
	err := os.MkdirAll(workDir, 0700)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"id":    job.ID,
		}).Error("Failed to create working directory")
		return err
	}

	return runDenss(job, workDir)
}

func RunClient() {
	ctx, err := app.NewAppContext()
	if err != nil {
		log.Fatal(err.Error())
	}

	job, err := model.FetchNextPending(ctx.DB)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = processJob(job)
	if err != nil {
		log.Fatal(err.Error())
	}

	/*
		job.DensityMap, err = ioutil.ReadFile("/home/ubuntu/mock-results/6lyz_averaged.ccp4")
		if err != nil {
			log.Fatal(err.Error())
		}

		job.FSCChart, err = ioutil.ReadFile("/home/ubuntu/mock-results/fsc.png")
		if err != nil {
			log.Fatal(err.Error())
		}

		job.RawData, err = ioutil.ReadFile("/home/ubuntu/mock-results/job1.zip")
		if err != nil {
			log.Fatal(err.Error())
		}

		err = model.CompleteJob(ctx.DB, job, model.StatusComplete)
		if err != nil {
			log.Fatal(err.Error())
		}

		log.Printf(job.URL())
	*/
}
