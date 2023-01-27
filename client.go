// Copyright (c) 2023, the WebKit for Windows project authors.  Please see the
// AUTHORS file for details. All rights reserved. Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package reqcheck

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/semver"
)

type (
	Release struct {
		Tag    string
		SemVer *semver.Version
	}

	ListOptions struct {
		// For paginated result sets, page of results to retrieve.
		Page int
		// For paginated result sets, the number of results to include per page.
		PerPage int
	}

	Client interface {
		ListReleases(ctx context.Context, owner, name string, opt ListOptions) ([]Release, error)

		ListTags(ctx context.Context, owner, name string, opt ListOptions) ([]Release, error)
	}
)

var ErrScmDriver = errors.New("scm driver error")

func NewClientFromDriver(driver, uri, token string) (Client, error) {
	if driver == "github" {
		return NewGitHub(uri, token)
	}

	if driver == "gitlab" {
		return NewGitLab(uri, token)
	}

	return nil, fmt.Errorf("unknown scm driver %s: %w", driver, ErrScmDriver)
}
