// Copyright (c) 2023, the WebKit for Windows project authors.  Please see the
// AUTHORS file for details. All rights reserved. Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/WebKitForWindows/reqcheck"
	"github.com/urfave/cli/v2"
)

func githubCmd() *cli.Command {
	settings := struct {
		URI        string
		Token      string
		Tags       bool
		Prerelease bool
		Constraint string
		LimitTo    int
	}{}

	return &cli.Command{
		Name:      "github",
		Usage:     "query github for requirements",
		ArgsUsage: "<owner> <repo>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "uri",
				Usage:       "uri for github instance",
				Value:       "https://github.com",
				Destination: &settings.URI,
			},
			&cli.StringFlag{
				Name:        "token",
				Usage:       "access token for github api",
				EnvVars:     []string{"GITHUB_TOKEN"},
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
		Action: func(c *cli.Context) error {
			if settings.Token == "" {
				return fmt.Errorf("no token provided: %w", ErrCli)
			}

			if c.NArg() != 2 {
				return fmt.Errorf("command takes two arguments <owner> <repo>: %w", ErrCli)
			}

			owner := c.Args().Get(0)
			repo := c.Args().Get(1)

			client, err := reqcheck.NewGitHub(settings.URI, settings.Token)
			if err != nil {
				return fmt.Errorf("could not connect to github server at %s: %w", settings.URI, err)
			}

			releaseOpts := reqcheck.ListReleaseOptions{
				Owner:   owner,
				Repo:    repo,
				Tags:    settings.Tags,
				LimitTo: settings.LimitTo,
			}

			observer := reqcheck.ListReleases(client, releaseOpts)

			if settings.Constraint != "" {
				constraint, err := semver.NewConstraint(settings.Constraint)
				if err != nil {
					return fmt.Errorf("could not parse constraint %s: %w", settings.Constraint, err)
				}

				observer = observer.Filter(reqcheck.FilterSemanticConstraint(constraint))
			} else if !settings.Prerelease {
				observer = observer.Filter(reqcheck.FilterStableRelease)
			}

			for item := range observer.Observe() {
				if item.Error() {
					return fmt.Errorf("error when getting releases for %s%s: %w", owner, repo, err)
				}

				release := item.V.(reqcheck.Release)
				if release.SemVer != nil {
					fmt.Printf("tag %s -> semver %s\n", release.Tag, release.SemVer.String())
				} else {
					fmt.Printf("tag %s -> semver ???\n", release.Tag)
				}
			}

			return nil
		},
	}
}
