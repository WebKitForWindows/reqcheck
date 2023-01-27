// Copyright (c) 2023, the WebKit for Windows project authors.  Please see the
// AUTHORS file for details. All rights reserved. Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package reqcheck

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/go-github/v50/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type githubClient struct {
	client *github.Client
}

func NewGitHub(uri, token string) (Client, error) {
	return NewGitHubClient(uri, token, http.DefaultClient)
}

func NewGitHubClient(uri, token string, cl *http.Client) (Client, error) {
	// Parse the url
	githubURL, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("could not parse GitHub link: %w", err)
	}

	// Create the client
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(
		context.WithValue(context.Background(), oauth2.HTTPClient, cl),
		ts,
	)

	client := github.NewClient(tc)

	if githubURL.Hostname() != "github.com" {
		relBaseURL, _ := url.Parse("./api/v3/")
		relUploadURL, _ := url.Parse("./api/uploads/")

		client.BaseURL = githubURL.ResolveReference(relBaseURL)
		client.UploadURL = githubURL.ResolveReference(relUploadURL)
	}

	logrus.WithFields(logrus.Fields{
		"github-url": githubURL.String(),
		"base-url":   client.BaseURL.String(),
		"upload-url": client.BaseURL.String(),
	}).Debug("connecting to github instance")

	return &githubClient{client: client}, nil
}

func (c *githubClient) ListReleases(ctx context.Context, owner, name string, opt ListOptions) ([]Release, error) {
	ghOpts := &github.ListOptions{
		Page:    opt.Page,
		PerPage: opt.PerPage,
	}

	releases, _, err := c.client.Repositories.ListReleases(ctx, owner, name, ghOpts)
	if err != nil {
		return nil, fmt.Errorf("error getting releases from repository %s/%s: %w", owner, name, err)
	}

	r := make([]Release, 0, ghOpts.PerPage)

	for _, release := range releases {
		tagName := release.GetTagName()

		logrus.WithFields(logrus.Fields{
			"tag":    tagName,
			"commit": release.GetTargetCommitish(),
		}).Debug("found release")

		r = append(r, Release{Tag: tagName, SemVer: generateVersion(tagName, versionMatcher)})
	}

	return r, nil
}

func (c *githubClient) ListTags(ctx context.Context, owner, name string, opt ListOptions) ([]Release, error) {
	ghOpts := &github.ListOptions{
		Page:    opt.Page,
		PerPage: opt.PerPage,
	}

	tags, _, err := c.client.Repositories.ListTags(ctx, owner, name, ghOpts)
	if err != nil {
		return nil, fmt.Errorf("error getting tags from repository %s/%s: %w", owner, name, err)
	}

	r := make([]Release, 0, ghOpts.PerPage)

	for _, tag := range tags {
		tagName := tag.GetName()

		logrus.WithFields(logrus.Fields{
			"tag":    tagName,
			"commit": tag.GetCommit().GetSHA(),
		}).Debug("found tag")

		r = append(r, Release{Tag: tagName, SemVer: generateVersion(tagName, versionMatcher)})
	}

	return r, nil
}
