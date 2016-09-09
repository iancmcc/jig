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
	"fmt"
	"path/filepath"

	"github.com/cheggaaa/pb"
	"github.com/iancmcc/jig/config"
	"github.com/iancmcc/jig/vcs"
	"github.com/spf13/cobra"
	"metis/src/golang/src/github.com/Sirupsen/logrus"
)

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		root, err := config.FindClosestJigRoot("")
		if err != nil {
			logrus.Fatal("No jig root found. Use 'jig init' to create one.")
		}
		manifest, err := config.DefaultManifest("")
		if err != nil {
			logrus.Fatal("No repo manifest to use to pull. `jig restore` a manifest first.")
		}
		pullchans := []<-chan vcs.Progress{}
		for _, repo := range manifest.Repos {
			log := logrus.WithField("repo", repo.Repo)
			dir, err := vcs.RepoToPath(repo.Repo)
			if err != nil {
				log.Error("Unable to parse repo")
			}
			dir = filepath.Join(root, dir)
			dir, err = filepath.Abs(dir)
			if err != nil {
				log.WithError(err).Error("Unable to pull repo")
			}
			pullchan, err := vcs.Git.Pull(repo, dir)
			if err != nil {
				log.WithError(err).Error("Unable to pull repo")
			}
			pullchans = append(pullchans, pullchan)
		}
		bar := pb.StartNew(0)
		go bar.Start()
		for prog := range vcs.CombinedProgress(pullchans...) {
			bar.Total = int64(prog.Total)
			bar.Set(prog.Current)
			bar.Prefix(fmt.Sprintf("%s (%s)", prog.Message, prog.Repo))
		}
		bar.Finish()
	},
}

func init() {
	RootCmd.AddCommand(pullCmd)
}
