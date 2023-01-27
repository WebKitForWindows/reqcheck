// Copyright (c) 2023, the WebKit for Windows project authors.  Please see the
// AUTHORS file for details. All rights reserved. Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package reqcheck

import (
	"fmt"
	"regexp"

	"github.com/Masterminds/semver"
	"github.com/sirupsen/logrus"
)

var versionMatcher = regexp.MustCompile(`^[a-zA-Z-_.]*(?P<major>\d+)[._-](?P<minor>\d*)[._-]*(?P<patch>\d*)[._-]*(?P<prerelease>[a-zA-Z-_.]*)(?P<preversion>\d*)$`)

func generateVersion(tag string, matcher *regexp.Regexp) *semver.Version {
	semVer, err := semver.NewVersion(tag)
	if err == nil {
		return semVer
	}

	match := matcher.FindStringSubmatch(tag)
	if match == nil {
		return nil
	}

	major := match[1]

	minor := match[2]
	if minor == "" {
		minor = "0"
	}

	patch := match[3]
	if patch == "" {
		patch = "0"
	}

	prerelease := match[4]
	if prerelease != "" {
		if prerelease == "." {
			prerelease = "build"
		}

		prerelease = "-" + prerelease

		preVersion := match[5]
		if preVersion != "" {
			prerelease += "." + preVersion
		}
	}

	semVer, err = semver.NewVersion(fmt.Sprintf("%s.%s.%s%s", major, minor, patch, prerelease))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"major":      major,
			"minor":      minor,
			"patch":      patch,
			"prerelease": match[4],
			"preversion": match[5],
		}).Warn("could not parse version")

		return nil
	}

	return semVer
}
