// Copyright (c) 2023, the WebKit for Windows project authors.  Please see the
// AUTHORS file for details. All rights reserved. Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/Masterminds/semver"
	"github.com/WebKitForWindows/reqcheck"
	"github.com/reactivex/rxgo/v2"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

func vcpkgCmd() *cli.Command {
	settings := struct {
		Output string
	}{}

	return &cli.Command{
		Name:      "vcpkg",
		Usage:     "query ",
		ArgsUsage: "<vcpkg-path>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "output-file",
				Usage:       "output results to file",
				Destination: &settings.Output,
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() > 1 {
				return fmt.Errorf("command takes one optional argument <vcpkg-path>: %w", ErrCli)
			}

			// Determine vcpkg directory
			vcpkgPath := c.Args().Get(0)
			if !filepath.IsAbs(vcpkgPath) {
				workingDir, err := os.Getwd()
				logrus.WithField("working-directory", workingDir).Debug("root")
				if err != nil {
					return fmt.Errorf("could not determine working directory: %w", ErrCli)
				}
				vcpkgPath = filepath.Join(workingDir, vcpkgPath)
			}

			logrus.WithField("vcpkg-path", vcpkgPath).Debug("path")

			// Parse and verify config
			cfg, err := loadConfig(filepath.Join(vcpkgPath, configFileName))
			if err != nil {
				return fmt.Errorf("could not open config file %s: %w", configFileName, err)
			}

			scms := make(map[string]reqcheck.Client)
			for name, scmConfig := range cfg.Scms {
				scm, err := reqcheck.NewClientFromDriver(scmConfig.Driver, scmConfig.URI, scmConfig.Token)
				if err != nil {
					return fmt.Errorf("could not connect to scm %s: %w", name, err)
				}
				scms[name] = scm
			}

			var tmpl string
			if cfg.Template != "" {
				tmpl = strings.TrimSpace(cfg.Template)
			} else {
				tmpl = defaultTmpl
			}
			t, err := template.New("vcpkg").Parse(tmpl)
			if err != nil {
				return fmt.Errorf("could not parse template: %w", err)
			}

			var output io.Writer
			if settings.Output != "" {
				output, err = os.Create(settings.Output)
				if err != nil {
					return fmt.Errorf("could not open file for writing %s: %w", settings.Output, err)
				}
			} else {
				output = os.Stdout
			}

			// Iterate over values
			type releaseUpdate struct {
				Name    string
				Current string
				Upgrade string
			}

			current := make([]releaseUpdate, 0)
			upgrade := make([]releaseUpdate, 0)

			for name, library := range cfg.Libraries {
				version, err := readVcpkgVersion(vcpkgPath, name)
				if err != nil {
					return fmt.Errorf("could not find version for %s: %w", name, err)
				}

				logrus.WithField("version", version).Debug("found config")

				var constraintFmt string
				if library.Constraint != "" {
					constraintFmt = library.Constraint
				} else {
					constraintFmt = ">= %s"
				}

				constraintStr := fmt.Sprintf(constraintFmt, version)
				constraint, err := semver.NewConstraint(constraintStr)
				if err != nil {
					return fmt.Errorf("could not create constraint for %s from %s: %w", name, version, err)
				}
				logrus.WithField("constraint", constraintStr).Debug("found constraint")

				scm, ok := scms[library.Host]
				if !ok {
					return fmt.Errorf("could not find scm assigned to %s: %w", library.Host, ErrCli)
				}

				releaseOpts := reqcheck.ListReleaseOptions{
					Owner:   library.Owner,
					Repo:    library.Repo,
					Tags:    library.Tags,
					LimitTo: library.LimitTo,
				}

				latestRelease, err := reqcheck.ListReleases(scm, releaseOpts).
					Filter(reqcheck.FilterSemanticConstraint(constraint)).
					Reduce(reqcheck.ReduceGreatestVersion).
					Get()
				if err != nil || latestRelease == rxgo.OptionalSingleEmpty {
					return fmt.Errorf("could not get releases %w", err)
				}

				release := releaseUpdate{
					Name:    name,
					Current: version,
					Upgrade: latestRelease.V.(reqcheck.Release).SemVer.String(),
				}

				if release.Current == release.Upgrade {
					current = append(current, release)
				} else {
					upgrade = append(upgrade, release)
				}
			}

			// Sort the results
			sort.Slice(current, func(i, j int) bool {
				return current[i].Name < current[j].Name
			})
			sort.Slice(upgrade, func(i, j int) bool {
				return upgrade[i].Name < upgrade[j].Name
			})

			// Output results to template
			td := struct {
				Current []releaseUpdate
				Upgrade []releaseUpdate
			}{
				Current: current,
				Upgrade: upgrade,
			}

			err = t.Execute(output, td)
			if err != nil {
				return fmt.Errorf("could not write results: %w", err)
			}

			return nil
		},
	}
}

const configFileName = ".reqcheck.yml"

func readVcpkgVersion(vcpkgPath, name string) (string, error) {
	file, err := os.ReadFile(filepath.Join(vcpkgPath, "ports", name, "vcpkg.json"))
	if err != nil {
		return "", fmt.Errorf("could not read %s config file: %w", name, err)
	}

	un := make(map[interface{}]interface{})
	err = yaml.Unmarshal(file, &un)
	if err != nil {
		return "", fmt.Errorf("could not read %s config file: %w", name, err)
	}

	if semVer, ok := un["version-semver"].(string); ok {
		return semVer, nil
	}

	if ver, ok := un["version"].(string); ok {
		return ver, nil
	}

	return "", fmt.Errorf("could not find version string for %s: %w", name, ErrCli)
}

const defaultTmpl = `The following libraries are up to date:
{{ range .Current }}  {{ .Name }}: {{ .Current }}
{{ else }}  No libraries are up to date{{ end }}
The following libraries have updates:
{{ range .Upgrade}}  {{ .Name }}: {{ .Current }} -> {{ .Upgrade }}
{{ else }}  All libraries are up to date{{ end }}`

type (
	config struct {
		Scms      map[string]sourceControl `yaml:"scm"`
		Libraries map[string]library       `yaml:"repos"`
		Template  string                   `yaml:"template"`
	}

	sourceControl struct {
		Driver string
		URI    string
		Token  string
	}

	library struct {
		Host       string `yaml:"host"`
		Owner      string `yaml:"owner"`
		Repo       string `yaml:"repo"`
		Tags       bool   `yaml:"tags"`
		Constraint string `yaml:"constraint"`
		LimitTo    int    `yaml:"limit"`
	}
)

func loadConfig(path string) (config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return config{}, fmt.Errorf("could not read config file %s: %w", path, err)
	}

	var c config

	err = yaml.Unmarshal(b, &c)
	if err != nil {
		return config{}, fmt.Errorf("error when loading config: %w", err)
	}

	return c, nil
}

func (s *sourceControl) UnmarshalYAML(unmarshal func(interface{}) error) error {
	un := make(map[interface{}]interface{})

	err := unmarshal(&un)
	if err != nil {
		return err
	}

	val, ok := un["driver"]
	if ok {
		s.Driver = val.(string)
	}

	val, ok = un["uri"]
	if ok {
		s.URI = val.(string)
	}

	val, ok = un["token"]
	if ok {
		switch t := val.(type) {
		case string:
			s.Token = t
		case map[string]interface{}:
			envVar, ok := t["from_environment"].(string)
			if ok {
				env, ok := os.LookupEnv(envVar)
				if ok {
					s.Token = env
				} else {
					return fmt.Errorf("could not find token in %s, %w", envVar, ErrCli)
				}
			}
		}
	}

	return nil
}
