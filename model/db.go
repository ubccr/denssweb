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
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const (
	JobSchema = `
		create table if not exists job 
		(id integer primary key, status_id integer, input_data blob, dmax real,
         density_map blob, fsc_chart blob, summary_chart bob, raw_data blob, oversampling real, token string,
         electrons integer, max_steps integer, max_runs integer, name string, num_samples integer,
         task string, percent_complete integer, log_message string, email string,
         voxel_size real, submitted datetime, started datetime, completed datetime)
	`
	JobStatusSchema = `
		create table if not exists job_status
		(id integer primary key, status string)
	`
)

func NewDB(driver, dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	if driver == "sqlite3" {
		err = initDB(db)
		if err != nil {
			return nil, err
		}
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func initDB(db *sqlx.DB) error {
	// Create tables if not exist
	_, err := db.Exec(JobSchema)
	if err != nil {
		return err
	}

	_, err = db.Exec(JobStatusSchema)
	if err != nil {
		return err
	}

	_, err = db.Exec(`replace into job_status (id,status) values (?,?)`, StatusPending, "Pending")
	if err != nil {
		return err
	}

	_, err = db.Exec(`replace into job_status (id,status) values (?,?)`, StatusRunning, "Running")
	if err != nil {
		return err
	}

	_, err = db.Exec(`replace into job_status (id,status) values (?,?)`, StatusComplete, "Complete")
	if err != nil {
		return err
	}

	_, err = db.Exec(`replace into job_status (id,status) values (?,?)`, StatusError, "Error")
	if err != nil {
		return err
	}

	return nil
}
