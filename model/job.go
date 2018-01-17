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
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

const (
	_              = iota // 0
	StatusPending         // 1
	StatusRunning         // 2
	StatusComplete        // 3
	StatusError           // 4
)

// A DENSS Job
type Job struct {
	// Unique ID for the Job
	ID int64 `db:"id" json:"id"`

	// Email Address
	Email string `db:"email" json:"-"`

	// Unique Job token
	Token string `db:"token" json:"-"`

	// Job Status ID
	StatusID int64 `db:"status_id" json:"-"`

	// Job Status string
	Status string `db:"status" json:"status"`

	// Task Name
	Task string `db:"task" json:"task"`

	// Percent complete
	PercentComplete int64 `db:"percent_complete" json:"percent_complete"`

	// Log message for task
	LogMessage string `db:"log_message" json:"log_message"`

	// Job Name
	Name string `db:"name" json:"name"`

	// Input data file (*.dat or GNOM *.out file)
	InputData []byte `db:"input_data" json:"-"`

	// Resulting density map in CCP4 format
	DensityMap []byte `db:"density_map" json:"-"`

	// Fourier SHell Correlation (FSC) Curve chart
	FSCChart []byte `db:"fsc_chart" json:"-"`

	// Summary stats chart
	SummaryChart []byte `db:"summary_chart" json:"-"`

	// A zip of the raw output from DENSS
	RawData []byte `db:"raw_data" json:"-"`

	// Maximum dimension of particle
	Dmax float64 `db:"dmax" json:"-"`

	// Number of samples. This represents the size of the grid in each
	// dimension. The grid is 3D so NumSamples=31 would be 31 x 31 x 31. The
	// grid size will determine the speed of the calculation and memory used.
	// More samples means greater resolution. This is calculated by DENSS, it's
	// not given to DENSS but we want to control the speed of calcuation so we
	// use this parameter to determine the voxel size.
	NumSamples int64 `db:"num_samples" json:"-"`

	// Oversampling size
	Oversampling float64 `db:"oversampling" json:"-"`

	// Voxel Size
	VoxelSize float64 `db:"voxel_size" json:"-"`

	// Number of electrons
	Electrons int64 `db:"electrons" json:"-"`

	// Maximum number of steps
	MaxSteps int64 `db:"max_steps" json:"-"`

	// Maximum number of times to run DENSS
	MaxRuns int64 `db:"max_runs" json:"-"`

	// Time the job was submitted
	Submitted *time.Time `db:"submitted" json:"-"`

	// Time the job started running
	Started *time.Time `db:"started" json:"-"`

	// Time the job completed
	Completed *time.Time `db:"completed" json:"-"`

	// Current running/wait time for the job. Only used in json
	Time string `db:"-" json:"time"`
}

func (j *Job) URL() string {
	return fmt.Sprintf("%s/job/%s", viper.GetString("base_url"), j.Token)
}

func (j *Job) RunTime() string {
	wt := ""

	if j.Started == nil {
		return wt
	} else if j.Completed != nil {
		wt = humanize.RelTime(*j.Started, *j.Completed, "", "")
	} else {
		now := time.Now()
		wt = humanize.RelTime(*j.Started, now, "", "")
	}

	if wt == "now" {
		wt = "0 seconds"
	}

	return wt
}

func (j *Job) WaitTime() string {
	wt := ""

	if j.Submitted == nil {
		return wt
	} else if j.Started != nil {
		wt = humanize.RelTime(*j.Submitted, *j.Started, "", "")
	} else {
		now := time.Now()
		wt = humanize.RelTime(*j.Submitted, now, "", "")
	}

	if wt == "now" {
		wt = "0 seconds"
	}

	return wt
}

