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
	"os"
	"path/filepath"

	"github.com/jhoonb/archivex"
	"github.com/spf13/viper"
	"github.com/ubccr/denssweb/model"
)

// Create zip archive file of DENSS output files
func createZIP(job *model.Job, workDir string) error {
	zipFile := filepath.Join(viper.GetString("work_dir"), fmt.Sprintf("denss-%d.zip", job.ID))
	os.Remove(zipFile)

	zip := new(archivex.ZipFile)

	err := zip.Create(zipFile)
	if err != nil {
		return err
	}

	err = zip.AddAll(workDir, true)
	if err != nil {
		return err
	}

	err = zip.Close()
	if err != nil {
		return err
	}

	job.RawData, err = ioutil.ReadFile(zipFile)
	if err != nil {
		return err
	}

	return nil
}
