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

type querySettings struct {
	URI        string
	Token      string
	Tags       bool
	Prerelease bool
	Constraint string
	LimitTo    int
}

func queryAction(driver string, settings *querySettings) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if settings.Token == "" {
			return fmt.Errorf("no token provided: %w", ErrCli)
		}

		if c.NArg() != 2 {
			return fmt.Errorf("command takes two arguments <owner> <repo>: %w", ErrCli)
		}

		owner := c.Args().Get(0)
		repo := c.Args().Get(1)

		client, err := reqcheck.NewClientFromDriver(driver, settings.URI, settings.Token)
		if err != nil {
			return fmt.Errorf("could not connect to %s server at %s: %w", driver, settings.URI, err)
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
				return fmt.Errorf("error when getting releases from %s/%s/%s: %w", settings.URI, owner, repo, err)
			}

			release := item.V.(reqcheck.Release)
			if release.SemVer != nil {
				fmt.Printf("tag %s -> semver %s\n", release.Tag, release.SemVer.String())
			} else {
				fmt.Printf("tag %s -> semver ???\n", release.Tag)
			}
		}

		return nil
	}
}
