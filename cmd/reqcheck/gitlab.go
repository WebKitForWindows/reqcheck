// Copyright (c) 2023, the WebKit for Windows project authors.  Please see the
// AUTHORS file for details. All rights reserved. Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package main

import (
	"github.com/WebKitForWindows/reqcheck"
	"github.com/urfave/cli/v3"
)

func gitlabCmd() *cli.Command {
	var settings querySettings

	return &cli.Command{
		Name:      "gitlab",
		Usage:     "query gitlab for requirements",
		ArgsUsage: "<owner> <repo>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "uri",
				Usage:       "uri for gitlab instance",
				Value:       "https://gitlab.com",
				Destination: &settings.URI,
			},
			&cli.StringFlag{
				Name:        "token",
				Usage:       "access token for gitlab api",
				Sources:     cli.EnvVars("GITLAB_TOKEN"),
				Destination: &settings.Token,
			},
			&cli.BoolFlag{
				Name:        "tags",
				Usage:       "use tags rather than releases",
				Destination: &settings.Tags,
			},
			&cli.BoolFlag{
				Name:        "prerelease",
				Usage:       "include pre-releases",
				Destination: &settings.Prerelease,
			},
			&cli.StringFlag{
				Name:        "constraint",
				Usage:       "semantic version constraint",
				Destination: &settings.Constraint,
			},
			&cli.IntFlag{
				Name:        "limit-to",
				Usage:       "limit the amount of results from the api",
				Destination: &settings.LimitTo,
			},
		},
		Action: queryAction(reqcheck.DriverGitLab, &settings),
	}
}
