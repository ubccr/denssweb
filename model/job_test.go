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

package model

import (
	"database/sql"
	"testing"
)

func TestJob(t *testing.T) {
	db, err := NewDB("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	_, err = FetchJob(db, "C0YfN48ruj10")
	if err != sql.ErrNoRows {
		t.Error(err)
	}

	email := "test@example.com"

	job := &Job{Email: email, InputData: []byte("test"), FileType: "out"}
	err = QueueJob(db, job)
	if err != nil {
		t.Fatal(err)
	}

	jobx, err := FetchJob(db, job.Token)
	if err != nil {
		t.Fatal(err)
	}

	if jobx.ID != job.ID {
		t.Errorf("Incorrect job ID: got %d should be %d", jobx.ID, job.ID)
	}

	if jobx.Email != email {
		t.Errorf("Incorrect job Email: got %s should be %s", jobx.Email, email)
	}

	jobx, err = FetchNextPending(db)
	if err != nil {
		t.Fatal(err)
	}

	if jobx.ID != job.ID {
		t.Errorf("Incorrect job ID: got %d should be %d", jobx.ID, job.ID)
	}

	if jobx.FileType != job.FileType {
		t.Errorf("Incorrect job File Type: got %s should be %s", jobx.FileType, job.FileType)
	}

	jobx, err = FetchJob(db, job.Token)
	if err != nil {
		t.Fatal(err)
	}

	if jobx.StatusID != StatusRunning {
		t.Errorf("Incorrect job status: got %d should be %d", jobx.StatusID, StatusRunning)
	}

	jobx.DensityMap = []byte("xxx")
	jobx.FSCChart = []byte("yyy")
	jobx.RawData = []byte("zzz")

	err = CompleteJob(db, jobx, StatusComplete)
	if err != nil {
		t.Fatal(err)
	}

	if jobx.StatusID != StatusComplete {
		t.Errorf("Incorrect job status: got %d should be %d", jobx.StatusID, StatusComplete)
	}

	jobs, err := FetchAllJobs(db, StatusComplete, 10, 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(jobs) != 1 {
		t.Errorf("Incorrect number of jobs: got %d should be %d", len(jobs), 1)
	}

	jobs, err = FetchAllJobs(db, StatusError, 10, 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(jobs) != 0 {
		t.Errorf("Incorrect number of jobs: got %d should be %d", len(jobs), 0)
	}
}
