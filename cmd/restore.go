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
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/cheggaaa/pb"
	"github.com/iancmcc/jig/config"
	"github.com/iancmcc/jig/utils"
	"github.com/iancmcc/jig/vcs"
	"github.com/spf13/cobra"
)

var appnd bool

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "A brief description of your command",
	Long:  `A`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			manifest *config.Manifest
			err      error
		)
		root, err := config.FindClosestJigRoot("")
		if err != nil {
			logrus.Fatal("No jig root found. Use 'jig init' to create one.")
		}
		if len(args) == 0 {
			// Restore existing manifest
			manifest, err = config.DefaultManifest("")
			if err != nil {
				logrus.Fatal("No repo manifest to restore. Pass a manifest file first.")
			}
		} else {
			manifestpath := args[0]
			switch manifestpath {
			case "-":
				manifest, err = config.FromJSON(os.Stdin)
			default:
				f, err := os.Open(manifestpath)
				if err != nil {
					logrus.WithField("manifest", manifest).WithError(err).Fatal("Unable to open manifest file")
				}
				defer f.Close()
				manifest, err = config.FromJSON(f)
			}

		}

		if err != nil {
			logrus.WithField("manifest", manifest).Fatal("Unable to parse manifest file")
		}

		if appnd {
			oldmanifest, err := config.JigRootManifest()
			if err == nil {
				for _, r := range manifest.Repos {
					oldmanifest.Add(r)
				}
				manifest = oldmanifest
			}
		}

		manifest.Save(root)

		pullchans := []<-chan vcs.Progress{}

		for _, repo := range manifest.Repos {
			pullchan, err := vcs.ApplyRepoConfig(root, vcs.Git, repo)
			if err != nil {
				short, e := utils.RepoToPath(repo.Repo)
				if e != nil {
					short = repo.Repo
				}
				logrus.WithError(err).WithFields(logrus.Fields{
					"repo": short,
					"ref":  repo.Ref,
				}).Error("Unable to update repository")
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
	RootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().BoolVarP(&appnd, "append", "a", false, "Merge manifest being restored with current manifest")
}
