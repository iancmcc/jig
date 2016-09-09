// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/iancmcc/jig/config"
	"github.com/spf13/cobra"
)

var force bool

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a Jig root",
	Long:  `Initialize the directory passed as the root of a Jig source tree.`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			path string
			err  error
		)
		if len(args) > 0 {
			path = args[0]
		}
		path, err = filepath.Abs(path)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"path": path,
			}).Fatal("Unable to determine initialization path")
		}

		// Validate that we are not nesting
		if p, err := config.FindClosestJigRoot(path); err == nil && !force {
			logrus.WithField("root", p).Fatal("You're already inside a Jig root. Pass -f to force creation anyway.")
		}

		if err := config.CreateJigRoot(path); err != nil {
			logrus.WithField("path", path).Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&force, "force", "f", false, "Force creation of a nested Jig root")
}
