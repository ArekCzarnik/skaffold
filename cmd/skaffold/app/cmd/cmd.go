/*
Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"io"
	"os"

	"github.com/GoogleCloudPlatform/skaffold/pkg/skaffold/config"
	"github.com/GoogleCloudPlatform/skaffold/pkg/skaffold/constants"
	"github.com/GoogleCloudPlatform/skaffold/pkg/skaffold/runner"
	"github.com/GoogleCloudPlatform/skaffold/pkg/skaffold/version"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	opts     = &config.SkaffoldOptions{}
	v        string
	filename string
)

func NewSkaffoldCommand(out, err io.Writer) *cobra.Command {
	c := &cobra.Command{
		Use: "skaffold",
		Short: "A tool that makes the onboarding of existing applications to Kubernetes Engine simple and repeatable.		",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := SetUpLogs(err, v); err != nil {
				return err
			}
			logrus.Infof("Skaffold %+v", version.Get())
			return nil
		},
	}

	c.AddCommand(NewCmdVersion(out))
	c.AddCommand(NewCmdRun(out))
	c.AddCommand(NewCmdDev(out))
	c.AddCommand(NewCmdDocker(out))

	c.PersistentFlags().StringVarP(&v, "verbosity", "v", constants.DefaultLogLevel.String(), "Log level (debug, info, warn, error, fatal, panic")
	return c
}

func AddRunDevFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&filename, "filename", "f", "skaffold.yaml", "Filename of pipeline file")
	cmd.Flags().BoolVar(&opts.Notification, "toot", false, "Emit a terminal beep after the deploy is complete")
}

func SetUpLogs(out io.Writer, level string) error {
	logrus.SetOutput(out)
	lvl, err := logrus.ParseLevel(v)
	if err != nil {
		return errors.Wrap(err, "parsing log level")
	}
	logrus.SetLevel(lvl)
	return nil
}

var defaultConfigForMode = map[bool]*config.SkaffoldConfig{
	true:  config.DefaultDevSkaffoldConfig,
	false: config.DefaultRunSkaffoldConfig,
}

func runSkaffold(out io.Writer, dev bool, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return errors.Wrap(err, "opening skaffold config")
	}
	defer f.Close()

	defaultConfig := defaultConfigForMode[dev]
	cfg, err := config.Parse(f, defaultConfig)
	if err != nil {
		return errors.Wrap(err, "parsing skaffold config")
	}

	opts.Output = out
	opts.DevMode = dev
	r, err := runner.NewForConfig(opts, cfg)
	if err != nil {
		return errors.Wrap(err, "getting skaffold config")
	}

	if err := r.Run(); err != nil {
		return errors.Wrap(err, "running skaffold steps")
	}

	return nil
}
