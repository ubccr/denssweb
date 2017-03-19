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
	db, err := newTestDb()
	if err != nil {
		t.Fatal(err)
	}

	_, err = FetchJob(db, 1)
	if err != sql.ErrNoRows {
		t.Error(err)
	}

	job := &Job{InputData: []byte("test")}
	err = QueueJob(db, job)
	if err != nil {
		t.Fatal(err)
	}

	jobx, err := FetchJob(db, job.ID)
	if err != nil {
		t.Fatal(err)
	}

	if jobx.ID != job.ID {
		t.Errorf("Incorrect job ID: got %d should be %d", jobx.ID, job.ID)
	}

	jobx, err = FetchNextPending(db)
	if err != nil {
		t.Fatal(err)
	}

	if jobx.ID != job.ID {
		t.Errorf("Incorrect job ID: got %d should be %d", jobx.ID, job.ID)
	}

	jobx, err = FetchJob(db, job.ID)
	if err != nil {
		t.Fatal(err)
	}

	if jobx.StatusID != StatusRunning {
		t.Errorf("Incorrect job status: got %d should be %d", jobx.StatusID, StatusRunning)
	}
}
