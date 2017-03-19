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
	"time"

	//log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
)

const (
	StatusPending  = iota // 0
	StatusRunning         // 1
	StatusComplete        // 2
	StatusError           // 3
)

type Job struct {
	ID        int64      `db:"id"`
	StatusID  int64      `db:"status_id"`
	Status    string     `db:"status"`
	WorkDir   string     `db:"work_dir"`
	Submitted *time.Time `db:"submitted"`
	Started   *time.Time `db:"started"`
	Completed *time.Time `db:"completed"`
}

func FetchJob(db *sqlx.DB, id int64) (*Job, error) {
	job := Job{}
	err := db.Get(&job, `
		select
			j.id,
			j.status_id,
			s.status,
            j.work_dir,
            j.submitted,
            j.started,
            j.completed
        from job as j 
        join job_status s on s.id = j.status_id
        where j.id = ?`, id)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func QueueJob(db *sqlx.DB, job *Job) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Commit()

	job.StatusID = StatusPending
	now := time.Now()
	job.Submitted = &now

	res, err := tx.NamedExec(`
        insert into job (
            status_id,
            work_dir,
            submitted
        ) values (
            :status_id,
            :work_dir,
            :submitted)`, job)
	if err != nil {
		return err
	}

	job.ID, err = res.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func FetchNextPending(db *sqlx.DB) (*Job, error) {
	tx, err := db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Commit()

	job := Job{}
	err = tx.Get(&job, `
		select
			j.id,
			j.status_id,
			s.status,
            j.work_dir,
            j.submitted,
            j.started,
            j.completed
        from job as j 
        join job_status s on s.id = j.status_id
        where j.status_id = ?
        order by j.submitted asc
        limit 1`, StatusPending)
	if err != nil {
		return nil, err
	}

	job.StatusID = StatusRunning
	job.Status = "Running"
	now := time.Now()
	job.Started = &now

	_, err = tx.NamedExec(`
        update job set status_id = :status_id, started = :started
        where id = :id`, job)
	if err != nil {
		return nil, err
	}

	return &job, nil
}
