// Copyright (c) 2023, the WebKit for Windows project authors.  Please see the
// AUTHORS file for details. All rights reserved. Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package reqcheck

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

type gitlabClient struct {
	client *gitlab.Client
}

func NewGitLab(uri, token string) (Client, error) {
	return NewGitLabClient(uri, token, http.DefaultClient)
}

func NewGitLabClient(uri, token string, cl *http.Client) (Client, error) {
	// Parse the url
	gitlabURL, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("could not parse gitlab link: %w", err)
	}

	// Get base url for API
	relBaseURL, _ := url.Parse("./api/v4")
	baseURL := gitlabURL.ResolveReference(relBaseURL)

	logrus.WithFields(logrus.Fields{
		"gitlab-url": gitlabURL.String(),
		"base-url":   baseURL.String(),
	}).Debug("connecting to gitlab instance")

	// Create the client
	client, err := gitlab.NewClient(
		token,
		gitlab.WithBaseURL(baseURL.String()),
		gitlab.WithHTTPClient(cl),
	)
	if err != nil {
		return nil, fmt.Errorf("could not connect to gitlab instance %s: %w", uri, err)
	}

	return &gitlabClient{client: client}, nil
}

func (c *gitlabClient) ListReleases(ctx context.Context, owner, name string, opt ListOptions) ([]Release, error) {
	glOpts := &gitlab.ListReleasesOptions{
		ListOptions: gitlab.ListOptions{
			Page:    opt.Page,
			PerPage: opt.PerPage,
		},
	}

	releases, _, err := c.client.Releases.ListReleases(fmt.Sprintf("%s/%s", owner, name), glOpts)
	if err != nil {
		return nil, fmt.Errorf("error getting releases from repository %s/%s: %w", owner, name, err)
	}

	r := make([]Release, 0, glOpts.PerPage)

	for _, release := range releases {
		tagName := release.TagName

		logrus.WithFields(logrus.Fields{
			"tag":    tagName,
			"commit": release.Commit.ID,
		}).Debug("found release")

		r = append(r, Release{Tag: tagName, SemVer: generateVersion(tagName, versionMatcher)})
	}

	return r, nil
}

func (c *gitlabClient) ListTags(ctx context.Context, owner, name string, opt ListOptions) ([]Release, error) {
	glOpts := &gitlab.ListTagsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    opt.Page,
			PerPage: opt.PerPage,
		},
	}

	tags, _, err := c.client.Tags.ListTags(fmt.Sprintf("%s/%s", owner, name), glOpts)
	if err != nil {
		return nil, fmt.Errorf("error getting tags from repository %s/%s: %w", owner, name, err)
	}

	r := make([]Release, 0, glOpts.PerPage)

	for _, tag := range tags {
		tagName := tag.Name

		logrus.WithFields(logrus.Fields{
			"tag":    tagName,
			"commit": tag.Commit.ID,
		}).Debug("found tag")

		r = append(r, Release{Tag: tagName, SemVer: generateVersion(tagName, versionMatcher)})
	}

	return r, nil
}
