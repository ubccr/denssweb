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

	log "github.com/Sirupsen/logrus"
	"github.com/ubccr/denssweb/app"
	"github.com/ubccr/denssweb/model"
)

func RunClient() {
	ctx, err := app.NewAppContext()
	if err != nil {
		log.Fatal(err.Error())
	}

	job, err := model.FetchNextPending(ctx.DB)
	if err != nil {
		log.Fatal(err.Error())
	}

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

	url, _ := job.URL()
	log.Printf(url)
}