// Fetch job by token. This is used for displaying the Job status in the web
// interface and no raw binary data is included
func FetchJob(db *sqlx.DB, token string) (*Job, error) {
	job := Job{}
	err := db.Get(&job, `
		select
			j.id,
			j.status_id,
			s.status,
            j.task,
            j.percent_complete,
            j.log_message,
            j.name,
            j.token,
            j.email,
            j.dmax,
            j.name,
            j.oversampling,
            j.num_samples,
            j.voxel_size,
            j.electrons,
            j.max_steps,
            j.max_runs,
            j.submitted,
            j.started,
            j.completed
        from job as j 
        join job_status s on s.id = j.status_id
        where j.token = ?`, token)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

// Queue a new DENSS Job
func QueueJob(db *sqlx.DB, job *Job) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Commit()

	job.StatusID = StatusPending
	now := time.Now()
	job.Submitted = &now

	job.Task = "Not started"
	job.PercentComplete = 0
	job.LogMessage = ""
	job.Token = randToken()

	// Set default values for params
	if job.Dmax <= 0 {
		job.Dmax = 50.0
	}
	if job.Oversampling <= 0 {
		job.Oversampling = 2.0
	}
	if job.Electrons <= 0 {
		job.Electrons = 10000
	}
	if job.MaxSteps <= 0 {
		job.MaxSteps = 3000
	}
	if job.MaxRuns <= 0 {
		job.MaxRuns = 20
	}
	if job.NumSamples <= 0 {
		job.NumSamples = 33
	}
	if job.VoxelSize <= 0 {
		job.VoxelSize = (job.Dmax * job.Oversampling) / float64(job.NumSamples)
	}

	res, err := tx.NamedExec(`
        insert into job (
            status_id,
            task,
            percent_complete,
            log_message,
            input_data,
            name,
            token,
            email,
            dmax,
            num_samples,
            oversampling,
            electrons,
            max_steps,
            max_runs,
            voxel_size,
            submitted
        ) values (
            :status_id,
            :task,
            :percent_complete,
            :log_message,
            :input_data,
            :name,
            :token,
            :email,
            :dmax,
            :num_samples,
            :oversampling,
            :electrons,
            :max_steps,
            :max_runs,
            :voxel_size,
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

// Fetch next job in pending status, update status to running and return job
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
            j.input_data,
            j.name,
            j.token,
            j.email,
            j.dmax,
            j.oversampling,
            j.num_samples,
            j.electrons,
            j.max_steps,
            j.max_runs,
            j.voxel_size,
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

// Fetch all jobs by status
func FetchAllJobs(db *sqlx.DB, status, limit, offset int) ([]*Job, error) {
	jobs := []*Job{}

	args := make([]interface{}, 0)
	query := `
        select
			j.id,
			j.status_id,
			s.status,
            j.name,
            j.token,
            j.email,
            j.dmax,
            j.oversampling,
            j.num_samples,
            j.electrons,
            j.max_steps,
            j.max_runs,
            j.voxel_size,
            j.submitted,
            j.started,
            j.completed
        from job as j 
        join job_status s on s.id = j.status_id`

	if status > 0 {
		query += ` where j.status_id = ?`
		args = append(args, status)
		if status == StatusComplete {
			query += ` order by j.completed desc`
		} else if status == StatusRunning {
			query += ` order by j.started desc`
		} else if status == StatusError {
			query += ` order by j.completed desc`
		} else {
			query += ` order by j.submitted desc`
		}
	} else {
		query += ` order by j.submitted desc`
	}

	query += ` limit ? offset ?`
	args = append(args, limit, offset)

	err := db.Select(&jobs, query, args...)
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

// Complete Job
func CompleteJob(db *sqlx.DB, job *Job, statusID int) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Commit()

	job.StatusID = int64(statusID)
	now := time.Now()
	job.Completed = &now

	_, err = tx.NamedExec(`
        update job set
            status_id = :status_id,
            density_map = :density_map,
            fsc_chart = :fsc_chart,
            summary_chart = :summary_chart,
            raw_data = :raw_data,
            completed = :completed
        where id = :id`, job)
	if err != nil {
		return err
	}

	return nil
}

// Fetch job density map by token.
func FetchDensityMap(db *sqlx.DB, token string) (*Job, error) {
	job := Job{}
	err := db.Get(&job, `
		select
			j.id,
			j.status_id,
            j.name,
            j.density_map,
            j.submitted,
            j.started,
            j.completed
        from job as j 
        where j.token = ?`, token)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

// Fetch job fsc chart by token.
func FetchFSCChart(db *sqlx.DB, token string) (*Job, error) {
	job := Job{}
	err := db.Get(&job, `
		select
			j.id,
			j.status_id,
            j.name,
            j.fsc_chart,
            j.submitted,
            j.started,
            j.completed
        from job as j 
        where j.token = ?`, token)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

// Fetch job summary chart by token.
func FetchSummaryChart(db *sqlx.DB, token string) (*Job, error) {
	job := Job{}
	err := db.Get(&job, `
		select
			j.id,
			j.status_id,
            j.name,
            j.summary_chart,
            j.submitted,
            j.started,
            j.completed
        from job as j 
        where j.token = ?`, token)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

// Fetch job raw data by token.
func FetchRawData(db *sqlx.DB, token string) (*Job, error) {
	job := Job{}
	err := db.Get(&job, `
		select
			j.id,
			j.status_id,
            j.name,
            j.raw_data,
            j.submitted,
            j.started,
            j.completed
        from job as j 
        where j.token = ?`, token)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

// Log message for job
func LogJobMessage(db *sqlx.DB, job *Job, task, message string, percent int) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Commit()

	job.Task = task
	job.LogMessage = message
	job.PercentComplete = int64(percent)

	_, err = tx.NamedExec(`
        update job set
            task = :task,
            log_message = :log_message,
            percent_complete = :percent_complete
        where id = :id`, job)
	if err != nil {
		return err
	}

	return nil
}

// Generate random tokens
func randToken() string {
	b := make([]byte, 9)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
