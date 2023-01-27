// Copyright (c) 2023, the WebKit for Windows project authors.  Please see the
// AUTHORS file for details. All rights reserved. Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	version = "unknown"
	ErrCli  = errors.New("cli error")
)

func main() {
	var logLevel string

	app := &cli.App{
		Name:                 "reqcheck",
		Usage:                "query releases",
		Version:              version,
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "log-level",
				Usage:       "logging level",
				Value:       "warning",
				Destination: &logLevel,
			},
		},
		Commands: []*cli.Command{
			githubCmd(),
			gitlabCmd(),
			vcpkgCmd(),
		},
		Before: func(c *cli.Context) error {
			lvl, err := logrus.ParseLevel(logLevel)
			if err != nil {
				return fmt.Errorf("invalid logging level %s: %w", logLevel, err)
			}

			logrus.SetLevel(lvl)

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
