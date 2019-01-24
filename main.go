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

package main

import (
	"runtime"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ubccr/denssweb/app"
	"github.com/ubccr/denssweb/client"
	"github.com/ubccr/denssweb/server"
	"github.com/urfave/cli"
)

var (
	Version = "dev"
)

func init() {
	viper.SetConfigName("denssweb")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
}

func main() {
	capp := cli.NewApp()
	capp.Name = "denssweb"
	capp.Copyright = `Copyright 2017 DENSSWeb Authors.  

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
   `
	capp.Authors = []cli.Author{
		{Name: "Andrew E. Bruno", Email: "aebruno2@buffalo.edu"},
		{Name: "Thomas D. Grant", Email: "tgrant@hwi.buffalo.edu"}}
	capp.Usage = "denssweb"
	capp.Version = Version
	capp.Flags = []cli.Flag{
		&cli.StringFlag{Name: "conf,c", Usage: "Path to conf file"},
		&cli.BoolFlag{Name: "debug,d", Usage: "Print debug messages"},
	}
	capp.Before = func(c *cli.Context) error {
		if c.GlobalBool("debug") {
			log.SetLevel(log.InfoLevel)
		} else {
			log.SetLevel(log.WarnLevel)
		}

		conf := c.GlobalString("conf")
		if len(conf) > 0 {
			viper.SetConfigFile(conf)
		}

		err := viper.ReadInConfig()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Warn("Failed to parse config file. Using defaults")
		}

		return nil
	}
	capp.Commands = []cli.Command{
		{
			Name:  "run",
			Usage: "Run both http server and client work",
			Flags: []cli.Flag{
				&cli.IntFlag{Name: "threads, t", Value: runtime.NumCPU(), Usage: "Max threads (default numcpu)"},
			},
			Action: func(c *cli.Context) {
				ctx, err := app.NewAppContext()
				if err != nil {
					log.Fatal(err.Error())
				}
				go client.RunClient(ctx, c.Int("threads"))
				server.RunServer(ctx)
			},
		},
		{
			Name:  "server",
			Usage: "Run http server only",
			Action: func(c *cli.Context) {
				ctx, err := app.NewAppContext()
				if err != nil {
					log.Fatal(err.Error())
				}
				server.RunServer(ctx)
			},
		},
		{
			Name:  "client",
			Usage: "Run client worker only",
			Flags: []cli.Flag{
				&cli.IntFlag{Name: "threads, t", Value: runtime.NumCPU(), Usage: "Max threads (default numcpu)"},
			},
			Action: func(c *cli.Context) {
				ctx, err := app.NewAppContext()
				if err != nil {
					log.Fatal(err.Error())
				}
				client.RunClient(ctx, c.Int("threads"))
			},
		}}

	capp.RunAndExitOnError()
}
