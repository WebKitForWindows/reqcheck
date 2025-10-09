// Copyright (c) 2023, the WebKit for Windows project authors.  Please see the
// AUTHORS file for details. All rights reserved. Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

var (
	version = "unknown"
	ErrCli  = errors.New("cli error")
)

func main() {
	var logLevel string

	app := &cli.Command{
		Name:                  "reqcheck",
		Usage:                 "query releases",
		Version:               version,
		EnableShellCompletion: true,
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
		Before: func(c context.Context, cmd *cli.Command) (context.Context, error) {
			lvl, err := logrus.ParseLevel(logLevel)
			if err != nil {
				return c, fmt.Errorf("invalid logging level %s: %w", logLevel, err)
			}

			logrus.SetLevel(lvl)

			return c, nil
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
