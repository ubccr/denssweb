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
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ubccr/denssweb/model"
)

// Combine DENSS output files into a single HDF file
func buildStack(log *logrus.Logger, job *model.Job, workDir string) error {
	stackFile := filepath.Join(workDir, "stack.hdf")

	args := []string{
		"--stackname",
		stackFile,
	}
	for i := 0; i < int(job.MaxRuns); i++ {
		args = append(args, filepath.Join(workDir, fmt.Sprintf("output_%d.mrc", i)))
	}

	log.WithFields(logrus.Fields{
		"id":        job.ID,
		"stackFile": stackFile,
	}).Info("Building stack hdf using EMAN2")

	e2stacks := filepath.Join(viper.GetString("eman2dir"), "bin", "e2buildstacks.py")
	cmd := exec.Command(e2stacks, args...)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(logrus.Fields{
			"id":        job.ID,
			"error":     err.Error(),
			"stackFile": stackFile,
			"output":    string(out),
		}).Error("e2buildstacks.py command failed")
		return err
	}

	log.WithFields(logrus.Fields{
		"id":        job.ID,
		"stackFile": stackFile,
	}).Info("stack hdf built successfully")

	return nil
}

// Run averaging
func runAveraging(log *logrus.Logger, job *model.Job, workDir string, threads int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(viper.GetInt64("max_seconds"))*time.Second)
	defer cancel()

	stackResizedFile := filepath.Join(workDir, "stack.hdf")

	args := []string{
		"--input",
		stackResizedFile,
		fmt.Sprintf("--parallel=thread:%d", threads),
		"--saveali",
		"--savesteps",
		"--keep",
		"3.0",
		"--keepsig",
	}

	log.WithFields(logrus.Fields{
		"id":               job.ID,
		"stackResizedFile": stackResizedFile,
		"threads":          threads,
	}).Info("Running averaging using EMAN2")

	e2spt := filepath.Join(viper.GetString("eman2dir"), "bin", "e2spt_classaverage.py")
	cmd := exec.CommandContext(ctx, e2spt, args...)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(logrus.Fields{
			"error":            err.Error(),
			"id":               job.ID,
			"stackResizedFile": stackResizedFile,
			"threads":          threads,
			"output":           string(out),
		}).Error("e2spt_classaverage.py command failed")
		return err
	}

	densityMapHDF := filepath.Join(workDir, "spt_01", "final_avg_ali2ref.hdf")

	// Ensure final_avg_ali2ref.hdf exists
	_, err = os.Stat(densityMapHDF)
	if os.IsNotExist(err) {
		log.WithFields(logrus.Fields{
			"id":               job.ID,
			"error":            err.Error(),
			"hdf":              densityMapHDF,
			"stackResizedFile": stackResizedFile,
			"output":           string(out),
		}).Error("Final averaging file does not exist. e2spt_classaverage.py command failed")
		return err
	} else if err != nil {
		log.WithFields(logrus.Fields{
			"id":               job.ID,
			"error":            err.Error(),
			"hdf":              densityMapHDF,
			"stackResizedFile": stackResizedFile,
			"output":           string(out),
		}).Error("Failed to read final averaging file. e2spt_classaverage.py command failed")
		return err
	}

	log.WithFields(logrus.Fields{
		"id":               job.ID,
		"stackResizedFile": stackResizedFile,
		"threads":          threads,
	}).Info("Averaging completed successfully")

	densityMapCCP4 := filepath.Join(workDir, "output_averaged.ccp4")

	log.WithFields(logrus.Fields{
		"id":   job.ID,
		"hdf":  densityMapHDF,
		"ccp4": densityMapCCP4,
	}).Info("Converting electron density map to CCP4")

	e2proc3d := filepath.Join(viper.GetString("eman2dir"), "bin", "e2proc3d.py")
	args = []string{
		densityMapHDF,
		densityMapCCP4,
	}

	cmd = exec.CommandContext(ctx, e2proc3d, args...)
	cmd.Dir = workDir
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.WithFields(logrus.Fields{
			"error":  err.Error(),
			"id":     job.ID,
			"hdf":    densityMapHDF,
			"ccp4":   densityMapCCP4,
			"output": string(out),
		}).Error("e2proc3d.py command failed to convert to CCP4")
		return err
	}

	job.DensityMap, err = ioutil.ReadFile(densityMapCCP4)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error":  err.Error(),
			"id":     job.ID,
			"hdf":    densityMapHDF,
			"ccp4":   densityMapCCP4,
			"output": string(out),
		}).Error("Failed to read ccp4 file")
		return err
	}

	log.WithFields(logrus.Fields{
		"id":   job.ID,
		"hdf":  densityMapHDF,
		"ccp4": densityMapCCP4,
	}).Info("Successfully converted electron density map to CCP4")

	return nil
}
