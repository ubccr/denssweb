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
	"archive/zip"
	"bytes"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/ubccr/denssweb/model"
)

// Create zip archive file of DENSS output files
func createZIP(log *logrus.Logger, job *model.Job, workDir string) error {

	log.WithFields(logrus.Fields{
		"id": job.ID,
	}).Info("Creating zip archive")

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	finalAvg, err := ioutil.ReadFile(filepath.Join(workDir, "spt_01", "final_avg_ali2ref.hdf"))
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    job.ID,
			"file":  finalAvg,
		}).Error("Failed to read hdf file")
		return err
	}

	fsc01, err := ioutil.ReadFile(filepath.Join(workDir, "spt_01", "fsc_0.txt"))
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    job.ID,
			"file":  fsc01,
		}).Error("Failed to read fsc file")
		return err
	}

	var files = []struct {
		Name string
		Body []byte
	}{
		{"density-map.ccp4", job.DensityMap},
		{"final_avg_ali2ref.hdf", finalAvg},
		{"fsc.png", job.FSCChart},
		{"fsc_01.txt", fsc01},
	}
	for _, file := range files {
		header := &zip.FileHeader{
			Name:               file.Name,
			Method:             zip.Store,
			UncompressedSize64: uint64(len(file.Body)),
		}
		header.SetModTime(time.Now().UTC())
		header.SetMode(0600)

		f, err := w.CreateHeader(header)
		if err != nil {
			log.WithFields(logrus.Fields{
				"error": err.Error(),
				"id":    job.ID,
				"file":  file.Name,
			}).Error("Failed to create file in zip archive")
			return err
		}
		_, err = f.Write([]byte(file.Body))
		if err != nil {
			log.WithFields(logrus.Fields{
				"error": err.Error(),
				"id":    job.ID,
				"file":  file.Name,
			}).Error("Failed to write file to zip archive")
			return err
		}
	}

	err = w.Close()
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    job.ID,
		}).Error("Failed to create zip archive")
		return err
	}

	job.RawData = buf.Bytes()

	log.WithFields(logrus.Fields{
		"id": job.ID,
	}).Info("Successfully created zip archive")

	return nil
}
