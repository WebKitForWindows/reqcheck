// Copyright (c) 2024, the WebKit for Windows project authors.  Please see the
// AUTHORS file for details. All rights reserved. Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package reqcheck

import "context"

func ReduceGreatestVersion(_ context.Context, acc interface{}, elem interface{}) (interface{}, error) {
	if acc == nil {
		return elem, nil
	}

	accRelease := acc.(Release)
	if accRelease.SemVer == nil {
		return elem, nil
	}

	elemRelease := elem.(Release)
	if elemRelease.SemVer == nil {
		return acc, nil
	}

	if accRelease.SemVer.GreaterThan(elemRelease.SemVer) {
		return acc, nil
	}

	return elem, nil
}
