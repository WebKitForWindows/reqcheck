// Copyright (c) 2023, the WebKit for Windows project authors.  Please see the
// AUTHORS file for details. All rights reserved. Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package reqcheck

import (
	"slices"

	"github.com/Masterminds/semver"
)

func FilterSemanticVersion(item interface{}) bool {
	release := item.(Release)

	return release.SemVer != nil
}

func FilterStableRelease(item interface{}) bool {
	release := item.(Release)
	if release.SemVer == nil {
		return false
	}

	return release.SemVer.Prerelease() == ""
}

func FilterRepeatVersions() func(interface{}) bool {
	seen := make([]string, 0)
	return func(item interface{}) bool {
		release := item.(Release)
		if release.SemVer == nil {
			return true
		}

		versionString := release.SemVer.String()

		if slices.Contains(seen, versionString) {
			return false
		}

		seen = append(seen, versionString)
		return true
	}
}

func FilterSemanticConstraint(c *semver.Constraints) func(interface{}) bool {
	return func(item interface{}) bool {
		release := item.(Release)
		if release.SemVer == nil {
			return false
		}

		return c.Check(release.SemVer)
	}
}
