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
	ID int64 `db:"id"`

	// Unique Job token
	Token string `db:"token"`

	// Job Status ID
	StatusID int64 `db:"status_id"`

	// Job Status string
	Status string `db:"status"`

	// Job Name
	Name string `db:"name"`

	// Input data file (*.dat or GNOM *.out file)
	InputData []byte `db:"input_data"`

	// Resulting density map in CCP4 format
	DensityMap []byte `db:"density_map"`

	// Fourier SHell Correlation (FSC) Curve chart
	FSCChart []byte `db:"fsc_chart"`

	// A zip of the raw output from DENSS
	RawData []byte `db:"raw_data"`

	// Maximum dimension of particle
	Dmax float64 `db:"dmax"`

	// Number of samples. This represents the size of the grid in each
	// dimension. The grid is 3D so NumSamples=31 would be 31 x 31 x 31. The
	// grid size will determine the speed of the calculation and memory used.
	// More samples means greater resolution. This is calculated by DENSS, it's
	// not given to DENSS but we want to control the speed of calcuation so we
	// use this parameter to determine the voxel size.
	NumSamples int `db:"num_samples"`

	// Oversampling size
	Oversampling float64 `db:"oversampling"`

	// Voxel Size
	VoxelSize float64 `db:"voxel_size"`

	// Number of electrons
	Electrons int64 `db:"electrons"`

	// Maximum number of steps
	MaxSteps int64 `db:"max_steps"`

	// Maximum number of times to run DENSS
	MaxRuns int64 `db:"max_runs"`

	// Time the job was submitted
	Submitted *time.Time `db:"submitted"`

	// Time the job started running
	Started *time.Time `db:"started"`

	// Time the job completed
	Completed *time.Time `db:"completed"`
}

func (j *Job) URL() string {
	return fmt.Sprintf("%s/job/%s", viper.GetString("base_url"), j.Token)
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
            j.name,
            j.token,
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

	if job.Dmax <= 0 {
		job.Dmax = 50.0
	}

	// XXX In future versions these will be adjustable parameters. For now we
	// hard code them
	job.Token = randToken()
	job.Oversampling = 2.0
	job.Electrons = 10000
	job.MaxSteps = 3000
	job.MaxRuns = 20
	job.NumSamples = 31
	job.VoxelSize = (job.Dmax * job.Oversampling) / float64(job.NumSamples)

	res, err := tx.NamedExec(`
        insert into job (
            status_id,
            input_data,
            name,
            token,
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
            :input_data,
            :name,
            :token,
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

// Generate random tokens
func randToken() string {
	b := make([]byte, 9)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
