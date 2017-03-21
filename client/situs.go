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
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ubccr/denssweb/model"
)

// Convert DENSS XPLOR output files to MRC
func convertToMRC(log *logrus.Logger, job *model.Job, workDir string, thread int) error {
	xplorFile := filepath.Join(workDir, fmt.Sprintf("output_%d.xplor", thread))
	mrcFile := filepath.Join(workDir, fmt.Sprintf("output_%d.mrc", thread))

	args := []string{
		xplorFile,
		mrcFile,
	}

	log.WithFields(logrus.Fields{
		"id":     job.ID,
		"xplor":  xplorFile,
		"mrc":    mrcFile,
		"thread": thread,
	}).Info("Converting denss output to mrc using map2map")

	cmd := exec.Command(viper.GetString("map2map_path"), args...)
	cmd.Dir = workDir

	// This sets input file type to xplor for map2map command
	cmd.Stdin = strings.NewReader("2\n")

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(logrus.Fields{
			"id":        job.ID,
			"error":     err.Error(),
			"xplorFile": xplorFile,
			"thread":    thread,
			"output":    string(out),
		}).Error("map2map command failed")
		return err
	}

	// Ensure mrc file exists
	_, err = os.Stat(mrcFile)
	if os.IsNotExist(err) {
		log.WithFields(logrus.Fields{
			"id":        job.ID,
			"error":     err.Error(),
			"xplorFile": xplorFile,
			"thread":    thread,
			"mrcFile":   mrcFile,
			"output":    string(out),
		}).Error("mrc file does not exists. map2map failed")
		return err
	} else if err != nil {
		log.WithFields(logrus.Fields{
			"id":        job.ID,
			"error":     err.Error(),
			"xplorFile": xplorFile,
			"thread":    thread,
			"mrcFile":   mrcFile,
			"output":    string(out),
		}).Error("Failed to read mrc. map2map failed")
		return err
	}

	log.WithFields(logrus.Fields{
		"id":     job.ID,
		"xplor":  xplorFile,
		"mrc":    mrcFile,
		"thread": thread,
	}).Info("map2map completed successfully")

	return nil
}
