// Copyright (c) 2023, the WebKit for Windows project authors.  Please see the
// AUTHORS file for details. All rights reserved. Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package reqcheck

import (
	"context"
	"fmt"

	"github.com/reactivex/rxgo/v2"
	"github.com/sirupsen/logrus"
)

type ListReleaseOptions struct {
	Owner   string
	Repo    string
	Tags    bool
	LimitTo int
}

const (
	startingPage   = 1
	perPageDefault = 30
	limitToDefault = 100000000
)

func ListReleases(client Client, opts ListReleaseOptions) rxgo.Observable {
	listOpts := ListOptions{Page: startingPage, PerPage: perPageDefault}

	var listFunc func(context.Context, string, string, ListOptions) ([]Release, error)
	if opts.Tags {
		listFunc = client.ListTags
	} else {
		listFunc = client.ListReleases
	}

	// Default limit to an arbitrarily high number
	if opts.LimitTo == 0 {
		opts.LimitTo = limitToDefault
	}

	itemCount := 0

	return rxgo.Create([]rxgo.Producer{
		func(ctx context.Context, next chan<- rxgo.Item) {
			for {
				items, err := listFunc(ctx, opts.Owner, opts.Repo, listOpts)
				if err != nil {
					next <- rxgo.Error(fmt.Errorf("could not access %s/%s releases: %w", opts.Owner, opts.Repo, err))

					break
				}

				for _, item := range items {
					next <- rxgo.Of(item)

					itemCount++
					if itemCount >= opts.LimitTo {
						logrus.WithField("limit-to", opts.LimitTo).Debug("reached query limit")

						break
					}
				}

				if len(items) < listOpts.PerPage {
					break
				}

				listOpts.Page++
			}
		},
	})
}
